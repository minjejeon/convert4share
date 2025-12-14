//go:build windows

package windows

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

const (
	menuName = "Convert with Convert4Share"
	keyName  = "Convert4Share"
	progID   = "Convert4Share.File"
)

// RegisterContextMenu adds the application to the Windows context menu.
func RegisterContextMenu() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %w", err)
	}

	extensions := []string{".mov", ".heic"}

	// 1. Register the ProgID (Convert4Share.File) - Required for Open With
	if err := registerProgID(exePath); err != nil {
		return fmt.Errorf("could not register ProgID: %w", err)
	}

	// 2. Register Generic Shell Extension (Classic Menu / Show More Options)
	// This uses "AppliesTo" on "*" to support mixed selection of .mov and .heic files.
	if err := registerGenericShellExtension(exePath); err != nil {
		return fmt.Errorf("could not register generic shell extension: %w", err)
	}

	// 3. Register OpenWithProgids (Modern Menu - via Open With submenu)
	// This part is disabled because it does not work correctly with multi-selection on Windows 11.
	// for _, ext := range extensions {
	// 	if err := registerOpenWith(ext); err != nil {
	// 		return fmt.Errorf("could not register OpenWith for %s: %w", ext, err)
	// 	}
	// }
	return nil
}

func registerProgID(exePath string) error {
	// Create HKCR\Convert4Share.File
	key, _, err := registry.CreateKey(registry.CLASSES_ROOT, progID, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	if err := key.SetStringValue("", "Convert4Share File"); err != nil {
		return err
	}

	// DefaultIcon
	iconKey, _, err := registry.CreateKey(key, "DefaultIcon", registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer iconKey.Close()
	if err := iconKey.SetStringValue("", fmt.Sprintf(`"%s"`, exePath)); err != nil {
		return err
	}

	// shell\open\command
	cmdKey, _, err := registry.CreateKey(key, `shell\open\command`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer cmdKey.Close()

	command := fmt.Sprintf(`"%s" "%%1"`, exePath)
	if err := cmdKey.SetStringValue("", command); err != nil {
		return err
	}

	// shell\open -> FriendlyAppName for "Open With" menu appearance
	openKey, _, err := registry.CreateKey(key, `shell\open`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer openKey.Close()
	if err := openKey.SetStringValue("FriendlyAppName", menuName); err != nil {
		// Ignore error if we can't set friendly name, it's cosmetic
	}

	return nil
}

func registerGenericShellExtension(exePath string) error {
	// Register under HKCR\*\shell\Convert4Share
	keyPath := fmt.Sprintf(`*\shell\%s`, keyName)
	key, _, err := registry.CreateKey(registry.CLASSES_ROOT, keyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	if err := key.SetStringValue("", menuName); err != nil {
		return err
	}
	if err := key.SetStringValue("Icon", fmt.Sprintf(`"%s"`, exePath)); err != nil {
		return err
	}

	// AppliesTo logic: Only show for .mov or .heic
	appliesTo := "System.FileExtension:=.mov OR System.FileExtension:=.heic"
	if err := key.SetStringValue("AppliesTo", appliesTo); err != nil {
		return err
	}

	cmdKeyPath := fmt.Sprintf(`%s\command`, keyPath)
	cmdKey, _, err := registry.CreateKey(registry.CLASSES_ROOT, cmdKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer cmdKey.Close()

	command := fmt.Sprintf(`"%s" "%%1"`, exePath)
	if err := cmdKey.SetStringValue("", command); err != nil {
		return err
	}
	return nil
}

// UnregisterContextMenu removes the application from the Windows context menu.
func UnregisterContextMenu() error {
	// Remove new Generic Shell Extension
	if err := deleteGenericShellExtension(); err != nil {
		return err
	}

	if err := deleteProgID(); err != nil {
		return err
	}

	return nil
}

func deleteGenericShellExtension() error {
	// Delete command first
	cmdPath := fmt.Sprintf(`*\shell\%s\command`, keyName)
	if err := registry.DeleteKey(registry.CLASSES_ROOT, cmdPath); err != nil && err != registry.ErrNotExist {
		return err
	}

	// Delete key
	keyPath := fmt.Sprintf(`*\shell\%s`, keyName)
	if err := registry.DeleteKey(registry.CLASSES_ROOT, keyPath); err != nil && err != registry.ErrNotExist {
		return err
	}
	return nil
}

func deleteProgID() error {
	keys := []string{
		progID + `\shell\open\command`,
		progID + `\shell\open`,
		progID + `\shell`,
		progID + `\DefaultIcon`,
		progID,
	}

	for _, k := range keys {
		if err := registry.DeleteKey(registry.CLASSES_ROOT, k); err != nil && err != registry.ErrNotExist {
			return err
		}
	}
	return nil
}

// IsContextMenuInstalled checks if the context menu is currently registered.
func IsContextMenuInstalled() bool {
	// Check the new Generic Shell Extension key
	keyPath := `*\shell\Convert4Share`
	k, err := registry.OpenKey(registry.CLASSES_ROOT, keyPath, registry.QUERY_VALUE)
	if err != nil {
		// Fallback check for legacy installation?
		// If new key is missing, check old key to avoid false negatives during transition?
		// But usually we just care if the *current* version is installed.
		// If old is installed but new isn't, we might return false so user reinstalls.
		return false
	}
	k.Close()
	return true
}

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

	for _, ext := range extensions {
		// 2. Register SystemFileAssociations (Classic Menu / Show More Options)
		if err := registerSystemFileAssociation(ext, exePath); err != nil {
			return fmt.Errorf("could not register system file association for %s: %w", ext, err)
		}

		// 3. Register OpenWithProgids (Modern Menu - via Open With submenu)
		if err := registerOpenWith(ext); err != nil {
			return fmt.Errorf("could not register OpenWith for %s: %w", ext, err)
		}
	}
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
	// Usually FriendlyAppName works better in HKCR\Applications\Convert4Share.exe,
	// but adding it here doesn't hurt.
	if err := openKey.SetStringValue("FriendlyAppName", menuName); err != nil {
		// Ignore error if we can't set friendly name, it's cosmetic
	}

	return nil
}

func registerSystemFileAssociation(ext, exePath string) error {
	keyPath := fmt.Sprintf(`SystemFileAssociations\%s\shell\%s`, ext, keyName)
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

func registerOpenWith(ext string) error {
	keyPath := fmt.Sprintf(`%s\OpenWithProgids`, ext)
	// Use CreateKey to ensure it exists
	key, _, err := registry.CreateKey(registry.CLASSES_ROOT, keyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	// Set value with name = progID, data = empty string
	if err := key.SetStringValue(progID, ""); err != nil {
		return err
	}
	return nil
}

// UnregisterContextMenu removes the application from the Windows context menu.
func UnregisterContextMenu() error {
	extensions := []string{".mov", ".heic"}
	for _, ext := range extensions {
		if err := deleteSystemFileAssociation(ext); err != nil {
			return err
		}
		if err := unregisterOpenWith(ext); err != nil {
			return err
		}
	}

	if err := deleteProgID(); err != nil {
		return err
	}

	return nil
}

func deleteSystemFileAssociation(ext string) error {
	// Delete command first
	cmdPath := fmt.Sprintf(`SystemFileAssociations\%s\shell\%s\command`, ext, keyName)
	if err := registry.DeleteKey(registry.CLASSES_ROOT, cmdPath); err != nil && err != registry.ErrNotExist {
		return err
	}

	// Delete key
	keyPath := fmt.Sprintf(`SystemFileAssociations\%s\shell\%s`, ext, keyName)
	if err := registry.DeleteKey(registry.CLASSES_ROOT, keyPath); err != nil && err != registry.ErrNotExist {
		return err
	}
	return nil
}

func unregisterOpenWith(ext string) error {
	keyPath := fmt.Sprintf(`%s\OpenWithProgids`, ext)
	k, err := registry.OpenKey(registry.CLASSES_ROOT, keyPath, registry.SET_VALUE)
	if err == nil {
		defer k.Close()
		if err := k.DeleteValue(progID); err != nil && err != registry.ErrNotExist {
			return err
		}
	} else if err != registry.ErrNotExist {
		return err
	}
	return nil
}

func deleteProgID() error {
	// Delete keys recursively-ish
	keys := []string{
		progID + `\shell\open\command`,
		progID + `\shell\open`,
		progID + `\shell`,
		progID + `\DefaultIcon`,
		progID,
	}

	for _, k := range keys {
		if err := registry.DeleteKey(registry.CLASSES_ROOT, k); err != nil && err != registry.ErrNotExist {
			// If we fail to delete a child, we might fail to delete parent.
			// But we try anyway.
			// For robustness, maybe we should return error?
			// But if one child fails, we still want to try deleting others?
			// Let's just return error for now to be safe.
			return err
		}
	}
	return nil
}

// IsContextMenuInstalled checks if the context menu is currently registered.
func IsContextMenuInstalled() bool {
	// Check one of the keys. If it exists, we assume it's installed.
	keyPath := `SystemFileAssociations\.mov\shell\Convert4Share`
	k, err := registry.OpenKey(registry.CLASSES_ROOT, keyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	k.Close()
	return true
}

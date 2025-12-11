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
)

// RegisterContextMenu adds the application to the Windows context menu.
func RegisterContextMenu() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %w", err)
	}

	command := fmt.Sprintf(`"%s" "%%1"`, exePath)
	extensions := []string{".mov", ".heic"}

	for _, ext := range extensions {
		// Create HKEY_CLASSES_ROOT\SystemFileAssociations\.ext\shell\Convert4Share
		keyPath := fmt.Sprintf(`SystemFileAssociations\%s\shell\%s`, ext, keyName)
		key, _, err := registry.CreateKey(registry.CLASSES_ROOT, keyPath, registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("could not create shell key for %s: %w", ext, err)
		}
		// Set the menu display text and icon
		if err := key.SetStringValue("", menuName); err != nil {
			key.Close()
			return fmt.Errorf("could not set default value for %s: %w", ext, err)
		}
		if err := key.SetStringValue("Icon", `"`+exePath+`"`); err != nil {
			key.Close()
			return fmt.Errorf("could not set icon for %s: %w", ext, err)
		}

		key.Close()

		// Create HKEY_CLASSES_ROOT\SystemFileAssociations\.ext\shell\Convert4Share\command
		cmdKeyPath := fmt.Sprintf(`%s\command`, keyPath)
		cmdKey, _, err := registry.CreateKey(registry.CLASSES_ROOT, cmdKeyPath, registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("could not create command key for %s: %w", ext, err)
		}
		if err := cmdKey.SetStringValue("", command); err != nil {
			cmdKey.Close()
			return fmt.Errorf("could not set command value for %s: %w", ext, err)
		}
		cmdKey.Close()
	}
	return nil
}

// UnregisterContextMenu removes the application from the Windows context menu.
func UnregisterContextMenu() error {
	extensions := []string{".mov", ".heic"}
	for _, ext := range extensions {
		keyPath := fmt.Sprintf(`SystemFileAssociations\%s\shell\Convert4Share`, ext)
		if err := registry.DeleteKey(registry.CLASSES_ROOT, keyPath); err != nil && err != registry.ErrNotExist {
			return fmt.Errorf("could not delete key for %s: %w", ext, err)
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

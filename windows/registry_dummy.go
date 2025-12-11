//go:build !windows

package windows

import "errors"

// RegisterContextMenu is a dummy implementation for non-Windows systems.
func RegisterContextMenu() error {
	return errors.New("registry modification is only supported on Windows")
}

// UnregisterContextMenu is a dummy implementation for non-Windows systems.
func UnregisterContextMenu() error {
	return errors.New("registry modification is only supported on Windows")
}

// IsContextMenuInstalled is a dummy implementation for non-Windows systems.
func IsContextMenuInstalled() bool {
	return false
}

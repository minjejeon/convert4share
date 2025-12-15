//go:build windows

package windows

import (
	"os/exec"
	"strings"
	"syscall"
)

// CopyFileToClipboard copies the given file path to the clipboard as a file drop list (CF_HDROP).
func CopyFileToClipboard(path string) error {
	// Escape single quotes for PowerShell
	escapedPath := strings.ReplaceAll(path, "'", "''")

	cmdStr := `Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.Clipboard]::SetFileDropList([System.Collections.Specialized.StringCollection]@('` + escapedPath + `'))`
	cmd := exec.Command("powershell", "-Command", cmdStr)

	// Hide the console window
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	return cmd.Run()
}

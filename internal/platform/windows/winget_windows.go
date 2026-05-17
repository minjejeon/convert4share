//go:build windows

package windows

import (
	"fmt"
	"os/exec"
	"syscall"
)

// InstallWingetPackage installs a package using winget.
func InstallWingetPackage(packageID string) error {
	path, err := exec.LookPath("winget")
	if err != nil {
		return fmt.Errorf("winget not found: please install App Installer from Microsoft Store")
	}

	cmd := exec.Command(path, "install", "--id", packageID, "--silent", "--accept-source-agreements", "--accept-package-agreements")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installation failed: %s\nOutput: %s", err, string(output))
	}
	return nil
}

//go:build !windows

package windows

import "fmt"

// InstallWingetPackage installs a package using winget.
func InstallWingetPackage(packageID string) error {
	return fmt.Errorf("winget installation is only supported on Windows")
}

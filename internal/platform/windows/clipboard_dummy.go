//go:build !windows && !linux

package windows

import (
	"fmt"
	"runtime"
)

// CopyFileToClipboard is a no-op stub on platforms without a clipboard implementation.
func CopyFileToClipboard(path string) error {
	return fmt.Errorf("CopyFileToClipboard not implemented for %s", runtime.GOOS)
}

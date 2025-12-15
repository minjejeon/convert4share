//go:build !windows

package windows

import (
	"fmt"
)

// CopyFileToClipboard is a no-op on non-Windows systems for now.
func CopyFileToClipboard(path string) error {
	// Alternatively, we could try to implement for Mac/Linux, but sticking to Windows per request context.
	// Maybe return error or print?
	fmt.Println("CopyFileToClipboard not implemented for this OS")
	return nil
}

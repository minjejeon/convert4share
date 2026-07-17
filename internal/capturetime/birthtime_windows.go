//go:build windows

package capturetime

import (
	"os"
	"syscall"
	"time"
)

// birthTime returns the file creation time on Windows.
func birthTime(info os.FileInfo) (time.Time, bool) {
	attr, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		return time.Time{}, false
	}
	return time.Unix(0, attr.CreationTime.Nanoseconds()), true
}

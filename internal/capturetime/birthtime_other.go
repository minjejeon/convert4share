//go:build !windows

package capturetime

import (
	"os"
	"time"
)

// birthTime is unsupported off Windows; callers fall back to mtime.
func birthTime(info os.FileInfo) (time.Time, bool) {
	return time.Time{}, false
}

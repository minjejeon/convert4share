package logging

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestLogFilePath_UsesPerUserCacheDir asserts the log file resolves to a
// per-user writable location rather than next to the executable, so an
// all-users install under Program Files (not writable by standard users)
// still records logs.
func TestLogFilePath_UsesPerUserCacheDir(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("path layout assertion is Windows-specific")
	}

	tmp := t.TempDir()
	// os.UserCacheDir reads %LocalAppData% on Windows.
	t.Setenv("LocalAppData", tmp)

	got, err := logFilePath()
	if err != nil {
		t.Fatalf("logFilePath: %v", err)
	}

	want := filepath.Join(tmp, "Convert4Share", "logs", "convert4share.log")
	if got != want {
		t.Errorf("logFilePath = %q, want %q", got, want)
	}

	// The directory must already exist and be writable.
	dir := filepath.Dir(got)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("expected log dir %q to exist as a directory: %v", dir, err)
	}
	if err := os.WriteFile(got, []byte("x"), 0o644); err != nil {
		t.Errorf("log dir not writable: %v", err)
	}
}

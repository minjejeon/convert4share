package jobs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestResolveDestination(t *testing.T) {
	s := New(nil)

	tempDir, err := os.MkdirTemp("", "convert4share-jobs-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	name := "testfile"
	ext := ".mp4"
	expected := filepath.Join(tempDir, "testfile.mp4")

	dest, err := s.resolveDestination(tempDir, name, ext, "rename")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if dest != expected {
		t.Errorf("Expected %s, got %s", expected, dest)
	}

	if err := os.WriteFile(expected, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	expectedRename := filepath.Join(tempDir, "testfile (1).mp4")
	dest, err = s.resolveDestination(tempDir, name, ext, "rename")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if dest != expectedRename {
		t.Errorf("Expected %s, got %s", expectedRename, dest)
	}

	if err := os.WriteFile(expectedRename, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	expectedRename2 := filepath.Join(tempDir, "testfile (2).mp4")
	dest, err = s.resolveDestination(tempDir, name, ext, "rename")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if dest != expectedRename2 {
		t.Errorf("Expected %s, got %s", expectedRename2, dest)
	}

	dest, err = s.resolveDestination(tempDir, name, ext, "overwrite")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if dest != expected {
		t.Errorf("Expected %s, got %s", expected, dest)
	}

	if _, err = s.resolveDestination(tempDir, name, ext, "error"); err == nil {
		t.Error("Expected error, got nil")
	}

	os.Remove(expected)
	dest, err = s.resolveDestination(tempDir, name, ext, "error")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if dest != expected {
		t.Errorf("Expected %s, got %s", expected, dest)
	}

	if err := os.WriteFile(expected, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create 0-byte file: %v", err)
	}
	dest, err = s.resolveDestination(tempDir, name, ext, "rename")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if dest != expected {
		t.Errorf("Expected %s (overwrite 0-byte), got %s", expected, dest)
	}
}

func TestApplySettingsResizesSemaphores(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	s := New(nil)
	if got := s.ffmpegSem.Cap(); got != 1 {
		t.Errorf("default ffmpegSem cap: want 1, got %d", got)
	}
	if got := s.magickSem.Cap(); got != 1 {
		t.Errorf("default magickSem cap: want 1, got %d", got)
	}

	viper.Set("maxFfmpegWorkers", 4)
	viper.Set("maxMagickWorkers", 8)
	s.applySettings()

	if got := s.ffmpegSem.Cap(); got != 4 {
		t.Errorf("after applySettings ffmpegSem cap: want 4, got %d", got)
	}
	if got := s.magickSem.Cap(); got != 8 {
		t.Errorf("after applySettings magickSem cap: want 8, got %d", got)
	}

	// Clamp: invalid values fall back to 1.
	viper.Set("maxFfmpegWorkers", 0)
	viper.Set("maxMagickWorkers", -3)
	s.applySettings()

	if got := s.ffmpegSem.Cap(); got != 1 {
		t.Errorf("clamp ffmpegSem cap: want 1, got %d", got)
	}
	if got := s.magickSem.Cap(); got != 1 {
		t.Errorf("clamp magickSem cap: want 1, got %d", got)
	}
}

func TestExtractFileArgs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "convert4share-extract-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	validFile := filepath.Join(tempDir, "valid.mov")
	if err := os.WriteFile(validFile, []byte("data"), 0644); err != nil {
		t.Fatalf("create valid: %v", err)
	}
	emptyFile := filepath.Join(tempDir, "empty.mov")
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatalf("create empty: %v", err)
	}
	exePath := filepath.Join(tempDir, "Convert4Share.exe")
	if err := os.WriteFile(exePath, []byte("MZ"), 0644); err != nil {
		t.Fatalf("create exe: %v", err)
	}
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	missingFile := filepath.Join(tempDir, "does-not-exist.mov")

	args := []string{
		exePath,         // own executable, must be filtered
		validFile,       // accepted
		emptyFile,       // 0 bytes, filtered
		subDir,          // directory, filtered
		missingFile,     // missing, filtered
		"",              // blank, filtered
		`"` + validFile + `"`, // quoted, but de-dups path to validFile
	}

	got := ExtractFileArgs(args, exePath)

	// validFile must appear (possibly twice — once direct, once dequoted).
	foundValid := 0
	for _, g := range got {
		if g == validFile {
			foundValid++
		}
		if g == exePath {
			t.Errorf("exe path was not filtered out: %v", got)
		}
		if g == emptyFile {
			t.Errorf("empty file was not filtered out: %v", got)
		}
		if g == subDir {
			t.Errorf("directory was not filtered out: %v", got)
		}
	}
	if foundValid == 0 {
		t.Errorf("valid file missing from result: %v", got)
	}
}

func TestCancelJobAndPauseQueue_NoApp(t *testing.T) {
	// Verifies the methods are safe to call without an associated
	// application (s.app == nil), which is how unit tests exercise the
	// service. Emit becomes a no-op.
	s := New(nil)

	s.PauseQueue()
	if !s.isPaused {
		t.Error("PauseQueue did not flip isPaused")
	}

	s.ResumeQueue()
	if s.isPaused {
		t.Error("ResumeQueue did not clear isPaused")
	}

	// CancelJob on an unknown ID is a no-op (must not panic).
	s.CancelJob("unknown-id")
}

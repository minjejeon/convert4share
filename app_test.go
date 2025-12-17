package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDestination(t *testing.T) {
	// Create a temp directory
	tempDir, err := os.MkdirTemp("", "convert4share-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Case 1: File does not exist
	name := "testfile"
	ext := ".mp4"
	expected := filepath.Join(tempDir, "testfile.mp4")

	dest, err := resolveDestination(tempDir, name, ext, "rename")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if dest != expected {
		t.Errorf("Expected %s, got %s", expected, dest)
	}

	// Case 2: File exists, rename
	// Create the file first
	if err := os.WriteFile(expected, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	expectedRename := filepath.Join(tempDir, "testfile (1).mp4")
	dest, err = resolveDestination(tempDir, name, ext, "rename")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if dest != expectedRename {
		t.Errorf("Expected %s, got %s", expectedRename, dest)
	}

    // Case 2b: File (1) exists, should be (2)
    if err := os.WriteFile(expectedRename, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
    expectedRename2 := filepath.Join(tempDir, "testfile (2).mp4")
	dest, err = resolveDestination(tempDir, name, ext, "rename")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if dest != expectedRename2 {
		t.Errorf("Expected %s, got %s", expectedRename2, dest)
	}


	// Case 3: File exists, overwrite
	dest, err = resolveDestination(tempDir, name, ext, "overwrite")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if dest != expected { // Should return original path even if exists
		t.Errorf("Expected %s, got %s", expected, dest)
	}

	// Case 4: File exists, error
	_, err = resolveDestination(tempDir, name, ext, "error")
	if err == nil {
		t.Error("Expected error, got nil")
	}

    // Case 5: File does not exist, error option (should proceed)
    // Delete the file first
    os.Remove(expected)
    dest, err = resolveDestination(tempDir, name, ext, "error")
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
    if dest != expected {
        t.Errorf("Expected %s, got %s", expected, dest)
    }
}

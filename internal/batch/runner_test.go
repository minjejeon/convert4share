package batch_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/minjejeon/convert4share/internal/batch"
	"github.com/minjejeon/convert4share/internal/converter"
)

func TestResolveDestination(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "c4s-batch-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	name := "sample"
	ext := ".mp4"
	expected := filepath.Join(tempDir, "sample.mp4")

	dest, err := batch.ResolveDestination(tempDir, name, ext, "rename")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if dest != expected {
		t.Errorf("Expected %s, got %s", expected, dest)
	}

	// Create non-empty file
	if err := os.WriteFile(expected, []byte("data"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	expectedRename := filepath.Join(tempDir, "sample (1).mp4")
	dest, err = batch.ResolveDestination(tempDir, name, ext, "rename")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if dest != expectedRename {
		t.Errorf("Expected %s, got %s", expectedRename, dest)
	}
}

func TestBatchRunnerBasic(t *testing.T) {
	runner := batch.NewRunner(&converter.Config{}, batch.Options{
		CollisionOption: "rename",
		AutoLivePhoto:   true,
	}, nil)

	if runner == nil {
		t.Fatal("expected runner to not be nil")
	}

	var statuses []batch.JobStatus
	reporter := func(status batch.JobStatus) {
		statuses = append(statuses, status)
	}

	// Test non-existent file reporting
	success, skipped, failed := runner.Run(context.Background(), []string{"/non/existent/file.mov"}, reporter)
	if success != 0 || failed != 1 || skipped != 0 {
		t.Errorf("expected 0 success, 1 failed, got success=%d, failed=%d, skipped=%d", success, failed, skipped)
	}
}

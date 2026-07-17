package capturetime

import (
	"testing"
	"time"
)

func TestVideoCreationTime(t *testing.T) {
	got, ok := videoCreationTime("testdata/sample.mp4")
	if !ok {
		t.Fatal("expected to extract mvhd creation time from sample.mp4")
	}
	// mp4 epoch is 1904; a real capture time must be after 2000.
	if got.Year() < 2000 || got.Year() > 2100 {
		t.Errorf("implausible year %d from mvhd: %v", got.Year(), got)
	}
}

func TestVideoCreationTimeMissing(t *testing.T) {
	if _, ok := videoCreationTime("testdata/does-not-exist.mp4"); ok {
		t.Error("expected ok=false for missing file")
	}
}

func TestImageCaptureTime(t *testing.T) {
	got, ok := imageCaptureTime("testdata/sample.jpg")
	if !ok {
		t.Fatal("expected to extract EXIF DateTimeOriginal from sample.jpg")
	}
	want := time.Date(2021, 3, 4, 5, 6, 7, 0, got.Location())
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestImageCaptureTimeMissing(t *testing.T) {
	if _, ok := imageCaptureTime("testdata/does-not-exist.jpg"); ok {
		t.Error("expected ok=false for missing file")
	}
}

// mp4Epoch is the reference used by the implementation.
var _ = time.UTC

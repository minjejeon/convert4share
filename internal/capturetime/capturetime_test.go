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

// mp4Epoch is the reference used by the implementation.
var _ = time.UTC

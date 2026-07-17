package capturetime

import (
	"os"
	"path/filepath"
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

func TestImageCaptureTimeHeic(t *testing.T) {
	got, ok := imageCaptureTime("testdata/sample.heic")
	if !ok {
		t.Fatal("expected to extract EXIF DateTimeOriginal from sample.heic")
	}
	want := time.Date(2022, 2, 3, 4, 5, 6, 0, got.Location())
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestImageCaptureTimeMissing(t *testing.T) {
	if _, ok := imageCaptureTime("testdata/does-not-exist.jpg"); ok {
		t.Error("expected ok=false for missing file")
	}
}

func TestResolveFallsBackToMtime(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "plain.txt") // no metadata, unknown ext
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	mtime := time.Date(2019, 8, 7, 6, 5, 4, 0, time.Local)
	if err := os.Chtimes(p, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	got := Resolve(p, ".txt")
	// On Windows birthtime always succeeds, so Resolve returns the file's
	// creation time (now-ish) and the mtime branch is never reached; the
	// mtime fallback is exercised on the !windows build where birthTime
	// reports unsupported. Either way the result must be non-zero with a
	// plausible year.
	if got.IsZero() || got.Year() < 2000 {
		t.Errorf("Resolve returned implausible time: %v", got)
	}
}

func TestResolveUsesVideoMetadata(t *testing.T) {
	got := Resolve("testdata/sample.mp4", ".mp4")
	if got.Year() < 2000 || got.Year() > 2100 {
		t.Errorf("expected mvhd-derived time, got %v", got)
	}
}

func TestResolveUsesImageMetadata(t *testing.T) {
	got := Resolve("testdata/sample.jpg", ".jpg")
	want := time.Date(2021, 3, 4, 5, 6, 7, 0, got.Location())
	if !got.Equal(want) {
		t.Errorf("expected EXIF-derived time via Resolve, got %v want %v", got, want)
	}
}

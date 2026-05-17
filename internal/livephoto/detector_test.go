package livephoto

import (
	"path/filepath"
	"testing"
)

func TestPairedHeicStems_CollectsHeicEntries(t *testing.T) {
	files := []string{
		"/photos/IMG_1234.HEIC",
		"/photos/IMG_1234.MOV",
		"/photos/notes.txt",
	}
	got := PairedHeicStems(files)
	if len(got) != 1 {
		t.Fatalf("expected 1 heic stem, got %d (%v)", len(got), got)
	}
	wantKey := filepath.Join("/photos", "IMG_1234")
	if !got[wantKey] {
		t.Errorf("expected key %q in %v", wantKey, got)
	}
}

func TestPairedHeicStems_TrimsSurroundingQuotes(t *testing.T) {
	files := []string{`"/photos/IMG_9999.heic"`}
	got := PairedHeicStems(files)
	want := filepath.Join("/photos", "IMG_9999")
	if !got[want] {
		t.Errorf("expected key %q in %v", want, got)
	}
}

func TestPairedHeicStems_IgnoresNonHeic(t *testing.T) {
	files := []string{"/photos/foo.jpg", "/photos/bar.mov"}
	got := PairedHeicStems(files)
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestPairedHeicStems_CaseInsensitiveExtension(t *testing.T) {
	files := []string{"/photos/a.Heic", "/photos/b.HEIC", "/photos/c.heic"}
	got := PairedHeicStems(files)
	if len(got) != 3 {
		t.Errorf("expected 3 entries, got %d (%v)", len(got), got)
	}
}

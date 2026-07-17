package naming

import (
	"testing"
	"time"
)

func TestBuildCaptureName(t *testing.T) {
	ts := time.Date(2026, 7, 5, 10, 22, 43, 0, time.Local)
	cases := []struct {
		name   string
		format string
		stem   string
		seq    int
		want   string
	}{
		{"default preset", "{YY}-{MM}-{DD} {HH}-{mm}-{ss} {num}", "IMG_3574", 1, "26-07-05 10-22-43 3574"},
		{"four digit year", "{YYYY}-{MM}-{DD}_{HH}-{mm}-{ss}", "IMG_3574", 1, "2026-07-05_10-22-43"},
		{"pad short number", "{num}", "IMG_12", 1, "0012"},
		{"keep long number", "{num}", "PXL_123456", 1, "123456"},
		{"no number falls back to seq", "{num}", "MyVideo", 7, "0007"},
		{"explicit seq token", "{seq}", "IMG_3574", 42, "0042"},
		{"name token", "{name}_{YYYY}", "MyVideo", 1, "MyVideo_2026"},
		{"unknown token literal", "x{foo}y", "a", 1, "x{foo}y"},
		{"sanitize illegal chars", "{HH}:{mm}", "a", 1, "10-22"},
	}
	for _, tc := range cases {
		if got := BuildCaptureName(tc.format, ts, tc.stem, tc.seq); got != tc.want {
			t.Errorf("%s: BuildCaptureName(%q,...,%q,%d) = %q, want %q",
				tc.name, tc.format, tc.stem, tc.seq, got, tc.want)
		}
	}
}

func TestBuildCaptureNameEmptyFallsBackToStem(t *testing.T) {
	ts := time.Date(2026, 7, 5, 10, 22, 43, 0, time.Local)
	if got := BuildCaptureName("", ts, "Original", 1); got != "Original" {
		t.Errorf("empty format should fall back to stem, got %q", got)
	}
}

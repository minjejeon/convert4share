package converter

import (
	"strings"
	"testing"
)

func TestHeicDecodeHint(t *testing.T) {
	// The exact stderr emitted by ImageMagick 6 + libheif 1.17.6 on a newer
	// Apple HEIC that carries multiple auxiliary images (HDR gain map, mattes).
	oldLibheifOutput := "convert-im6.q16: Invalid input: Unspecified: Too many auxiliary image references (2.0) `/x/a.heic' @ error/heic.c/IsHEIFSuccess/139.\n" +
		"convert-im6.q16: no images defined `/x/a.jpg' @ error/convert.c/ConvertImageCommand/3234."

	hint := heicDecodeHint(oldLibheifOutput)
	if hint == "" {
		t.Fatalf("expected a hint for the old-libheif auxiliary-image failure, got empty string")
	}
	if !strings.Contains(strings.ToLower(hint), "libheif") {
		t.Errorf("hint should name libheif so the user knows what to upgrade; got %q", hint)
	}

	// A generic, unrelated magick failure must NOT trigger the HEIC hint.
	unrelated := "convert-im6.q16: unable to open image `/x/a.jpg': No such file or directory @ error/blob.c/OpenBlob/2964."
	if h := heicDecodeHint(unrelated); h != "" {
		t.Errorf("expected no hint for unrelated error, got %q", h)
	}

	// Empty output → no hint.
	if h := heicDecodeHint(""); h != "" {
		t.Errorf("expected no hint for empty output, got %q", h)
	}
}

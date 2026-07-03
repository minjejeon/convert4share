package converter

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// heicDecodeHint returns an actionable, user-facing hint when magick's output
// indicates a libheif HEIC-decode failure. The common cause is an outdated
// system libheif (e.g. the 1.17.x shipped by Ubuntu 24.04) that cannot decode
// newer Apple HEIC files carrying multiple auxiliary images (HDR gain map,
// depth/segmentation mattes). The bug is fixed in libheif 1.18+. Returns "" when
// the output does not look like this failure.
func heicDecodeHint(output string) string {
	lower := strings.ToLower(output)
	isHeicFailure := strings.Contains(lower, "too many auxiliary image references") ||
		(strings.Contains(lower, "heic.c") && strings.Contains(lower, "isheifsuccess"))
	if !isHeicFailure {
		return ""
	}
	return "HEIC decode failed: your system's libheif is too old to read this file's " +
		"auxiliary images (HDR gain map / depth). Upgrade libheif to 1.18 or newer and retry."
}

func (c *Config) Magick(ctx context.Context, orig, dest string) error {
	args := []string{orig}
	if c.MaxImageSize > 0 {
		args = append(args, "-resize", fmt.Sprintf("%dx%d>", c.MaxImageSize, c.MaxImageSize))
	}
	args = append(args, "-quality", "90", dest)

	cmd := prepareCommandContext(ctx, c.MagickBinary, args...)
	// Ensure standard input is closed to prevent magick from waiting for input
	cmd.Stdin = nil
	slog.Info("Running magick command", "command", cmd.String())

	// Use CombinedOutput to avoid hanging on Windows GUI if stdout/stderr are not consumed.
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("magick failed", "error", err, "command", cmd.String(), "output", string(output))
		if hint := heicDecodeHint(string(output)); hint != "" {
			return fmt.Errorf("%s (magick failed: %w). Output: %s", hint, err, string(output))
		}
		return fmt.Errorf("magick failed: %w. Output: %s", err, string(output))
	}
	return nil
}

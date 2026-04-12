package converter

import (
	"context"
	"fmt"
	"log"
)

func (c *Config) Magick(ctx context.Context, orig, dest string) error {
	args := []string{orig}
	if c.MaxImageSize > 0 {
		args = append(args, "-resize", fmt.Sprintf("%dx%d>", c.MaxImageSize, c.MaxImageSize))
	}
	args = append(args, "-quality", "90", dest)

	cmd := prepareCommandContext(ctx, c.MagickBinary, args...)
	// Ensure standard input is closed to prevent magick from waiting for input
	cmd.Stdin = nil
	log.Printf("Running magick command: %s", cmd.String())

	// Use CombinedOutput to avoid hanging on Windows GUI if stdout/stderr are not consumed.
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("magick failed: %w. Output: %s", err, string(output))
	}
	return nil
}

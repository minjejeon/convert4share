package converter

import (
	"context"
	"fmt"
	"log"
)

func (c *Config) Magick(ctx context.Context, orig, dest string) error {
	cmd := prepareCommandContext(ctx, c.MagickBinary, orig, dest)
	log.Printf("Running magick command: %s", cmd.String())

	// Use CombinedOutput to avoid hanging on Windows GUI if stdout/stderr are not consumed.
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("magick failed: %w. Output: %s", err, string(output))
	}
	return nil
}

package converter

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type Job struct{ Orig, Dest string }

type Config struct {
	MagickBinary        string
	FfmpegBinary        string
	MaxSize             int
	HardwareAccelerator string
	FfmpegCustomArgs    string
	VideoQuality        string // "high", "medium", "low"
}

type ProgressCallback func(progress int, speed string)

var (
	durationRegex = regexp.MustCompile(`Duration: (\d+):(\d{2}):(\d{2})\.(\d+)`)
	timeRegex     = regexp.MustCompile(`time=(\d+):(\d{2}):(\d{2})\.(\d+)`)
	speedRegex    = regexp.MustCompile(`speed=\s*([\d\.]+)x`)
)

func (c *Config) GenerateThumbnail(ctx context.Context, inputFile string) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(inputFile))
	var cmd *exec.Cmd
	var stdout bytes.Buffer

	// Preview size: 200px width, aspect ratio preserved

	if ext == ".mov" || ext == ".mp4" || ext == ".mkv" || ext == ".avi" {
		args := []string{
			"-hide_banner",
			"-loglevel", "error",
			"-ss", "00:00:00",
			"-i", inputFile,
			"-vframes", "1",
			"-vf", "scale=200:-1",
			"-f", "image2",
			"-c:v", "mjpeg",
			"pipe:1",
		}
		cmd = prepareCommandContext(ctx, c.FfmpegBinary, args...)
	} else {
		// Note: For HEIC, magick handles it if delegates are present.
		// We use input[0] to get the first frame/page.
		args := []string{
			inputFile + "[0]",
			"-resize", "200x200",
			"-quality", "80",
			"jpeg:-",
		}
		cmd = prepareCommandContext(ctx, c.MagickBinary, args...)
	}

	// Ensure standard input is closed
	cmd.Stdin = nil
	cmd.Stdout = &stdout
	// We can ignore stderr or capture it for debug
	// cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("thumbnail generation failed: %w", err)
	}

	return stdout.Bytes(), nil
}

func scanCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		return i + 1, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

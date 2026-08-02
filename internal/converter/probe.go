package converter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ResolveFfprobe picks the ffprobe executable to run.
//
// A configured value that names a path is trusted as-is. Otherwise the
// directory holding ffmpeg is tried first — ffprobe ships alongside
// ffmpeg in every distribution, so this resolves correctly even when
// only ffmpeg was configured by hand — before falling back to a PATH
// lookup.
func ResolveFfprobe(ffprobeBinary, ffmpegBinary string) (string, error) {
	if strings.ContainsAny(ffprobeBinary, `/\`) {
		return ffprobeBinary, nil
	}

	name := ffprobeBinary
	if name == "" {
		name = "ffprobe"
	}

	if strings.ContainsAny(ffmpegBinary, `/\`) {
		sibling := filepath.Join(filepath.Dir(ffmpegBinary), name)
		if runtime.GOOS == "windows" && filepath.Ext(sibling) == "" {
			sibling += ".exe"
		}
		if info, err := os.Stat(sibling); err == nil && !info.IsDir() {
			return sibling, nil
		}
	}

	return exec.LookPath(name)
}

// Probe reads the stream layout of a media file with ffprobe.
func (c *Config) Probe(ctx context.Context, path string) (MediaInfo, error) {
	bin, err := ResolveFfprobe(c.FfprobeBinary, c.FfmpegBinary)
	if err != nil {
		return MediaInfo{}, fmt.Errorf("ffprobe not found: %w", err)
	}

	args := []string{
		"-v", "error",
		"-show_streams",
		"-show_format",
		"-print_format", "json",
		path,
	}

	cmd := prepareCommandContext(ctx, bin, args...)
	cmd.Stdin = nil

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return MediaInfo{}, fmt.Errorf("ffprobe failed: %w. Log: %s", err, strings.TrimSpace(stderr.String()))
	}

	var info MediaInfo
	if err := json.Unmarshal(stdout.Bytes(), &info); err != nil {
		return MediaInfo{}, fmt.Errorf("could not parse ffprobe output: %w", err)
	}

	return info, nil
}

package converter

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

func TestBuildFfmpegArgsCopyAudio(t *testing.T) {
	c := &Config{MaxSize: 1920, VideoQuality: "high"}

	reencode := c.BuildFfmpegArgs("in.mp4", "out.mp4", false)
	if i := slices.Index(reencode, "-c:a"); i < 0 || reencode[i+1] != "aac" {
		t.Errorf("copyAudio=false: want -c:a aac, got %v", reencode)
	}

	copied := c.BuildFfmpegArgs("in.mp4", "out.mp4", true)
	if i := slices.Index(copied, "-c:a"); i < 0 || copied[i+1] != "copy" {
		t.Errorf("copyAudio=true: want -c:a copy, got %v", copied)
	}
}

func TestResolveFfprobePrefersConfiguredPath(t *testing.T) {
	configured := filepath.Join(t.TempDir(), "custom-ffprobe.exe")
	got, err := ResolveFfprobe(configured, "ffmpeg")
	if err != nil {
		t.Fatalf("ResolveFfprobe() error = %v", err)
	}
	if got != configured {
		t.Errorf("ResolveFfprobe() = %q, want the configured path %q", got, configured)
	}
}

func TestResolveFfprobeFindsSiblingOfFfmpeg(t *testing.T) {
	dir := t.TempDir()
	name := "ffprobe"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	sibling := filepath.Join(dir, name)
	if err := os.WriteFile(sibling, []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}

	ffmpegName := "ffmpeg"
	if runtime.GOOS == "windows" {
		ffmpegName += ".exe"
	}
	got, err := ResolveFfprobe("ffprobe", filepath.Join(dir, ffmpegName))
	if err != nil {
		t.Fatalf("ResolveFfprobe() error = %v", err)
	}
	if got != sibling {
		t.Errorf("ResolveFfprobe() = %q, want the sibling %q", got, sibling)
	}
}

// ffprobeForTest locates a usable ffprobe or skips the test.
func ffprobeForTest(t *testing.T) string {
	t.Helper()
	bin, err := exec.LookPath("ffprobe")
	if err != nil {
		t.Skip("ffprobe not on PATH")
	}
	return bin
}

func ffmpegForTest(t *testing.T) string {
	t.Helper()
	bin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not on PATH")
	}
	return bin
}

// synthesise renders a one-second clip with the given encoder args so
// the probe path is exercised against real container bytes.
func synthesise(t *testing.T, ffmpeg, dest string, encodeArgs ...string) {
	t.Helper()
	args := []string{
		"-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "testsrc=size=320x240:rate=10:duration=1",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=1",
	}
	args = append(args, encodeArgs...)
	args = append(args, dest)

	cmd := exec.Command(ffmpeg, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("could not synthesise fixture (%v): %s", err, out)
	}
}

func TestProbeRealFiles(t *testing.T) {
	ffprobe := ffprobeForTest(t)
	ffmpeg := ffmpegForTest(t)
	dir := t.TempDir()
	c := &Config{FfprobeBinary: ffprobe, FfmpegBinary: ffmpeg}

	t.Run("h264 aac is copied", func(t *testing.T) {
		dest := filepath.Join(dir, "h264.mp4")
		synthesise(t, ffmpeg, dest, "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac")

		info, err := c.Probe(context.Background(), dest)
		if err != nil {
			t.Fatalf("Probe() error = %v", err)
		}
		if need, reason := NeedsConversion(info, 1920); need {
			t.Errorf("NeedsConversion() = true (%s), want false", reason)
		}
		if !HasAACAudio(info) {
			t.Error("HasAACAudio() = false, want true")
		}
	})

	t.Run("hevc is converted", func(t *testing.T) {
		dest := filepath.Join(dir, "hevc.mp4")
		synthesise(t, ffmpeg, dest, "-c:v", "libx265", "-pix_fmt", "yuv420p", "-tag:v", "hvc1", "-c:a", "aac")

		info, err := c.Probe(context.Background(), dest)
		if err != nil {
			t.Fatalf("Probe() error = %v", err)
		}
		need, reason := NeedsConversion(info, 1920)
		if !need {
			t.Fatal("NeedsConversion() = false, want true for HEVC")
		}
		if reason == "" {
			t.Error("expected a reason explaining the HEVC verdict")
		}
	})

	t.Run("oversized h264 is converted", func(t *testing.T) {
		dest := filepath.Join(dir, "big.mp4")
		synthesise(t, ffmpeg, dest,
			"-c:v", "libx264", "-pix_fmt", "yuv420p", "-vf", "scale=3840:2160", "-c:a", "aac")

		info, err := c.Probe(context.Background(), dest)
		if err != nil {
			t.Fatalf("Probe() error = %v", err)
		}
		if need, _ := NeedsConversion(info, 1920); !need {
			t.Error("NeedsConversion() = false, want true for a 3840x2160 source")
		}
	})

	t.Run("missing file errors", func(t *testing.T) {
		if _, err := c.Probe(context.Background(), filepath.Join(dir, "nope.mp4")); err == nil {
			t.Error("Probe() on a missing file returned nil error")
		}
	})
}

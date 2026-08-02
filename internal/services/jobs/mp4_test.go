package jobs

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/minjejeon/convert4share/internal/converter"
)

// mp4Fixtures builds the converter config used by the mp4 tests, or
// skips when the external tools are unavailable.
func mp4Fixtures(t *testing.T) *converter.Config {
	t.Helper()
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not on PATH")
	}
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		t.Skip("ffprobe not on PATH")
	}
	return &converter.Config{
		FfmpegBinary:  ffmpeg,
		FfprobeBinary: ffprobe,
		MaxSize:       1920,
		VideoQuality:  "low",
	}
}

// writeClip renders a short clip with the given encoder settings.
func writeClip(t *testing.T, cfg *converter.Config, dest string, encodeArgs ...string) {
	t.Helper()
	args := []string{
		"-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "testsrc=size=320x240:rate=10:duration=1",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=1",
	}
	args = append(args, encodeArgs...)
	args = append(args, dest)
	if out, err := exec.Command(cfg.FfmpegBinary, args...).CombinedOutput(); err != nil {
		t.Skipf("could not build fixture (%v): %s", err, out)
	}
}

func videoCodecOf(t *testing.T, cfg *converter.Config, path string) string {
	t.Helper()
	info, err := cfg.Probe(context.Background(), path)
	if err != nil {
		t.Fatalf("Probe(%s) error = %v", path, err)
	}
	for _, s := range info.Streams {
		if s.CodecType == "video" {
			return s.CodecName
		}
	}
	return ""
}

// inputFor assembles a runOneInput for a source file, with a reporter
// that discards events.
func inputFor(cfg *converter.Config, src, destDir, collision string) runOneInput {
	stem := filepath.Base(src)
	stem = stem[:len(stem)-len(filepath.Ext(stem))]
	return runOneInput{
		jobID:           src,
		src:             src,
		ext:             ".mp4",
		stem:            stem,
		destDir:         destDir,
		collisionOption: collision,
		convConfig:      cfg,
		report:          func(id, dest string, percent int, status, errMsg, speed string) {},
	}
}

func TestProcessMP4CopiesCompatibleFile(t *testing.T) {
	cfg := mp4Fixtures(t)
	srcDir, destDir := t.TempDir(), t.TempDir()
	src := filepath.Join(srcDir, "clip.mp4")
	writeClip(t, cfg, src, "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac")

	before, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}

	s := New(nil)
	dest, note, err := s.processMP4(context.Background(), inputFor(cfg, src, destDir, "rename"))
	if err != nil {
		t.Fatalf("processMP4() error = %v", err)
	}
	if note == "" {
		t.Error("expected a note explaining the copy")
	}

	after, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading destination: %v", err)
	}
	if len(after) != len(before) {
		t.Errorf("destination is %d bytes, want a byte-for-byte copy of %d", len(after), len(before))
	}
}

func TestProcessMP4ConvertsHEVC(t *testing.T) {
	cfg := mp4Fixtures(t)
	srcDir, destDir := t.TempDir(), t.TempDir()
	src := filepath.Join(srcDir, "clip.mp4")
	writeClip(t, cfg, src, "-c:v", "libx265", "-pix_fmt", "yuv420p", "-tag:v", "hvc1", "-c:a", "aac")

	if got := videoCodecOf(t, cfg, src); got != "hevc" {
		t.Fatalf("fixture is %s, want hevc", got)
	}

	s := New(nil)
	dest, _, err := s.processMP4(context.Background(), inputFor(cfg, src, destDir, "rename"))
	if err != nil {
		t.Fatalf("processMP4() error = %v", err)
	}
	if got := videoCodecOf(t, cfg, dest); got != "h264" {
		t.Errorf("destination codec = %s, want h264", got)
	}

	entries, err := os.ReadDir(destDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".mp4" && len(e.Name()) > len("c4s-tmp") &&
			filepath.Base(dest) != e.Name() {
			t.Errorf("leftover file in destination: %s", e.Name())
		}
	}
}

// The regression this feature could most easily introduce: with
// collisionOption "overwrite" and the output landing in the source
// directory, dest resolves to src. Handing ffmpeg one path as both
// input and output destroys the original.
func TestProcessMP4InPlaceOverwriteKeepsOriginal(t *testing.T) {
	cfg := mp4Fixtures(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "clip.mp4")
	writeClip(t, cfg, src, "-c:v", "libx265", "-pix_fmt", "yuv420p", "-tag:v", "hvc1", "-c:a", "aac")

	s := New(nil)
	dest, _, err := s.processMP4(context.Background(), inputFor(cfg, src, dir, "overwrite"))
	if err != nil {
		t.Fatalf("processMP4() error = %v", err)
	}
	if dest != src {
		t.Fatalf("expected dest to resolve onto src (%s), got %s", src, dest)
	}

	info, err := os.Stat(src)
	if err != nil {
		t.Fatalf("source vanished: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("source was truncated to 0 bytes")
	}
	if got := videoCodecOf(t, cfg, src); got != "h264" {
		t.Errorf("in-place result codec = %s, want h264", got)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Errorf("expected only the converted file to remain, got %v", names)
	}
}

// A compatible file whose destination is itself must be left strictly
// alone — copyFile would truncate it through its own handle.
func TestProcessMP4InPlaceCompatibleIsUntouched(t *testing.T) {
	cfg := mp4Fixtures(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "clip.mp4")
	writeClip(t, cfg, src, "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac")

	before, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}

	s := New(nil)
	dest, _, err := s.processMP4(context.Background(), inputFor(cfg, src, dir, "overwrite"))
	if err != nil {
		t.Fatalf("processMP4() error = %v", err)
	}
	if dest != src {
		t.Fatalf("expected dest to resolve onto src (%s), got %s", src, dest)
	}

	after, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Errorf("file changed: %d bytes now, %d before", len(after), len(before))
	}
}

// With no usable ffprobe the verdict cannot be trusted, so the file is
// converted rather than passed through unverified.
func TestProcessMP4ConvertsWhenProbeFails(t *testing.T) {
	cfg := mp4Fixtures(t)
	dir, destDir := t.TempDir(), t.TempDir()
	src := filepath.Join(dir, "clip.mp4")
	writeClip(t, cfg, src, "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac")

	broken := *cfg
	broken.FfprobeBinary = filepath.Join(dir, "no-such-ffprobe")
	broken.FfmpegBinary = cfg.FfmpegBinary

	s := New(nil)
	dest, note, err := s.processMP4(context.Background(), inputFor(&broken, src, destDir, "rename"))
	if err != nil {
		t.Fatalf("processMP4() error = %v", err)
	}
	if got := videoCodecOf(t, cfg, dest); got != "h264" {
		t.Errorf("destination codec = %s, want h264", got)
	}
	if note == "" {
		t.Error("expected a note explaining the conversion")
	}
}

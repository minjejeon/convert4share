package converter

import (
	"testing"
)

// BenchmarkFfmpegOutputParsing benchmarks the regex matching used in Ffmpeg parsing.
// This confirms the performance of the package-level compiled regexes.
func BenchmarkFfmpegOutputParsing(b *testing.B) {
	line := "frame=  123 fps=0.0 q=0.0 size=       0kB time=00:00:05.12 bitrate=   0.0kbits/s speed=10.2x"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Accessing package-level private variables directly
		timeRegex.FindStringSubmatch(line)
		durationRegex.FindStringSubmatch(line)
	}
}

func TestBuildFfmpegArgs(t *testing.T) {
	c := Config{
		MagickBinary:        "magick",
		FfmpegBinary:        "ffmpeg",
		MaxSize:             1920,
		HardwareAccelerator: "amd",
		VideoQuality:        "high",
	}

	args := c.BuildFfmpegArgs("input.mov", "output.mp4")

	// Helper to check if arg exists
	hasArg := func(arg string) bool {
		for _, a := range args {
			if a == arg {
				return true
			}
		}
		return false
	}

	if hasArg("-preanalysis") {
		t.Error("Expected -preanalysis to be removed for AMD High quality")
	}

	if !hasArg("-rc") {
		t.Error("Expected -rc for AMD High quality")
	}

	foundMaxRate := false
	for i, a := range args {
		if a == "-maxrate" && i+1 < len(args) && args[i+1] == "10M" {
			foundMaxRate = true
			break
		}
	}
	if !foundMaxRate {
		t.Error("Expected -maxrate 10M for AMD High quality")
	}
}

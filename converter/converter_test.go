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

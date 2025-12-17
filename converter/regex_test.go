package converter

import (
	"testing"
)

func TestRegexParsing(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool // expect a match
	}{
		{
			name:     "Standard Duration",
			line:     "Duration: 01:30:15.50, start: 0.000000, bitrate: 1234 kb/s",
			expected: true,
		},
		{
			name:     "Long Duration (100 hours)",
			line:     "Duration: 100:00:05.00, start: 0.000000, bitrate: 1234 kb/s",
			expected: true,
		},
		{
			name:     "Standard Time",
			line:     "frame= 100 fps= 25 q=28.0 size=    1024kB time=00:00:04.50 bitrate=2000.0kbits/s speed=1.0x",
			expected: true,
		},
		{
			name:     "Long Time (100 hours)",
			line:     "frame=99999 fps= 25 q=28.0 size=9999999kB time=100:00:04.50 bitrate=2000.0kbits/s speed=1.0x",
			expected: true,
		},
		{
			name:     "Duration with 1 digit ms",
			line:     "Duration: 00:00:01.5, start: 0.000000",
			expected: true,
		},
		{
			name:     "Time with 3 digit ms",
			line:     "time=00:00:01.500 bitrate=...",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var match bool
			if tt.name == "Standard Duration" || tt.name == "Long Duration (100 hours)" || tt.name == "Duration with 1 digit ms" {
				matches := durationRegex.FindStringSubmatch(tt.line)
				match = len(matches) == 5
			} else {
				matches := timeRegex.FindStringSubmatch(tt.line)
				match = len(matches) == 5
			}

			if match != tt.expected {
				t.Errorf("expected match=%v, got %v for line: %s", tt.expected, match, tt.line)
			}
		})
	}
}

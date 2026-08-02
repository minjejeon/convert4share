package converter

import "testing"

func videoStream(codec, pixFmt string, w, h int) Stream {
	return Stream{CodecType: "video", CodecName: codec, PixFmt: pixFmt, Width: w, Height: h}
}

func audioStream(codec string) Stream {
	return Stream{CodecType: "audio", CodecName: codec}
}

func TestNeedsConversion(t *testing.T) {
	coverArt := Stream{CodecType: "video", CodecName: "mjpeg", PixFmt: "yuvj420p", Width: 600, Height: 600}
	coverArt.Disposition.AttachedPic = 1

	tests := []struct {
		name    string
		streams []Stream
		maxSize int
		want    bool
	}{
		{
			name:    "h264 yuv420p within maxSize with aac copies",
			streams: []Stream{videoStream("h264", "yuv420p", 1080, 1920), audioStream("aac")},
			maxSize: 1920,
			want:    false,
		},
		{
			// The file that motivated this feature: KakaoTalk HEVC export.
			name:    "hevc converts",
			streams: []Stream{videoStream("hevc", "yuv420p", 1080, 1920), audioStream("aac")},
			maxSize: 1920,
			want:    true,
		},
		{
			name:    "vp9 converts",
			streams: []Stream{videoStream("vp9", "yuv420p", 640, 480)},
			maxSize: 1920,
			want:    true,
		},
		{
			name:    "h264 high 10 converts",
			streams: []Stream{videoStream("h264", "yuv420p10le", 640, 480)},
			maxSize: 1920,
			want:    true,
		},
		{
			name:    "h264 4:2:2 converts",
			streams: []Stream{videoStream("h264", "yuv422p", 640, 480)},
			maxSize: 1920,
			want:    true,
		},
		{
			name:    "landscape wider than maxSize converts",
			streams: []Stream{videoStream("h264", "yuv420p", 3840, 2160)},
			maxSize: 1920,
			want:    true,
		},
		{
			name:    "portrait taller than maxSize converts",
			streams: []Stream{videoStream("h264", "yuv420p", 2160, 3840)},
			maxSize: 1920,
			want:    true,
		},
		{
			name:    "exactly maxSize copies",
			streams: []Stream{videoStream("h264", "yuv420p", 1920, 1080)},
			maxSize: 1920,
			want:    false,
		},
		{
			name:    "maxSize unset skips the resolution check",
			streams: []Stream{videoStream("h264", "yuv420p", 3840, 2160)},
			maxSize: 0,
			want:    false,
		},
		{
			name:    "mp3 audio copies",
			streams: []Stream{videoStream("h264", "yuv420p", 640, 480), audioStream("mp3")},
			maxSize: 1920,
			want:    false,
		},
		{
			name:    "ac3 audio converts",
			streams: []Stream{videoStream("h264", "yuv420p", 640, 480), audioStream("ac3")},
			maxSize: 1920,
			want:    true,
		},
		{
			name:    "any incompatible audio track converts",
			streams: []Stream{videoStream("h264", "yuv420p", 640, 480), audioStream("aac"), audioStream("ac3")},
			maxSize: 1920,
			want:    true,
		},
		{
			name:    "no audio copies",
			streams: []Stream{videoStream("h264", "yuv420p", 640, 480)},
			maxSize: 1920,
			want:    false,
		},
		{
			// Nothing for the scale filter to work on; leave it alone.
			name:    "audio only copies",
			streams: []Stream{audioStream("aac")},
			maxSize: 1920,
			want:    false,
		},
		{
			name:    "no streams copies",
			streams: nil,
			maxSize: 1920,
			want:    false,
		},
		{
			// Cover art is an attached picture, not the video track.
			name:    "cover art is not treated as video",
			streams: []Stream{coverArt, videoStream("hevc", "yuv420p", 640, 480)},
			maxSize: 1920,
			want:    true,
		},
		{
			name:    "cover art alongside compatible video copies",
			streams: []Stream{coverArt, videoStream("h264", "yuv420p", 640, 480)},
			maxSize: 1920,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, reason := NeedsConversion(MediaInfo{Streams: tt.streams}, tt.maxSize)
			if got != tt.want {
				t.Errorf("NeedsConversion() = %v (%q), want %v", got, reason, tt.want)
			}
			if got && reason == "" {
				t.Error("NeedsConversion() returned true with an empty reason")
			}
			if !got && reason != "" {
				t.Errorf("NeedsConversion() returned false with reason %q, want empty", reason)
			}
		})
	}
}

func TestHasAACAudio(t *testing.T) {
	tests := []struct {
		name    string
		streams []Stream
		want    bool
	}{
		{"aac", []Stream{videoStream("h264", "yuv420p", 640, 480), audioStream("aac")}, true},
		{"mp3", []Stream{videoStream("h264", "yuv420p", 640, 480), audioStream("mp3")}, false},
		{"no audio", []Stream{videoStream("h264", "yuv420p", 640, 480)}, false},
		{"mixed", []Stream{audioStream("aac"), audioStream("ac3")}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasAACAudio(MediaInfo{Streams: tt.streams}); got != tt.want {
				t.Errorf("HasAACAudio() = %v, want %v", got, tt.want)
			}
		})
	}
}

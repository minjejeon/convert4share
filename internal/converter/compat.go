package converter

import (
	"fmt"
	"strings"
)

// Stream is the subset of an ffprobe stream entry the compatibility
// rule needs.
type Stream struct {
	CodecType   string `json:"codec_type"`
	CodecName   string `json:"codec_name"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	PixFmt      string `json:"pix_fmt"`
	Disposition struct {
		AttachedPic int `json:"attached_pic"`
	} `json:"disposition"`
}

// MediaInfo is a decoded `ffprobe -show_streams -print_format json`
// result.
type MediaInfo struct {
	Streams []Stream `json:"streams"`
}

// compatibleVideoCodec is the only video codec that plays everywhere
// Convert4Share targets.
const compatibleVideoCodec = "h264"

// compatiblePixelFormat excludes High 10 (yuv420p10le) and 4:2:2
// H.264 variants, which fail on the same players HEVC does.
const compatiblePixelFormat = "yuv420p"

// compatibleAudioCodecs are the audio codecs an mp4 may carry without
// forcing a re-encode.
var compatibleAudioCodecs = map[string]bool{"aac": true, "mp3": true}

// videoStream returns the first real video stream, ignoring cover art
// (attached_pic), and reports whether one was found.
func (m MediaInfo) videoStream() (Stream, bool) {
	for _, s := range m.Streams {
		if s.CodecType == "video" && s.Disposition.AttachedPic == 0 {
			return s, true
		}
	}
	return Stream{}, false
}

// NeedsConversion reports whether a container must be re-encoded to be
// widely shareable, along with a human-readable reason when it must.
// The reason is empty when the file can be copied as-is.
//
// A file is copied only when its video track is H.264 in yuv420p, fits
// within maxSize on its longest side, and every audio track is AAC or
// MP3. maxSize <= 0 disables the resolution check. A file with no video
// track (audio-only) is copied: there is nothing to scale or re-encode.
func NeedsConversion(info MediaInfo, maxSize int) (bool, string) {
	video, ok := info.videoStream()
	if !ok {
		return false, ""
	}

	if !strings.EqualFold(video.CodecName, compatibleVideoCodec) {
		return true, fmt.Sprintf("video codec %s (want %s)", video.CodecName, compatibleVideoCodec)
	}

	if !strings.EqualFold(video.PixFmt, compatiblePixelFormat) {
		return true, fmt.Sprintf("pixel format %s (want %s)", video.PixFmt, compatiblePixelFormat)
	}

	// Display dimensions, not coded_width/coded_height: encoders pad to
	// macroblock boundaries (1080 displays as 1088 coded).
	longest := video.Width
	if video.Height > longest {
		longest = video.Height
	}
	if maxSize > 0 && longest > maxSize {
		return true, fmt.Sprintf("resolution %dx%d exceeds %dpx", video.Width, video.Height, maxSize)
	}

	for _, s := range info.Streams {
		if s.CodecType != "audio" {
			continue
		}
		if !compatibleAudioCodecs[strings.ToLower(s.CodecName)] {
			return true, fmt.Sprintf("audio codec %s (want aac or mp3)", s.CodecName)
		}
	}

	return false, ""
}

// HasAACAudio reports whether every audio track is already AAC, which
// lets the ffmpeg call stream-copy the audio instead of re-encoding it.
// False when there is no audio at all — there is nothing to copy.
func HasAACAudio(info MediaInfo) bool {
	found := false
	for _, s := range info.Streams {
		if s.CodecType != "audio" {
			continue
		}
		if !strings.EqualFold(s.CodecName, "aac") {
			return false
		}
		found = true
	}
	return found
}

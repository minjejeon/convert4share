// Package capturetime resolves a best-effort "capture time" for a media
// file: the moment it was recorded, from embedded metadata where
// available, otherwise from filesystem timestamps.
package capturetime

import (
	"os"
	"strings"
	"time"

	"github.com/abema/go-mp4"
	exif "github.com/dsoprea/go-exif/v3"
)

// mp4Epoch is the MP4/QuickTime time base: 1904-01-01 00:00:00 UTC.
var mp4Epoch = time.Date(1904, 1, 1, 0, 0, 0, 0, time.UTC)

// videoCreationTime reads the moov/mvhd creation time from an MP4/MOV
// file. The value is treated as wall-clock (formatted by field later),
// so it is constructed in UTC and never zone-shifted.
func videoCreationTime(path string) (time.Time, bool) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, false
	}
	defer f.Close()

	boxes, err := mp4.ExtractBoxWithPayload(f, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeMvhd()})
	if err != nil || len(boxes) == 0 {
		return time.Time{}, false
	}
	mvhd, ok := boxes[0].Payload.(*mp4.Mvhd)
	if !ok {
		return time.Time{}, false
	}

	var secs uint64
	if mvhd.GetVersion() == 0 {
		secs = uint64(mvhd.CreationTimeV0)
	} else {
		secs = mvhd.CreationTimeV1
	}
	if secs == 0 {
		return time.Time{}, false
	}
	return mp4Epoch.Add(time.Duration(secs) * time.Second), true
}

// imageCaptureTime reads EXIF DateTimeOriginal from an image file
// (JPEG or HEIC — the EXIF TIFF block is located by signature scan, so
// both containers work). The EXIF value has no zone; it is parsed as
// local wall-clock.
func imageCaptureTime(path string) (time.Time, bool) {
	rawExif, err := exif.SearchFileAndExtractExif(path)
	if err != nil {
		return time.Time{}, false
	}
	entries, _, err := exif.GetFlatExifData(rawExif, nil)
	if err != nil {
		return time.Time{}, false
	}
	for _, e := range entries {
		if e.TagName == "DateTimeOriginal" {
			if t, perr := time.ParseInLocation("2006:01:02 15:04:05", e.Formatted, time.Local); perr == nil {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

// Resolve returns the best-effort capture time for path, whose lowercase
// extension is ext (e.g. ".mp4"). Order: embedded metadata, then file
// creation time (birthtime), then modification time. It always returns
// a usable time for an existing file.
func Resolve(path, ext string) time.Time {
	switch strings.ToLower(ext) {
	case ".mov", ".mp4":
		if t, ok := videoCreationTime(path); ok {
			return t
		}
	case ".heic", ".jpg", ".jpeg":
		if t, ok := imageCaptureTime(path); ok {
			return t
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	if t, ok := birthTime(info); ok {
		return t
	}
	return info.ModTime()
}

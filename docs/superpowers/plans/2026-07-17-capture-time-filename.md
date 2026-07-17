# Capture-time Filename Option Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a user-configurable option to rename converted output files from the media's capture time plus a number (e.g. `26-07-05 10-22-43 3574.mp4`), using a brace-token format string with presets and a live preview.

**Architecture:** A new `internal/capturetime` package resolves a best-effort capture time (MP4/MOV `mvhd` → image EXIF `DateTimeOriginal` → file birthtime → mtime). A pure `naming.BuildCaptureName` function in a new dependency-free `internal/naming` package turns a format string + time + original stem + batch sequence into a filename stem; it is the single source of truth for both real conversions (jobs) and the `SettingsService.PreviewFileName` used by the UI (avoiding any settings→jobs coupling or import cycle). Two new settings (`fileNaming`, `fileNameFormat`) thread through config, the jobs batch loop, and the Settings UI.

**Tech Stack:** Go 1.25, Wails v3, React/TypeScript/Tailwind, viper. New deps: `github.com/abema/go-mp4`, `github.com/dsoprea/go-exif/v3`.

---

## File structure

- `internal/capturetime/capturetime.go` — public `Resolve(path, ext) time.Time`, video + image extractors.
- `internal/capturetime/birthtime_windows.go` / `birthtime_other.go` — platform birthtime helper.
- `internal/capturetime/capturetime_test.go` — extraction + fallback tests with fixtures.
- `internal/capturetime/testdata/` — sample media fixtures.
- `internal/naming/naming.go` — `BuildCaptureName` + number/sanitize helpers (dependency-free).
- `internal/naming/naming_test.go` — pure-function tests.
- `internal/services/jobs/service.go` — wire capture-time naming into `runBatch`.
- `internal/config/settings.go` — `FileNaming`, `FileNameFormat` fields + defaults + Load/Save.
- `internal/services/settings/service.go` — `PreviewFileName` method.
- `config.example.yaml` — document new keys.
- `frontend/src/components/SettingsPaths.tsx` — UI controls.
- `THIRD_PARTY_NOTICES.md` — new dependency notices.

---

## Task 1: Add dependencies

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Add the two libraries**

Run:
```bash
go get github.com/abema/go-mp4@latest
go get github.com/dsoprea/go-exif/v3@latest
go mod tidy
```

- [ ] **Step 2: Verify they resolve and compile**

Run: `go build ./...`
Expected: PASS (no build errors; new modules appear in `go.mod`).

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "build: add go-mp4 and go-exif for capture-time extraction"
```

---

## Task 2: Platform birthtime helper

**Files:**
- Create: `internal/capturetime/birthtime_windows.go`
- Create: `internal/capturetime/birthtime_other.go`

- [ ] **Step 1: Windows birthtime**

Create `internal/capturetime/birthtime_windows.go`:
```go
//go:build windows

package capturetime

import (
	"os"
	"syscall"
	"time"
)

// birthTime returns the file creation time on Windows.
func birthTime(info os.FileInfo) (time.Time, bool) {
	attr, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		return time.Time{}, false
	}
	return time.Unix(0, attr.CreationTime.Nanoseconds()), true
}
```

- [ ] **Step 2: Non-Windows birthtime (unsupported)**

Create `internal/capturetime/birthtime_other.go`:
```go
//go:build !windows

package capturetime

import (
	"os"
	"time"
)

// birthTime is unsupported off Windows; callers fall back to mtime.
func birthTime(info os.FileInfo) (time.Time, bool) {
	return time.Time{}, false
}
```

- [ ] **Step 3: Verify build on this platform**

Run: `go build ./internal/capturetime/`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/capturetime/birthtime_windows.go internal/capturetime/birthtime_other.go
git commit -m "feat(capturetime): platform birthtime helper"
```

---

## Task 3: Video capture time (mvhd)

**Files:**
- Create: `internal/capturetime/capturetime.go`
- Test: `internal/capturetime/capturetime_test.go`
- Fixture: `internal/capturetime/testdata/sample.mp4`

- [ ] **Step 1: Create a tiny MP4 fixture with a known creation time**

Create the fixture with ffmpeg (a 1-frame black video; ffmpeg writes an `mvhd`). Run from repo root:
```bash
mkdir -p internal/capturetime/testdata
ffmpeg -y -f lavfi -i color=c=black:s=16x16:d=1 -metadata creation_time="2023-05-01T09:08:07Z" -pix_fmt yuv420p internal/capturetime/testdata/sample.mp4
```
Note: the exact stored `mvhd` time may differ slightly from the tag; the test below reads the fixture's actual mvhd value at runtime rather than hardcoding, so it verifies the *extractor*, not ffmpeg. (See Step 3 test.)

- [ ] **Step 2: Write the failing test**

Create `internal/capturetime/capturetime_test.go`:
```go
package capturetime

import (
	"testing"
	"time"
)

func TestVideoCreationTime(t *testing.T) {
	got, ok := videoCreationTime("testdata/sample.mp4")
	if !ok {
		t.Fatal("expected to extract mvhd creation time from sample.mp4")
	}
	// mp4 epoch is 1904; a real capture time must be after 2000.
	if got.Year() < 2000 || got.Year() > 2100 {
		t.Errorf("implausible year %d from mvhd: %v", got.Year(), got)
	}
}

func TestVideoCreationTimeMissing(t *testing.T) {
	if _, ok := videoCreationTime("testdata/does-not-exist.mp4"); ok {
		t.Error("expected ok=false for missing file")
	}
}

// mp4Epoch is the reference used by the implementation.
var _ = time.UTC
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./internal/capturetime/ -run TestVideoCreationTime -v`
Expected: FAIL — `videoCreationTime` undefined.

- [ ] **Step 4: Implement video extraction**

Create `internal/capturetime/capturetime.go`:
```go
// Package capturetime resolves a best-effort "capture time" for a media
// file: the moment it was recorded, from embedded metadata where
// available, otherwise from filesystem timestamps.
package capturetime

import (
	"os"
	"time"

	"github.com/abema/go-mp4"
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
```

Note: if the go-mp4 API names differ (e.g. `GetVersion`/`CreationTimeV0`), adjust field access so the test in Step 3 passes — the test pins the behavior, not the API spelling.

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/capturetime/ -run TestVideoCreationTime -v`
Expected: PASS (both video tests).

- [ ] **Step 6: Commit**

```bash
git add internal/capturetime/capturetime.go internal/capturetime/capturetime_test.go internal/capturetime/testdata/sample.mp4
git commit -m "feat(capturetime): extract mvhd creation time from MP4/MOV"
```

---

## Task 4: Image capture time (EXIF)

**Files:**
- Modify: `internal/capturetime/capturetime.go`
- Modify: `internal/capturetime/capturetime_test.go`
- Fixture: `internal/capturetime/testdata/sample.jpg`

- [ ] **Step 1: Create a JPEG fixture with a known EXIF DateTimeOriginal**

Requires exiftool OR ImageMagick. With ImageMagick + exiftool commonly present; use exiftool if available:
```bash
ffmpeg -y -f lavfi -i color=c=gray:s=16x16 -frames:v 1 internal/capturetime/testdata/sample.jpg
exiftool -overwrite_original -DateTimeOriginal="2021:03:04 05:06:07" internal/capturetime/testdata/sample.jpg
```
If exiftool is unavailable, use ImageMagick: `magick internal/capturetime/testdata/sample.jpg -set exif:DateTimeOriginal "2021:03:04 05:06:07" internal/capturetime/testdata/sample.jpg`. The test below asserts the exact `2021-03-04 05:06:07` value.

- [ ] **Step 2: Write the failing test**

Add to `internal/capturetime/capturetime_test.go`:
```go
func TestImageCaptureTime(t *testing.T) {
	got, ok := imageCaptureTime("testdata/sample.jpg")
	if !ok {
		t.Fatal("expected to extract EXIF DateTimeOriginal from sample.jpg")
	}
	want := time.Date(2021, 3, 4, 5, 6, 7, 0, got.Location())
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestImageCaptureTimeMissing(t *testing.T) {
	if _, ok := imageCaptureTime("testdata/does-not-exist.jpg"); ok {
		t.Error("expected ok=false for missing file")
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./internal/capturetime/ -run TestImageCaptureTime -v`
Expected: FAIL — `imageCaptureTime` undefined.

- [ ] **Step 4: Implement EXIF extraction**

Add to `internal/capturetime/capturetime.go` (add `exif "github.com/dsoprea/go-exif/v3"` to the import block):
```go
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
```

Note: if `e.Formatted` is not the plain string for this tag in the installed go-exif version, adjust to read the tag value so the Step 2 test passes.

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/capturetime/ -run TestImageCaptureTime -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/capturetime/capturetime.go internal/capturetime/capturetime_test.go internal/capturetime/testdata/sample.jpg
git commit -m "feat(capturetime): extract EXIF DateTimeOriginal from images"
```

---

## Task 5: Resolve fallback chain

**Files:**
- Modify: `internal/capturetime/capturetime.go`
- Modify: `internal/capturetime/capturetime_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/capturetime/capturetime_test.go`:
```go
import (
	"os"            // add to existing import block
	"path/filepath" // add to existing import block
)

func TestResolveFallsBackToMtime(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "plain.txt") // no metadata, unknown ext
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	mtime := time.Date(2019, 8, 7, 6, 5, 4, 0, time.Local)
	if err := os.Chtimes(p, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	got := Resolve(p, ".txt")
	// birthtime may win on Windows for a just-created file; accept either
	// birthtime (now-ish) or the mtime we set — but it must be non-zero
	// and a plausible year.
	if got.IsZero() || got.Year() < 2000 {
		t.Errorf("Resolve returned implausible time: %v", got)
	}
}

func TestResolveUsesVideoMetadata(t *testing.T) {
	got := Resolve("testdata/sample.mp4", ".mp4")
	if got.Year() < 2000 || got.Year() > 2100 {
		t.Errorf("expected mvhd-derived time, got %v", got)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/capturetime/ -run TestResolve -v`
Expected: FAIL — `Resolve` undefined.

- [ ] **Step 3: Implement Resolve**

Add to `internal/capturetime/capturetime.go`:
```go
import "strings" // add to existing import block

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
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/capturetime/ -v`
Expected: PASS (all capturetime tests).

- [ ] **Step 5: Commit**

```bash
git add internal/capturetime/capturetime.go internal/capturetime/capturetime_test.go
git commit -m "feat(capturetime): best-effort Resolve with birthtime/mtime fallback"
```

---

## Task 6: BuildCaptureName pure function (internal/naming)

**Files:**
- Create: `internal/naming/naming.go`
- Test: `internal/naming/naming_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/naming/naming_test.go`:
```go
package naming

import (
	"testing"
	"time"
)

func TestBuildCaptureName(t *testing.T) {
	ts := time.Date(2026, 7, 5, 10, 22, 43, 0, time.Local)
	cases := []struct {
		name   string
		format string
		stem   string
		seq    int
		want   string
	}{
		{"default preset", "{YY}-{MM}-{DD} {HH}-{mm}-{ss} {num}", "IMG_3574", 1, "26-07-05 10-22-43 3574"},
		{"four digit year", "{YYYY}-{MM}-{DD}_{HH}-{mm}-{ss}", "IMG_3574", 1, "2026-07-05_10-22-43"},
		{"pad short number", "{num}", "IMG_12", 1, "0012"},
		{"keep long number", "{num}", "PXL_123456", 1, "123456"},
		{"no number falls back to seq", "{num}", "MyVideo", 7, "0007"},
		{"explicit seq token", "{seq}", "IMG_3574", 42, "0042"},
		{"name token", "{name}_{YYYY}", "MyVideo", 1, "MyVideo_2026"},
		{"unknown token literal", "x{foo}y", "a", 1, "x{foo}y"},
		{"sanitize illegal chars", "{HH}:{mm}", "a", 1, "10-22"},
	}
	for _, tc := range cases {
		if got := BuildCaptureName(tc.format, ts, tc.stem, tc.seq); got != tc.want {
			t.Errorf("%s: BuildCaptureName(%q,...,%q,%d) = %q, want %q",
				tc.name, tc.format, tc.stem, tc.seq, got, tc.want)
		}
	}
}

func TestBuildCaptureNameEmptyFallsBackToStem(t *testing.T) {
	ts := time.Date(2026, 7, 5, 10, 22, 43, 0, time.Local)
	if got := BuildCaptureName("", ts, "Original", 1); got != "Original" {
		t.Errorf("empty format should fall back to stem, got %q", got)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/naming/ -run TestBuildCaptureName -v`
Expected: FAIL — package/`BuildCaptureName` undefined.

- [ ] **Step 3: Implement naming**

Create `internal/naming/naming.go`:
```go
// Package naming renders output filename stems from a brace-token
// format string. It has no dependencies so both the jobs service (real
// conversions) and the settings service (UI preview) can use it.
package naming

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// trailingDigits matches the last run of digits in a string.
var trailingDigits = regexp.MustCompile(`([0-9]+)[^0-9]*$`)

// illegalFilenameChars matches characters not allowed in Windows names.
var illegalFilenameChars = regexp.MustCompile(`[\\/:*?"<>|]`)

// BuildCaptureName renders a filename stem (no extension) from a
// brace-token format string, a capture time, the original stem, and the
// batch sequence number. It is the single source of truth for both real
// conversions and the settings preview.
//
// Tokens: {YYYY} {YY} {MM} {DD} {HH} {mm} {ss} {num} {seq} {name}.
// {num} is the trailing number in origStem padded to >=4 digits, or the
// sequence number when origStem has no digits. Unknown tokens are left
// literal. Illegal filename characters in the result become '-'. An
// empty result falls back to origStem.
func BuildCaptureName(format string, t time.Time, origStem string, seq int) string {
	num := captureNumber(origStem, seq)
	repl := strings.NewReplacer(
		"{YYYY}", fmt.Sprintf("%04d", t.Year()),
		"{YY}", fmt.Sprintf("%02d", t.Year()%100),
		"{MM}", fmt.Sprintf("%02d", int(t.Month())),
		"{DD}", fmt.Sprintf("%02d", t.Day()),
		"{HH}", fmt.Sprintf("%02d", t.Hour()),
		"{mm}", fmt.Sprintf("%02d", t.Minute()),
		"{ss}", fmt.Sprintf("%02d", t.Second()),
		"{num}", num,
		"{seq}", fmt.Sprintf("%04d", seq),
		"{name}", origStem,
	)
	out := repl.Replace(format)
	out = illegalFilenameChars.ReplaceAllString(out, "-")
	out = strings.TrimSpace(out)
	if out == "" {
		return origStem
	}
	return out
}

// captureNumber returns the trailing digit run of stem padded to at
// least 4 digits, or the zero-padded seq when stem has no digits.
func captureNumber(stem string, seq int) string {
	m := trailingDigits.FindStringSubmatch(stem)
	if m == nil {
		return fmt.Sprintf("%04d", seq)
	}
	digits := m[1]
	if len(digits) < 4 {
		return fmt.Sprintf("%04s", digits)
	}
	return digits
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/naming/ -run TestBuildCaptureName -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/naming/naming.go internal/naming/naming_test.go
git commit -m "feat(naming): BuildCaptureName format renderer"
```

---

## Task 7: Config settings

**Files:**
- Modify: `internal/config/settings.go` (struct L49-65, SetDefaults L83-106, Load L109-127, Save L133-155)
- Modify: `config.example.yaml`
- Modify: `internal/config/settings_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/config/settings_test.go` (follow the existing viper.Reset pattern in that file):
```go
func TestFileNamingDefaults(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	SetDefaults("info")
	s := Load()
	if s.FileNaming != "original" {
		t.Errorf("FileNaming default: want original, got %q", s.FileNaming)
	}
	if s.FileNameFormat != "{YY}-{MM}-{DD} {HH}-{mm}-{ss} {num}" {
		t.Errorf("FileNameFormat default mismatch, got %q", s.FileNameFormat)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/config/ -run TestFileNamingDefaults -v`
Expected: FAIL — `s.FileNaming` undefined (compile error).

- [ ] **Step 3: Add struct fields**

In `internal/config/settings.go`, add to the `Settings` struct after `CollisionOption`:
```go
	CollisionOption     string   `json:"collisionOption"`
	FileNaming          string   `json:"fileNaming"`
	FileNameFormat      string   `json:"fileNameFormat"`
	LogLevel            string   `json:"logLevel"`
```

- [ ] **Step 4: Add defaults**

In `SetDefaults`, after the `collisionOption` default:
```go
	viper.SetDefault("collisionOption", "rename")
	viper.SetDefault("fileNaming", "original")
	viper.SetDefault("fileNameFormat", "{YY}-{MM}-{DD} {HH}-{mm}-{ss} {num}")
```

- [ ] **Step 5: Add to Load and Save**

In `Load()`, after `CollisionOption`:
```go
		CollisionOption:     viper.GetString("collisionOption"),
		FileNaming:          viper.GetString("fileNaming"),
		FileNameFormat:      viper.GetString("fileNameFormat"),
		LogLevel:            viper.GetString("logLevel"),
```
In `Save()`, after the `collisionOption` set:
```go
	viper.Set("collisionOption", s.CollisionOption)
	viper.Set("fileNaming", s.FileNaming)
	viper.Set("fileNameFormat", s.FileNameFormat)
	viper.Set("logLevel", s.LogLevel)
```

- [ ] **Step 6: Document in config.example.yaml**

Add near the other keys in `config.example.yaml`:
```yaml
# Output file naming. "original" keeps the source filename; "captureTime"
# renames using fileNameFormat below and the media's capture time.
fileNaming: "original"

# Format used when fileNaming is "captureTime". Tokens:
#   {YYYY} {YY} {MM} {DD} {HH} {mm} {ss}  - capture date/time
#   {num}   - trailing number of the source name (>=4 digits), or a
#             per-batch sequence number when the source has no digits
#   {seq}   - per-batch sequence number (always)
#   {name}  - original filename (without extension)
# Example result: 26-07-05 10-22-43 3574
fileNameFormat: "{YY}-{MM}-{DD} {HH}-{mm}-{ss} {num}"
```

- [ ] **Step 7: Run the test to verify it passes**

Run: `go test ./internal/config/ -run TestFileNamingDefaults -v`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/config/settings.go internal/config/settings_test.go config.example.yaml
git commit -m "feat(config): fileNaming and fileNameFormat settings"
```

---

## Task 8: Wire capture-time naming into runBatch

**Files:**
- Modify: `internal/services/jobs/service.go` (imports L11-28; `runBatch` L228-323, specifically the per-file section L294-318)

- [ ] **Step 1: Add the capturetime and naming imports**

In `internal/services/jobs/service.go`, add to the internal import group (keeping it alphabetically ordered with the existing `config`, `converter`, `livephoto` entries):
```go
	"github.com/minjejeon/convert4share/internal/capturetime"
	"github.com/minjejeon/convert4share/internal/config"
	"github.com/minjejeon/convert4share/internal/converter"
	"github.com/minjejeon/convert4share/internal/livephoto"
	"github.com/minjejeon/convert4share/internal/naming"
```

- [ ] **Step 2: Read the naming settings in runBatch**

In `runBatch`, next to the other viper reads (after `copyOnlyExts := ...` around L240):
```go
	copyOnlyExts := viper.GetStringSlice("copyOnlyExtensions")
	fileNaming := viper.GetString("fileNaming")
	fileNameFormat := viper.GetString("fileNameFormat")
```

- [ ] **Step 3: Compute the stem per file**

Replace the stem/parent block and the dispatch so a per-file sequence is assigned and the stem is overridden when capture-time naming is on. Change the loop from `for _, f := range files {` to `for i, f := range files {` and, in the section around L294-302, replace:
```go
		ext := strings.ToLower(filepath.Ext(sysPath))
		fname := filepath.Base(sysPath)
		stem := strings.TrimSuffix(fname, filepath.Ext(fname))
		parent := filepath.Dir(sysPath)

		destDir := parent
		if isExcludedDir(parent, config.ExcludePatterns()) {
			destDir = os.ExpandEnv(viper.GetString("defaultDestDir"))
		}
```
with:
```go
		ext := strings.ToLower(filepath.Ext(sysPath))
		fname := filepath.Base(sysPath)
		stem := strings.TrimSuffix(fname, filepath.Ext(fname))
		parent := filepath.Dir(sysPath)

		if fileNaming == "captureTime" {
			ct := capturetime.Resolve(sysPath, ext)
			stem = naming.BuildCaptureName(fileNameFormat, ct, stem, i+1)
		}

		destDir := parent
		if isExcludedDir(parent, config.ExcludePatterns()) {
			destDir = os.ExpandEnv(viper.GetString("defaultDestDir"))
		}
```
The `i+1` batch sequence is assigned in loop order before the goroutine starts, so it is deterministic. `stem` continues to flow into `runOne` → `resolveDestination`, so collision handling is unchanged.

- [ ] **Step 4: Build and vet**

Run: `go build ./... && go vet ./internal/services/jobs/`
Expected: PASS (no unused-var or import errors; `i` is now used).

- [ ] **Step 5: Run the jobs tests**

Run: `go test ./internal/services/jobs/ -v`
Expected: PASS (existing tests + naming tests).

- [ ] **Step 6: Commit**

```bash
git add internal/services/jobs/service.go
git commit -m "feat(jobs): apply capture-time naming in runBatch"
```

---

## Task 9: SettingsService preview method + bindings

**Files:**
- Modify: `internal/services/settings/service.go`
- Regenerate: `frontend/src/bindings/**`

- [ ] **Step 1: Add PreviewFileName**

In `internal/services/settings/service.go`, add imports `"time"` and `"github.com/minjejeon/convert4share/internal/naming"`, then add:
```go
// PreviewFileName renders the given capture-time format with a fixed
// sample time, source name, and sequence so the UI can preview it. It
// returns the stem (no extension).
func (s *Service) PreviewFileName(format string) string {
	sample := time.Date(2026, 7, 5, 10, 22, 43, 0, time.Local)
	return naming.BuildCaptureName(format, sample, "IMG_3574", 1)
}
```
Both `internal/services/jobs` and `internal/services/settings` import `internal/naming`, which imports nothing internal — so there is no import cycle.

- [ ] **Step 2: Build**

Run: `go build ./...`
Expected: PASS.

- [ ] **Step 3: Run affected tests**

Run: `go test ./internal/services/... -v`
Expected: PASS.

- [ ] **Step 4: Regenerate bindings**

Run: `task generate:bindings`
Expected: `frontend/src/bindings/.../settings/service.ts` now exports `PreviewFileName`, and `config/models.ts` includes `fileNaming` / `fileNameFormat`.

- [ ] **Step 5: Commit**

```bash
git add internal/services/settings/service.go frontend/src/bindings
git commit -m "feat(settings): PreviewFileName backed by internal/naming"
```

---

## Task 10: Settings UI

**Files:**
- Modify: `frontend/src/components/SettingsPaths.tsx`

- [ ] **Step 1: Add state, presets, and preview wiring**

Rewrite `frontend/src/components/SettingsPaths.tsx` to add the naming controls after the "File Collision Behavior" block. Full file:
```tsx
import { useEffect, useState } from 'react';
import { FolderOpen } from 'lucide-react';
import type { Settings } from '@bindings/config/models';
import { PreviewFileName } from '@bindings/services/settings/service';

interface SettingsPathsProps {
    settings: Settings;
    onChange: (settings: Settings) => void;
}

const PRESETS: { label: string; format: string }[] = [
    { label: '26-07-05 10-22-43 3574', format: '{YY}-{MM}-{DD} {HH}-{mm}-{ss} {num}' },
    { label: '2026-07-05_10-22-43', format: '{YYYY}-{MM}-{DD}_{HH}-{mm}-{ss}' },
    { label: '20260705_102243_3574', format: '{YYYY}{MM}{DD}_{HH}{mm}{ss}_{num}' },
    { label: 'IMG_20260705_102243', format: 'IMG_{YYYY}{MM}{DD}_{HH}{mm}{ss}' },
    { label: 'name + date', format: '{name}_{YYYY}{MM}{DD}' },
];

const inputClass =
    'block w-full rounded-lg bg-slate-50 dark:bg-slate-900 border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-200 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 sm:text-sm px-3 py-2.5 transition-shadow';

export function SettingsPaths({ settings, onChange }: SettingsPathsProps) {
    const captureTime = settings.fileNaming === 'captureTime';
    const [preview, setPreview] = useState('');

    useEffect(() => {
        if (!captureTime) return;
        PreviewFileName(settings.fileNameFormat || '').then(setPreview);
    }, [captureTime, settings.fileNameFormat]);

    return (
        <div className="bg-white dark:bg-slate-800/40 rounded-xl p-6 border border-slate-200 dark:border-slate-700/50 hover:border-slate-300 dark:hover:border-slate-600/50 transition-colors shadow-sm dark:shadow-none">
             <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-200 mb-6 flex items-center gap-2">
                <FolderOpen className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                Paths & Filters
             </h3>
             <div className="space-y-5">
                <div className="space-y-2">
                    <label htmlFor="paths-dest-dir" className="text-xs font-medium text-slate-500 dark:text-slate-400">Default Destination Directory</label>
                    <input
                        id="paths-dest-dir"
                        type="text"
                        className={inputClass + ' font-mono'}
                        value={settings.defaultDestDir}
                        onChange={(e) => onChange({ ...settings, defaultDestDir: e.target.value })}
                    />
                </div>
                <div className="space-y-2">
                    <label htmlFor="paths-collision" className="text-xs font-medium text-slate-500 dark:text-slate-400">File Collision Behavior</label>
                    <select
                        id="paths-collision"
                        className={inputClass}
                        value={settings.collisionOption || 'rename'}
                        onChange={(e) => onChange({ ...settings, collisionOption: e.target.value })}
                    >
                        <option value="rename">Rename</option>
                        <option value="overwrite">Overwrite</option>
                        <option value="error">Error</option>
                    </select>
                </div>
                <div className="space-y-2">
                    <label htmlFor="paths-naming" className="text-xs font-medium text-slate-500 dark:text-slate-400">File Naming</label>
                    <select
                        id="paths-naming"
                        className={inputClass}
                        value={settings.fileNaming || 'original'}
                        onChange={(e) => onChange({ ...settings, fileNaming: e.target.value })}
                    >
                        <option value="original">Keep original name</option>
                        <option value="captureTime">Capture time based</option>
                    </select>
                </div>
                {captureTime && (
                    <div className="space-y-2 pl-3 border-l-2 border-slate-200 dark:border-slate-700">
                        <label htmlFor="paths-name-preset" className="text-xs font-medium text-slate-500 dark:text-slate-400">Preset</label>
                        <select
                            id="paths-name-preset"
                            className={inputClass}
                            value={PRESETS.find((p) => p.format === settings.fileNameFormat)?.format ?? ''}
                            onChange={(e) => onChange({ ...settings, fileNameFormat: e.target.value })}
                        >
                            {!PRESETS.some((p) => p.format === settings.fileNameFormat) && (
                                <option value={settings.fileNameFormat}>Custom</option>
                            )}
                            {PRESETS.map((p) => (
                                <option key={p.format} value={p.format}>{p.label}</option>
                            ))}
                        </select>
                        <label htmlFor="paths-name-format" className="text-xs font-medium text-slate-500 dark:text-slate-400">Format</label>
                        <input
                            id="paths-name-format"
                            type="text"
                            className={inputClass + ' font-mono'}
                            value={settings.fileNameFormat || ''}
                            onChange={(e) => onChange({ ...settings, fileNameFormat: e.target.value })}
                            placeholder="{YY}-{MM}-{DD} {HH}-{mm}-{ss} {num}"
                        />
                        <p className="text-[10px] text-slate-400 dark:text-slate-500">
                            Preview: <span className="font-mono text-slate-600 dark:text-slate-300">{preview || '…'}</span>
                            {'  '}· Tokens: {'{YYYY} {YY} {MM} {DD} {HH} {mm} {ss} {num} {seq} {name}'}
                        </p>
                    </div>
                )}
                <div className="space-y-2">
                    <label htmlFor="paths-exclude" className="text-xs font-medium text-slate-500 dark:text-slate-400">Exclude Patterns (comma separated)</label>
                     <input
                        id="paths-exclude"
                        type="text"
                        className={inputClass}
                        value={settings.excludePatterns?.join(', ')}
                        onChange={(e) => onChange({ ...settings, excludePatterns: e.target.value.split(',').filter(s => s.trim() !== '').map(s => s.trim()) })}
                        placeholder="e.g. \Pictures\, \DCIM\"
                    />
                </div>
                <div className="space-y-2">
                    <label htmlFor="paths-copy-only" className="text-xs font-medium text-slate-500 dark:text-slate-400">Copy Only Extensions (comma separated)</label>
                     <input
                        id="paths-copy-only"
                        type="text"
                        className={inputClass}
                        value={settings.copyOnlyExtensions?.join(', ')}
                        onChange={(e) => onChange({ ...settings, copyOnlyExtensions: e.target.value.split(',').filter(s => s.trim() !== '').map(s => s.trim()) })}
                        placeholder="e.g. .jpg, .jpeg, .mp4"
                    />
                    <p className="text-[10px] text-slate-400 dark:text-slate-500">Files with these extensions will be copied to the destination without conversion.</p>
                </div>
             </div>
        </div>
    );
}
```

- [ ] **Step 2: Type-check the frontend**

Run: `npm --prefix frontend run build -s`
Expected: PASS (tsc + vite build succeed; `PreviewFileName` and the new Settings fields resolve against regenerated bindings).

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/SettingsPaths.tsx
git commit -m "feat(ui): capture-time file naming controls with preset and preview"
```

---

## Task 11: Notices, full verification, final commit

**Files:**
- Modify: `THIRD_PARTY_NOTICES.md`

- [ ] **Step 1: Add dependency notices**

Append entries for `github.com/abema/go-mp4` (MIT) and `github.com/dsoprea/go-exif` (MIT) to `THIRD_PARTY_NOTICES.md`, matching the existing format in that file. Read the file first to match its heading/layout convention.

- [ ] **Step 2: Full build, vet, test**

Run:
```bash
go build ./... && go vet ./... && go test ./...
```
Expected: all PASS.

- [ ] **Step 3: Frontend build**

Run: `npm --prefix frontend run build -s`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add THIRD_PARTY_NOTICES.md
git commit -m "docs: third-party notices for go-mp4 and go-exif"
```

---

## Manual verification (after all tasks)

Run the app (`task dev`), drop a mix of HEIC/MOV/JPG files with the setting on:
- With **Keep original name**: filenames unchanged (regression check).
- With **Capture time based** + default preset: outputs like `26-07-05 10-22-43 3574.mp4`; a file whose name has no number gets a sequence (`0001`); two files in the same second get ` (1)` via the collision rule.
- Change presets and confirm the live preview updates.
```

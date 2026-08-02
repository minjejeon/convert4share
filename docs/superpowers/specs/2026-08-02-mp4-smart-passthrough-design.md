# MP4 Smart Passthrough — Design

**Date:** 2026-08-02
**Status:** Approved

## Problem

`.mp4` ships in the `copyOnlyExtensions` default, so every mp4 is copied
verbatim. The extension says nothing about the codec inside, and a real
file exposed the gap:

```
Y:\2026\2026-07\_talkv_dJMcbsRURyp_pWgWENzlsNYXrHBKZMH6Q1_talkv_high.mp4
  container: mp4 (isom/iso2/mp41)
  video:     hevc (H.265), hvc1, Main, 1080x1920, 3.07 Mbps, 30fps
  audio:     aac (LC), 128 kbps, 44.1 kHz stereo
  67 s / 26.8 MB
```

H.265 in an mp4 wrapper does not play on KakaoTalk, older devices, or
several web players. Convert4Share exists to produce widely compatible
output, so this file should have been re-encoded to H.264 and was not.

Resolution (1080x1920) is within `maxSize` and the 3 Mbps bitrate is
below the `high` preset's 5 Mbps, so the codec is the only reason this
particular file needs conversion.

## Decisions

Settled during brainstorming:

- **Criteria:** codec *and* resolution. Bitrate is deliberately excluded
  — re-encoding a 3 Mbps source at a 5 Mbps target inflates the file and
  makes the verdict unstable.
- **`copyOnlyExtensions` wins absolutely.** An extension listed there is
  copied without probing. Only the default list changes.
- **No config migration.** Users whose `config.yaml` already contains
  `.mp4` must remove it themselves; the release notes say so.
- **Probe with ffprobe**, exposed as a first-class `ffprobeBinary`
  setting alongside ffmpeg and magick.
- **Probe failure falls back to converting**, prioritising compatibility
  over avoiding a needless re-encode.

## Compatibility rule

`internal/converter/compat.go`, a pure function over a probe result:

```go
func NeedsConversion(info MediaInfo, maxSize int) (bool, string)
```

Copy only when every condition holds:

| Aspect | Condition |
|---|---|
| Video codec | `h264` |
| Video pixel format | `yuv420p` |
| Resolution | `max(width, height) <= maxSize` |
| Audio codec | absent, or `aac` / `mp3` |

Notes:

- `pix_fmt` is checked because H.264 High 10 (`yuv420p10le`) and 4:2:2
  variants fail on the same players HEVC does. Codec alone misses them.
- Use `width`/`height` (display), not `coded_width`/`coded_height` —
  the sample file reports 1088 coded vs 1080 display.
- Streams with `disposition.attached_pic = 1` are cover art, not video.
- No video stream at all (audio-only mp4) copies: the scale filter has
  nothing to work on.
- The second return value is a human-readable reason such as
  `video codec hevc (want h264)`, used in logs and the job status line.

## ffprobe integration

- `config.Settings` gains `FfprobeBinary string \`json:"ffprobeBinary"\``
  with viper default `"ffprobe"`.
- `tools.DetectBinaries` grows an `ffprobe` key, discovered the same way
  as ffmpeg and magick: PATH, then WinGet `Links`, then a walk of WinGet
  `Packages`.
- Resolution order at call time: configured value if it contains a path
  separator → `ffprobe(.exe)` next to `ffmpegBinary` → PATH lookup.
  ffprobe ships with ffmpeg, so rule two covers nearly every install.
- `internal/converter/probe.go` runs
  `-v error -show_streams -show_format -print_format json` and unmarshals
  only the fields the rule needs.
- Settings UI gets an ffprobe row in `SettingsTools.tsx` matching the
  existing ffmpeg/magick rows (text input + browse). Auto-Detect fills
  it in.

Probe failure converts, per the decision above. Because a missing
ffprobe would then re-encode *every* mp4, Auto-Detect surfaces ffprobe
alongside the other tools so the gap is visible.

## Job routing

`internal/services/jobs/service.go` gains an `.mp4` case, ordered after
`isCopyOnly` and before `.mov`:

```
case isCopyOnly:        // untouched, never probes
case in.ext == ".mp4":  // probe, then copy or convert
case in.ext == ".mov":
case in.ext == ".heic":
```

Compatible files take the existing copy path and report
`Copied (already compatible)`. Incompatible files acquire `ffmpegSem`
and run the existing `Ffmpeg()`.

### In-place safety

`resolveDestination` returns `dest == src` when `collisionOption` is
`overwrite` and the output lands in the source directory
(`service.go:515-517`). mov→mp4 changed the extension and dodged this;
mp4→mp4 does not.

- **Convert path:** always encode to `<dest>.c4s-tmp.mp4` in the
  destination directory, then `os.Rename` into place. Remove the temp
  file on failure or cancellation. A same-volume rename costs nothing
  and structurally prevents handing ffmpeg one path as both input and
  output.
- **Copy path:** if `src == dest`, succeed without touching the file.
  `copyFile`'s `os.Create` would otherwise truncate the original.

The identical `copyFile` hazard on other extensions is pre-existing and
out of scope.

## Audio passthrough

`BuildFfmpegArgs` hardcodes `-c:a aac`. It gains a parameter so the mp4
path can emit `-c:a copy` when the source audio is already `aac`.
Re-encoding 128 kbps AAC to AAC is pure generation loss; the sample file
benefits directly.

## Settings default

`copyOnlyExtensions`: `[".jpg", ".jpeg", ".mp4"]` → `[".jpg", ".jpeg"]`.

## Testing

- `compat_test.go` — the rule table as a table-driven test, no ffprobe
  needed.
- `probe_test.go` — small H.264 and HEVC fixtures under `testdata`,
  generated with ffmpeg; `t.Skip` when ffprobe is absent.
- `service_test.go` — `src == dest` leaves the original intact on both
  the copy and convert paths.

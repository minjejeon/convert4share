# Capture-time filename option — design

## Goal

Add an option to rename converted output files based on the media's
capture time plus a number, e.g. `26-07-05 10-22-43 3574.mp4`. The
filename format is user-configurable via a format string, with a set of
built-in presets. Capture time is resolved best-effort from available
metadata, falling back to filesystem timestamps.

## User-facing behavior

### Naming mode

A new setting `fileNaming` with two values:

- `original` (default) — current behavior; output keeps the source stem.
- `captureTime` — output stem is built from the capture time and the
  configured format.

Extension conversion is unaffected (`.mov`→`.mp4`, `.heic`→`.jpg`,
copy-only keeps its extension). The mode applies to **all** outputs:
video, image, and copy-only files.

### Format string and tokens

Setting `fileNameFormat` holds a template string used when
`fileNaming == captureTime`. Tokens are brace-delimited; any other
characters are literals.

| Token     | Meaning                                                        |
|-----------|----------------------------------------------------------------|
| `{YYYY}`  | 4-digit year                                                   |
| `{YY}`    | 2-digit year                                                   |
| `{MM}`    | 2-digit month                                                  |
| `{DD}`    | 2-digit day                                                    |
| `{HH}`    | 2-digit hour (24h)                                             |
| `{mm}`    | 2-digit minute                                                 |
| `{ss}`    | 2-digit second                                                 |
| `{num}`   | Trailing number from the source filename, min 4 digits (zero-padded); if the source has no number, the batch sequence number. |
| `{seq}`   | Batch sequence number (always), zero-padded to 4 digits.       |
| `{name}`  | Original source stem (for preserving the original name).       |

`{num}` rules:
- Take the last run of digits in the source stem (`IMG_3574` → `3574`,
  `PXL_20230705_102243123` → `102243123`).
- If it has fewer than 4 digits, left-pad with zeros (`12` → `0012`).
- If it has more than 4 digits, keep all digits (no truncation — avoids
  information loss and collisions).
- If the source stem has no digits at all, substitute the batch sequence
  number (zero-padded to 4). This preserves the earlier decision that a
  missing number falls back to a sequence.

An unknown token (e.g. `{foo}`) is treated as a literal and left as-is.

### Presets

The Settings UI offers a preset dropdown; selecting one fills the format
field, which remains freely editable afterward.

1. `{YY}-{MM}-{DD} {HH}-{mm}-{ss} {num}` → `26-07-05 10-22-43 3574` **(default)**
2. `{YYYY}-{MM}-{DD}_{HH}-{mm}-{ss}` → `2026-07-05_10-22-43`
3. `{YYYY}{MM}{DD}_{HH}{mm}{ss}_{num}` → `20260705_102243_3574`
4. `IMG_{YYYY}{MM}{DD}_{HH}{mm}{ss}` → `IMG_20260705_102243`
5. `{name}_{YYYY}{MM}{DD}` → original name + date

### Filename safety

After token substitution, characters illegal in Windows filenames
(`\ / : * ? " < > |`) are replaced with `-`. If the result is empty
after trimming, fall back to the original stem so a file is never left
without a name.

## Capture-time resolution (best effort)

`internal/capturetime.Resolve(path, ext string) time.Time` returns the
best available capture time, trying in order:

1. **Metadata**
   - Video (`.mov`, `.mp4`): MP4/MOV `moov/mvhd` creation time via
     `github.com/abema/go-mp4`.
   - Image (`.heic`, `.jpg`, `.jpeg`): EXIF `DateTimeOriginal` via
     `github.com/dsoprea/go-exif/v3`, extracting the EXIF blob with the
     HEIC / JPEG structure extractors (see Dependencies).
2. **File creation time (birthtime)** — available on Windows via
   `os.Stat` → `syscall.Win32FileAttributeData.CreationTime`.
3. **File modification time (mtime)** — `os.Stat` `ModTime()`.

Resolve always returns a usable time (step 3 cannot fail for an existing
file), so a capture-time name is always produced (best effort).

### Time zone

All sources are treated as **wall-clock local time**, formatted directly
with no time-zone conversion:
- EXIF `DateTimeOriginal` has no zone; it is local wall-clock by
  convention.
- `mvhd` creation time is read as its stored wall-clock value (Apple
  devices in practice write local time here); no UTC→local shift.
- Filesystem times are formatted in the machine's local zone.

This is a deliberate best-effort simplification; sub-second/precise-zone
correctness is out of scope.

## Architecture

### New package: `internal/capturetime`

- `func Resolve(path, ext string) time.Time` — the fallback chain above.
- Internal helpers `videoCreationTime(path)`, `imageCaptureTime(path)`
  returning `(time.Time, bool)`.
- Unit-tested with small sample fixtures (a tiny MP4 with a known mvhd
  time, a JPEG with EXIF, a HEIC with EXIF) plus fallback paths.

### Filename builder: `internal/services/jobs`

- `func buildCaptureName(format string, t time.Time, origStem string, seq int) string`
  — pure function: tokenize `format`, substitute, sanitize. The single
  source of truth for both real conversions and the UI preview.
- Pure and side-effect free → directly unit-tested for every token,
  padding, overflow, missing-number→seq, unknown token, and sanitize
  cases.

### Wiring in `runBatch` / `runOne`

- Read `fileNaming` and `fileNameFormat` from viper in `runBatch`.
- When `fileNaming == captureTime`, assign the batch sequence number in
  the dispatch loop (loop order, before goroutines, so it is
  deterministic), resolve the capture time, compute the new stem with
  `buildCaptureName`, and pass it as the `stem` to `runOne`. Otherwise
  pass the original stem as today.
- Collisions continue to be handled by the existing
  `resolveDestination` + `collisionOption` (rename/overwrite/error). Two
  outputs that compute the same name get ` (1)` etc. as today.
- Live Photo pairing (`heicStems`) is computed from original stems before
  any rename, so it is unaffected.

### Settings service — preview

- Expose `PreviewFileName(format string) string` on the **SettingsService**
  (it owns settings and is already bound to the frontend), calling
  `buildCaptureName` with a fixed sample time
  (`2026-07-05 10:22:43`), sample original stem (`IMG_3574`), and sample
  seq (`1`). The frontend calls this for the live preview so the format
  semantics have a single backend source of truth.

## Settings & config

- `config.Settings` gains `FileNaming string` and `FileNameFormat
  string`; viper keys `fileNaming`, `fileNameFormat`. Defaults:
  `fileNaming: "original"`, `fileNameFormat: "{YY}-{MM}-{DD} {HH}-{mm}-{ss} {num}"`.
- Document both keys in `config.example.yaml`.
- Load/save in `settings.go` alongside the existing fields.

## Frontend (Settings UI)

- Add a "File naming" dropdown: *Keep original name* / *Capture time
  based*.
- When *Capture time based* is selected, show:
  - a preset dropdown (fills the format field on select),
  - a format text input,
  - a live preview line rendered by calling `PreviewFileName`.
- Regenerate Wails bindings (`task generate:bindings`) after the backend
  method/struct changes.

## Dependencies (new)

- `github.com/abema/go-mp4` — MP4/MOV box parsing (mvhd creation time).
- `github.com/dsoprea/go-exif/v3` — EXIF tag parsing.
- `github.com/dsoprea/go-heic-exif-extractor/v2` — pull the EXIF blob out
  of a HEIC container.
- `github.com/dsoprea/go-jpeg-image-structure/v2` — pull the EXIF blob out
  of a JPEG.

All are permissively licensed; add to `THIRD_PARTY_NOTICES.md`.

## Testing

- `internal/capturetime`: metadata extraction for MP4/HEIC/JPEG fixtures;
  birthtime and mtime fallbacks (fixture with stripped metadata).
- `buildCaptureName`: every token; `{num}` padding (`12`→`0012`),
  overflow (`12345`→`12345`), missing number → seq; `{seq}`, `{name}`,
  unknown token as literal; illegal-character sanitize; empty-result
  fallback to original stem.
- Existing jobs tests remain green; `go build ./...`, `go vet`, full
  `go test ./...`.

## Out of scope

- Precise time-zone / sub-second correctness.
- Reading capture time from formats other than MOV/MP4/HEIC/JPEG.
- Renaming files that are not being converted/copied by the app.

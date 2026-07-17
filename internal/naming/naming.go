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

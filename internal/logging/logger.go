// Package logging centralises Convert4Share's slog + lumberjack setup.
// The logger writes to a rotating log file in a per-user writable
// location (see logFilePath) and also mirrors output to stderr when
// stderr is actually usable (i.e. not when running under windowsgui
// where stderr is detached).
package logging

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// New constructs the application logger. The provided level var lets
// callers adjust the log level at runtime via LevelVar.Set. The
// returned logger is also installed as slog's default and as the
// destination of the standard library log package, matching the
// pre-refactor behavior.
func New(level *slog.LevelVar) *slog.Logger {
	logPath, err := logFilePath()
	if err != nil {
		// Fallback to stderr if available, otherwise discard so we
		// never panic writing logs when running under windowsgui with
		// stderr detached.
		var w io.Writer = os.Stderr
		if runtime.GOOS == "windows" {
			if _, statErr := os.Stderr.Stat(); statErr != nil {
				w = io.Discard
			}
		}
		logger := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))
		slog.SetDefault(logger)
		logger.Error("Could not resolve log directory", "error", err)
		return logger
	}

	rotator := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   // days
		Compress:   true, // disabled by default
	}

	// In windowsgui mode, stderr is unavailable. Log only to file in that case.
	multi := io.Writer(rotator)
	if runtime.GOOS != "windows" {
		multi = io.MultiWriter(os.Stderr, rotator)
	} else if _, err := os.Stderr.Stat(); err == nil {
		multi = io.MultiWriter(os.Stderr, rotator)
	}

	handler := slog.NewTextHandler(multi, &slog.HandlerOptions{Level: level})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	log.SetOutput(rotator) // Redirect standard log to lumberjack

	return logger
}

// logFilePath returns the rotating log file path in a per-user writable
// location (%LOCALAPPDATA%\Convert4Share\logs on Windows), creating the
// directory if missing. This is independent of the install directory:
// for an all-users install the executable lives under Program Files,
// where standard users cannot write, so logging next to the binary
// would silently fail.
func logFilePath() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "Convert4Share", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "convert4share.log"), nil
}

// ParseLevel converts a textual level (debug/info/warn/error) into an
// slog.Level. Unknown values fall back to LevelInfo.
func ParseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Package settings hosts the Wails v3 service that exposes user
// settings (GetSettings / SaveSettings) backed by viper and the
// helpers in internal/config.
//
// On successful Save, the service emits a `settings-changed` event so
// other services (notably JobsService for semaphore resizing) can
// react without a direct dependency.
package settings

import (
	"context"
	"log/slog"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/minjejeon/convert4share/internal/config"
	"github.com/minjejeon/convert4share/internal/logging"
	"github.com/minjejeon/convert4share/internal/naming"
)

// LevelSetter is the minimal interface the SettingsService needs to
// adjust the running logger level when the user changes LogLevel.
// *slog.LevelVar satisfies it.
type LevelSetter interface {
	Set(slog.Level)
}

// Service is the v3 service binding for application settings.
type Service struct {
	logger *slog.Logger
	level  LevelSetter
	app    *application.App
}

// New constructs a SettingsService. `level` may be nil if the caller
// does not wish to track log-level changes (e.g. tests).
func New(logger *slog.Logger, level LevelSetter) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{logger: logger, level: level}
}

// ServiceStartup grabs a reference to the running application so
// SaveSettings can emit the `settings-changed` event.
func (s *Service) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.app = application.Get()
	return nil
}

// GetSettings returns the current Settings view from viper.
func (s *Service) GetSettings() config.Settings {
	return config.Load()
}

// SaveSettings persists the given Settings to viper + the config file,
// then notifies the rest of the app. The log-level adjustment is done
// synchronously here so frontend feedback is immediate.
func (s *Service) SaveSettings(settings config.Settings) error {
	if err := config.Save(settings); err != nil {
		s.logger.Error("Failed to save settings", "error", err)
		return err
	}

	if s.level != nil {
		s.level.Set(logging.ParseLevel(settings.LogLevel))
		s.logger.Info("Logger level updated", "level", settings.LogLevel)
	}

	if s.app != nil {
		s.app.Event.Emit("settings-changed", settings)
	}
	return nil
}

// PreviewFileName renders the given capture-time format with a fixed
// sample time, source name, and sequence so the UI can preview it. It
// returns the stem (no extension).
func (s *Service) PreviewFileName(format string) string {
	sample := time.Date(2026, 7, 5, 10, 22, 43, 0, time.Local)
	return naming.BuildCaptureName(format, sample, "IMG_3574", 1)
}

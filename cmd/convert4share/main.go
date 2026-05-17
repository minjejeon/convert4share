// Convert4Share entry point for the Wails v3 migration.
//
// Phase 3 wires the real Jobs / Settings / Tools services into the v3
// application. Viper defaults are registered and config.yaml (next to
// the executable) is read before `app.Run()` so the services see
// populated settings during ServiceStartup. The frontend bindings
// under `frontend/src/wailsjs/` still target Wails v2 and will not
// work at runtime until Phase 5 swaps them for `@wailsio/runtime`.
package main

import (
	"embed"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/minjejeon/convert4share/internal/config"
	"github.com/minjejeon/convert4share/internal/logging"
	"github.com/minjejeon/convert4share/internal/services/jobs"
	"github.com/minjejeon/convert4share/internal/services/settings"
	"github.com/minjejeon/convert4share/internal/services/tools"
)

// Assets are embedded from a sibling `dist/` directory so that
// `go:embed` patterns work relative to this source file. Phase 6 wires
// the build pipeline to copy `frontend/dist` into this location.
//
//go:embed all:dist
var assets embed.FS

func main() {
	levelVar := new(slog.LevelVar)
	levelVar.Set(slog.LevelInfo)
	logger := logging.New(levelVar)

	loadConfig(logger)

	// Pull the configured log level into the live LevelVar so subsequent
	// log calls honour the user's preference. SettingsService updates
	// the same LevelVar on SaveSettings.
	levelVar.Set(logging.ParseLevel(viper.GetString("logLevel")))

	toolsService := tools.New(logger)
	autoDetectBinaries(logger, toolsService)

	jobsService := jobs.New(logger)
	settingsService := settings.New(logger, levelVar)

	app := application.New(application.Options{
		Name:        "Convert4Share",
		Description: "Converts MOV/HEIC to MP4/JPG",
		Services: []application.Service{
			application.NewService(jobsService),
			application.NewService(settingsService),
			application.NewService(toolsService),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(assets),
		},
		Logger:   logger,
		LogLevel: slog.LevelInfo,
	})

	// Window reference is retained for Phase 4 lifecycle hooks
	// (WindowDOMReady, second-instance focus). v3 currently constructs
	// the window lazily during app.Run() so creation here only stages
	// options.
	_ = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:          "Convert4Share",
		URL:            "/",
		Width:          1024,
		Height:         768,
		EnableFileDrop: true,
	})

	if err := app.Run(); err != nil {
		logger.Error("application exited with error", "error", err)
		os.Exit(1)
	}
}

// loadConfig registers viper defaults and reads <exeDir>/config.yaml
// when present. A missing file is logged but not fatal — defaults are
// sufficient for first launch.
func loadConfig(logger *slog.Logger) {
	exePath, err := os.Executable()
	if err != nil {
		logger.Error("Error getting executable path", "error", err)
		return
	}
	exeDir := filepath.Dir(exePath)

	viper.AddConfigPath(exeDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	defaultLogLevel := "info"
	if isDev() {
		defaultLogLevel = "debug"
	}
	config.SetDefaults(defaultLogLevel)

	if err := viper.ReadInConfig(); err != nil {
		logger.Info("Config file not found, using defaults", "error", err)
	}
}

// autoDetectBinaries fills in ffmpegBinary / magickBinary in viper if
// the currently configured paths do not resolve. ToolsService.
// DetectBinaries already inspects PATH and known WinGet locations.
func autoDetectBinaries(logger *slog.Logger, ts *tools.Service) {
	checkBinary := func(key string) bool {
		val := viper.GetString(key)
		if val == "" {
			return false
		}
		if _, err := exec.LookPath(val); err != nil {
			logger.Info("Binary path invalid or not found", "key", key, "path", val, "error", err)
			return false
		}
		return true
	}

	detected := ts.DetectBinaries()

	if !checkBinary("ffmpegBinary") {
		if path, ok := detected["ffmpeg"]; ok {
			logger.Info("Auto-detected ffmpeg binary", "path", path)
			viper.Set("ffmpegBinary", path)
		}
	}

	if !checkBinary("magickBinary") {
		if path, ok := detected["magick"]; ok {
			logger.Info("Auto-detected magick binary", "path", path)
			viper.Set("magickBinary", path)
		}
	}
}

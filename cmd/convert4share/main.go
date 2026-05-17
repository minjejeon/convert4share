// Convert4Share entry point for the Wails v3 migration.
//
// Phase 4 wires the v3 lifecycle: SingleInstance with window
// activation, ApplicationStarted (forwarding CLI file args),
// ApplicationOpenedWithFile (Windows file-association launches),
// and WindowRuntimeReady logging. v2's OnShutdown / OnBeforeClose
// callbacks were no-ops so no v3 hook is registered for them.
package main

import (
	"embed"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

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

	exePath, err := os.Executable()
	if err != nil {
		logger.Warn("Could not resolve executable path", "error", err)
		exePath = ""
	}

	toolsService := tools.New(logger)
	autoDetectBinaries(logger, toolsService)

	jobsService := jobs.New(logger)
	settingsService := settings.New(logger, levelVar)

	// pendingFiles holds CLI/file-association args until the frontend
	// reports it is mounted and subscribed (`frontend-ready` custom
	// event). Mirrors v2's processPendingFiles handshake.
	var (
		pendingMu    sync.Mutex
		pendingFiles []string
	)
	queuePending := func(files []string) {
		if len(files) == 0 {
			return
		}
		pendingMu.Lock()
		pendingFiles = append(pendingFiles, files...)
		pendingMu.Unlock()
	}
	drainPending := func() []string {
		pendingMu.Lock()
		defer pendingMu.Unlock()
		out := pendingFiles
		pendingFiles = nil
		return out
	}

	// window must be declared up-front so the SingleInstance closure
	// can call Restore()/Focus() on a captured reference once it is
	// assigned below. See examples/single-instance/main.go.
	var window *application.WebviewWindow

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
		// Registering extensions here is what allows Windows to deliver
		// ApplicationOpenedWithFile when the OS opens a .mov/.heic with
		// us (one-arg launch). Multi-file launches still arrive as plain
		// CLI args and are handled by the ApplicationStarted hook below.
		FileAssociations: []string{".mov", ".heic"},
		SingleInstance: &application.SingleInstanceOptions{
			// Re-use the v2 UUID so existing v2 installations migrate
			// cleanly to the v3 build without a stale lock blocking
			// startup. Phase 7 (installer) covers the rest of the
			// migration.
			UniqueID: "3310d829-dc96-4613-af1a-a5353d9f07a6",
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				logger.Info("Second instance launched", "args", data.Args, "workingDir", data.WorkingDir)

				if window != nil {
					window.UnMinimise()
					window.Show()
					window.Restore()
					window.Focus()
				}

				// data.Args is verbatim os.Args of the second instance,
				// so element 0 is its executable path. ExtractFileArgs
				// strips that (and any non-file noise) before forwarding.
				// The first instance's frontend is already mounted by
				// the time a second instance can launch, so we emit
				// directly rather than queuing.
				files := jobs.ExtractFileArgs(data.Args, exePath)
				if len(files) > 0 {
					logger.Info("Forwarding files from second instance", "files", files)
					jobsService.EmitFilesReceived(files)
				}
			},
		},
	})

	window = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:          "Convert4Share",
		URL:            "/",
		Width:          1024,
		Height:         768,
		EnableFileDrop: true,
	})

	// ApplicationStarted fires once after the services have started
	// and the main loop is running. This is the v3 successor to v2's
	// App.startup(ctx) — we queue any CLI file arguments the user
	// dropped on the launcher icon and flush them once the frontend
	// reports it is subscribed (`frontend-ready`).
	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		logger.Info("ApplicationStarted")
		args := os.Args
		if len(args) > 1 {
			files := jobs.ExtractFileArgs(args[1:], exePath)
			if len(files) > 0 {
				logger.Info("Queuing CLI file args", "files", files)
				queuePending(files)
			}
		}
	})

	// ApplicationOpenedWithFile fires on Windows when the OS opens a
	// single file with our registered extension. The filename arrives
	// via the event context.
	app.Event.OnApplicationEvent(events.Common.ApplicationOpenedWithFile, func(e *application.ApplicationEvent) {
		filename := e.Context().Filename()
		logger.Info("ApplicationOpenedWithFile", "file", filename)
		if filename == "" {
			return
		}
		files := jobs.ExtractFileArgs([]string{filename}, exePath)
		if len(files) == 0 {
			return
		}
		// If the frontend has already started, emit immediately;
		// otherwise queue until `frontend-ready`. The race is decided
		// by the listener registered below: pendingFiles is drained
		// each time frontend-ready fires, so adding to it here is safe.
		queuePending(files)
		// Best-effort flush — harmless if the frontend has not yet
		// subscribed (the event is dropped). The frontend-ready
		// listener will re-emit if needed.
		if drained := drainPending(); len(drained) > 0 {
			jobsService.EmitFilesReceived(drained)
		}
	})

	// The frontend emits this custom event after mounting its event
	// listeners (v2 handshake, preserved in v3). When it arrives we
	// flush any pending CLI / file-association files so they reach the
	// queue. Frontend remounts (HMR, navigation) re-emit, which is why
	// we drain rather than only flushing once.
	app.Event.On("frontend-ready", func(*application.CustomEvent) {
		logger.Info("Frontend reported ready")
		if files := drainPending(); len(files) > 0 {
			logger.Info("Flushing pending files to frontend", "files", files)
			jobsService.EmitFilesReceived(files)
		}
	})

	// WindowRuntimeReady is v3's analogue of v2's OnDomReady. It fires
	// after the webview signals that the JS runtime has finished its
	// handshake; the frontend-ready custom event is what we actually
	// use for the flush handshake but logging the runtime-ready edge
	// is useful for diagnosing startup hangs.
	if window != nil {
		window.OnWindowEvent(events.Common.WindowRuntimeReady, func(*application.WindowEvent) {
			logger.Info("Window runtime ready")
		})
	}

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

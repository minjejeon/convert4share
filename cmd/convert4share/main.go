// Convert4Share entry point for the Wails v3 migration.
//
// Phase 2.1 establishes the minimal v3 skeleton: a single window
// served from the embedded `frontend/dist`, with empty Jobs, Settings,
// and Tools services registered against the application. Lifecycle
// events, second-instance handling, dialogs, and file associations
// arrive in later phases; the frontend bindings under
// `frontend/src/wailsjs/` still target Wails v2 and will not work at
// runtime until Phase 5 swaps them for `@wailsio/runtime`.
package main

import (
	"embed"
	"log/slog"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"

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

	app := application.New(application.Options{
		Name:        "Convert4Share",
		Description: "Converts MOV/HEIC to MP4/JPG",
		Services: []application.Service{
			application.NewService(jobs.New()),
			application.NewService(settings.New()),
			application.NewService(tools.New()),
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

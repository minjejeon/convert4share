package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/minjejeon/convert4share/converter"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx          context.Context
	cfg          *converter.Config
	pendingFiles []string
	mu           sync.Mutex
	destMu       sync.Mutex
	isReady      bool
	processTimer *time.Timer
	jobCancels   map[string]context.CancelFunc
	isPaused     bool
	pauseCond    *sync.Cond
	ffmpegSem    chan struct{}
	magickSem    chan struct{}
}

func NewApp() *App {
	app := &App{
		jobCancels: make(map[string]context.CancelFunc),
	}
	app.pauseCond = sync.NewCond(&app.mu)
	return app
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.initConfig()
	runtime.EventsOn(ctx, "frontend-ready", func(optionalData ...interface{}) {
		logger.Info("Frontend reported ready")
		a.processPendingFiles()
	})
}

func (a *App) domReady(ctx context.Context) {
	a.mu.Lock()
	a.isReady = true
	a.mu.Unlock()

	logger.Info("DOM is ready.")
}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

func (a *App) shutdown(ctx context.Context) {
}

func (a *App) OnSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	logger.Info("Second instance launched", "args", secondInstanceData.Args)

	exePath, err := os.Executable()
	if err != nil {
		logger.Error("Error getting executable path during second instance launch", "error", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	logger.Info("Processing second instance args", "count", len(secondInstanceData.Args))

	if len(secondInstanceData.Args) > 0 {
		files := secondInstanceData.Args
		var actualFiles []string
		for _, arg := range files {
			if exePath != "" {
				if absArg, err := filepath.Abs(arg); err == nil && strings.EqualFold(absArg, exePath) {
					logger.Info("Skipping executable path in args", "arg", arg)
					continue
				}
			}

			if info, err := os.Stat(arg); err == nil && !info.IsDir() && info.Size() > 0 {
				if absArg, err := filepath.Abs(arg); err == nil {
					actualFiles = append(actualFiles, absArg)
				} else {
					actualFiles = append(actualFiles, arg)
				}
			} else {
				logger.Info("Skipping invalid file in second instance args", "arg", arg, "err", err)
			}
		}
		if len(actualFiles) > 0 {
			logger.Info("Adding files from second instance", "files", actualFiles)
			a.pendingFiles = append(a.pendingFiles, actualFiles...)
		} else {
			logger.Info("No valid files found in second instance args.")
		}
	} else {
		logger.Info("Second instance launched with no args.")
	}

	// Debounce the processing to handle rapid-fire calls (e.g. from multi-file selection)
	if a.processTimer != nil {
		a.processTimer.Stop()
	}
	a.processTimer = time.AfterFunc(500*time.Millisecond, func() {
		a.handleSecondInstance()
	})
}

func (a *App) handleSecondInstance() {
	// If the app is still starting up (ctx is nil), the window will appear naturally.
	// We only need to force focus if the app is already running.
	if a.ctx != nil {
		runtime.WindowUnminimise(a.ctx)
		runtime.WindowShow(a.ctx)
		runtime.WindowSetAlwaysOnTop(a.ctx, true)
		runtime.WindowSetAlwaysOnTop(a.ctx, false)
	}

	a.processPendingFiles()
}

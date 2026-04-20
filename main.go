package main

import (
	"embed"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/minjejeon/convert4share/cmd"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"gopkg.in/natefinch/lumberjack.v2"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed config.example.yaml
var configTemplate []byte

var logger *slog.Logger
var levelVar = &slog.LevelVar{}

func initLogger() {
	exePath, err := os.Executable()
	if err != nil {
		// Fallback to stderr if available, otherwise discard
		var w io.Writer = os.Stderr
		if runtime.GOOS == "windows" {
			if _, err := os.Stderr.Stat(); err != nil {
				w = io.Discard
			}
		}
		logger = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: levelVar}))
		slog.SetDefault(logger)
		logger.Error("Could not get executable path", "error", err)
		return
	}

	logPath := filepath.Join(filepath.Dir(exePath), "convert4share.log")
	
	rotator := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}

	// In windowsgui mode, stderr is unavailable. Log only to file in that case.
	multi := io.Writer(rotator)
	if runtime.GOOS != "windows" {
		multi = io.MultiWriter(os.Stderr, rotator)
	} else if _, err := os.Stderr.Stat(); err == nil {
		multi = io.MultiWriter(os.Stderr, rotator)
	}

	handler := slog.NewTextHandler(multi, &slog.HandlerOptions{
		Level: levelVar,
	})
	logger = slog.New(handler)
	slog.SetDefault(logger)
	log.SetOutput(rotator) // Redirect standard log to lumberjack

	// Initial level
	if isDev() {
		levelVar.Set(slog.LevelDebug)
	} else {
		levelVar.Set(slog.LevelInfo)
	}
}

func updateLoggerLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		levelVar.Set(slog.LevelDebug)
	case "info":
		levelVar.Set(slog.LevelInfo)
	case "warn":
		levelVar.Set(slog.LevelWarn)
	case "error":
		levelVar.Set(slog.LevelError)
	default:
		levelVar.Set(slog.LevelInfo)
	}
	logger.Info("Logger level updated", "level", level)
}

func init() {
	cmd.ConfigTemplate = configTemplate
}

func main() {
	initLogger()
	logger.Info("--------------------")
	logger.Info("App launched", "time", time.Now().String())
	logger.Info("Arguments", "args", os.Args)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install", "uninstall", "help", "--help":
			cmd.Execute()
			return
		}
	}

	app := NewApp()

	exePath, err := os.Executable()
	if err != nil {
		logger.Error("Error getting executable path", "error", err)
	}

	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			if exePath != "" {
				if absArg, err := filepath.Abs(arg); err == nil && strings.EqualFold(absArg, exePath) {
					logger.Info("Skipping executable path in args", "arg", arg)
					continue
				}
			}

			if absArg, err := filepath.Abs(arg); err == nil {
				app.pendingFiles = append(app.pendingFiles, absArg)
			} else {
				app.pendingFiles = append(app.pendingFiles, arg)
			}
		}
	}

	err = wails.Run(&options.App{
		Title:  "Convert4Share",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
		},
		OnStartup:     app.startup,
		OnDomReady:    app.domReady,
		OnShutdown:    app.shutdown,
		OnBeforeClose: app.beforeClose,
		Bind: []interface{}{
			app,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "3310d829-dc96-4613-af1a-a5353d9f07a6",
			OnSecondInstanceLaunch: app.OnSecondInstanceLaunch,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			BackdropType:         windows.Mica,
		},
	})

	if err != nil {
		logger.Error("Wails run error", "error", err.Error())
	}
}

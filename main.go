package main

import (
	"embed"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minjejeon/convert4share/cmd"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed config.example.yaml
var configTemplate []byte

var logger *slog.Logger

func initLogger() {
	if !isDev() {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		log.SetOutput(io.Discard)
		return
	}

	exePath, err := os.Executable()
	if err != nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
		logger.Error("Could not get executable path", "error", err)
		return
	}

	logPath := filepath.Join(filepath.Dir(exePath), fmt.Sprintf("convert4share-debug-%d.log", os.Getpid()))
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
		logger.Error("Could not open log file", "path", logPath, "error", err)
		return
	}

	handler := slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger = slog.New(handler)
	log.SetOutput(logFile)
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
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnShutdown:       app.shutdown,
		OnBeforeClose:    app.beforeClose,
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

package main

import (
	"embed"
	"log"
	"os"

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

func init() {
	cmd.ConfigTemplate = configTemplate
}

func main() {
	// Check if this is a CLI command (install/uninstall)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install", "uninstall", "help", "--help":
			cmd.Execute()
			return
		}
	}

	// Create an instance of the app structure
	app := NewApp()

	// Capture initial arguments (files dropped or passed via context menu)
	// Filter out the executable path (os.Args[0]) and flags if any
	// In Windows context menu usage, args are usually just file paths.
	// But Wails might capture flags too.
	if len(os.Args) > 1 {
		// Just take everything from 1 onwards as potential files
		for _, arg := range os.Args[1:] {
			if _, err := os.Stat(arg); err == nil {
				app.pendingFiles = append(app.pendingFiles, arg)
			}
		}
	}

	// Create application with options
	err := wails.Run(&options.App{
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
			UniqueId:               "50bfe626-4f09-4128-bbf1-c2612babf410",
			OnSecondInstanceLaunch: app.OnSecondInstanceLaunch,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			BackdropType:         windows.Mica,
		},
	})

	if err != nil {
		log.Println("Error:", err.Error())
	}
}

package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/minjejeon/convert4share/converter"
	"github.com/minjejeon/convert4share/windows"
	"github.com/spf13/viper"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx          context.Context
	cfg          *converter.Config
	pendingFiles []string
	mu           sync.Mutex
	isReady      bool
}

type Settings struct {
	MagickBinary        string   `json:"magickBinary"`
	FfmpegBinary        string   `json:"ffmpegBinary"`
	MaxSize             int      `json:"maxSize"`
	HardwareAccelerator string   `json:"hardwareAccelerator"`
	FfmpegCustomArgs    string   `json:"ffmpegCustomArgs"`
	DefaultDestDir      string   `json:"defaultDestDir"`
	ExcludePatterns     []string `json:"excludePatterns"`
}

type JobStatus struct {
	ID       string `json:"id"`
	File     string `json:"file"`
	Status   string `json:"status"` // "queued", "processing", "done", "error"
	Progress int    `json:"progress"`
	Error    string `json:"error,omitempty"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.initConfig()
}

func (a *App) initConfig() {
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

	viper.SetDefault("magickBinary", "magick")
	viper.SetDefault("ffmpegBinary", "ffmpeg")
	viper.SetDefault("maxSize", 1920)
	viper.SetDefault("maxMagickWorkers", 5)
	viper.SetDefault("maxFfmpegWorkers", 1)
	viper.SetDefault("hardwareAccelerator", "none")

	defaultDest := "$HOMEDRIVE/$HOMEPATH/Pictures"
	if home, err := os.UserHomeDir(); err == nil {
		defaultDest = filepath.Join(home, "Pictures")
	}
	viper.SetDefault("defaultDestDir", defaultDest)

	if err := viper.ReadInConfig(); err != nil {
		logger.Info("Config file not found, using defaults", "error", err)
	}
}

func (a *App) processPendingFiles() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isReady {
		logger.Info("processPendingFiles called but DOM not ready yet.")
		return
	}

	if len(a.pendingFiles) > 0 {
		logger.Info("Processing pending files", "files", a.pendingFiles)
		runtime.EventsEmit(a.ctx, "files-received", a.pendingFiles)
		a.pendingFiles = nil
	}
}

func (a *App) GetSettings() Settings {
	return Settings{
		MagickBinary:        viper.GetString("magickBinary"),
		FfmpegBinary:        viper.GetString("ffmpegBinary"),
		MaxSize:             viper.GetInt("maxSize"),
		HardwareAccelerator: viper.GetString("hardwareAccelerator"),
		FfmpegCustomArgs:    viper.GetString("ffmpegCustomArgs"),
		DefaultDestDir:      viper.GetString("defaultDestDir"),
		ExcludePatterns:     viper.GetStringSlice("excludeStringPatterns"),
	}
}

func (a *App) GetContextMenuStatus() bool {
	return windows.IsContextMenuInstalled()
}

func (a *App) InstallContextMenu() error {
	return windows.RunCommandAsAdmin("install")
}

func (a *App) UninstallContextMenu() error {
	return windows.RunCommandAsAdmin("uninstall")
}

func (a *App) SaveSettings(s Settings) error {
	viper.Set("magickBinary", s.MagickBinary)
	viper.Set("ffmpegBinary", s.FfmpegBinary)
	viper.Set("maxSize", s.MaxSize)
	viper.Set("hardwareAccelerator", s.HardwareAccelerator)
	viper.Set("ffmpegCustomArgs", s.FfmpegCustomArgs)
	viper.Set("defaultDestDir", s.DefaultDestDir)
	viper.Set("excludeStringPatterns", s.ExcludePatterns)

	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.yaml")

	return viper.WriteConfigAs(configPath)
}

func (a *App) ConvertFiles(files []string) {
	go func() {
		var wg sync.WaitGroup
		convConfig := &converter.Config{
			MagickBinary:        viper.GetString("magickBinary"),
			FfmpegBinary:        viper.GetString("ffmpegBinary"),
			MaxSize:             viper.GetInt("maxSize"),
			HardwareAccelerator: viper.GetString("hardwareAccelerator"),
			FfmpegCustomArgs:    viper.GetString("ffmpegCustomArgs"),
		}

		reporter := func(file string, percent int, status string, errMsg string) {
			runtime.EventsEmit(a.ctx, "conversion-progress", JobStatus{
				ID:       file,
				File:     file,
				Status:   status,
				Progress: percent,
				Error:    errMsg,
			})
		}

		maxFfmpeg := viper.GetInt("maxFfmpegWorkers")
		maxMagick := viper.GetInt("maxMagickWorkers")
		if maxFfmpeg < 1 {
			maxFfmpeg = 1
		}
		if maxMagick < 1 {
			maxMagick = 1
		}

		ffmpegSem := make(chan struct{}, maxFfmpeg)
		magickSem := make(chan struct{}, maxMagick)

		for _, f := range files {
			fpath := f

			if info, err := os.Stat(fpath); err != nil || info.IsDir() {
				continue
			}

			ext := strings.ToLower(filepath.Ext(fpath))

			fname := filepath.Base(fpath)
			stem := strings.TrimSuffix(fname, filepath.Ext(fname))
			parent := filepath.Dir(fpath)
			cleanedParent := filepath.Clean(parent)

			destDir := parent
			for _, pat := range viper.GetStringSlice("excludeStringPatterns") {
				cleanedPat := filepath.Clean(pat)
				if strings.Contains(cleanedParent, cleanedPat) {
					destDir = os.ExpandEnv(viper.GetString("defaultDestDir"))
					break
				}
			}

			wg.Add(1)
			go func(src string, extension string) {
				defer wg.Done()

				var err error
				reporter(src, 0, "processing", "")

				if extension == ".mov" {
					ffmpegSem <- struct{}{}
					defer func() { <-ffmpegSem }()
					dest := filepath.Join(destDir, stem+".mp4")

					err = convConfig.Ffmpeg(src, dest, func(progress int) {
						reporter(src, progress, "processing", "")
					})
				} else if extension == ".heic" {
					magickSem <- struct{}{}
					defer func() { <-magickSem }()
					dest := filepath.Join(destDir, stem+".jpg")
					err = convConfig.Magick(src, dest)
				} else {
					reporter(src, 0, "error", "Unsupported format")
					return
				}

				if err != nil {
					reporter(src, 100, "error", err.Error())
				} else {
					reporter(src, 100, "done", "")
				}
			}(fpath, ext)
		}

		wg.Wait()
		runtime.EventsEmit(a.ctx, "all-jobs-done", true)
	}()
}

func (a *App) AddFiles(files []string) {
	for _, f := range files {
		if info, err := os.Stat(f); err == nil && !info.IsDir() {
			runtime.EventsEmit(a.ctx, "file-added", f)
		}
	}
}

func (a *App) domReady(ctx context.Context) {
	a.mu.Lock()
	a.isReady = true
	a.mu.Unlock()

	logger.Info("DOM is ready.")
	a.processPendingFiles()
}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

func (a *App) shutdown(ctx context.Context) {
}

func (a *App) OnSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	logger.Info("Second instance launched", "args", secondInstanceData.Args)
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowShow(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowSetAlwaysOnTop(a.ctx, false)

	if len(secondInstanceData.Args) > 0 {
		files := secondInstanceData.Args
		var actualFiles []string
		for _, arg := range files {
			if info, err := os.Stat(arg); err == nil && !info.IsDir() {
				actualFiles = append(actualFiles, arg)
			}
		}
		if len(actualFiles) > 0 {
			logger.Info("Adding files from second instance", "files", actualFiles)
			a.mu.Lock()
			a.pendingFiles = append(a.pendingFiles, actualFiles...)
			a.mu.Unlock()

			a.processPendingFiles()
		} else {
			logger.Info("No valid files found in second instance args.")
		}
	} else {
		logger.Info("Second instance launched with no args.")
	}
}

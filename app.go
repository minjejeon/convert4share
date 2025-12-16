package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
	processTimer *time.Timer
}

type Settings struct {
	MagickBinary        string   `json:"magickBinary"`
	FfmpegBinary        string   `json:"ffmpegBinary"`
	MaxSize             int      `json:"maxSize"`
	HardwareAccelerator string   `json:"hardwareAccelerator"`
	FfmpegCustomArgs    string   `json:"ffmpegCustomArgs"`
	DefaultDestDir      string   `json:"defaultDestDir"`
	ExcludePatterns     []string `json:"excludePatterns"`
	VideoQuality        string   `json:"videoQuality"`
	MaxFfmpegWorkers    int      `json:"maxFfmpegWorkers"`
}

type JobStatus struct {
	ID       string `json:"id"`
	File     string `json:"file"`
	DestFile string `json:"destFile,omitempty"`
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
	viper.SetDefault("videoQuality", "high")

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
		VideoQuality:        viper.GetString("videoQuality"),
		MaxFfmpegWorkers:    viper.GetInt("maxFfmpegWorkers"),
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

func (a *App) CopyFileToClipboard(path string) error {
	return windows.CopyFileToClipboard(path)
}

func (a *App) SaveSettings(s Settings) error {
	viper.Set("magickBinary", s.MagickBinary)
	viper.Set("ffmpegBinary", s.FfmpegBinary)
	viper.Set("maxSize", s.MaxSize)
	viper.Set("hardwareAccelerator", s.HardwareAccelerator)
	viper.Set("ffmpegCustomArgs", s.FfmpegCustomArgs)
	viper.Set("defaultDestDir", s.DefaultDestDir)
	viper.Set("excludeStringPatterns", s.ExcludePatterns)
	viper.Set("videoQuality", s.VideoQuality)
	viper.Set("maxFfmpegWorkers", s.MaxFfmpegWorkers)

	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.yaml")

	return viper.WriteConfigAs(configPath)
}

func (a *App) SelectBinaryDialog() string {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Binary",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Executables",
				Pattern:     "*.exe;*.bat;*.cmd",
			},
			{
				DisplayName: "All Files",
				Pattern:     "*",
			},
		},
	})
	if err != nil {
		return ""
	}
	return selection
}

func (a *App) DetectBinaries() map[string]string {
	results := make(map[string]string)

	if path, err := exec.LookPath("ffmpeg"); err == nil {
		results["ffmpeg"] = path
	}

	if path, err := exec.LookPath("magick"); err == nil {
		results["magick"] = path
	}

	// Helper to check if file exists
	exists := func(p string) bool {
		info, err := os.Stat(p)
		return err == nil && !info.IsDir()
	}

	// Check WinGet locations if not found
	home, err := os.UserHomeDir()
	if err == nil {
		localAppData := filepath.Join(home, "AppData", "Local")
		wingetBase := filepath.Join(localAppData, "Microsoft", "WinGet")

		// 1. Check Links (Symlinks)
		linksDir := filepath.Join(wingetBase, "Links")
		if _, ok := results["ffmpeg"]; !ok {
			if p := filepath.Join(linksDir, "ffmpeg.exe"); exists(p) {
				results["ffmpeg"] = p
			}
		}
		if _, ok := results["magick"]; !ok {
			if p := filepath.Join(linksDir, "magick.exe"); exists(p) {
				results["magick"] = p
			}
		}

		// 2. Check Packages (Actual installation dirs)
		packagesDir := filepath.Join(wingetBase, "Packages")
		entries, err := os.ReadDir(packagesDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				lowerName := strings.ToLower(entry.Name())

				// Helper to search recursively in a package dir
				findInDir := func(dir, binName string) string {
					var found string
					filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
						if err != nil {
							return nil
						}
						if !d.IsDir() && strings.EqualFold(d.Name(), binName) {
							found = path
							return filepath.SkipAll
						}
						return nil
					})
					return found
				}

				if _, ok := results["ffmpeg"]; !ok && strings.Contains(lowerName, "ffmpeg") {
					if p := findInDir(filepath.Join(packagesDir, entry.Name()), "ffmpeg.exe"); p != "" {
						results["ffmpeg"] = p
					}
				}

				if _, ok := results["magick"]; !ok && (strings.Contains(lowerName, "imagemagick") || strings.Contains(lowerName, "magick")) {
					if p := findInDir(filepath.Join(packagesDir, entry.Name()), "magick.exe"); p != "" {
						results["magick"] = p
					}
				}
			}
		}
	}

	return results
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
			VideoQuality:        viper.GetString("videoQuality"),
		}

		reporter := func(file string, destFile string, percent int, status string, errMsg string) {
			runtime.EventsEmit(a.ctx, "conversion-progress", JobStatus{
				ID:       file,
				File:     file,
				DestFile: destFile,
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
				reporter(fpath, "", 0, "error", "File not found or invalid")
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
				var dest string

				if extension == ".mov" {
					dest = filepath.Join(destDir, stem+".mp4")
					reporter(src, dest, 0, "processing", "")

					ffmpegSem <- struct{}{}
					defer func() { <-ffmpegSem }()

					err = convConfig.Ffmpeg(src, dest, func(progress int) {
						reporter(src, dest, progress, "processing", "")
					})
				} else if extension == ".heic" {
					dest = filepath.Join(destDir, stem+".jpg")
					reporter(src, dest, 0, "processing", "")

					magickSem <- struct{}{}
					defer func() { <-magickSem }()

					err = convConfig.Magick(src, dest)
				} else {
					reporter(src, "", 0, "error", "Unsupported format")
					return
				}

				if err != nil {
					reporter(src, dest, 100, "error", err.Error())
				} else {
					reporter(src, dest, 100, "done", "")
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

func (a *App) GetThumbnail(path string) (string, error) {
	convConfig := &converter.Config{
		MagickBinary: viper.GetString("magickBinary"),
		FfmpegBinary: viper.GetString("ffmpegBinary"),
	}

	data, err := convConfig.GenerateThumbnail(path)
	if err != nil {
		logger.Error("Failed to generate thumbnail", "path", path, "error", err)
		return "", fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	base64Str := base64.StdEncoding.EncodeToString(data)
	return "data:image/jpeg;base64," + base64Str, nil
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

	exePath, err := os.Executable()
	if err != nil {
		logger.Error("Error getting executable path during second instance launch", "error", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if len(secondInstanceData.Args) > 0 {
		files := secondInstanceData.Args
		var actualFiles []string
		for _, arg := range files {
			// Skip the argument if it matches the executable path (self-reference)
			if exePath != "" {
				if absArg, err := filepath.Abs(arg); err == nil && strings.EqualFold(absArg, exePath) {
					continue
				}
			}

			if info, err := os.Stat(arg); err == nil && !info.IsDir() {
				actualFiles = append(actualFiles, arg)
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
	a.processTimer = time.AfterFunc(200*time.Millisecond, func() {
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

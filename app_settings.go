//go:build wails2_legacy

package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/minjejeon/convert4share/internal/config"
	"github.com/spf13/viper"
)

// Settings remains declared in the main package so the Wails-generated
// TypeScript bindings keep the `main.Settings` namespace. A future
// phase (post-frontend migration) will move the struct into
// internal/config.
type Settings struct {
	MagickBinary        string   `json:"magickBinary"`
	FfmpegBinary        string   `json:"ffmpegBinary"`
	MaxSize             int      `json:"maxSize"`
	MaxImageSize        int      `json:"maxImageSize"`
	AutoLivePhoto       bool     `json:"autoLivePhoto"`
	CopyOnlyExtensions  []string `json:"copyOnlyExtensions"`
	HardwareAccelerator string   `json:"hardwareAccelerator"`
	FfmpegCustomArgs    string   `json:"ffmpegCustomArgs"`
	DefaultDestDir      string   `json:"defaultDestDir"`
	ExcludePatterns     []string `json:"excludePatterns"`
	VideoQuality        string   `json:"videoQuality"`
	MaxFfmpegWorkers    int      `json:"maxFfmpegWorkers"`
	MaxMagickWorkers    int      `json:"maxMagickWorkers"`
	CollisionOption     string   `json:"collisionOption"`
	LogLevel            string   `json:"logLevel"`
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
	viper.SetDefault("maxImageSize", 2560)
	viper.SetDefault("autoLivePhoto", true)
	viper.SetDefault("copyOnlyExtensions", []string{".jpg", ".jpeg", ".mp4"})
	viper.SetDefault("maxMagickWorkers", 5)
	viper.SetDefault("maxFfmpegWorkers", 1)
	viper.SetDefault("hardwareAccelerator", "none")
	viper.SetDefault("videoQuality", "high")
	viper.SetDefault("collisionOption", "rename")

	defaultLogLevel := "info"
	if isDev() {
		defaultLogLevel = "debug"
	}
	viper.SetDefault("logLevel", defaultLogLevel)

	defaultDest := "$HOMEDRIVE/$HOMEPATH/Pictures"
	if home, err := os.UserHomeDir(); err == nil {
		defaultDest = filepath.Join(home, "Pictures")
	}
	viper.SetDefault("defaultDestDir", defaultDest)

	if err := viper.ReadInConfig(); err != nil {
		logger.Info("Config file not found, using defaults", "error", err)
	}

	// Update logger level from config
	updateLoggerLevel(viper.GetString("logLevel"))

	detected := a.DetectBinaries()

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

	a.updateSemaphores()
}

func (a *App) updateSemaphores() {
	maxFfmpeg := viper.GetInt("maxFfmpegWorkers")
	if maxFfmpeg < 1 {
		maxFfmpeg = 1
	}

	maxMagick := viper.GetInt("maxMagickWorkers")
	if maxMagick < 1 {
		maxMagick = 1
	}

	a.ffmpegSem.Resize(maxFfmpeg)
	a.magickSem.Resize(maxMagick)
}

func (a *App) GetSettings() Settings {
	return Settings{
		MagickBinary:        viper.GetString("magickBinary"),
		FfmpegBinary:        viper.GetString("ffmpegBinary"),
		MaxSize:             viper.GetInt("maxSize"),
		MaxImageSize:        viper.GetInt("maxImageSize"),
		AutoLivePhoto:       viper.GetBool("autoLivePhoto"),
		CopyOnlyExtensions:  viper.GetStringSlice("copyOnlyExtensions"),
		HardwareAccelerator: viper.GetString("hardwareAccelerator"),
		FfmpegCustomArgs:    viper.GetString("ffmpegCustomArgs"),
		DefaultDestDir:      viper.GetString("defaultDestDir"),
		ExcludePatterns:     config.ExcludePatterns(),
		VideoQuality:        viper.GetString("videoQuality"),
		MaxFfmpegWorkers:    viper.GetInt("maxFfmpegWorkers"),
		MaxMagickWorkers:    viper.GetInt("maxMagickWorkers"),
		CollisionOption:     viper.GetString("collisionOption"),
		LogLevel:            viper.GetString("logLevel"),
	}
}

// getExcludePatterns is kept as a thin wrapper for legacy callers within
// the main package; new code should call config.ExcludePatterns
// directly.
func getExcludePatterns() []string {
	return config.ExcludePatterns()
}

func (a *App) SaveSettings(s Settings) error {
	viper.Set("magickBinary", s.MagickBinary)
	viper.Set("ffmpegBinary", s.FfmpegBinary)
	viper.Set("maxSize", s.MaxSize)
	viper.Set("maxImageSize", s.MaxImageSize)
	viper.Set("autoLivePhoto", s.AutoLivePhoto)
	viper.Set("copyOnlyExtensions", s.CopyOnlyExtensions)
	viper.Set("hardwareAccelerator", s.HardwareAccelerator)
	viper.Set("ffmpegCustomArgs", s.FfmpegCustomArgs)
	viper.Set("defaultDestDir", s.DefaultDestDir)
	viper.Set("excludePatterns", s.ExcludePatterns)
	viper.Set("videoQuality", s.VideoQuality)
	viper.Set("maxFfmpegWorkers", s.MaxFfmpegWorkers)
	viper.Set("maxMagickWorkers", s.MaxMagickWorkers)
	viper.Set("collisionOption", s.CollisionOption)
	viper.Set("logLevel", s.LogLevel)

	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.yaml")

	err = viper.WriteConfigAs(configPath)
	if err == nil {
		a.updateSemaphores()
		updateLoggerLevel(s.LogLevel)
	}
	return err
}

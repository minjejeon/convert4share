package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/minjejeon/convert4share/internal/config"
	"github.com/spf13/viper"
)

// Settings is re-exported via a type alias so existing Wails bindings keep
// the same signature when frontend code is regenerated.
type Settings = config.Settings

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
	return config.Load()
}

// getExcludePatterns is kept as a thin wrapper for legacy callers within the
// main package; new code should call config.ExcludePatterns directly.
func getExcludePatterns() []string {
	return config.ExcludePatterns()
}

func (a *App) SaveSettings(s Settings) error {
	if err := config.Save(s); err != nil {
		return err
	}
	a.updateSemaphores()
	updateLoggerLevel(s.LogLevel)
	return nil
}

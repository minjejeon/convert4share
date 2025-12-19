package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/viper"
)

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
	CollisionOption     string   `json:"collisionOption"`
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
	viper.SetDefault("collisionOption", "rename")

	defaultDest := "$HOMEDRIVE/$HOMEPATH/Pictures"
	if home, err := os.UserHomeDir(); err == nil {
		defaultDest = filepath.Join(home, "Pictures")
	}
	viper.SetDefault("defaultDestDir", defaultDest)

	if err := viper.ReadInConfig(); err != nil {
		logger.Info("Config file not found, using defaults", "error", err)
	}

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

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.ffmpegSem == nil || cap(a.ffmpegSem) != maxFfmpeg {
		a.ffmpegSem = make(chan struct{}, maxFfmpeg)
	}
	if a.magickSem == nil || cap(a.magickSem) != maxMagick {
		a.magickSem = make(chan struct{}, maxMagick)
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
		CollisionOption:     viper.GetString("collisionOption"),
	}
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
	viper.Set("collisionOption", s.CollisionOption)

	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.yaml")

	err = viper.WriteConfigAs(configPath)
	if err == nil {
		a.updateSemaphores()
	}
	return err
}

package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

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

// Load reads the current viper state into a Settings struct.
func Load() Settings {
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
		ExcludePatterns:     ExcludePatterns(),
		VideoQuality:        viper.GetString("videoQuality"),
		MaxFfmpegWorkers:    viper.GetInt("maxFfmpegWorkers"),
		MaxMagickWorkers:    viper.GetInt("maxMagickWorkers"),
		CollisionOption:     viper.GetString("collisionOption"),
		LogLevel:            viper.GetString("logLevel"),
	}
}

// Save writes the given Settings into viper and persists them to
// `config.yaml` next to the running executable.
func Save(s Settings) error {
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

	return viper.WriteConfigAs(configPath)
}

// ExcludePatterns reads the exclude patterns list, preferring the canonical
// `excludePatterns` key but falling back to the legacy `excludeStringPatterns`
// key for backward compatibility with existing config.yaml files.
func ExcludePatterns() []string {
	patterns := viper.GetStringSlice("excludePatterns")
	if len(patterns) == 0 {
		patterns = viper.GetStringSlice("excludeStringPatterns")
	}
	return patterns
}

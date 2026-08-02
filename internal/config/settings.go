// Package config owns the Settings struct and the viper-backed helpers
// that load it from / write it to the user's config file. The struct
// definition lived in the main package during the Wails v2 era so that
// the generated TypeScript bindings would track the `main.Settings`
// namespace; the Wails v3 migration regenerates bindings against the
// new services, so the struct is now defined here for clarity.
package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// appDirName is the per-user folder Convert4Share owns under the OS
// config root.
const appDirName = "Convert4Share"

// Dir returns the per-user configuration directory for Convert4Share
// (%APPDATA%\Convert4Share on Windows), creating it if missing. This
// location is always writable by the current user — unlike the install
// directory under Program Files for an all-users install, where storing
// config next to the executable would silently fail for standard users.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, appDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// FilePath returns the absolute path to config.yaml within Dir().
func FilePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Settings is the user-facing configuration surface exposed by the
// SettingsService. The JSON tags are preserved verbatim from the v2
// binding so existing config.yaml files round-trip unchanged.
type Settings struct {
	MagickBinary        string   `json:"magickBinary"`
	FfmpegBinary        string   `json:"ffmpegBinary"`
	FfprobeBinary       string   `json:"ffprobeBinary"`
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
	FileNaming          string   `json:"fileNaming"`
	FileNameFormat      string   `json:"fileNameFormat"`
	LogLevel            string   `json:"logLevel"`
}

// ExcludePatterns reads the exclude patterns list, preferring the
// canonical `excludePatterns` key but falling back to the legacy
// `excludeStringPatterns` key for backward compatibility with existing
// config.yaml files.
func ExcludePatterns() []string {
	patterns := viper.GetStringSlice("excludePatterns")
	if len(patterns) == 0 {
		patterns = viper.GetStringSlice("excludeStringPatterns")
	}
	return patterns
}

// SetDefaults registers viper defaults for every Settings field. The
// defaultLogLevel argument lets callers vary the baseline ("debug" in
// dev builds, "info" otherwise) without dragging build-tag plumbing
// into this package.
func SetDefaults(defaultLogLevel string) {
	viper.SetDefault("magickBinary", "magick")
	viper.SetDefault("ffmpegBinary", "ffmpeg")
	viper.SetDefault("ffprobeBinary", "ffprobe")
	viper.SetDefault("maxSize", 1920)
	viper.SetDefault("maxImageSize", 2560)
	viper.SetDefault("autoLivePhoto", true)
	// .mp4 is deliberately absent: mp4 inputs are probed and copied
	// only when their contents are already shareable. Listing it here
	// would short-circuit that check.
	viper.SetDefault("copyOnlyExtensions", []string{".jpg", ".jpeg"})
	viper.SetDefault("maxMagickWorkers", 5)
	viper.SetDefault("maxFfmpegWorkers", 1)
	viper.SetDefault("hardwareAccelerator", "none")
	viper.SetDefault("videoQuality", "high")
	viper.SetDefault("collisionOption", "rename")
	viper.SetDefault("fileNaming", "original")
	viper.SetDefault("fileNameFormat", "{YY}-{MM}-{DD} {HH}-{mm}-{ss} {num}")

	if defaultLogLevel == "" {
		defaultLogLevel = "info"
	}
	viper.SetDefault("logLevel", defaultLogLevel)

	defaultDest := "$HOMEDRIVE/$HOMEPATH/Pictures"
	if home, err := os.UserHomeDir(); err == nil {
		defaultDest = filepath.Join(home, "Pictures")
	}
	viper.SetDefault("defaultDestDir", defaultDest)
}

// Load materialises the current viper view as a Settings value.
func Load() Settings {
	return Settings{
		MagickBinary:        viper.GetString("magickBinary"),
		FfmpegBinary:        viper.GetString("ffmpegBinary"),
		FfprobeBinary:       viper.GetString("ffprobeBinary"),
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
		FileNaming:          viper.GetString("fileNaming"),
		FileNameFormat:      viper.GetString("fileNameFormat"),
		LogLevel:            viper.GetString("logLevel"),
	}
}

// Save writes the provided Settings back into viper and persists the
// view to the per-user config file (see FilePath) via
// viper.WriteConfigAs. The explicit path is required because
// viper.WriteConfig fails when there is no pre-existing config file.
func Save(s Settings) error {
	viper.Set("magickBinary", s.MagickBinary)
	viper.Set("ffmpegBinary", s.FfmpegBinary)
	viper.Set("ffprobeBinary", s.FfprobeBinary)
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
	viper.Set("fileNaming", s.FileNaming)
	viper.Set("fileNameFormat", s.FileNameFormat)
	viper.Set("logLevel", s.LogLevel)

	configPath, err := FilePath()
	if err != nil {
		return err
	}
	return viper.WriteConfigAs(configPath)
}

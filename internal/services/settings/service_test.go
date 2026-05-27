package settings

import (
	"log/slog"
	"os"
	"testing"

	"github.com/spf13/viper"

	"github.com/minjejeon/convert4share/internal/config"
)

// fakeLevel captures Set() invocations so we can assert SaveSettings
// propagates the requested log level.
type fakeLevel struct {
	last slog.Level
	set  bool
}

func (f *fakeLevel) Set(l slog.Level) {
	f.last = l
	f.set = true
}

func TestGetSettings_ReflectsViperState(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	viper.Set("magickBinary", "C:/tools/magick.exe")
	viper.Set("ffmpegBinary", "C:/tools/ffmpeg.exe")
	viper.Set("maxSize", 4096)
	viper.Set("autoLivePhoto", false)

	s := New(nil, nil)
	got := s.GetSettings()

	if got.MagickBinary != "C:/tools/magick.exe" {
		t.Errorf("MagickBinary: want C:/tools/magick.exe, got %s", got.MagickBinary)
	}
	if got.FfmpegBinary != "C:/tools/ffmpeg.exe" {
		t.Errorf("FfmpegBinary: want C:/tools/ffmpeg.exe, got %s", got.FfmpegBinary)
	}
	if got.MaxSize != 4096 {
		t.Errorf("MaxSize: want 4096, got %d", got.MaxSize)
	}
	if got.AutoLivePhoto {
		t.Errorf("AutoLivePhoto: want false, got true")
	}
}

func TestSaveSettings_UpdatesLevelAndWritesConfig(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	// Save() writes to the per-user config dir (%APPDATA%\Convert4Share
	// on Windows) so an all-users install under Program Files stays
	// writable for standard users. Redirect that base to a temp dir so
	// the test never touches the real profile.
	t.Setenv("AppData", t.TempDir())
	configPath, err := config.FilePath()
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}

	lv := &fakeLevel{}
	s := New(nil, lv)

	in := config.Settings{
		MagickBinary:        "magick",
		FfmpegBinary:        "ffmpeg",
		MaxSize:             1920,
		MaxImageSize:        2560,
		AutoLivePhoto:       true,
		CopyOnlyExtensions:  []string{".jpg", ".mp4"},
		HardwareAccelerator: "none",
		VideoQuality:        "high",
		MaxFfmpegWorkers:    2,
		MaxMagickWorkers:    3,
		CollisionOption:     "rename",
		LogLevel:            "debug",
	}

	if err := s.SaveSettings(in); err != nil {
		t.Fatalf("SaveSettings returned error: %v", err)
	}

	if viper.GetInt("maxFfmpegWorkers") != 2 {
		t.Errorf("viper maxFfmpegWorkers: want 2, got %d", viper.GetInt("maxFfmpegWorkers"))
	}
	if viper.GetInt("maxMagickWorkers") != 3 {
		t.Errorf("viper maxMagickWorkers: want 3, got %d", viper.GetInt("maxMagickWorkers"))
	}
	if !lv.set {
		t.Errorf("level setter not invoked")
	}
	if lv.last != slog.LevelDebug {
		t.Errorf("level: want Debug, got %v", lv.last)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("expected config file at %s, got error: %v", configPath, err)
	}
}

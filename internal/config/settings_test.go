package config

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func TestExcludePatterns_PrefersCanonicalKey(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	want := []string{"Cloud/Photos", "Drive/Pictures"}
	viper.Set("excludePatterns", want)
	viper.Set("excludeStringPatterns", []string{"legacy"})

	got := ExcludePatterns()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Expected canonical key value %v, got %v", want, got)
	}
}

func TestExcludePatterns_FallsBackToLegacyKey(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	legacy := []string{"Cloud/Photos", "Drive/Pictures"}
	viper.Set("excludeStringPatterns", legacy)

	got := ExcludePatterns()
	if !reflect.DeepEqual(got, legacy) {
		t.Errorf("Expected fallback to legacy key %v, got %v", legacy, got)
	}
}

func TestLoad_ReflectsCanonicalAndLegacyKeys(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	want := []string{"Some Cloud/Photos", "Google Drive/My Pictures"}
	viper.Set("excludePatterns", want)

	if got := Load().ExcludePatterns; !reflect.DeepEqual(got, want) {
		t.Errorf("Canonical key not honored by Load: want %v, got %v", want, got)
	}

	viper.Reset()
	legacy := []string{"Legacy Path/Photos"}
	viper.Set("excludeStringPatterns", legacy)
	if got := Load().ExcludePatterns; !reflect.DeepEqual(got, legacy) {
		t.Errorf("Legacy key not honored by Load fallback: want %v, got %v", legacy, got)
	}
}

func TestSetDefaults_AppliesExpectedValues(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	SetDefaults("debug")

	s := Load()
	if s.MagickBinary != "magick" {
		t.Errorf("MagickBinary default: want magick, got %s", s.MagickBinary)
	}
	if s.FfmpegBinary != "ffmpeg" {
		t.Errorf("FfmpegBinary default: want ffmpeg, got %s", s.FfmpegBinary)
	}
	if s.MaxSize != 1920 {
		t.Errorf("MaxSize default: want 1920, got %d", s.MaxSize)
	}
	if s.MaxImageSize != 2560 {
		t.Errorf("MaxImageSize default: want 2560, got %d", s.MaxImageSize)
	}
	if !s.AutoLivePhoto {
		t.Errorf("AutoLivePhoto default: want true, got false")
	}
	if s.MaxFfmpegWorkers != 1 {
		t.Errorf("MaxFfmpegWorkers default: want 1, got %d", s.MaxFfmpegWorkers)
	}
	if s.MaxMagickWorkers != 5 {
		t.Errorf("MaxMagickWorkers default: want 5, got %d", s.MaxMagickWorkers)
	}
	if s.HardwareAccelerator != "none" {
		t.Errorf("HardwareAccelerator default: want none, got %s", s.HardwareAccelerator)
	}
	if s.VideoQuality != "high" {
		t.Errorf("VideoQuality default: want high, got %s", s.VideoQuality)
	}
	if s.CollisionOption != "rename" {
		t.Errorf("CollisionOption default: want rename, got %s", s.CollisionOption)
	}
	if s.LogLevel != "debug" {
		t.Errorf("LogLevel default: want debug (caller-supplied), got %s", s.LogLevel)
	}
	wantCopy := []string{".jpg", ".jpeg", ".mp4"}
	if !reflect.DeepEqual(s.CopyOnlyExtensions, wantCopy) {
		t.Errorf("CopyOnlyExtensions default: want %v, got %v", wantCopy, s.CopyOnlyExtensions)
	}
}

func TestFileNamingDefaults(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	SetDefaults("info")
	s := Load()
	if s.FileNaming != "original" {
		t.Errorf("FileNaming default: want original, got %q", s.FileNaming)
	}
	if s.FileNameFormat != "{YY}-{MM}-{DD} {HH}-{mm}-{ss} {num}" {
		t.Errorf("FileNameFormat default mismatch, got %q", s.FileNameFormat)
	}
}

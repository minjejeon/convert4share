package cmd

import (
	"bytes"
	_ "embed"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ConfigTemplate []byte

var (
	RootCmd = &cobra.Command{
		Use:   "convert4share [file]",
		Short: "Converts .mov and .heic files to .mp4 and .jpg.",
		Long:  `A simple utility to convert media files for better compatibility.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

func initConfig() {

	exePath, err := os.Executable()
	cobra.CheckErr(err)
	exeDir := filepath.Dir(exePath)

	viper.AddConfigPath(exeDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	configReadErr := viper.ReadInConfig()
	if configReadErr == nil {
		slog.Info("Using config file", "path", viper.ConfigFileUsed())
	}

	viper.SetDefault("magickBinary", "magick")
	viper.SetDefault("ffmpegBinary", "ffmpeg")

	defaultDest := "$HOMEPATH/Pictures"
	if home, err := os.UserHomeDir(); err == nil {
		defaultDest = filepath.Join(home, "Pictures")
	}
	viper.SetDefault("defaultDestDir", defaultDest)

	viper.SetDefault("excludePatterns", []string{})
	viper.SetDefault("maxSize", 1920)
	viper.SetDefault("maxMagickWorkers", 5)
	viper.SetDefault("maxFfmpegWorkers", 1)
	viper.SetDefault("ffmpegCustomArgs", "")

	if !viper.IsSet("hardwareAccelerator") {
		slog.Info("hardwareAccelerator not set. Detecting GPU...")
		detectedAccelerator := "none"
		if isNvidiaGpu() {
			slog.Info("NVIDIA GPU detected.")
			detectedAccelerator = "nvidia"
		} else if isAmdGpu() {
			slog.Info("AMD GPU detected.")
			detectedAccelerator = "amd"
		} else {
			slog.Info("No supported GPU detected, defaulting to software encoding.")
		}
		viper.Set("hardwareAccelerator", detectedAccelerator)

		if _, ok := configReadErr.(viper.ConfigFileNotFoundError); ok {
			createDefaultConfig(exeDir)
		}
	}
}

func createDefaultConfig(dir string) {
	configPath := filepath.Join(dir, "config.yaml")
	slog.Info("Config file not found. Creating a new one", "path", configPath)

	detectedAccelerator := viper.GetString("hardwareAccelerator")
	content := strings.Replace(string(ConfigTemplate), `hardwareAccelerator: "none"`, `hardwareAccelerator: "`+detectedAccelerator+`"`, 1)

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		slog.Error("Error creating config file", "error", err)
	}
}

func isNvidiaGpu() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "Get-CimInstance Win32_VideoController | Select-Object -ExpandProperty Caption")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		slog.Error("Failed to detect GPU using PowerShell", "error", err)
		return false
	}
	return strings.Contains(strings.ToUpper(out.String()), "NVIDIA")
}

func isAmdGpu() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "Get-CimInstance Win32_VideoController | Select-Object -ExpandProperty Caption")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		slog.Error("Failed to detect GPU using PowerShell", "error", err)
		return false
	}
	return strings.Contains(strings.ToUpper(out.String()), "AMD")
}

func Execute() {
	cobra.MousetrapHelpText = ""
	cobra.OnInitialize(initConfig)
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

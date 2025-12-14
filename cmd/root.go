package cmd

import (
	"bytes"
	_ "embed"
	"log"
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
		log.Println("Using config file:", viper.ConfigFileUsed())
	}

	viper.SetDefault("magickBinary", "magick")
	viper.SetDefault("ffmpegBinary", "ffmpeg")

	defaultDest := "$HOMEPATH/Pictures"
	if home, err := os.UserHomeDir(); err == nil {
		defaultDest = filepath.Join(home, "Pictures")
	}
	viper.SetDefault("defaultDestDir", defaultDest)

	viper.SetDefault("excludeStringPatterns", []string{})
	viper.SetDefault("maxSize", 1920)
	viper.SetDefault("maxMagickWorkers", 5)
	viper.SetDefault("maxFfmpegWorkers", 1)
	viper.SetDefault("ffmpegCustomArgs", "")

	if !viper.IsSet("hardwareAccelerator") {
		log.Println("hardwareAccelerator not set. Detecting GPU...")
		detectedAccelerator := "none"
		if isNvidiaGpu() {
			log.Println("NVIDIA GPU detected.")
			detectedAccelerator = "nvidia"
		} else if isAmdGpu() {
			log.Println("AMD GPU detected.")
			detectedAccelerator = "amd"
		} else {
			log.Println("No supported GPU detected, defaulting to software encoding.")
		}
		viper.Set("hardwareAccelerator", detectedAccelerator)

		if _, ok := configReadErr.(viper.ConfigFileNotFoundError); ok {
			createDefaultConfig(exeDir)
		}
	}
}

func createDefaultConfig(dir string) {
	configPath := filepath.Join(dir, "config.yaml")
	log.Printf("Config file not found. Creating a new one at: %s", configPath)

	detectedAccelerator := viper.GetString("hardwareAccelerator")
	content := strings.Replace(string(ConfigTemplate), `hardwareAccelerator: "none"`, `hardwareAccelerator: "`+detectedAccelerator+`"`, 1)

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		log.Printf("Error creating config file: %v", err)
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
		log.Printf("Failed to detect GPU using PowerShell: %v", err)
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
		log.Printf("Failed to detect GPU using PowerShell: %v", err)
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

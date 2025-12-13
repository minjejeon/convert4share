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
	// rootCmd represents the base command when called without any subcommands
	RootCmd = &cobra.Command{
		Use:   "convert4share [file]",
		Short: "Converts .mov and .heic files to .mp4 and .jpg.",
		Long:  `A simple utility to convert media files for better compatibility.`,
		// The default action is now handled by the main application logic (Wails),
		// so this Command is mainly a container for install/uninstall subcommands.
		// If users run the binary without subcommands, it should probably show help if no args,
		// but main.go handles the "GUI mode" if args are present or not.
		// We only reach here if main.go explicitly calls Execute.
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	exePath, err := os.Executable()
	cobra.CheckErr(err)
	exeDir := filepath.Dir(exePath)

	// Search config in home directory, executable directory with name "config.yaml".
	viper.AddConfigPath(exeDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	configReadErr := viper.ReadInConfig()
	if configReadErr == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}

	// Set default values
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

	// Auto-detect and set hardwareAccelerator if not present
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

		// Create a new config file if it wasn't found
		if _, ok := configReadErr.(viper.ConfigFileNotFoundError); ok {
			createDefaultConfig(exeDir)
		}
	}
}

// createDefaultConfig writes a new config.yaml in the executable's directory using the embedded template.
func createDefaultConfig(dir string) {
	configPath := filepath.Join(dir, "config.yaml")
	log.Printf("Config file not found. Creating a new one at: %s", configPath)

	// Get the detected accelerator value from viper
	detectedAccelerator := viper.GetString("hardwareAccelerator")
	// Replace the placeholder in the template
	content := strings.Replace(string(ConfigTemplate), `hardwareAccelerator: "none"`, `hardwareAccelerator: "`+detectedAccelerator+`"`, 1)

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		log.Printf("Error creating config file: %v", err)
	}
}

// isNvidiaGpu checks if an NVIDIA GPU is present by checking video controller descriptions.
func isNvidiaGpu() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	// Use PowerShell as it's more reliable than `wmic` which may be deprecated.
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "Get-CimInstance Win32_VideoController | Select-Object -ExpandProperty Caption")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to detect GPU using PowerShell: %v", err)
		return false
	}
	return strings.Contains(strings.ToUpper(out.String()), "NVIDIA")
}

// isAmdGpu checks if an AMD GPU is present by checking video controller descriptions.
func isAmdGpu() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	// Use PowerShell as it's more reliable than `wmic` which may be deprecated.
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "Get-CimInstance Win32_VideoController | Select-Object -ExpandProperty Caption")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to detect GPU using PowerShell: %v", err)
		return false
	}
	return strings.Contains(strings.ToUpper(out.String()), "AMD")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.MousetrapHelpText = ""
	cobra.OnInitialize(initConfig)
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

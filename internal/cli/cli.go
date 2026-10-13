package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/viper"

	"github.com/minjejeon/convert4share/internal/batch"
	"github.com/minjejeon/convert4share/internal/config"
	"github.com/minjejeon/convert4share/internal/converter"
	"github.com/minjejeon/convert4share/internal/logging"
	"github.com/minjejeon/convert4share/internal/services/tools"
)

const (
	Version   = "0.7.0"
	usageText = `Usage: c4s [options] <file1> [file2 ...]

Convert MOV and HEIC files to MP4 and JPG using hardware acceleration.

Options:
  -o, --output <dir>        Output directory for converted files
  -q, --quality <preset>    Video quality preset: high, medium, low (default: from config)
      --hwaccel <type>      Hardware accelerator: auto, nvidia, amd, vaapi, cpu (default: from config)
      --max-size <px>       Max dimension for video resolution in pixels (default: 0, no limit)
      --max-image-size <px> Max dimension for HEIC image in pixels (default: 2560)
      --collision <mode>    Collision resolution: rename, overwrite, skip, error (default: rename)
      --live-photo          Auto-skip .mov paired with .heic (default: true)
  -j, --concurrency <n>     Max parallel conversions (default: 2)
  -v, --verbose             Enable detailed debug logging to stderr
  -h, --help                Show this help message
  -V, --version             Print version information
`
)

// Run runs the CLI command with standard OS I/O streams and returns the exit code.
func Run(args []string) int {
	return RunWithIO(args, os.Stdout, os.Stderr)
}

// RunWithIO executes the CLI command with custom stdout and stderr writers.
func RunWithIO(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("c4s", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		outputDir       string
		quality         string
		hwaccel         string
		maxSize         int
		maxImageSize    int
		collision       string
		autoLivePhoto   bool
		concurrency     int
		verbose         bool
		showHelp        bool
		showVersion     bool
	)

	fs.StringVar(&outputDir, "o", "", "Output directory")
	fs.StringVar(&outputDir, "output", "", "Output directory")
	fs.StringVar(&quality, "q", "", "Video quality preset: high, medium, low")
	fs.StringVar(&quality, "quality", "", "Video quality preset: high, medium, low")
	fs.StringVar(&hwaccel, "hwaccel", "", "Hardware accelerator")
	fs.IntVar(&maxSize, "max-size", 0, "Max dimension for video")
	fs.IntVar(&maxImageSize, "max-image-size", 0, "Max dimension for HEIC")
	fs.StringVar(&collision, "collision", "", "Collision resolution: rename, overwrite, skip, error")
	fs.BoolVar(&autoLivePhoto, "live-photo", true, "Auto-skip Live Photo MOV")
	fs.IntVar(&concurrency, "j", 0, "Max parallel conversions")
	fs.IntVar(&concurrency, "concurrency", 0, "Max parallel conversions")
	fs.BoolVar(&verbose, "v", false, "Enable verbose logging")
	fs.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	fs.BoolVar(&showHelp, "h", false, "Show help")
	fs.BoolVar(&showHelp, "help", false, "Show help")
	fs.BoolVar(&showVersion, "V", false, "Show version")
	fs.BoolVar(&showVersion, "version", false, "Show version")

	fs.Usage = func() {
		fmt.Fprint(stdout, usageText)
	}

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if showHelp {
		fs.Usage()
		return 0
	}

	if showVersion {
		fmt.Fprintf(stdout, "c4s (Convert4Share) version %s\n", Version)
		return 0
	}

	files := fs.Args()
	if len(files) == 0 {
		fmt.Fprintln(stderr, "c4s: error: no input files provided")
		fmt.Fprintln(stderr, "Try 'c4s --help' for more information.")
		return 1
	}

	// Initialize logger
	levelVar := new(slog.LevelVar)
	if verbose {
		levelVar.Set(slog.LevelDebug)
	} else {
		levelVar.Set(slog.LevelWarn)
	}
	logger := logging.New(levelVar)

	// Load configuration
	loadConfig(logger)

	// Binary detection
	toolsService := tools.New(logger)
	autoDetectBinaries(logger, toolsService)

	// Resolve effective options
	effQuality := viper.GetString("videoQuality")
	if quality != "" {
		effQuality = quality
	}
	if effQuality == "" {
		effQuality = "high"
	}

	effHwaccel := viper.GetString("hardwareAccelerator")
	if hwaccel != "" {
		effHwaccel = hwaccel
	}

	effMaxSize := viper.GetInt("maxSize")
	if maxSize > 0 {
		effMaxSize = maxSize
	}

	effMaxImageSize := viper.GetInt("maxImageSize")
	if maxImageSize > 0 {
		effMaxImageSize = maxImageSize
	}
	if effMaxImageSize <= 0 {
		effMaxImageSize = 2560
	}

	effCollision := viper.GetString("collisionOption")
	if collision != "" {
		effCollision = collision
	}
	if effCollision == "" {
		effCollision = "rename"
	}

	effConcurrency := viper.GetInt("ffmpegConcurrency")
	if concurrency > 0 {
		effConcurrency = concurrency
	}
	if effConcurrency <= 0 {
		effConcurrency = 2
	}

	convConfig := &converter.Config{
		MagickBinary:        viper.GetString("magickBinary"),
		FfmpegBinary:        viper.GetString("ffmpegBinary"),
		MaxSize:             effMaxSize,
		MaxImageSize:        effMaxImageSize,
		HardwareAccelerator: effHwaccel,
		FfmpegCustomArgs:    viper.GetString("ffmpegCustomArgs"),
		VideoQuality:        effQuality,
	}

	batchOpts := batch.Options{
		CollisionOption:    effCollision,
		AutoLivePhoto:      autoLivePhoto,
		CopyOnlyExtensions: viper.GetStringSlice("copyOnlyExtensions"),
		DefaultDestDir:     viper.GetString("defaultDestDir"),
		ExcludePatterns:    config.ExcludePatterns(),
		CustomDestDir:      outputDir,
		FFmpegConcurrency:  effConcurrency,
		MagickConcurrency:  effConcurrency,
	}

	runner := batch.NewRunner(convConfig, batchOpts, logger)

	// Context with interrupt cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(stderr, "\nOperation cancelled by user.")
		cancel()
	}()

	startTime := time.Now()
	totalFiles := len(files)

	reporter := func(status batch.JobStatus) {
		base := filepath.Base(status.File)
		destBase := filepath.Base(status.DestFile)
		switch status.Status {
		case "done":
			if status.Error != "" {
				fmt.Fprintf(stdout, "[SKIP] %s (%s)\n", base, status.Error)
			} else {
				fmt.Fprintf(stdout, "[OK]   %s -> %s\n", base, destBase)
			}
		case "error":
			fmt.Fprintf(stderr, "[FAIL] %s: %s\n", base, status.Error)
		case "processing":
			if verbose && status.Progress > 0 {
				fmt.Fprintf(stdout, "       %s: %d%% (%s)\n", base, status.Progress, status.Speed)
			}
		}
	}

	success, skipped, failed := runner.Run(ctx, files, reporter)
	elapsed := time.Since(startTime).Round(time.Millisecond)

	fmt.Fprintf(stdout, "\nFinished %d files: %d converted, %d skipped, %d failed (took %s)\n",
		totalFiles, success, skipped, failed, elapsed)

	if failed > 0 {
		return 1
	}
	return 0
}

func loadConfig(logger *slog.Logger) {
	configDir, err := config.Dir()
	if err != nil {
		logger.Debug("Could not determine config directory", "error", err)
		return
	}

	viper.AddConfigPath(configDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	config.SetDefaults("info")
	_ = viper.ReadInConfig()
}

func autoDetectBinaries(logger *slog.Logger, ts *tools.Service) {
	checkBinary := func(key string) bool {
		val := viper.GetString(key)
		if val == "" {
			return false
		}
		if _, err := exec.LookPath(val); err != nil {
			return false
		}
		return true
	}

	detected := ts.DetectBinaries()

	if !checkBinary("ffmpegBinary") {
		if path, ok := detected["ffmpeg"]; ok {
			viper.Set("ffmpegBinary", path)
		}
	}

	if !checkBinary("magickBinary") {
		if path, ok := detected["magick"]; ok {
			viper.Set("magickBinary", path)
		}
	}
}

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
	"sync"
	"time"

	"github.com/minjejeon/convert4share/converter"
	"github.com/minjejeon/convert4share/windows"
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
		Args:  cobra.ArbitraryArgs,
		Run:   run,
	}
)

const (
	progUUID = "50bfe626-4f09-4128-bbf1-c2612babf410"
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
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}

	// Set default values
	viper.SetDefault("magickBinary", "magick")
	viper.SetDefault("ffmpegBinary", "ffmpeg")
	viper.SetDefault("defaultDestDir", "$HOMEPATH/Downloads")
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
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
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

// activeJobs tracks the number of files currently being processed.
// It must be accessed using atomic operations to ensure thread safety.
var isFirstInstance bool

// processFile determines the conversion type and creates a job.
func processFile(fpath string, ffmpegJobs chan<- converter.Job, magickJobs chan<- converter.Job, wg *sync.WaitGroup) {
	fname := filepath.Base(fpath)
	log.Printf("Processing file: %s", fname)
	ext := strings.ToLower(filepath.Ext(fname))
	stem := strings.TrimSuffix(fname, filepath.Ext(fname))

	parent := filepath.Dir(fpath)
	cleanedParent := filepath.Clean(parent)

	destDir := parent
	for _, pat := range viper.GetStringSlice("excludeStringPatterns") {
		cleanedPat := filepath.Clean(pat)
		if strings.Contains(cleanedParent, cleanedPat) {
			destDir = os.ExpandEnv(viper.GetString("defaultDestDir"))
			break
		}
	}

	var job converter.Job
	switch ext {
	case ".mov":
		wg.Add(1)
		job = converter.Job{Orig: fpath, Dest: filepath.Join(destDir, stem+".mp4")}
		log.Printf("Queueing ffmpeg job: %s -> %s", job.Orig, job.Dest)
		ffmpegJobs <- job
	case ".heic":
		wg.Add(1)
		job = converter.Job{Orig: fpath, Dest: filepath.Join(destDir, stem+".jpg")}
		log.Printf("Queueing magick job: %s -> %s", job.Orig, job.Dest)
		magickJobs <- job
	default:
		log.Printf("%s is an unsupported format: %s", fpath, ext)
	}
}

// startWorkers launches goroutines that will process jobs from the channels.
func startWorkers(conv *converter.Config, ffmpegJobs <-chan converter.Job, magickJobs <-chan converter.Job, wg *sync.WaitGroup, cond *sync.Cond) {
	jobDone := func() {
		cond.Signal()
		wg.Done()
	}
	// Ffmpeg workers
	for i := 0; i < viper.GetInt("maxFfmpegWorkers"); i++ {
		go func(workerID int) {
			for job := range ffmpegJobs {
				log.Printf("[ffmpeg-worker-%d] Starting job for %s", workerID, job.Orig)
				err := conv.Ffmpeg(job.Orig, job.Dest)
				if err != nil {
					log.Printf("[ffmpeg-worker-%d] ERROR: %v", workerID, err)
				} else {
					log.Printf("[ffmpeg-worker-%d] Finished job for %s", workerID, job.Orig)
				}
				jobDone()
			}
		}(i)
	}

	// Magick workers
	for i := 0; i < viper.GetInt("maxMagickWorkers"); i++ {
		go func(workerID int) {
			for job := range magickJobs {
				log.Printf("[magick-worker-%d] Starting job for %s", workerID, job.Orig)
				err := conv.Magick(job.Orig, job.Dest)
				if err != nil {
					log.Printf("[magick-worker-%d] ERROR: %v", workerID, err)
				} else {
					log.Printf("[magick-worker-%d] Finished job for %s", workerID, job.Orig)
				}
				jobDone()
			}
		}(i)
	}
}

func run(cmd *cobra.Command, args []string) {
	// If no arguments are provided, show the help message and exit.
	if len(args) == 0 {
		cmd.Help()
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// This channel will receive file paths from other instances.
	secondInstanceBuffer := make(chan string, 100)

	// Setup single instance IPC. This will return true for the first instance,
	// and exit if another instance is found.
	// If this is the first instance, it will continue and start listening for messages.
	isFirstInstance = windows.SetupSingleInstance(progUUID, secondInstanceBuffer)
	if isFirstInstance {
		log.Println("This is the first instance. Starting master process and workers...")
	}

	// Create job channels
	ffmpegJobs := make(chan converter.Job, 100)
	magickJobs := make(chan converter.Job, 100)

	// Create a converter config from viper settings
	convConfig := &converter.Config{
		MagickBinary:        viper.GetString("magickBinary"),
		FfmpegBinary:        viper.GetString("ffmpegBinary"),
		MaxSize:             viper.GetInt("maxSize"),
		HardwareAccelerator: viper.GetString("hardwareAccelerator"),
		FfmpegCustomArgs:    viper.GetString("ffmpegCustomArgs"),
	}

	cond := sync.NewCond(&mu)
	// Start worker pools
	startWorkers(convConfig, ffmpegJobs, magickJobs, &wg, cond)

	// Process the file arguments from the first instance itself
	for _, arg := range args {
		fpath, err := filepath.Abs(arg)
		if err != nil {
			log.Printf("Could not get absolute path for %s: %v", arg, err)
			continue
		}
		processFile(fpath, ffmpegJobs, magickJobs, &wg)
	}

	// Loop, processing files from other instances, and exit on idle.
	const idleTimeout = 10 * time.Second
	idleTimer := time.NewTimer(idleTimeout)
	// Stop the timer immediately. It will be reset only when all work is done.
	if !idleTimer.Stop() {
		<-idleTimer.C
	}

	// This goroutine waits for all jobs to finish, then starts the idle timer.
	go func() {
		mu.Lock()
		defer mu.Unlock()
		for {
			wg.Wait()
			log.Println("All active jobs finished. Starting idle timer.")
			idleTimer.Reset(idleTimeout)
			// Wait for a signal that a new job has finished.
			cond.Wait()
		}
	}()

	for {
		select {
		case fpath := <-secondInstanceBuffer:
			log.Printf("Received new file %s. Stopping idle timer until work is done.", fpath)
			// Stop the timer as soon as new work arrives.
			if !idleTimer.Stop() {
				// If the timer has already fired, drain the channel to prevent a premature exit.
				select {
				case <-idleTimer.C:
				default:
				}
			}

			processFile(fpath, ffmpegJobs, magickJobs, &wg)

		case <-idleTimer.C:
			log.Println("Idle timeout reached. Waiting for any remaining jobs to finish...")
			// Final check to ensure no new jobs came in while the timer was firing.
			wg.Wait()
			log.Println("All jobs finished. Exiting.")
			// Close channels to terminate worker goroutines gracefully.
			close(ffmpegJobs)
			close(magickJobs)
			return // Exit main
		}
	}
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

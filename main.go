package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/minjejeon/convert4share/windows"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sys/windows/registry"
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "convert4share [file]",
		Short: "Converts .mov and .heic files to .mp4 and .jpg.",
		Long:  `A simple utility to convert media files for better compatibility.`,
		Args:  cobra.ArbitraryArgs,
		Run:   run,
	}

	installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install the application to the Windows context menu.",
		Long: `Adds a 'Convert with Convert4Share' option to the context menu
for .mov and .heic files. This command must be run with administrator privileges.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !windows.IsElevated() {
				windows.RunAsAdmin()
				return
			}
			if err := registerContextMenu(); err != nil {
				log.Fatalf("Failed to install context menu: %v. Please ensure you are running this command as an administrator.", err)
			}
			log.Println("Context menu installed successfully for .mov and .heic files.")
		},
	}

	uninstallCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the application from the Windows context menu.",
		Long: `Removes the 'Convert with Convert4Share' option from the context menu.
This command must be run with administrator privileges.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !windows.IsElevated() {
				windows.RunAsAdmin()
				return
			}
			if err := unregisterContextMenu(); err != nil {
				log.Fatalf("Failed to uninstall context menu: %v. Please ensure you are running this command as an administrator.", err)
			}
			log.Println("Context menu uninstalled successfully.")
		},
	}
)

const (
	progUUID = "50bfe626-4f09-4128-bbf1-c2612babf410"
)

//go:embed config.example.yaml
var configTemplate []byte

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
	content := strings.Replace(string(configTemplate), `hardwareAccelerator: "none"`, `hardwareAccelerator: "`+detectedAccelerator+`"`, 1)

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

// Job defines a conversion task.
type Job struct{ Orig, Dest string }

func magick(orig, dest string) error {
	cmd := exec.Command(viper.GetString("magickBinary"), orig, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("Running magick command: %s", cmd.String())
	return cmd.Run()
}

func ffmpeg(orig, dest string) error {
	ffmpegBinary := viper.GetString("ffmpegBinary")
	args := []string{
		"-hide_banner",
		"-loglevel", "warning",
		"-stats",
		"-y", // Overwrite output files without asking
	}

	maxSize := viper.GetInt("maxSize")
	scaleArg := fmt.Sprintf("scale='w=%d:h=%d:force_original_aspect_ratio=decrease'", maxSize, maxSize)

	// Set video codec based on configuration
	accelerator := strings.ToLower(viper.GetString("hardwareAccelerator"))
	switch accelerator {
	case "amd":
		log.Println("Using 'amd' hardware accelerator (h264_amf) from config.")
		amdScaleArg := fmt.Sprintf("vpp_amf='w=%d:h=%d:force_original_aspect_ratio=decrease'", maxSize, maxSize)
		args = append(args,
			"-i", orig,
			"-c:v", "h264_amf",
			"-vf", amdScaleArg,
		)
	case "nvidia":
		log.Println("Using 'nvidia' hardware accelerator (h264_nvenc) from config.")
		args = append(args, "-hwaccel", "cuda", "-i", orig, "-c:v", "h264_nvenc", "-vf", scaleArg)
	case "none", "": // Handles 'none' or null/empty value from yaml
		log.Println("Using software encoder (libx264).")
		args = append(args, "-i", orig, "-c:v", "libx264", "-vf", scaleArg)
	default:
		log.Printf("Unknown hardwareAccelerator '%s', falling back to software encoder (libx264).", accelerator)
		args = append(args, "-i", orig, "-c:v", "libx264", "-vf", scaleArg)
	}

	// Add custom ffmpeg arguments from config
	customArgs := viper.GetString("ffmpegCustomArgs")
	if customArgs != "" {
		// Split the string by spaces to get individual arguments
		// This handles multiple arguments in the string correctly.
		log.Printf("Adding custom ffmpeg arguments: %s", customArgs)
		args = append(args, strings.Fields(customArgs)...)
	}

	// Add scaling filter and audio codec

	args = append(args,
		"-c:a", "aac",
		dest,
	)

	cmd := exec.Command(ffmpegBinary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("Running ffmpeg command: %s", cmd.String())
	return cmd.Run()
}

// activeJobs tracks the number of files currently being processed.
// It must be accessed using atomic operations to ensure thread safety.
var isFirstInstance bool

// processFile determines the conversion type and creates a job.
func processFile(fpath string, ffmpegJobs chan<- Job, magickJobs chan<- Job, wg *sync.WaitGroup) {
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

	var job Job
	switch ext {
	case ".mov":
		wg.Add(1)
		job = Job{Orig: fpath, Dest: filepath.Join(destDir, stem+".mp4")}
		log.Printf("Queueing ffmpeg job: %s -> %s", job.Orig, job.Dest)
		ffmpegJobs <- job
	case ".heic":
		wg.Add(1)
		job = Job{Orig: fpath, Dest: filepath.Join(destDir, stem+".jpg")}
		log.Printf("Queueing magick job: %s -> %s", job.Orig, job.Dest)
		magickJobs <- job
	default:
		log.Printf("%s is an unsupported format: %s", fpath, ext)
	}
}

// startWorkers launches goroutines that will process jobs from the channels.
func startWorkers(ffmpegJobs <-chan Job, magickJobs <-chan Job, wg *sync.WaitGroup, cond *sync.Cond) {
	jobDone := func() {
		cond.Signal()
		wg.Done()
	}
	// Ffmpeg workers
	for i := 0; i < viper.GetInt("maxFfmpegWorkers"); i++ {
		go func(workerID int) {
			for job := range ffmpegJobs {
				log.Printf("[ffmpeg-worker-%d] Starting job for %s", workerID, job.Orig)
				err := ffmpeg(job.Orig, job.Dest)
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
				err := magick(job.Orig, job.Dest)
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
	ffmpegJobs := make(chan Job, 100)
	magickJobs := make(chan Job, 100)

	cond := sync.NewCond(&mu)
	// Start worker pools
	startWorkers(ffmpegJobs, magickJobs, &wg, cond)

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

func registerContextMenu() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %w", err)
	}

	menuName := "Convert with Convert4Share"
	keyName := "Convert4Share"
	command := fmt.Sprintf(`"%s" "%%1"`, exePath) // Reverted to simple command
	extensions := []string{".mov", ".heic"}

	for _, ext := range extensions {
		// Create HKEY_CLASSES_ROOT\SystemFileAssociations\.ext\shell\Convert4Share
		keyPath := fmt.Sprintf(`SystemFileAssociations\%s\shell\%s`, ext, keyName)
		key, _, err := registry.CreateKey(registry.CLASSES_ROOT, keyPath, registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("could not create shell key for %s: %w", ext, err)
		}
		// Set the menu display text and icon
		if err := key.SetStringValue("", menuName); err != nil {
			key.Close()
			return fmt.Errorf("could not set default value for %s: %w", ext, err)
		}
		if err := key.SetStringValue("Icon", `"`+exePath+`"`); err != nil {
			key.Close()
			return fmt.Errorf("could not set icon for %s: %w", ext, err)
		}

		key.Close()

		// Create HKEY_CLASSES_ROOT\SystemFileAssociations\.ext\shell\Convert4Share\command
		cmdKeyPath := fmt.Sprintf(`%s\command`, keyPath)
		cmdKey, _, err := registry.CreateKey(registry.CLASSES_ROOT, cmdKeyPath, registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("could not create command key for %s: %w", ext, err)
		}
		if err := cmdKey.SetStringValue("", command); err != nil {
			cmdKey.Close()
			return fmt.Errorf("could not set command value for %s: %w", ext, err)
		}
		cmdKey.Close()
	}
	return nil
}

func unregisterContextMenu() error {
	extensions := []string{".mov", ".heic"}
	for _, ext := range extensions {
		keyPath := fmt.Sprintf(`SystemFileAssociations\%s\shell\Convert4Share`, ext)
		if err := registry.DeleteKey(registry.CLASSES_ROOT, keyPath); err != nil && err != registry.ErrNotExist {
			return fmt.Errorf("could not delete key for %s: %w", ext, err)
		}
	}
	return nil
}

func main() {
	cobra.MousetrapHelpText = ""
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(installCmd, uninstallCmd)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

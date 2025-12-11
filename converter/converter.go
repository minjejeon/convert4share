package converter

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Job defines a conversion task.
type Job struct{ Orig, Dest string }

// Config holds the configuration for the converters.
type Config struct {
	MagickBinary        string
	FfmpegBinary        string
	MaxSize             int
	HardwareAccelerator string
	FfmpegCustomArgs    string
}

// ProgressCallback is a function that reports progress percentage (0-100).
type ProgressCallback func(progress int)

// Magick runs the ImageMagick conversion command.
func (c *Config) Magick(orig, dest string) error {
	cmd := exec.Command(c.MagickBinary, orig, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("Running magick command: %s", cmd.String())
	return cmd.Run()
}

// Ffmpeg runs the FFmpeg conversion command with progress reporting.
func (c *Config) Ffmpeg(orig, dest string, onProgress ProgressCallback) error {
	args := []string{
		"-hide_banner",
		"-loglevel", "info", // Need info to see duration and stats
		"-stats", // Ensure stats are printed
		"-y",     // Overwrite output files without asking
	}

	scaleArg := fmt.Sprintf("scale='w=%d:h=%d:force_original_aspect_ratio=decrease'", c.MaxSize, c.MaxSize)

	// Set video codec based on configuration
	accelerator := strings.ToLower(c.HardwareAccelerator)
	switch accelerator {
	case "amd":
		log.Println("Using 'amd' hardware accelerator (h264_amf) from config.")
		args = append(args,
			"-i", orig,
			"-c:v", "h264_amf",
			"-vf", strings.Replace(scaleArg, "scale", "vpp_amf", 1),
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
	if c.FfmpegCustomArgs != "" {
		// Split the string by spaces to get individual arguments
		log.Printf("Adding custom ffmpeg arguments: %s", c.FfmpegCustomArgs)
		args = append(args, strings.Fields(c.FfmpegCustomArgs)...)
	}

	// Add audio codec and destination
	args = append(args,
		"-c:a", "aac",
		dest,
	)

	cmd := exec.Command(c.FfmpegBinary, args...)

	// We need to read stderr for ffmpeg output
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("could not get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start ffmpeg: %w", err)
	}

	// Parsing logic
	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanLines)

		var duration time.Duration
		durationRegex := regexp.MustCompile(`Duration: (\d{2}):(\d{2}):(\d{2})\.(\d{2})`)
		timeRegex := regexp.MustCompile(`time=(\d{2}):(\d{2}):(\d{2})\.(\d{2})`)

		for scanner.Scan() {
			line := scanner.Text()
			// log.Println("ffmpeg output:", line) // Debug

			// Parse Duration
			if duration == 0 {
				matches := durationRegex.FindStringSubmatch(line)
				if len(matches) == 5 {
					h, _ := strconv.Atoi(matches[1])
					m, _ := strconv.Atoi(matches[2])
					s, _ := strconv.Atoi(matches[3])
					ms, _ := strconv.Atoi(matches[4])
					duration = time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second + time.Duration(ms*10)*time.Millisecond
					log.Printf("Detected video duration: %s", duration)
				}
			}

			// Parse time= to calculate progress
			if duration > 0 {
				matches := timeRegex.FindStringSubmatch(line)
				if len(matches) == 5 {
					h, _ := strconv.Atoi(matches[1])
					m, _ := strconv.Atoi(matches[2])
					s, _ := strconv.Atoi(matches[3])
					ms, _ := strconv.Atoi(matches[4])
					currentTime := time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second + time.Duration(ms*10)*time.Millisecond

					progress := int((float64(currentTime) / float64(duration)) * 100)
					if progress > 100 {
						progress = 100
					}
					if onProgress != nil {
						onProgress(progress)
					}
				}
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg finished with error: %w", err)
	}
	return nil
}

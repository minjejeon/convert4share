package converter

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
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

// Magick runs the ImageMagick conversion command.
func (c *Config) Magick(orig, dest string) error {
	cmd := exec.Command(c.MagickBinary, orig, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("Running magick command: %s", cmd.String())
	return cmd.Run()
}

// Ffmpeg runs the FFmpeg conversion command.
func (c *Config) Ffmpeg(orig, dest string) error {
	args := []string{
		"-hide_banner",
		"-loglevel", "warning",
		"-stats",
		"-y", // Overwrite output files without asking
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
		// This handles multiple arguments in the string correctly.
		log.Printf("Adding custom ffmpeg arguments: %s", c.FfmpegCustomArgs)
		args = append(args, strings.Fields(c.FfmpegCustomArgs)...)
	}

	// Add audio codec and destination
	args = append(args,
		"-c:a", "aac",
		dest,
	)

	cmd := exec.Command(c.FfmpegBinary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("Running ffmpeg command: %s", cmd.String())
	return cmd.Run()
}

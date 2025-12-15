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
	"sync"
	"time"
)

type Job struct{ Orig, Dest string }

type Config struct {
	MagickBinary        string
	FfmpegBinary        string
	MaxSize             int
	HardwareAccelerator string
	FfmpegCustomArgs    string
	VideoQuality        string // "high", "medium", "low"
}

type ProgressCallback func(progress int)

var (
	durationRegex = regexp.MustCompile(`Duration: (\d{2}):(\d{2}):(\d{2})\.(\d{2})`)
	timeRegex     = regexp.MustCompile(`time=(\d{2}):(\d{2}):(\d{2})\.(\d{2})`)
)

func (c *Config) Magick(orig, dest string) error {
	cmd := exec.Command(c.MagickBinary, orig, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("Running magick command: %s", cmd.String())
	return cmd.Run()
}

func (c *Config) Ffmpeg(orig, dest string, onProgress ProgressCallback) error {
	args := []string{
		"-hide_banner",
		"-loglevel", "info",
		"-stats",
		"-y",
	}

	scaleArg := fmt.Sprintf("scale='w=%d:h=%d:force_original_aspect_ratio=decrease'", c.MaxSize, c.MaxSize)

	// Determine bitrate and preset based on VideoQuality
	var bitrate string
	var amdQuality string
	var nvidiaPreset string

	switch strings.ToLower(c.VideoQuality) {
	case "low":
		bitrate = "1M"
		amdQuality = "speed"
		nvidiaPreset = "fast"
	case "medium":
		bitrate = "2.5M"
		amdQuality = "balanced"
		nvidiaPreset = "medium"
	case "high":
		fallthrough
	default:
		bitrate = "5M"
		amdQuality = "quality"
		nvidiaPreset = "slow"
	}

	accelerator := strings.ToLower(c.HardwareAccelerator)
	switch accelerator {
	case "amd":
		log.Println("Using 'amd' hardware accelerator (h264_amf) from config.")
		args = append(args,
			"-i", orig,
			"-c:v", "h264_amf",
			"-b:v", bitrate,
			"-quality", amdQuality,
			"-vf", strings.Replace(scaleArg, "scale", "vpp_amf", 1),
		)
	case "nvidia":
		log.Println("Using 'nvidia' hardware accelerator (h264_nvenc) from config.")
		args = append(args,
			"-hwaccel", "cuda",
			"-i", orig,
			"-c:v", "h264_nvenc",
			"-preset", nvidiaPreset,
			"-b:v", bitrate,
			"-vf", scaleArg,
		)
	case "none", "":
		log.Println("Using software encoder (libx264).")
		args = append(args, "-i", orig, "-c:v", "libx264", "-vf", scaleArg)
	default:
		log.Printf("Unknown hardwareAccelerator '%s', falling back to software encoder (libx264).", accelerator)
		args = append(args, "-i", orig, "-c:v", "libx264", "-vf", scaleArg)
	}

	if c.FfmpegCustomArgs != "" {
		log.Printf("Adding custom ffmpeg arguments: %s", c.FfmpegCustomArgs)
		args = append(args, strings.Fields(c.FfmpegCustomArgs)...)
	}

	args = append(args,
		"-c:a", "aac",
		dest,
	)

	cmd := exec.Command(c.FfmpegBinary, args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("could not get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start ffmpeg: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanLines)

		var duration time.Duration

		for scanner.Scan() {
			line := scanner.Text()

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

	err = cmd.Wait()
	wg.Wait()
	if err != nil {
		return fmt.Errorf("ffmpeg finished with error: %w", err)
	}
	return nil
}

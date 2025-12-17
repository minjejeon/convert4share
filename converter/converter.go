package converter

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	durationRegex = regexp.MustCompile(`Duration: (\d+):(\d{2}):(\d{2})\.(\d{2})`)
	timeRegex     = regexp.MustCompile(`time=(\d+):(\d{2}):(\d{2})\.(\d{2})`)
)

func (c *Config) Magick(ctx context.Context, orig, dest string) error {
	cmd := prepareCommandContext(ctx, c.MagickBinary, orig, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("Running magick command: %s", cmd.String())
	return cmd.Run()
}

func (c *Config) Ffmpeg(ctx context.Context, orig, dest string, onProgress ProgressCallback) error {
	args := []string{
		"-hide_banner",
		"-loglevel", "info",
		"-stats",
		"-y",
	}

	scaleArg := fmt.Sprintf("scale='w=%d:h=%d:force_original_aspect_ratio=decrease'", c.MaxSize, c.MaxSize)

	// Determine bitrate and preset based on VideoQuality
	var bitrate string
	var maxBitrate string
	var bufSize string
	var amdQuality string
	var nvidiaPreset string

	switch strings.ToLower(c.VideoQuality) {
	case "low":
		bitrate = "1M"
		maxBitrate = "2M"
		bufSize = "2M"
		amdQuality = "speed"
		nvidiaPreset = "fast"
	case "medium":
		bitrate = "2.5M"
		maxBitrate = "5M"
		bufSize = "5M"
		amdQuality = "balanced"
		nvidiaPreset = "medium"
	case "high":
		fallthrough
	default:
		bitrate = "5M"
		maxBitrate = "10M"
		bufSize = "10M"
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

		// Recommended settings from https://github.com/GPUOpen-LibrariesAndSDKs/AMF/wiki/Recommended-FFmpeg-Encoder-Settings
		switch amdQuality {
		case "quality": // High
			args = append(args,
				"-rc", "vbr_peak",
				"-maxrate", maxBitrate,
				"-bufsize", bufSize,
				"-vbaq", "true",
				"-preencode", "true",
				"-high_motion_quality_boost_enable", "true",
				"-preanalysis", "true",
				"-pa_adaptive_mini_gop", "true",
				"-pa_lookahead_buffer_depth", "40",
				"-pa_taq_mode", "2",
				"-bf", "3",
			)
		case "balanced": // Medium
			args = append(args,
				"-rc", "vbr_peak",
				"-maxrate", maxBitrate,
				"-bufsize", bufSize,
				"-vbaq", "true",
				"-preencode", "true",
				"-high_motion_quality_boost_enable", "true",
				"-bf", "3",
			)
		case "speed": // Low
			// Keep it simple for speed
		}
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

	cmd := prepareCommandContext(ctx, c.FfmpegBinary, args...)

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
		scanner.Split(scanCR)

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

func (c *Config) GenerateThumbnail(inputFile string) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(inputFile))
	var cmd *exec.Cmd
	var stdout bytes.Buffer

	// Preview size: 200px width, aspect ratio preserved

	if ext == ".mov" || ext == ".mp4" || ext == ".mkv" || ext == ".avi" {
		// FFMPEG
		// -ss 00:00:00 -i input -vframes 1 -vf scale=200:-1 -f image2 -c:v mjpeg pipe:1
		args := []string{
			"-hide_banner",
			"-loglevel", "error",
			"-ss", "00:00:00",
			"-i", inputFile,
			"-vframes", "1",
			"-vf", "scale=200:-1",
			"-f", "image2",
			"-c:v", "mjpeg",
			"pipe:1",
		}
		cmd = prepareCommand(c.FfmpegBinary, args...)
	} else {
		// Magick
		// input[0] -resize 200x200 jpeg:-
		// Note: For HEIC, magick handles it if delegates are present.
		// We use input[0] to get the first frame/page.
		args := []string{
			inputFile + "[0]",
			"-resize", "200x200",
			"-quality", "80",
			"jpeg:-",
		}
		cmd = prepareCommand(c.MagickBinary, args...)
	}

	cmd.Stdout = &stdout
	// We can ignore stderr or capture it for debug
	// cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("thumbnail generation failed: %w", err)
	}

	return stdout.Bytes(), nil
}

func scanCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		return i + 1, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

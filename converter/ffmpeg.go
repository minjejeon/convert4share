package converter

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (c *Config) BuildFfmpegArgs(orig, dest string) []string {
	args := []string{
		"-hide_banner",
		"-loglevel", "info",
		"-stats",
		"-y",
	}

	scaleArg := fmt.Sprintf("scale='w=%d:h=%d:force_original_aspect_ratio=decrease'", c.MaxSize, c.MaxSize)

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

	return args
}

func (c *Config) Ffmpeg(ctx context.Context, orig, dest string, onProgress ProgressCallback) error {
	args := c.BuildFfmpegArgs(orig, dest)
	cmd := prepareCommandContext(ctx, c.FfmpegBinary, args...)

	// Ensure standard input is closed to prevent ffmpeg from waiting for input
	cmd.Stdin = nil

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("could not get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start ffmpeg: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	var stderrLog []string
	var stderrMu sync.Mutex

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		scanner.Split(scanCR)

		var duration time.Duration

		for scanner.Scan() {
			line := scanner.Text()

			stderrMu.Lock()
			stderrLog = append(stderrLog, line)
			if len(stderrLog) > 20 {
				stderrLog = stderrLog[1:]
			}
			stderrMu.Unlock()

			// Debug logging for ffmpeg output to diagnose hangs/errors
			// Only log lines that don't look like standard progress to avoid flooding logs too much.
			isProgress := timeRegex.MatchString(line)
			if !isProgress {
				log.Printf("ffmpeg: %s", line)
			}

			if duration == 0 {
				matches := durationRegex.FindStringSubmatch(line)
				if len(matches) == 5 {
					h, _ := strconv.Atoi(matches[1])
					m, _ := strconv.Atoi(matches[2])
					s, _ := strconv.Atoi(matches[3])

					fractionStr := matches[4]
					if len(fractionStr) > 9 {
						fractionStr = fractionStr[:9]
					}
					for len(fractionStr) < 9 {
						fractionStr += "0"
					}
					nanos, _ := strconv.Atoi(fractionStr)

					duration = time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second + time.Duration(nanos)*time.Nanosecond
					log.Printf("Detected video duration: %s", duration)
				}
			}

			if duration > 0 {
				matches := timeRegex.FindStringSubmatch(line)
				if len(matches) == 5 {
					h, _ := strconv.Atoi(matches[1])
					m, _ := strconv.Atoi(matches[2])
					s, _ := strconv.Atoi(matches[3])

					fractionStr := matches[4]
					if len(fractionStr) > 9 {
						fractionStr = fractionStr[:9]
					}
					for len(fractionStr) < 9 {
						fractionStr += "0"
					}
					nanos, _ := strconv.Atoi(fractionStr)

					currentTime := time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second + time.Duration(nanos)*time.Nanosecond

					progress := int((float64(currentTime) / float64(duration)) * 100)
					if progress > 100 {
						progress = 100
					}

					var speed string
					speedMatch := speedRegex.FindStringSubmatch(line)
					if len(speedMatch) > 1 {
						speed = speedMatch[1] + "x"
					}

					if onProgress != nil {
						onProgress(progress, speed)
					}
				}
			}
		}
	}()

	err = cmd.Wait()
	wg.Wait()
	if err != nil {
		stderrMu.Lock()
		logs := strings.Join(stderrLog, "\n")
		stderrMu.Unlock()
		return fmt.Errorf("ffmpeg finished with error: %w. Log: %s", err, logs)
	}
	return nil
}

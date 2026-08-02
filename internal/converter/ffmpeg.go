package converter

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"
)

// BuildFfmpegArgs assembles the encode command line. copyAudio
// stream-copies the audio track instead of re-encoding it, which the
// caller should set when the source audio is already AAC — running
// AAC through the encoder again is pure generation loss.
func (c *Config) BuildFfmpegArgs(orig, dest string, copyAudio bool) []string {
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
		slog.Info("Using 'amd' hardware accelerator (h264_amf) from config.")
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
		slog.Info("Using 'nvidia' hardware accelerator (h264_nvenc) from config.")
		// User reported success with software scale + format=yuv420p
		args = append(args,
			"-hwaccel", "cuda",
			"-i", orig,
			"-c:v", "h264_nvenc",
			"-preset", nvidiaPreset,
			"-b:v", bitrate,
			"-vf", scaleArg+",format=yuv420p",
		)
	case "vaapi":
		slog.Info("Using 'vaapi' hardware accelerator (h264_vaapi) from config.")
		vaapiFilter := fmt.Sprintf("format=nv12,hwupload,scale_vaapi=w=%d:h=%d:force_original_aspect_ratio=decrease", c.MaxSize, c.MaxSize)
		args = append(args,
			"-vaapi_device", "/dev/dri/renderD128",
			"-i", orig,
			"-vf", vaapiFilter,
			"-c:v", "h264_vaapi",
			"-b:v", bitrate,
			"-maxrate", maxBitrate,
		)
	case "none", "":
		slog.Info("Using software encoder (libx264).")
		args = append(args, "-i", orig, "-c:v", "libx264", "-vf", scaleArg)
	default:
		slog.Warn("Unknown hardwareAccelerator, falling back to software encoder (libx264).", "accelerator", accelerator)
		args = append(args, "-i", orig, "-c:v", "libx264", "-vf", scaleArg)
	}

	if c.FfmpegCustomArgs != "" {
		slog.Info("Adding custom ffmpeg arguments", "args", c.FfmpegCustomArgs)
		args = append(args, strings.Fields(c.FfmpegCustomArgs)...)
	}

	audioCodec := "aac"
	if copyAudio {
		audioCodec = "copy"
	}

	args = append(args,
		"-c:a", audioCodec,
		dest,
	)

	return args
}

func (c *Config) Ffmpeg(ctx context.Context, orig, dest string, copyAudio bool, onProgress ProgressCallback) error {
	args := c.BuildFfmpegArgs(orig, dest, copyAudio)
	cmd := prepareCommandContext(ctx, c.FfmpegBinary, args...)

	slog.Info("Launching ffmpeg", "command", cmd.String())

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
				slog.Debug("ffmpeg output", "line", line)
			}

			if duration == 0 {
				matches := durationRegex.FindStringSubmatch(line)
				if len(matches) == 5 {
					h, _ := strconv.Atoi(matches[1])
					m, _ := strconv.Atoi(matches[2])
					s, _ := strconv.Atoi(matches[3])

					nanos := parseFractionToNanos(matches[4])
					duration = time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second + time.Duration(nanos)*time.Nanosecond
					slog.Info("Detected video duration", "duration", duration)
				}
			}

			if duration > 0 {
				matches := timeRegex.FindStringSubmatch(line)
				if len(matches) == 5 {
					h, _ := strconv.Atoi(matches[1])
					m, _ := strconv.Atoi(matches[2])
					s, _ := strconv.Atoi(matches[3])

					nanos := parseFractionToNanos(matches[4])
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
		slog.Error("ffmpeg failed", "error", err, "command", cmd.String(), "last_logs", logs)
		return fmt.Errorf("ffmpeg finished with error: %w. Log: %s", err, logs)
	}
	return nil
}

// parseFractionToNanos converts a fractional second string (e.g. "50") to nanoseconds.
// It pads or truncates the string to 9 digits to represent nanoseconds.
// For example: "5" -> 500000000 (500ms), "123" -> 123000000 (123ms).
func parseFractionToNanos(s string) int {
	if len(s) > 9 {
		s = s[:9]
	} else if len(s) < 9 {
		s = s + strings.Repeat("0", 9-len(s))
	}
	nanos, _ := strconv.Atoi(s)
	return nanos
}

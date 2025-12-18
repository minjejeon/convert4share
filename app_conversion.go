package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/minjejeon/convert4share/converter"
	"github.com/spf13/viper"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) processPendingFiles() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isReady {
		logger.Info("processPendingFiles called but DOM not ready yet.")
		return
	}

	if len(a.pendingFiles) > 0 {
		logger.Info("Processing pending files", "files", a.pendingFiles)
		runtime.EventsEmit(a.ctx, "files-received", a.pendingFiles)
		a.pendingFiles = nil
	}
}

func (a *App) CancelJob(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if cancel, ok := a.jobCancels[id]; ok {
		cancel()
		delete(a.jobCancels, id)
	}
	a.pauseCond.Broadcast()
}

func (a *App) PauseQueue() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.isPaused = true
	runtime.EventsEmit(a.ctx, "queue-paused", true)
}

func (a *App) ResumeQueue() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.isPaused = false
	a.pauseCond.Broadcast()
	runtime.EventsEmit(a.ctx, "queue-resumed", true)
}

func (a *App) ConvertFiles(files []string) {
	go func() {
		var wg sync.WaitGroup
		convConfig := &converter.Config{
			MagickBinary:        viper.GetString("magickBinary"),
			FfmpegBinary:        viper.GetString("ffmpegBinary"),
			MaxSize:             viper.GetInt("maxSize"),
			HardwareAccelerator: viper.GetString("hardwareAccelerator"),
			FfmpegCustomArgs:    viper.GetString("ffmpegCustomArgs"),
			VideoQuality:        viper.GetString("videoQuality"),
		}
		collisionOption := viper.GetString("collisionOption")

		reporter := func(file string, destFile string, percent int, status string, errMsg string) {
			runtime.EventsEmit(a.ctx, "conversion-progress", JobStatus{
				ID:       file,
				File:     file,
				DestFile: destFile,
				Status:   status,
				Progress: percent,
				Error:    errMsg,
			})
		}

		a.mu.Lock()
		ffmpegSem := a.ffmpegSem
		magickSem := a.magickSem
		a.mu.Unlock()

		for _, f := range files {
			fpath := f

			if info, err := os.Stat(fpath); err != nil || info.IsDir() || info.Size() == 0 {
				reporter(fpath, "", 0, "error", "File is empty or invalid")
				continue
			}

			ext := strings.ToLower(filepath.Ext(fpath))

			fname := filepath.Base(fpath)
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

			wg.Add(1)
			go func(src string, extension string) {
				defer wg.Done()

				a.mu.Lock()
				jobCtx, cancel := context.WithCancel(a.ctx)
				a.jobCancels[src] = cancel

				for a.isPaused {
					if jobCtx.Err() != nil {
						delete(a.jobCancels, src)
						a.mu.Unlock()
						cancel()
						return
					}
					a.pauseCond.Wait()
				}
				if jobCtx.Err() != nil {
					delete(a.jobCancels, src)
					a.mu.Unlock()
					cancel()
					return
				}
				a.mu.Unlock()

				defer func() {
					cancel()
					a.mu.Lock()
					delete(a.jobCancels, src)
					a.mu.Unlock()
				}()

				var err error
				var dest string

				if extension == ".mov" {
					dest, err = a.resolveDestination(destDir, stem, ".mp4", collisionOption)
					if err != nil {
						reporter(src, "", 100, "error", err.Error())
						return
					}

					reporter(src, dest, 0, "processing", "")

					select {
					case ffmpegSem <- struct{}{}:
					case <-jobCtx.Done():
						return
					}
					defer func() { <-ffmpegSem }()

					err = convConfig.Ffmpeg(jobCtx, src, dest, func(progress int) {
						reporter(src, dest, progress, "processing", "")
					})
				} else if extension == ".heic" {
					dest, err = a.resolveDestination(destDir, stem, ".jpg", collisionOption)
					if err != nil {
						reporter(src, "", 100, "error", err.Error())
						return
					}

					reporter(src, dest, 0, "processing", "")

					select {
					case magickSem <- struct{}{}:
					case <-jobCtx.Done():
						return
					}
					defer func() { <-magickSem }()

					err = convConfig.Magick(jobCtx, src, dest)
				} else {
					reporter(src, "", 0, "error", "Unsupported format")
					return
				}

				if err != nil {
					if dest != "" {
						os.Remove(dest)
					}
					reporter(src, dest, 100, "error", err.Error())
				} else {
					reporter(src, dest, 100, "done", "")
				}
			}(fpath, ext)
		}

		wg.Wait()
		runtime.EventsEmit(a.ctx, "all-jobs-done", true)
	}()
}

func (a *App) resolveDestination(dir, name, ext, collisionOption string) (string, error) {
	a.destMu.Lock()
	defer a.destMu.Unlock()

	dest := filepath.Join(dir, name+ext)

	// Helper to create empty file atomically
	createPlaceholder := func(path string) bool {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL, 0666)
		if err == nil {
			f.Close()
			return true
		}
		return false
	}

	if collisionOption == "overwrite" {
		return dest, nil
	}

	info, err := os.Stat(dest)
	if os.IsNotExist(err) {
		if createPlaceholder(dest) {
			return dest, nil
		}
		// Refresh info if creation failed (race lost)
		info, err = os.Stat(dest)
	}

	// If file exists and is 0 bytes, overwrite it
	if err == nil && info != nil && info.Size() == 0 {
		return dest, nil
	}

	if collisionOption == "error" {
		return "", fmt.Errorf("file already exists: %s", dest)
	}

	// Default to "rename"
	for i := 1; ; i++ {
		d := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", name, i, ext))
		if createPlaceholder(d) {
			return d, nil
		}
	}
}

func (a *App) AddFiles(files []string) {
	logger.Info("AddFiles called", "files", files)
	for _, f := range files {
		if info, err := os.Stat(f); err == nil && !info.IsDir() && info.Size() > 0 {
			if absArg, err := filepath.Abs(f); err == nil {
				runtime.EventsEmit(a.ctx, "file-added", absArg)
			} else {
				runtime.EventsEmit(a.ctx, "file-added", f)
			}
		}
	}
}

func (a *App) GetThumbnail(path string) (string, error) {
	convConfig := &converter.Config{
		MagickBinary: viper.GetString("magickBinary"),
		FfmpegBinary: viper.GetString("ffmpegBinary"),
	}

	data, err := convConfig.GenerateThumbnail(path)
	if err != nil {
		logger.Error("Failed to generate thumbnail", "path", path, "ffmpeg", convConfig.FfmpegBinary, "magick", convConfig.MagickBinary, "error", err)
		return "", fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	base64Str := base64.StdEncoding.EncodeToString(data)
	return "data:image/jpeg;base64," + base64Str, nil
}

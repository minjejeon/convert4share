package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/minjejeon/convert4share/converter"
	"github.com/spf13/viper"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type JobStatus struct {
	ID       string `json:"id"`
	File     string `json:"file"`
	DestFile string `json:"destFile,omitempty"`
	Status   string `json:"status"` // "queued", "processing", "done", "error"
	Progress int    `json:"progress"`
	Speed    string `json:"speed,omitempty"`
	Error    string `json:"error,omitempty"`
}

func (a *App) processPendingFiles() {
	a.mu.Lock()
	if !a.isReady {
		a.mu.Unlock()
		logger.Info("processPendingFiles called but DOM not ready yet.")
		return
	}

	filesToProcess := a.pendingFiles
	a.pendingFiles = nil
	a.mu.Unlock()

	if len(filesToProcess) == 0 {
		return
	}

	go func(files []string) {
		var validFiles []string
		for _, f := range files {
			if info, err := os.Stat(f); err == nil && !info.IsDir() {
				if absArg, err := filepath.Abs(f); err == nil {
					validFiles = append(validFiles, absArg)
				} else {
					validFiles = append(validFiles, f)
				}
			} else {
				logger.Info("Skipping invalid file", "file", f)
			}
		}

		if len(validFiles) > 0 {
			logger.Info("Processing pending files", "files", validFiles)
			runtime.EventsEmit(a.ctx, "files-received", validFiles)
		}
	}(filesToProcess)
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

		reporter := func(file string, destFile string, percent int, status string, errMsg string, speed string) {
			runtime.EventsEmit(a.ctx, "conversion-progress", JobStatus{
				ID:       file,
				File:     file,
				DestFile: destFile,
				Status:   status,
				Progress: percent,
				Error:    errMsg,
				Speed:    speed,
			})
		}

		a.mu.Lock()
		ffmpegSem := a.ffmpegSem
		magickSem := a.magickSem
		a.mu.Unlock()

		for _, f := range files {
			// Trim surrounding quotes if present
			cleanPath := strings.Trim(f, "\"")
			jobID := cleanPath
			sysPath := cleanPath
			if abs, err := filepath.Abs(sysPath); err == nil {
				sysPath = abs
			}

			info, err := os.Stat(sysPath)
			if err != nil {
				if os.IsNotExist(err) {
					reporter(jobID, "", 0, "error", "File not found", "")
				} else {
					reporter(jobID, "", 0, "error", fmt.Sprintf("File access error: %s", err.Error()), "")
				}
				continue
			}
			if info.IsDir() {
				reporter(jobID, "", 0, "error", "Path is a directory", "")
				continue
			}
			if info.Size() == 0 {
				reporter(jobID, "", 0, "error", "File is empty (0 bytes)", "")
				continue
			}

			a.mu.Lock()
			if _, ok := a.jobCancels[jobID]; ok {
				a.mu.Unlock()
				logger.Warn("Skipping file as it is already being processed", "file", jobID)
				continue
			}
			a.mu.Unlock()

			ext := strings.ToLower(filepath.Ext(sysPath))

			fname := filepath.Base(sysPath)
			stem := strings.TrimSuffix(fname, filepath.Ext(fname))
			parent := filepath.Dir(sysPath)
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
			go func(id string, src string, extension string) {
				defer wg.Done()

				a.mu.Lock()
				jobCtx, cancel := context.WithCancel(a.ctx)
				a.jobCancels[id] = cancel

				for a.isPaused {
					if jobCtx.Err() != nil {
						delete(a.jobCancels, id)
						a.mu.Unlock()
						cancel()
						return
					}
					a.pauseCond.Wait()
				}
				if jobCtx.Err() != nil {
					delete(a.jobCancels, id)
					a.mu.Unlock()
					cancel()
					return
				}
				a.mu.Unlock()

				defer func() {
					cancel()
					a.mu.Lock()
					delete(a.jobCancels, id)
					a.mu.Unlock()
				}()

				var err error
				var dest string

				if extension == ".mov" {
					dest, err = a.resolveDestination(destDir, stem, ".mp4", collisionOption)
					if err != nil {
						reporter(id, "", 100, "error", err.Error(), "")
						return
					}

					reporter(id, dest, 0, "pending", "", "")

					select {
					case ffmpegSem <- struct{}{}:
						defer func() { <-ffmpegSem }()
					case <-jobCtx.Done():
						return
					}

					reporter(id, dest, 0, "processing", "", "")

					err = convConfig.Ffmpeg(jobCtx, src, dest, func(progress int, speed string) {
						reporter(id, dest, progress, "processing", "", speed)
					})
				} else if extension == ".heic" {
					dest, err = a.resolveDestination(destDir, stem, ".jpg", collisionOption)
					if err != nil {
						reporter(id, "", 100, "error", err.Error(), "")
						return
					}

					reporter(id, dest, 0, "pending", "", "")

					select {
					case magickSem <- struct{}{}:
						defer func() { <-magickSem }()
					case <-jobCtx.Done():
						return
					}

					reporter(id, dest, 0, "processing", "", "")

					err = convConfig.Magick(jobCtx, src, dest)
				} else {
					reporter(id, "", 0, "error", "Unsupported format", "")
					return
				}

				if err != nil {
					if dest != "" {
						os.Remove(dest)
					}
					reporter(id, dest, 100, "error", err.Error(), "")
				} else {
					reporter(id, dest, 100, "done", "", "")
				}
			}(jobID, sysPath, ext)
		}

		wg.Wait()
		runtime.EventsEmit(a.ctx, "all-jobs-done", true)
	}()
}

func (a *App) resolveDestination(dir, name, ext, collisionOption string) (string, error) {
	a.destMu.Lock()
	defer a.destMu.Unlock()

	dest := filepath.Join(dir, name+ext)

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

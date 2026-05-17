// Package jobs hosts the Wails v3 service that owns the conversion
// queue, job lifecycle, and progress events. It is the v3 successor to
// the App.AddFiles / App.ConvertFiles / App.CancelJob / App.PauseQueue
// / App.ResumeQueue methods from the v2 implementation.
//
// Settings changes are consumed via the application-wide
// `settings-changed` custom event (emitted by the SettingsService) so
// the two services stay decoupled.
package jobs

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/minjejeon/convert4share/internal/concurrency"
	"github.com/minjejeon/convert4share/internal/config"
	"github.com/minjejeon/convert4share/internal/converter"
	"github.com/minjejeon/convert4share/internal/livephoto"
)

// JobStatus mirrors the v2 progress payload so the existing frontend
// event handler keeps working once it switches to the v3 bindings.
type JobStatus struct {
	ID       string `json:"id"`
	File     string `json:"file"`
	DestFile string `json:"destFile,omitempty"`
	Status   string `json:"status"` // "queued", "processing", "done", "error"
	Progress int    `json:"progress"`
	Speed    string `json:"speed,omitempty"`
	Error    string `json:"error,omitempty"`
}

// Service owns the conversion queue plus its concurrency primitives.
// All exported methods are bound by Wails v3.
type Service struct {
	logger *slog.Logger
	app    *application.App

	mu         sync.Mutex
	destMu     sync.Mutex
	jobCancels map[string]context.CancelFunc
	isPaused   bool
	pauseCond  *sync.Cond

	ffmpegSem *concurrency.Sem
	magickSem *concurrency.Sem
}

// New constructs a Service with safe-default semaphore capacities (1
// each). The real capacities arrive from settings via
// applySettings(...) after ServiceStartup, or via the
// `settings-changed` event after a SaveSettings call.
func New(logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	s := &Service{
		logger:     logger,
		jobCancels: make(map[string]context.CancelFunc),
		// Initialize with capacity 1 so any acquire that happens before
		// the first settings load does not deadlock (BUG-001 lineage).
		ffmpegSem: concurrency.New(1),
		magickSem: concurrency.New(1),
	}
	s.pauseCond = sync.NewCond(&s.mu)
	return s
}

// ServiceStartup is called by Wails v3 once during application boot.
// We seize a reference to the app for event emission and context, then
// subscribe to the `settings-changed` event so semaphore capacities
// track user updates without a hard dependency on SettingsService.
func (s *Service) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.app = application.Get()

	// Initial sync from viper (defaults + config.yaml have already been
	// loaded by main.go before app.Run).
	s.applySettings()

	if s.app != nil {
		s.app.Event.On("settings-changed", func(*application.CustomEvent) {
			s.applySettings()
		})
	}
	return nil
}

// applySettings reads the current viper view and resizes the
// semaphores. Called both at startup and whenever SettingsService
// emits `settings-changed`.
func (s *Service) applySettings() {
	maxFfmpeg := viper.GetInt("maxFfmpegWorkers")
	if maxFfmpeg < 1 {
		maxFfmpeg = 1
	}
	maxMagick := viper.GetInt("maxMagickWorkers")
	if maxMagick < 1 {
		maxMagick = 1
	}
	s.ffmpegSem.Resize(maxFfmpeg)
	s.magickSem.Resize(maxMagick)
}

func (s *Service) emit(name string, data any) {
	if s.app == nil {
		return
	}
	s.app.Event.Emit(name, data)
}

// rootCtx returns the application's lifetime context, or
// context.Background when running outside an application (tests).
func (s *Service) rootCtx() context.Context {
	if s.app == nil {
		return context.Background()
	}
	return s.app.Context()
}

// AddFiles is the v3 equivalent of the v2 App.AddFiles: validates each
// path and emits a `file-added` event per accepted file.
func (s *Service) AddFiles(files []string) {
	s.logger.Info("AddFiles called", "files", files)
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil || info.IsDir() || info.Size() == 0 {
			continue
		}
		absArg, err := filepath.Abs(f)
		if err != nil {
			absArg = f
		}
		s.emit("file-added", absArg)
	}
}

// CancelJob cancels the in-flight conversion for the given job id (the
// absolute source path). If the queue is paused, the pause cond is
// broadcast so the cancelled goroutine can observe the cancellation.
func (s *Service) CancelJob(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cancel, ok := s.jobCancels[id]; ok {
		cancel()
		delete(s.jobCancels, id)
	}
	s.pauseCond.Broadcast()
}

// PauseQueue prevents any further jobs from leaving the pause loop
// until ResumeQueue is called. Already-running ffmpeg/magick processes
// continue to completion.
func (s *Service) PauseQueue() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isPaused = true
	s.emit("queue-paused", true)
}

// ResumeQueue clears the pause flag and broadcasts so all blocked
// workers re-check their state.
func (s *Service) ResumeQueue() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isPaused = false
	s.pauseCond.Broadcast()
	s.emit("queue-resumed", true)
}

// EmitFilesReceived fans out a `files-received` event with the given
// validated paths. Exposed for the main package to use when forwarding
// files from CLI args or second-instance launches; the legacy v2
// processPendingFiles path is gone.
func (s *Service) EmitFilesReceived(files []string) {
	if len(files) == 0 {
		return
	}
	s.emit("files-received", files)
}

// ExtractFileArgs filters a list of CLI arguments down to existing,
// non-empty regular files (resolved to absolute paths). The path
// matching exePath (case-insensitive) is skipped so we never feed the
// app its own binary as input — Windows sometimes hands it back in
// argv. exePath may be empty to disable that filter (tests).
func ExtractFileArgs(args []string, exePath string) []string {
	var out []string
	for _, arg := range args {
		if arg == "" {
			continue
		}
		clean := strings.Trim(arg, "\"")
		absArg, err := filepath.Abs(clean)
		if err != nil {
			absArg = clean
		}
		if exePath != "" {
			if strings.EqualFold(absArg, exePath) {
				continue
			}
		}
		info, err := os.Stat(absArg)
		if err != nil || info.IsDir() || info.Size() == 0 {
			continue
		}
		out = append(out, absArg)
	}
	return out
}

// ConvertFiles is the main entry point invoked by the frontend. It
// runs entirely in a background goroutine so the IPC call returns
// immediately; per-file progress is reported via `conversion-progress`
// events and the batch terminates with `all-jobs-done`.
func (s *Service) ConvertFiles(files []string) {
	go s.runBatch(files)
}

func (s *Service) runBatch(files []string) {
	convConfig := &converter.Config{
		MagickBinary:        viper.GetString("magickBinary"),
		FfmpegBinary:        viper.GetString("ffmpegBinary"),
		MaxSize:             viper.GetInt("maxSize"),
		MaxImageSize:        viper.GetInt("maxImageSize"),
		HardwareAccelerator: viper.GetString("hardwareAccelerator"),
		FfmpegCustomArgs:    viper.GetString("ffmpegCustomArgs"),
		VideoQuality:        viper.GetString("videoQuality"),
	}
	collisionOption := viper.GetString("collisionOption")
	autoLivePhoto := viper.GetBool("autoLivePhoto")
	copyOnlyExts := viper.GetStringSlice("copyOnlyExtensions")

	reporter := func(id, destFile string, percent int, status, errMsg, speed string) {
		s.emit("conversion-progress", JobStatus{
			ID:       id,
			File:     id,
			DestFile: destFile,
			Status:   status,
			Progress: percent,
			Error:    errMsg,
			Speed:    speed,
		})
	}

	var heicStems map[string]bool
	if autoLivePhoto {
		heicStems = livephoto.PairedHeicStems(files)
	}

	var wg sync.WaitGroup
	for _, f := range files {
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

		s.mu.Lock()
		if _, ok := s.jobCancels[jobID]; ok {
			s.mu.Unlock()
			s.logger.Warn("Skipping file as it is already being processed", "file", jobID)
			continue
		}
		s.mu.Unlock()

		ext := strings.ToLower(filepath.Ext(sysPath))
		fname := filepath.Base(sysPath)
		stem := strings.TrimSuffix(fname, filepath.Ext(fname))
		parent := filepath.Dir(sysPath)
		cleanedParent := filepath.Clean(parent)

		destDir := parent
		for _, pat := range config.ExcludePatterns() {
			cleanedPat := filepath.Clean(pat)
			if strings.Contains(cleanedParent, cleanedPat) {
				destDir = os.ExpandEnv(viper.GetString("defaultDestDir"))
				break
			}
		}

		wg.Add(1)
		go s.runOne(&wg, runOneInput{
			jobID:           jobID,
			src:             sysPath,
			ext:             ext,
			stem:            stem,
			parent:          parent,
			destDir:         destDir,
			collisionOption: collisionOption,
			autoLivePhoto:   autoLivePhoto,
			copyOnlyExts:    copyOnlyExts,
			heicStems:       heicStems,
			convConfig:      convConfig,
			report:          reporter,
		})
	}

	wg.Wait()
	s.emit("all-jobs-done", true)
}

type runOneInput struct {
	jobID           string
	src             string
	ext             string
	stem            string
	parent          string
	destDir         string
	collisionOption string
	autoLivePhoto   bool
	copyOnlyExts    []string
	heicStems       map[string]bool
	convConfig      *converter.Config
	report          func(id, dest string, percent int, status, errMsg, speed string)
}

func (s *Service) runOne(wg *sync.WaitGroup, in runOneInput) {
	defer wg.Done()

	s.mu.Lock()
	jobCtx, cancel := context.WithCancel(s.rootCtx())
	s.jobCancels[in.jobID] = cancel

	for s.isPaused {
		if jobCtx.Err() != nil {
			delete(s.jobCancels, in.jobID)
			s.mu.Unlock()
			cancel()
			return
		}
		s.pauseCond.Wait()
	}
	if jobCtx.Err() != nil {
		delete(s.jobCancels, in.jobID)
		s.mu.Unlock()
		cancel()
		return
	}
	s.mu.Unlock()

	defer func() {
		cancel()
		s.mu.Lock()
		delete(s.jobCancels, in.jobID)
		s.mu.Unlock()
	}()

	var (
		err  error
		dest string
	)

	isCopyOnly := false
	for _, copyExt := range in.copyOnlyExts {
		if strings.EqualFold(in.ext, copyExt) {
			isCopyOnly = true
			break
		}
	}

	switch {
	case isCopyOnly:
		dest, err = s.resolveDestination(in.destDir, in.stem, in.ext, in.collisionOption)
		if err != nil {
			in.report(in.jobID, "", 100, "error", err.Error(), "")
			return
		}
		in.report(in.jobID, dest, 0, "processing", "", "")
		err = copyFile(in.src, dest)

	case in.ext == ".mov":
		if in.autoLivePhoto && in.heicStems != nil {
			if _, ok := in.heicStems[filepath.Join(in.parent, in.stem)]; ok {
				s.logger.Info("Skipping .mov as it is a Live Photo (paired with .heic)", "file", in.src)
				in.report(in.jobID, "", 100, "done", "Skipped (Live Photo)", "")
				return
			}
		}

		dest, err = s.resolveDestination(in.destDir, in.stem, ".mp4", in.collisionOption)
		if err != nil {
			in.report(in.jobID, "", 100, "error", err.Error(), "")
			return
		}

		in.report(in.jobID, dest, 0, "pending", "", "")

		release, acqErr := s.ffmpegSem.Acquire(jobCtx)
		if acqErr != nil {
			return
		}
		defer release()

		in.report(in.jobID, dest, 0, "processing", "", "")
		err = in.convConfig.Ffmpeg(jobCtx, in.src, dest, func(progress int, speed string) {
			in.report(in.jobID, dest, progress, "processing", "", speed)
		})

	case in.ext == ".heic":
		dest, err = s.resolveDestination(in.destDir, in.stem, ".jpg", in.collisionOption)
		if err != nil {
			in.report(in.jobID, "", 100, "error", err.Error(), "")
			return
		}

		in.report(in.jobID, dest, 0, "pending", "", "")

		release, acqErr := s.magickSem.Acquire(jobCtx)
		if acqErr != nil {
			return
		}
		defer release()

		in.report(in.jobID, dest, 0, "processing", "", "")
		err = in.convConfig.Magick(jobCtx, in.src, dest)

	default:
		in.report(in.jobID, "", 0, "error", "Unsupported format", "")
		return
	}

	if err != nil {
		if dest != "" {
			os.Remove(dest)
		}
		in.report(in.jobID, dest, 100, "error", err.Error(), "")
		return
	}
	in.report(in.jobID, dest, 100, "done", "", "")
}

// resolveDestination picks a destination path that satisfies the
// collision policy. The destMu lock serialises placeholder creation so
// two concurrent goroutines cannot pick the same numbered suffix.
func (s *Service) resolveDestination(dir, name, ext, collisionOption string) (string, error) {
	s.destMu.Lock()
	defer s.destMu.Unlock()

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
		info, err = os.Stat(dest)
	}

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

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

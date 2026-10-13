package batch

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/minjejeon/convert4share/internal/concurrency"
	"github.com/minjejeon/convert4share/internal/converter"
	"github.com/minjejeon/convert4share/internal/livephoto"
)

// JobStatus describes the current status of an individual file conversion.
type JobStatus struct {
	ID       string `json:"id"`
	File     string `json:"file"`
	DestFile string `json:"destFile,omitempty"`
	Status   string `json:"status"` // "queued", "pending", "processing", "done", "error"
	Progress int    `json:"progress"`
	Speed    string `json:"speed,omitempty"`
	Error    string `json:"error,omitempty"`
}

// ProgressReporter is invoked whenever a job changes status or reports progress.
type ProgressReporter func(status JobStatus)

// Options holds configuration options for running a batch of files.
type Options struct {
	CollisionOption     string   // "rename", "overwrite", "error", "skip"
	AutoLivePhoto       bool     // skip .mov paired with .heic
	CopyOnlyExtensions  []string // e.g. [".jpg", ".jpeg", ".mp4"]
	DefaultDestDir      string   // destination dir when matching exclude patterns
	ExcludePatterns     []string // paths that trigger DefaultDestDir
	CustomDestDir       string   // explicit target dir overriding parent directory
	FFmpegConcurrency   int
	MagickConcurrency   int
}

// Runner coordinates concurrent file conversions with semaphores and collision handling.
type Runner struct {
	convConfig *converter.Config
	options    Options
	logger     *slog.Logger
	ffmpegSem  *concurrency.Sem
	magickSem  *concurrency.Sem
	destMu     sync.Mutex
}

// NewRunner creates a new batch Runner.
func NewRunner(cfg *converter.Config, opts Options, logger *slog.Logger) *Runner {
	if logger == nil {
		logger = slog.Default()
	}
	fCap := opts.FFmpegConcurrency
	if fCap <= 0 {
		fCap = 1
	}
	mCap := opts.MagickConcurrency
	if mCap <= 0 {
		mCap = 1
	}

	return &Runner{
		convConfig: cfg,
		options:    opts,
		logger:     logger,
		ffmpegSem:  concurrency.New(fCap),
		magickSem:  concurrency.New(mCap),
	}
}

// SetSemaphores adjusts the concurrency capacity.
func (r *Runner) SetSemaphores(ffmpegCap, magickCap int) {
	if ffmpegCap > 0 {
		r.ffmpegSem.Resize(ffmpegCap)
	}
	if magickCap > 0 {
		r.magickSem.Resize(magickCap)
	}
}

// ResolveDestination picks a collision-free destination file path.
// It creates an empty placeholder when necessary to guarantee uniqueness under concurrency.
func ResolveDestination(dir, name, ext, collisionOption string) (string, error) {
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
	if collisionOption == "skip" {
		return "", fmt.Errorf("skipped: %s already exists", dest)
	}

	for i := 1; ; i++ {
		d := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", name, i, ext))
		if createPlaceholder(d) {
			return d, nil
		}
	}
}

// Run executes the batch conversion synchronously for the given list of files,
// invoking the progress reporter on updates.
// Returns counts of (success, skipped, failed).
func (r *Runner) Run(ctx context.Context, files []string, reporter ProgressReporter) (int, int, int) {
	if reporter == nil {
		reporter = func(JobStatus) {}
	}

	var heicStems map[string]bool
	if r.options.AutoLivePhoto {
		heicStems = livephoto.PairedHeicStems(files)
	}

	var (
		successCount int32
		skipCount    int32
		errorCount   int32
		wg           sync.WaitGroup
	)

	report := func(id, destFile string, percent int, status, errMsg, speed string) {
		reporter(JobStatus{
			ID:       id,
			File:     id,
			DestFile: destFile,
			Status:   status,
			Progress: percent,
			Speed:    speed,
			Error:    errMsg,
		})
	}

	for _, f := range files {
		cleanPath := strings.Trim(f, "\"")
		jobID := cleanPath
		sysPath := cleanPath
		if abs, err := filepath.Abs(sysPath); err == nil {
			sysPath = abs
		}

		info, err := os.Stat(sysPath)
		if err != nil {
			atomic.AddInt32(&errorCount, 1)
			if os.IsNotExist(err) {
				report(jobID, "", 0, "error", "File not found", "")
			} else {
				report(jobID, "", 0, "error", fmt.Sprintf("File access error: %s", err.Error()), "")
			}
			continue
		}
		if info.IsDir() {
			atomic.AddInt32(&errorCount, 1)
			report(jobID, "", 0, "error", "Path is a directory", "")
			continue
		}
		if info.Size() == 0 {
			atomic.AddInt32(&errorCount, 1)
			report(jobID, "", 0, "error", "File is empty (0 bytes)", "")
			continue
		}

		ext := strings.ToLower(filepath.Ext(sysPath))
		fname := filepath.Base(sysPath)
		stem := strings.TrimSuffix(fname, filepath.Ext(fname))
		parent := filepath.Dir(sysPath)
		cleanedParent := filepath.Clean(parent)

		destDir := parent
		if r.options.CustomDestDir != "" {
			destDir = r.options.CustomDestDir
		} else {
			for _, pat := range r.options.ExcludePatterns {
				cleanedPat := filepath.Clean(pat)
				if strings.Contains(cleanedParent, cleanedPat) {
					destDir = os.ExpandEnv(r.options.DefaultDestDir)
					break
				}
			}
		}

		wg.Add(1)
		go func(src, jobExt, jobStem, jobParent, targetDir string) {
			defer wg.Done()

			if ctx.Err() != nil {
				atomic.AddInt32(&errorCount, 1)
				report(src, "", 0, "error", ctx.Err().Error(), "")
				return
			}

			isCopyOnly := false
			for _, copyExt := range r.options.CopyOnlyExtensions {
				if strings.EqualFold(jobExt, copyExt) {
					isCopyOnly = true
					break
				}
			}

			var (
				dest string
				dErr error
			)

			switch {
			case isCopyOnly:
				r.destMu.Lock()
				dest, dErr = ResolveDestination(targetDir, jobStem, jobExt, r.options.CollisionOption)
				r.destMu.Unlock()
				if dErr != nil {
					atomic.AddInt32(&errorCount, 1)
					report(src, "", 100, "error", dErr.Error(), "")
					return
				}
				report(src, dest, 0, "processing", "", "")
				if err := copyFile(src, dest); err != nil {
					if dest != "" {
						os.Remove(dest)
					}
					atomic.AddInt32(&errorCount, 1)
					report(src, dest, 100, "error", err.Error(), "")
					return
				}
				atomic.AddInt32(&successCount, 1)
				report(src, dest, 100, "done", "", "")

			case jobExt == ".mov":
				if r.options.AutoLivePhoto && heicStems != nil {
					if _, ok := heicStems[filepath.Join(jobParent, jobStem)]; ok {
						r.logger.Info("Skipping .mov as it is a Live Photo (paired with .heic)", "file", src)
						atomic.AddInt32(&skipCount, 1)
						report(src, "", 100, "done", "Skipped (Live Photo)", "")
						return
					}
				}

				r.destMu.Lock()
				dest, dErr = ResolveDestination(targetDir, jobStem, ".mp4", r.options.CollisionOption)
				r.destMu.Unlock()
				if dErr != nil {
					atomic.AddInt32(&errorCount, 1)
					report(src, "", 100, "error", dErr.Error(), "")
					return
				}

				report(src, dest, 0, "pending", "", "")

				release, acqErr := r.ffmpegSem.Acquire(ctx)
				if acqErr != nil {
					if dest != "" {
						os.Remove(dest)
					}
					atomic.AddInt32(&errorCount, 1)
					report(src, dest, 0, "error", acqErr.Error(), "")
					return
				}

				report(src, dest, 0, "processing", "", "")
				err = r.convConfig.Ffmpeg(ctx, src, dest, func(progress int, speed string) {
					report(src, dest, progress, "processing", "", speed)
				})
				release()

				if err != nil {
					if dest != "" {
						os.Remove(dest)
					}
					atomic.AddInt32(&errorCount, 1)
					report(src, dest, 100, "error", err.Error(), "")
					return
				}
				atomic.AddInt32(&successCount, 1)
				report(src, dest, 100, "done", "", "")

			case jobExt == ".heic":
				r.destMu.Lock()
				dest, dErr = ResolveDestination(targetDir, jobStem, ".jpg", r.options.CollisionOption)
				r.destMu.Unlock()
				if dErr != nil {
					atomic.AddInt32(&errorCount, 1)
					report(src, "", 100, "error", dErr.Error(), "")
					return
				}

				report(src, dest, 0, "pending", "", "")

				release, acqErr := r.magickSem.Acquire(ctx)
				if acqErr != nil {
					if dest != "" {
						os.Remove(dest)
					}
					atomic.AddInt32(&errorCount, 1)
					report(src, dest, 0, "error", acqErr.Error(), "")
					return
				}

				report(src, dest, 0, "processing", "", "")
				err = r.convConfig.Magick(ctx, src, dest)
				release()

				if err != nil {
					if dest != "" {
						os.Remove(dest)
					}
					atomic.AddInt32(&errorCount, 1)
					report(src, dest, 100, "error", err.Error(), "")
					return
				}
				atomic.AddInt32(&successCount, 1)
				report(src, dest, 100, "done", "", "")

			default:
				atomic.AddInt32(&errorCount, 1)
				report(src, "", 0, "error", "Unsupported format", "")
			}
		}(sysPath, ext, stem, parent, destDir)
	}

	wg.Wait()
	return int(successCount), int(skipCount), int(errorCount)
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

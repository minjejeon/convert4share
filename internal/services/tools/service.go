// Package tools hosts the Wails v3 service for ancillary helpers:
// file/binary selection dialogs, thumbnail generation, clipboard copy,
// winget installation, and binary auto-detection. It is the v3
// successor to the v2 App.SelectFiles / SelectBinaryDialog /
// GetThumbnail / CopyFileToClipboard / DetectBinaries / InstallTool
// methods.
//
// Phase 7 will reintroduce context-menu install/uninstall as a v3
// FileAssociations feature; the v2 InstallContextMenu /
// UninstallContextMenu / GetContextMenuStatus methods are deliberately
// not ported here.
package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"

	"github.com/spf13/viper"
	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/minjejeon/convert4share/internal/concurrency"
	"github.com/minjejeon/convert4share/internal/converter"
	platformwindows "github.com/minjejeon/convert4share/internal/platform/windows"
)

// Service is the v3 service binding for tool helpers.
type Service struct {
	logger   *slog.Logger
	app      *application.App
	thumbSem *concurrency.Sem
}

// New constructs a ToolsService. The thumbnail semaphore is fixed at
// capacity 1 to match the v2 behaviour (one thumbnail at a time keeps
// ffmpeg/magick spawn pressure low).
func New(logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		logger:   logger,
		thumbSem: concurrency.New(1),
	}
}

// ServiceStartup captures a reference to the running application, used
// only for the Dialog API and for the long-lived context passed to
// thumbnail generation.
func (s *Service) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.app = application.Get()
	return nil
}

func (s *Service) rootCtx() context.Context {
	if s.app == nil {
		return context.Background()
	}
	return s.app.Context()
}

// SelectFiles opens a multi-select file dialog filtered to the media
// formats Convert4Share understands. An empty slice (not an error) is
// returned if the user cancels.
func (s *Service) SelectFiles() ([]string, error) {
	if s.app == nil {
		return nil, fmt.Errorf("application not initialised")
	}
	dialog := s.app.Dialog.OpenFile().
		SetTitle("Select Files to Convert").
		AddFilter("Media Files", "*.mov;*.mp4;*.heic;*.png;*.jpg;*.jpeg").
		AddFilter("All Files", "*.*")
	files, err := dialog.PromptForMultipleSelection()
	if err != nil {
		return nil, err
	}
	if files == nil {
		return []string{}, nil
	}
	return files, nil
}

// SelectBinaryDialog opens a single-select file dialog filtered to
// executable extensions. Returns an empty string (not an error) if the
// user cancels.
func (s *Service) SelectBinaryDialog() (string, error) {
	if s.app == nil {
		return "", fmt.Errorf("application not initialised")
	}
	dialog := s.app.Dialog.OpenFile().
		SetTitle("Select Binary")
	if runtime.GOOS == "windows" {
		dialog.AddFilter("Executables", "*.exe;*.bat;*.cmd")
	}
	dialog.AddFilter("All Files", "*")
	return dialog.PromptForSingleSelection()
}

// CopyFileToClipboard places the given file on the system clipboard as
// a CF_HDROP-style file drop list (PowerShell on Windows; no-op on
// other platforms).
func (s *Service) CopyFileToClipboard(path string) error {
	return platformwindows.CopyFileToClipboard(path)
}

// DetectBinaries searches PATH (and platform-specific install
// locations) for ffmpeg / ffprobe / magick and returns a map keyed by
// tool name. On ImageMagick v6, where the binary is named 'convert'
// rather than 'magick', the 'convert' path is used for the "magick" key.
func (s *Service) DetectBinaries() map[string]string {
	results := make(map[string]string)

	if path, err := exec.LookPath("ffmpeg"); err == nil {
		results["ffmpeg"] = path
	}
	if path, err := exec.LookPath("ffprobe"); err == nil {
		results["ffprobe"] = path
	}
	if path, err := exec.LookPath("magick"); err == nil {
		results["magick"] = path
	} else if path, err := exec.LookPath("convert"); err == nil {
		results["magick"] = path
	}

	detectPlatformBinaries(results)

	return results
}

// InstallTool installs the requested tool via winget (ffmpeg ->
// Gyan.FFmpeg, magick -> ImageMagick.ImageMagick) and, if successful,
// persists the freshly detected binary path back into viper.
func (s *Service) InstallTool(toolName string) error {
	var packageID string
	switch toolName {
	case "ffmpeg":
		packageID = "Gyan.FFmpeg"
	case "magick":
		packageID = "ImageMagick.ImageMagick"
	default:
		return fmt.Errorf("unknown tool: %s", toolName)
	}

	if runtime.GOOS != "windows" {
		return fmt.Errorf("automatic install is not supported on this OS; install ffmpeg and ImageMagick with your package manager (e.g. 'sudo apt install ffmpeg imagemagick', 'sudo dnf install ffmpeg ImageMagick', or 'sudo pacman -S ffmpeg imagemagick')")
	}

	if err := platformwindows.InstallWingetPackage(packageID); err != nil {
		return err
	}

	detected := s.DetectBinaries()
	path, ok := detected[toolName]
	if !ok {
		return fmt.Errorf("installation completed but binary not found. You may need to restart the application")
	}

	switch toolName {
	case "ffmpeg":
		viper.Set("ffmpegBinary", path)
		// The Gyan.FFmpeg package carries ffprobe too; pick it up in
		// the same pass so mp4 probing works without a second install.
		if probePath, ok := detected["ffprobe"]; ok {
			viper.Set("ffprobeBinary", probePath)
		}
	case "magick":
		viper.Set("magickBinary", path)
	}

	// Notify other services (JobsService) so they re-read viper. The
	// SettingsService is the canonical settings-changed emitter, but
	// the binary auto-update happens outside its surface and still
	// needs to propagate.
	if s.app != nil {
		s.app.Event.Emit("settings-changed", nil)
	}
	return nil
}

// GetThumbnail generates a 200px-wide JPEG preview for the given file
// and returns it as a `data:image/jpeg;base64,...` URL. Concurrent
// invocations are serialised via thumbSem (capacity 1) to keep
// ffmpeg/magick spawn pressure low.
func (s *Service) GetThumbnail(path string) (string, error) {
	release, err := s.thumbSem.Acquire(s.rootCtx())
	if err != nil {
		return "", err
	}
	defer release()

	convConfig := &converter.Config{
		MagickBinary: viper.GetString("magickBinary"),
		FfmpegBinary: viper.GetString("ffmpegBinary"),
	}

	data, err := convConfig.GenerateThumbnail(s.rootCtx(), path)
	if err != nil {
		s.logger.Error("Failed to generate thumbnail",
			"path", path,
			"ffmpeg", convConfig.FfmpegBinary,
			"magick", convConfig.MagickBinary,
			"error", err,
		)
		return "", fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data), nil
}

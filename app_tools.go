package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/minjejeon/convert4share/converter"
	"github.com/minjejeon/convert4share/windows"
	"github.com/spf13/viper"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetContextMenuStatus() bool {
	return windows.IsContextMenuInstalled()
}

func (a *App) InstallContextMenu() error {
	return windows.RunCommandAsAdmin("install")
}

func (a *App) UninstallContextMenu() error {
	return windows.RunCommandAsAdmin("uninstall")
}

func (a *App) CopyFileToClipboard(path string) error {
	return windows.CopyFileToClipboard(path)
}

func (a *App) InstallTool(toolName string) error {
	var packageID string
	switch toolName {
	case "ffmpeg":
		packageID = "Gyan.FFmpeg"
	case "magick":
		packageID = "ImageMagick.ImageMagick"
	default:
		return fmt.Errorf("unknown tool: %s", toolName)
	}

	if err := windows.InstallWingetPackage(packageID); err != nil {
		return err
	}

	// Re-detect
	detected := a.DetectBinaries()
	if path, ok := detected[toolName]; ok {
		if toolName == "ffmpeg" {
			viper.Set("ffmpegBinary", path)
		} else if toolName == "magick" {
			viper.Set("magickBinary", path)
		}
		// Force save to persist
		return a.SaveSettings(a.GetSettings())
	} else {
		return fmt.Errorf("installation completed but binary not found. You may need to restart the application")
	}
}

func (a *App) SelectFiles() []string {
	selections, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Files to Convert",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Media Files",
				Pattern:     "*.mov;*.heic;*.png;*.jpg;*.jpeg",
			},
			{
				DisplayName: "All Files",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return []string{}
	}
	return selections
}

func (a *App) SelectBinaryDialog() string {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Binary",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Executables",
				Pattern:     "*.exe;*.bat;*.cmd",
			},
			{
				DisplayName: "All Files",
				Pattern:     "*",
			},
		},
	})
	if err != nil {
		return ""
	}
	return selection
}

func (a *App) DetectBinaries() map[string]string {
	results := make(map[string]string)

	if path, err := exec.LookPath("ffmpeg"); err == nil {
		results["ffmpeg"] = path
	}

	if path, err := exec.LookPath("magick"); err == nil {
		results["magick"] = path
	}

	exists := func(p string) bool {
		info, err := os.Stat(p)
		return err == nil && !info.IsDir()
	}

	// Check WinGet locations if not found
	home, err := os.UserHomeDir()
	if err == nil {
		localAppData := filepath.Join(home, "AppData", "Local")
		wingetBase := filepath.Join(localAppData, "Microsoft", "WinGet")

		// 1. Check Links (Symlinks)
		linksDir := filepath.Join(wingetBase, "Links")
		if _, ok := results["ffmpeg"]; !ok {
			if p := filepath.Join(linksDir, "ffmpeg.exe"); exists(p) {
				results["ffmpeg"] = p
			}
		}
		if _, ok := results["magick"]; !ok {
			if p := filepath.Join(linksDir, "magick.exe"); exists(p) {
				results["magick"] = p
			}
		}

		// 2. Check Packages (Actual installation dirs)
		packagesDir := filepath.Join(wingetBase, "Packages")
		entries, err := os.ReadDir(packagesDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				lowerName := strings.ToLower(entry.Name())

				findInDir := func(dir, binName string) string {
					var found string
					filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
						if err != nil {
							return nil
						}
						if !d.IsDir() && strings.EqualFold(d.Name(), binName) {
							found = path
							return filepath.SkipAll
						}
						return nil
					})
					return found
				}

				if _, ok := results["ffmpeg"]; !ok && strings.Contains(lowerName, "ffmpeg") {
					if p := findInDir(filepath.Join(packagesDir, entry.Name()), "ffmpeg.exe"); p != "" {
						results["ffmpeg"] = p
					}
				}

				if _, ok := results["magick"]; !ok && (strings.Contains(lowerName, "imagemagick") || strings.Contains(lowerName, "magick")) {
					if p := findInDir(filepath.Join(packagesDir, entry.Name()), "magick.exe"); p != "" {
						results["magick"] = p
					}
				}
			}
		}
	}

	return results
}

func (a *App) GetThumbnail(path string) (string, error) {
	convConfig := &converter.Config{
		MagickBinary: viper.GetString("magickBinary"),
		FfmpegBinary: viper.GetString("ffmpegBinary"),
	}

	data, err := convConfig.GenerateThumbnail(a.ctx, path)
	if err != nil {
		logger.Error("Failed to generate thumbnail", "path", path, "ffmpeg", convConfig.FfmpegBinary, "magick", convConfig.MagickBinary, "error", err)
		return "", fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	base64Str := base64.StdEncoding.EncodeToString(data)
	return "data:image/jpeg;base64," + base64Str, nil
}

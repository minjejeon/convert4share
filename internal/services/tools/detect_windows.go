//go:build windows

package tools

import (
	"os"
	"path/filepath"
	"strings"
)

// detectPlatformBinaries augments PATH-based detection with standard
// WinGet install locations (Links shims and unpacked Packages),
// filling only the tool keys not already resolved.
func detectPlatformBinaries(results map[string]string) {
	exists := func(p string) bool {
		info, err := os.Stat(p)
		return err == nil && !info.IsDir()
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	localAppData := filepath.Join(home, "AppData", "Local")
	wingetBase := filepath.Join(localAppData, "Microsoft", "WinGet")

	linksDir := filepath.Join(wingetBase, "Links")
	if _, ok := results["ffmpeg"]; !ok {
		if p := filepath.Join(linksDir, "ffmpeg.exe"); exists(p) {
			results["ffmpeg"] = p
		}
	}
	if _, ok := results["ffprobe"]; !ok {
		if p := filepath.Join(linksDir, "ffprobe.exe"); exists(p) {
			results["ffprobe"] = p
		}
	}
	if _, ok := results["magick"]; !ok {
		if p := filepath.Join(linksDir, "magick.exe"); exists(p) {
			results["magick"] = p
		}
	}

	packagesDir := filepath.Join(wingetBase, "Packages")
	entries, err := os.ReadDir(packagesDir)
	if err != nil {
		return
	}
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

		if strings.Contains(lowerName, "ffmpeg") {
			// ffprobe ships in the same package as ffmpeg.
			if _, ok := results["ffmpeg"]; !ok {
				if p := findInDir(filepath.Join(packagesDir, entry.Name()), "ffmpeg.exe"); p != "" {
					results["ffmpeg"] = p
				}
			}
			if _, ok := results["ffprobe"]; !ok {
				if p := findInDir(filepath.Join(packagesDir, entry.Name()), "ffprobe.exe"); p != "" {
					results["ffprobe"] = p
				}
			}
		}

		if _, ok := results["magick"]; !ok && (strings.Contains(lowerName, "imagemagick") || strings.Contains(lowerName, "magick")) {
			if p := findInDir(filepath.Join(packagesDir, entry.Name()), "magick.exe"); p != "" {
				results["magick"] = p
			}
		}
	}
}

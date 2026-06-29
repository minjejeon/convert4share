//go:build !windows

package tools

import "os"

// detectPlatformBinaries fills tool keys not already resolved via PATH
// by probing common install locations. Most installs land on PATH
// (/usr/bin, /usr/local/bin, /snap/bin, flatpak-exported bins), which
// exec.LookPath already covers; these are a minimal fallback.
func detectPlatformBinaries(results map[string]string) {
	exists := func(p string) bool {
		info, err := os.Stat(p)
		return err == nil && !info.IsDir()
	}

	probe := func(key string, candidates ...string) {
		if _, ok := results[key]; ok {
			return
		}
		for _, p := range candidates {
			if exists(p) {
				results[key] = p
				return
			}
		}
	}

	probe("ffmpeg", "/usr/bin/ffmpeg", "/usr/local/bin/ffmpeg")
	probe("magick", "/usr/bin/magick", "/usr/local/bin/magick", "/usr/bin/convert", "/usr/local/bin/convert")
}

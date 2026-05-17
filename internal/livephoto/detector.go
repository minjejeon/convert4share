// Package livephoto detects iOS Live Photo pairs in a list of file
// paths. A Live Photo on disk is a .heic still and a sidecar .mov with
// the same stem (filename without extension) in the same directory. The
// .mov component is essentially the still's motion clip and is usually
// not worth re-encoding on its own; callers use PairedHeicStems to
// decide whether to skip the .mov half.
package livephoto

import (
	"path/filepath"
	"strings"
)

// PairedHeicStems returns the set of "<dir>/<stem>" keys for every
// .heic file in paths. Paths surrounded by literal double quotes are
// trimmed before inspection (matching the legacy behavior).
//
// Callers detect a paired .mov by computing filepath.Join(parent, stem)
// for the .mov and checking membership in the returned map.
func PairedHeicStems(paths []string) map[string]bool {
	stems := make(map[string]bool)
	for _, f := range paths {
		cleanPath := strings.Trim(f, "\"")
		if strings.ToLower(filepath.Ext(cleanPath)) != ".heic" {
			continue
		}
		stem := strings.TrimSuffix(filepath.Base(cleanPath), filepath.Ext(cleanPath))
		dir := filepath.Dir(cleanPath)
		stems[filepath.Join(dir, stem)] = true
	}
	return stems
}

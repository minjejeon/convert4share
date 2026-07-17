package jobs

import "path/filepath"

// isPairedLivePhotoMov reports whether a .mov should be skipped because it is
// the video half of a Live Photo paired with a .heic. The lookup must use the
// ORIGINAL stem, since heicStems is keyed by original (pre-rename) stems.
func isPairedLivePhotoMov(autoLivePhoto bool, heicStems map[string]bool, parent, pairStem string) bool {
	if !autoLivePhoto || heicStems == nil {
		return false
	}
	_, ok := heicStems[filepath.Join(parent, pairStem)]
	return ok
}

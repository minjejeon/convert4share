//go:build !windows || !winrt

package share

import "log"

// CheckActivation checks if the app was activated via Share Target.
// It returns true if handled (files added to args), false otherwise.
func CheckActivation() bool {
	log.Println("Share Target activation check skipped (build without 'winrt' tag).")
	return false
}

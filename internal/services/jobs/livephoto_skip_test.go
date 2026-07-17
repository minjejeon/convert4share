package jobs

import (
	"path/filepath"
	"testing"
)

func TestIsPairedLivePhotoMov(t *testing.T) {
	const parent = `C:\photos`
	heicStems := map[string]bool{
		filepath.Join(parent, "IMG_1234"): true,
	}

	tests := []struct {
		name          string
		autoLivePhoto bool
		heicStems     map[string]bool
		pairStem      string
		want          bool
	}{
		{
			name:          "original stem matches",
			autoLivePhoto: true,
			heicStems:     heicStems,
			pairStem:      "IMG_1234",
			want:          true,
		},
		{
			name:          "rendered capture-time stem does not match",
			autoLivePhoto: true,
			heicStems:     heicStems,
			pairStem:      "26-07-05 10-22-43 1234",
			want:          false,
		},
		{
			name:          "autoLivePhoto disabled",
			autoLivePhoto: false,
			heicStems:     heicStems,
			pairStem:      "IMG_1234",
			want:          false,
		},
		{
			name:          "nil heicStems",
			autoLivePhoto: true,
			heicStems:     nil,
			pairStem:      "IMG_1234",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPairedLivePhotoMov(tt.autoLivePhoto, tt.heicStems, parent, tt.pairStem)
			if got != tt.want {
				t.Errorf("isPairedLivePhotoMov(%v, ..., %q, %q) = %v, want %v",
					tt.autoLivePhoto, parent, tt.pairStem, got, tt.want)
			}
		})
	}
}

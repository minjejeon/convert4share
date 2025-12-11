//go:build !windows

package windows

func RunAsAdmin() {
	// No-op on non-windows
}

func IsElevated() bool {
	return false
}

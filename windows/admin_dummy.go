//go:build wails2_legacy && !windows

package windows

func RunAsAdmin() {
	// No-op on non-windows
}

func IsElevated() bool {
	return false
}

func RunCommandAsAdmin(args string) error {
	return nil
}

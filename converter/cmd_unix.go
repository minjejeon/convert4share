//go:build !windows

package converter

import "os/exec"

func prepareCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

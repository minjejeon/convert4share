//go:build !windows

package converter

import (
	"context"
	"os/exec"
)

func prepareCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

func prepareCommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

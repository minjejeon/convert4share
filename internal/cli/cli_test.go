package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/minjejeon/convert4share/internal/cli"
)

func TestHelpFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.RunWithIO([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0 for --help, got %d", code)
	}
	out := stdout.String()
	if !strings.Contains(out, "Usage: c4s") {
		t.Errorf("expected usage info in stdout, got:\n%s", out)
	}
}

func TestVersionFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.RunWithIO([]string{"--version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0 for --version, got %d", code)
	}
	out := stdout.String()
	if !strings.Contains(out, "c4s (Convert4Share)") {
		t.Errorf("expected version info in stdout, got:\n%s", out)
	}
}

func TestNoFilesProvided(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.RunWithIO([]string{}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1 when no files provided, got %d", code)
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "no input files") {
		t.Errorf("expected 'no input files' error in stderr, got:\n%s", errOut)
	}
}

func TestUnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.RunWithIO([]string{"--nonexistent-flag"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1 for unknown flag, got %d", code)
	}
}

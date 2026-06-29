//go:build linux

package windows

import (
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
)

// CopyFileToClipboard copies the given file to the clipboard as a text/uri-list
// so it can be pasted into a file manager or chat. Uses wl-copy (Wayland) when
// available, otherwise xclip (X11).
func CopyFileToClipboard(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve absolute path: %w", err)
	}

	uri := (&url.URL{Scheme: "file", Path: abs}).String()
	uriList := uri + "\n"

	var cmd *exec.Cmd
	switch {
	case lookPath("wl-copy"):
		cmd = exec.Command("wl-copy", "--type", "text/uri-list")
	case lookPath("xclip"):
		cmd = exec.Command("xclip", "-selection", "clipboard", "-t", "text/uri-list")
	default:
		return fmt.Errorf("no clipboard tool found (install wl-clipboard or xclip)")
	}

	cmd.Stdin = strings.NewReader(uriList)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", cmd.Path, err)
	}
	return nil
}

func lookPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

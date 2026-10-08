//go:build !windows

package theme

import (
	"fmt"
	"os/exec"
)

// ReadClipboard returns the text currently on the clipboard.
func ReadClipboard() (string, error) {
	if path, err := exec.LookPath("pbpaste"); err == nil {
		out, err := exec.Command(path).Output()
		return string(out), err
	}
	if path, err := exec.LookPath("wl-paste"); err == nil {
		out, err := exec.Command(path, "--no-newline").Output()
		return string(out), err
	}
	if path, err := exec.LookPath("xclip"); err == nil {
		out, err := exec.Command(path, "-selection", "clipboard", "-o").Output()
		return string(out), err
	}
	if path, err := exec.LookPath("xsel"); err == nil {
		out, err := exec.Command(path, "--clipboard", "--output").Output()
		return string(out), err
	}
	return "", fmt.Errorf("no clipboard utility found (pbpaste, wl-paste, xclip, xsel)")
}

package theme

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type limitedBuffer struct{ buffer bytes.Buffer }

func (b *limitedBuffer) String() string { return b.buffer.String() }

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.buffer.Len()+len(p) > maxPaletteBytes {
		return 0, fmt.Errorf("output exceeds %d bytes", maxPaletteBytes)
	}
	return b.buffer.Write(p)
}

// fzf runs previews in a shell. Explicitly select one so inherited SHELL and
// FZF_DEFAULT_OPTS cannot change how our executable path is interpreted.
func previewCommand(exe string, args ...string) (string, string) {
	shell := "sh -c"
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
	prefix := ""
	if runtime.GOOS == "windows" {
		shell = "powershell.exe -NoProfile -NonInteractive -Command"
		if _, err := exec.LookPath("pwsh.exe"); err == nil {
			shell = "pwsh.exe -NoProfile -NonInteractive -Command"
		}
		quote = func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
		prefix = "& "
	}
	parts := []string{prefix + quote(exe)}
	for _, arg := range args {
		parts = append(parts, quote(arg))
	}
	return strings.Join(parts, " ") + " {}", shell
}

func runFZF(input, prompt, header, expect string, previewArgs ...string) (string, string, bool, error) {
	fzf, err := exec.LookPath("fzf")
	if err != nil {
		return "", "", false, fmt.Errorf("fzf is not installed: %w", err)
	}
	exe, err := os.Executable()
	if err != nil {
		return "", "", false, err
	}
	preview, shell := previewCommand(strings.TrimPrefix(exe, `\\?\`), previewArgs...)
	cmd := exec.Command(fzf, "--layout=reverse", "--prompt="+prompt, "--header="+header,
		"--info=inline", "--height=100%", "--expect="+expect, "--with-shell="+shell,
		"--preview="+preview, "--preview-window=right,62%,border-left")
	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr
	var output limitedBuffer
	cmd.Stdout = &output
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && (exit.ExitCode() == 130 || exit.ExitCode() == 1) {
			return "", "", true, nil
		}
		return "", "", false, fmt.Errorf("fzf failed: %w", err)
	}
	lines := strings.Split(strings.TrimRight(output.String(), "\r\n"), "\n")
	if len(lines) < 2 {
		return "", "", false, fmt.Errorf("fzf returned malformed selection")
	}
	key := strings.TrimSpace(lines[0])
	return key, strings.TrimSpace(lines[1]), key == "esc", nil
}

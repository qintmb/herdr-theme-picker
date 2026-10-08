//go:build windows

package theme

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ReadClipboard uses PowerShell's supported clipboard API instead of constructing
// an oversized unsafe slice over an allocation whose size is unknown.
func ReadClipboard() (string, error) {
	ps := "powershell.exe"
	if path, err := exec.LookPath("pwsh.exe"); err == nil {
		ps = path
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, ps, "-NoProfile", "-NonInteractive", "-Command",
		"$ErrorActionPreference='Stop'; [Console]::OutputEncoding=[System.Text.UTF8Encoding]::new($false); Get-Clipboard -Raw")
	var out limitedBuffer
	cmd.Stdout = &out
	var stderr limitedBuffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("read clipboard: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return strings.ReplaceAll(out.String(), "\r\n", "\n"), nil
}

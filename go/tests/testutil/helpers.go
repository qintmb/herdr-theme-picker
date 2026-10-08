package testutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"herdr-theme-picker/internal/terminal"
)

func WriteFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

type FZFStep struct {
	Selection string
	ExitCode  int
}

// The same test binary acts as a fake fzf or Herdr in subprocess tests. This
// avoids assuming a POSIX shell or using platform-specific executable scripts.
func Main(m *testing.M) {
	if len(os.Args) > 2 && os.Args[1] == "terminal-sync" {
		if err := terminal.RunHelper(os.Args[2]); err != nil {
			if errors.Is(err, terminal.ErrUnavailable) {
				os.Exit(3)
			}
			os.Exit(1)
		}
		os.Exit(0)
	}
	if os.Getenv("THEME_TEST_CONSOLE_CLIENT") == "1" {
		fmt.Fprintln(os.Stdout, "ready")
		io.Copy(io.Discard, os.Stdin)
		os.Exit(0)
	}
	if os.Getenv("THEME_TEST_ARGV") == "1" {
		json.NewEncoder(os.Stdout).Encode(os.Args[1:])
		os.Exit(0)
	}
	if script := os.Getenv("THEME_TEST_FZF_SCRIPT"); script != "" {
		exe, _ := os.Executable()
		if strings.TrimSuffix(filepath.Base(exe), ".exe") == "fzf" {
			data, err := os.ReadFile(script)
			if err != nil {
				os.Exit(2)
			}
			var steps []FZFStep
			if json.Unmarshal(data, &steps) != nil {
				os.Exit(2)
			}
			counter, _ := os.ReadFile(script + ".count")
			n, _ := strconv.Atoi(string(counter))
			if n >= len(steps) {
				os.Exit(2)
			}
			if os.WriteFile(script+".count", []byte(strconv.Itoa(n+1)), 0o600) != nil {
				os.Exit(2)
			}
			if steps[n].ExitCode != 0 {
				os.Exit(steps[n].ExitCode)
			}
			fmt.Print(steps[n].Selection)
			os.Exit(0)
		}
		if strings.TrimSuffix(filepath.Base(exe), ".exe") == "herdr" {
			fmt.Fprintln(os.Stderr, "simulated reload failure")
			os.Exit(2)
		}
	}
	cleanup := isolate()
	code := m.Run()
	cleanup()
	os.Exit(code)
}

// isolate points every config, state, and home lookup at a temporary tree and
// replaces herdr with a missing binary. A test that forgets to set a path can
// then never edit the user's config, reload a live session, update their
// Ghostty fragment, or recolor a real terminal (the client PID is invalid).
func isolate() func() {
	root, err := os.MkdirTemp("", "herdr-theme-test-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	for _, key := range []string{"HERDR_ENV", "HERDR_PLUGIN_ID", "HERDR_PLUGIN_CONTEXT_JSON", "HERDR_SOCKET_PATH", "HERDR_PLUGIN_ROOT"} {
		os.Unsetenv(key)
	}
	for key, dir := range map[string]string{
		"HOME": "home", "USERPROFILE": "home", "APPDATA": "appdata", "LOCALAPPDATA": "localappdata",
		"XDG_CONFIG_HOME": "xdg-config", "XDG_STATE_HOME": "xdg-state", "XDG_CACHE_HOME": "xdg-cache",
		"HERDR_PLUGIN_STATE_DIR": "state",
	} {
		os.Setenv(key, filepath.Join(root, dir))
	}
	os.Setenv("HERDR_CONFIG_PATH", filepath.Join(root, "config", "config.toml"))
	os.Setenv("HERDR_BIN_PATH", filepath.Join(root, "missing-herdr"))
	os.Setenv("HERDR_THEME_CLIENT_PID", "invalid")
	return func() { os.RemoveAll(root) }
}

func CopyExecutable(t *testing.T, path string) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o755); err != nil {
		t.Fatal(err)
	}
}

func FakeFZF(t *testing.T, steps ...FZFStep) string {
	t.Helper()
	dir := t.TempDir()
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	CopyExecutable(t, filepath.Join(dir, "fzf"+suffix))
	CopyExecutable(t, filepath.Join(dir, "herdr"+suffix))
	script := filepath.Join(dir, "steps.json")
	data, err := json.Marshal(steps)
	if err != nil {
		t.Fatal(err)
	}
	WriteFile(t, script, string(data))
	t.Setenv("THEME_TEST_FZF_SCRIPT", script)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HERDR_BIN_PATH", filepath.Join(dir, "herdr"+suffix))
	return script
}

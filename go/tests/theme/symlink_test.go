package theme_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "herdr-theme-picker/internal/theme"
)

func symlinkOrSkip(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
}

func assertSymlink(t *testing.T, path string) {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s was replaced by a regular file", path)
	}
}

func assertAccent(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "[theme.custom]") || !strings.Contains(string(data), `"#123456"`) {
		t.Fatalf("%s missing written block:\n%s", path, data)
	}
}

// Issue #3: dotfile setups symlink config.toml into the config dir; applying a
// theme must edit the target and leave the link in place.
func TestConfigSymlinkPreserved(t *testing.T) {
	tokens := map[string]string{"accent": "#123456"}

	t.Run("plain file", func(t *testing.T) {
		cfg := filepath.Join(t.TempDir(), "config.toml")
		writeTestFile(t, cfg, "[ui]\nfoo = 1\n")
		if err := WriteCustomBlock(cfg, tokens); err != nil {
			t.Fatal(err)
		}
		assertAccent(t, cfg)
	})

	t.Run("relative link", func(t *testing.T) {
		dir := t.TempDir()
		writeTestFile(t, filepath.Join(dir, "dotfiles", "herdr.toml"), "[ui]\nfoo = 1\n")
		cfg := filepath.Join(dir, "config.toml")
		symlinkOrSkip(t, filepath.Join("dotfiles", "herdr.toml"), cfg)
		if err := WriteCustomBlock(cfg, tokens); err != nil {
			t.Fatal(err)
		}
		assertSymlink(t, cfg)
		assertAccent(t, filepath.Join(dir, "dotfiles", "herdr.toml"))
	})

	t.Run("two hop chain", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "real.toml")
		writeTestFile(t, target, "[ui]\nfoo = 1\n")
		middle := filepath.Join(dir, "middle.toml")
		symlinkOrSkip(t, target, middle)
		cfg := filepath.Join(dir, "config.toml")
		symlinkOrSkip(t, middle, cfg)
		if err := WriteCustomBlock(cfg, tokens); err != nil {
			t.Fatal(err)
		}
		assertSymlink(t, cfg)
		assertSymlink(t, middle)
		assertAccent(t, target)
	})

	t.Run("dangling link", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "missing.toml")
		cfg := filepath.Join(dir, "config.toml")
		symlinkOrSkip(t, target, cfg)
		if err := WriteCustomBlock(cfg, tokens); err != nil {
			t.Fatal(err)
		}
		assertSymlink(t, cfg)
		assertAccent(t, target)
	})

	t.Run("loop", func(t *testing.T) {
		dir := t.TempDir()
		a := filepath.Join(dir, "a.toml")
		b := filepath.Join(dir, "b.toml")
		symlinkOrSkip(t, b, a)
		symlinkOrSkip(t, a, b)
		if err := WriteCustomBlock(a, tokens); err == nil {
			t.Fatal("expected an error for a symlink loop")
		}
		assertSymlink(t, a)
		assertSymlink(t, b)
	})
}

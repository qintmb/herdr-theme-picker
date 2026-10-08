package theme

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeleteGuardsAndConfirmation(t *testing.T) {
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	writeTestFile(t, filepath.Join(StateDir(), "keep.txt"), "keep")
	for _, slug := range []string{"../keep.txt", "dracula-default", "CON", "con", "com1", "-flag", strings.Repeat("a", 121)} {
		if err := deleteTheme(slug, strings.NewReader("y\n")); err == nil {
			t.Errorf("delete accepted %q", slug)
		}
		if err := EditTheme(slug); err == nil {
			t.Errorf("edit accepted %q", slug)
		}
	}
	if _, err := os.Stat(filepath.Join(StateDir(), "keep.txt")); err != nil {
		t.Fatal("guard removed unrelated file", err)
	}
	if err := saveUserTheme("custom", fixturePalette(t)); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, AppliedFile(), "custom\n")
	for _, answer := range []string{"n\n", "\n", ""} {
		if err := deleteTheme("custom", strings.NewReader(answer)); err != nil {
			t.Fatal(err)
		}
		if !IsUserTheme("custom") {
			t.Fatalf("deleted theme without affirmative confirmation %q", answer)
		}
	}
	if err := deleteTheme("custom", strings.NewReader("y\n")); err != nil {
		t.Fatal(err)
	}
	if IsUserTheme("custom") {
		t.Fatal("theme remains indexed")
	}
	if _, err := os.Stat(AppliedFile()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("active marker remains: %v", err)
	}
}

func TestUserThemeSymlinkGuard(t *testing.T) {
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	outside := filepath.Join(t.TempDir(), "outside")
	writeTestFile(t, outside, fixturePalette(t))
	if err := os.MkdirAll(UserThemesDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(UserThemesDir(), "linked")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	writeTestFile(t, UserIndexFile(), "linked\n")
	if err := deleteTheme("linked", strings.NewReader("y\n")); err == nil {
		t.Fatal("followed linked theme")
	}
	if err := saveUserTheme("linked", fixturePalette(t)); err == nil {
		t.Fatal("overwrote linked theme")
	}
	if _, err := ResolvePalette("linked"); err == nil {
		t.Fatal("resolved linked user theme")
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatal(err)
	}
}

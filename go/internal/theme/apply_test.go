package theme

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Herdr sets HERDR_ENV in every pane, including the plugin popup. Applying
// from inside Herdr is the primary local workflow and must keep working.
func TestApplyWorksInsideHerdrPane(t *testing.T) {
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	cfg := filepath.Join(t.TempDir(), "config.toml")
	t.Setenv("HERDR_CONFIG_PATH", cfg)
	t.Setenv("HERDR_ENV", "1")
	t.Setenv("HERDR_PLUGIN_ID", "herdr-theme-picker")
	t.Setenv("HERDR_PLUGIN_CONTEXT_JSON", "{}")
	if err := ApplyTheme("dracula-default"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(cfg)
	if err != nil || !strings.Contains(string(data), "[theme.custom]") {
		t.Fatalf("config not written: %q %v", data, err)
	}
	data, err = os.ReadFile(AppliedFile())
	if err != nil || string(data) != "dracula-default\n" {
		t.Fatalf("selection not saved: %q %v", data, err)
	}
}

// Ghostty reads the fragment only when the user includes it, so an existing
// file is the opt-in. Pane stdout never receives color setters.
func TestApplyUpdatesGhosttyFragmentOnlyWhenPresent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	t.Setenv("HERDR_CONFIG_PATH", filepath.Join(t.TempDir(), "config.toml"))
	fragment := filepath.Join(home, ".config", "ghostty", "herdr-theme")
	out, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	old := os.Stdout
	os.Stdout = out
	t.Cleanup(func() { os.Stdout = old })

	if err := ApplyTheme("dracula-default"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(fragment); !os.IsNotExist(err) {
		t.Fatalf("created a Ghostty fragment the user never set up: %v", err)
	}
	writeTestFile(t, fragment, "old colors\n")
	if err := ApplyTheme("dracula-default"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(fragment)
	if err != nil || !strings.Contains(string(data), "background = #282a36") {
		t.Fatalf("fragment not refreshed: %q %v", data, err)
	}
	data, err = os.ReadFile(out.Name())
	if err != nil || strings.ContainsRune(string(data), '\033') {
		t.Fatalf("apply wrote terminal control bytes to stdout: %q %v", data, err)
	}
}

func TestSyncRejectsStaleSelectionBeforeTerminalWrite(t *testing.T) {
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	cfg := filepath.Join(t.TempDir(), "config.toml")
	t.Setenv("HERDR_CONFIG_PATH", cfg)
	if err := ApplyTheme("dracula-default"); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, cfg, "[theme]\nname = 'nord'\n")
	if err := SyncAppliedTheme("invalid"); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("stale selection reached terminal sync: %v", err)
	}
}

// Auto-switch layers are tables inside [theme.custom]; the stale check must
// still read the base tokens.
func TestSyncAcceptsConfigWithAutoSwitchLayers(t *testing.T) {
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	cfg := filepath.Join(t.TempDir(), "config.toml")
	writeTestFile(t, cfg, "[theme.custom.dark]\npanel_bg = '#000000'\n")
	t.Setenv("HERDR_CONFIG_PATH", cfg)
	if err := ApplyTheme("dracula-default"); err != nil {
		t.Fatal(err)
	}
	if err := SyncAppliedTheme("invalid"); err == nil || strings.Contains(err.Error(), "does not match") || strings.Contains(err.Error(), "toml") {
		t.Fatalf("expected only the invalid test client to fail sync, got: %v", err)
	}
}

func TestMarkerWriteFailureIsReported(t *testing.T) {
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	t.Setenv("HERDR_CONFIG_PATH", filepath.Join(t.TempDir(), "config.toml"))
	if err := os.MkdirAll(AppliedFile(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := ApplyTheme("nord-default"); err == nil || !strings.Contains(err.Error(), "selection marker failed") {
		t.Fatalf("marker failure swallowed: %v", err)
	}
	if err := SyncAppliedTheme("invalid"); err == nil {
		t.Fatal("synced an unreadable selection")
	}
}

func TestLegacyNamesDoNotDisablePickerOrLoseIndexEntries(t *testing.T) {
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	writeTestFile(t, UserIndexFile(), "con\n../outside\n")
	// Windows cannot create con. Its entry must be retained for Unix recovery.
	if runtime.GOOS != "windows" {
		writeTestFile(t, filepath.Join(UserThemesDir(), "con"), fixturePalette(t))
	}
	items, err := pickerItems()
	if err != nil || !strings.Contains(items, "dracula-default") {
		t.Fatalf("legacy entry disabled picker: %v", err)
	}
	if runtime.GOOS != "windows" {
		if !strings.Contains(items, "★ con") {
			t.Fatal("existing Unix theme disappeared")
		}
		if _, err := ResolvePalette("con"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := ResolvePalette("../outside"); err == nil {
		t.Fatal("legacy handling allowed traversal")
	}
	if err := saveUserTheme("new-theme", fixturePalette(t)); err != nil {
		t.Fatal(err)
	}
	if err := deleteTheme("new-theme", strings.NewReader("y\n")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(UserIndexFile())
	if err != nil || string(data) != "con\n../outside\n" {
		t.Fatalf("legacy entries lost: %q %v", data, err)
	}
}

func TestEmptySelectionDoesNotDisablePicker(t *testing.T) {
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	writeTestFile(t, AppliedFile(), "  \n")
	if _, err := pickerItems(); err != nil {
		t.Fatal(err)
	}
}

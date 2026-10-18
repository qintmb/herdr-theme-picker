package theme_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "herdr-theme-picker/internal/theme"
)

// testConfig points the plugin at a throwaway config and state dir, matching
// how the other suites isolate a run.
func testConfig(t *testing.T) string {
	t.Helper()
	cfg := filepath.Join(t.TempDir(), "config.toml")
	t.Setenv("HERDR_CONFIG_PATH", cfg)
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	return cfg
}

// captureOutput swallows stdout and stderr for the test and returns a reader
// for what was written, so assertions stay independent of console noise.
func captureOutput(t *testing.T) func() string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatal(err)
	}
	originalOut, originalErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = f, f
	t.Cleanup(func() {
		os.Stdout, os.Stderr = originalOut, originalErr
		f.Close()
	})
	return func() string {
		data, err := os.ReadFile(f.Name())
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestRemoveCustomBlockRoundTrip(t *testing.T) {
	cfg := testConfig(t)
	original := "[keys]\nprefix = \"cmd+b\"\n\n[theme]\nname = \"nord\"\nauto_switch = false\n\n[ui]\nstatus_indicators = \"dots\"\n"
	writeTestFile(t, cfg, original)

	if err := WriteCustomBlock(cfg, map[string]string{"panel_bg": "#282828", "accent": "#458588"}); err != nil {
		t.Fatal(err)
	}
	if !HasCustomBlock(cfg) {
		t.Fatal("block not detected after write")
	}

	if err := RemoveCustomBlock(cfg); err != nil {
		t.Fatal(err)
	}
	if HasCustomBlock(cfg) {
		t.Fatal("block still detected after removal")
	}
	if got := readFile(t, cfg); got != original {
		t.Fatalf("round trip changed the file:\n%q\nwant:\n%q", got, original)
	}
}

func TestRemoveCustomBlockKeepsUserLayers(t *testing.T) {
	cfg := testConfig(t)
	writeTestFile(t, cfg, "[theme]\nname = \"nord\"\n\n[theme.custom]\npanel_bg = \"#111111\"\n\n[theme.custom.dark]\npanel_bg = \"#000000\"\n\n[ui]\nstatus_indicators = \"dots\"\n")

	if err := RemoveCustomBlock(cfg); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, cfg)
	if strings.Contains(got, "#111111") {
		t.Fatalf("plugin tokens survived:\n%s", got)
	}
	if !strings.Contains(got, "[theme.custom.dark]") || !strings.Contains(got, "#000000") {
		t.Fatalf("user-written layer removed:\n%s", got)
	}
	if !strings.Contains(got, "[ui]") || !strings.Contains(got, "status_indicators") {
		t.Fatalf("following section removed:\n%s", got)
	}
	if HasCustomBlock(cfg) {
		t.Fatal("a layers-only custom table is not a plugin block")
	}
	if !HasCustomLayers(cfg) {
		t.Fatal("HasCustomLayers did not report the dark layer")
	}
}

func TestRemoveCustomBlockIsIdempotent(t *testing.T) {
	cfg := testConfig(t)
	writeTestFile(t, cfg, "[theme]\nname = \"nord\"\n")
	if err := WriteCustomBlock(cfg, map[string]string{"panel_bg": "#282828"}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := RemoveCustomBlock(cfg); err != nil {
			t.Fatal(err)
		}
	}
	if got := readFile(t, cfg); got != "[theme]\nname = \"nord\"\n" {
		t.Fatalf("repeated removal changed the file: %q", got)
	}
	stray, err := filepath.Glob(filepath.Join(filepath.Dir(cfg), "*.tmp*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(stray) != 0 {
		t.Fatalf("stray temp files left behind: %v", stray)
	}
}

func TestRemoveCustomBlockLeavesMissingConfigAbsent(t *testing.T) {
	cfg := testConfig(t)
	if err := RemoveCustomBlock(cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cfg); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("reset created a config Herdr did not have")
	}
}

func TestBaseThemeName(t *testing.T) {
	dir := t.TempDir()
	cases := []struct{ label, body, want string }{
		{"explicit", "[theme]\nname = \"vesper\"\nauto_switch = false\n", "vesper"},
		{"unset", "[theme]\nauto_switch = false\n", DefaultThemeName},
		{"no theme table", "[ui]\nwidth = 30\n", DefaultThemeName},
		{"custom only", "[theme.custom]\npanel_bg = \"#282828\"\n", DefaultThemeName},
		{"invalid toml", "[theme\nname = \"x\"\n", DefaultThemeName},
	}
	for _, c := range cases {
		path := filepath.Join(dir, strings.ReplaceAll(c.label, " ", "-")+".toml")
		writeTestFile(t, path, c.body)
		if got := BaseThemeName(path); got != c.want {
			t.Errorf("BaseThemeName(%s) = %q, want %q", c.label, got, c.want)
		}
	}
	if got := BaseThemeName(filepath.Join(dir, "missing.toml")); got != DefaultThemeName {
		t.Errorf("missing config = %q, want %q", got, DefaultThemeName)
	}
}

func TestResetToBuiltinThemeClearsTokensAndMarker(t *testing.T) {
	cfg := testConfig(t)
	writeTestFile(t, cfg, "[theme]\nname = \"nord\"\n")
	if err := WriteCustomBlock(cfg, map[string]string{"panel_bg": "#282828"}); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, AppliedFile(), "dracula-default\n")
	captureOutput(t)

	if err := ResetToBuiltinTheme(); err != nil {
		t.Fatalf("reset failed: %v", err)
	}
	if HasCustomBlock(cfg) {
		t.Fatal("custom tokens survived the reset")
	}
	if _, err := os.Stat(AppliedFile()); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("selection marker survived the reset")
	}
	if got := readFile(t, cfg); got != "[theme]\nname = \"nord\"\n" {
		t.Fatalf("reset changed unrelated config: %q", got)
	}
}

func TestResetToBuiltinThemeWhenAlreadyClean(t *testing.T) {
	cfg := testConfig(t)
	writeTestFile(t, cfg, "[theme]\nname = \"vesper\"\n")
	out := captureOutput(t)

	if err := ResetToBuiltinTheme(); err != nil {
		t.Fatal(err)
	}
	if got := out(); !strings.Contains(got, `Already on the built-in theme "vesper"`) {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestResetToBuiltinThemeLeavesMissingConfigAbsent(t *testing.T) {
	cfg := testConfig(t)
	captureOutput(t)
	if err := ResetToBuiltinTheme(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cfg); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("reset created a config Herdr did not have")
	}
}

func TestPickerDefaultRowResets(t *testing.T) {
	cfg := testConfig(t)
	writeTestFile(t, cfg, "[theme]\nname = \"nord\"\n")
	if err := WriteCustomBlock(cfg, map[string]string{"panel_bg": "#282828"}); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, AppliedFile(), "dracula-default\n")
	captureOutput(t)
	// With tokens live the row is unmarked, so this is the exact string fzf
	// hands back when Enter lands on it.
	fakeFZF(t, fzfStep{Selection: "\n" + DefaultRow() + "\n"})

	if err := RunPicker(); err != nil {
		t.Fatalf("picking the default row failed: %v", err)
	}
	if HasCustomBlock(cfg) {
		t.Fatal("the picker's default row did not clear the custom block")
	}
	if _, err := os.Stat(AppliedFile()); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("the picker's default row did not clear the selection marker")
	}
}

func TestPickerDefaultRowIsMarkedWhenNoCustomBlock(t *testing.T) {
	cfg := testConfig(t)
	writeTestFile(t, cfg, "[theme]\nname = \"nord\"\n")
	out := captureOutput(t)
	// No tokens live, so the row carries the ✓ and Enter on it is a no-op.
	fakeFZF(t, fzfStep{Selection: "\n✓ " + DefaultRow() + "\n"})

	if err := RunPicker(); err != nil {
		t.Fatalf("picking the marked default row failed: %v", err)
	}
	if got := out(); !strings.Contains(got, `Already on the built-in theme "nord"`) {
		t.Fatalf("marked row did not report the built-in theme: %q", got)
	}
	if got := readFile(t, cfg); got != "[theme]\nname = \"nord\"\n" {
		t.Fatalf("a no-op reset changed the config: %q", got)
	}
}

func TestDefaultRowPreview(t *testing.T) {
	cfg := testConfig(t)
	writeTestFile(t, cfg, "[theme]\nname = \"vesper\"\n")
	got := DefaultRowPreview()
	if !strings.Contains(got, "vesper") {
		t.Fatalf("preview does not name the built-in theme: %q", got)
	}
	if strings.Contains(got, "custom.light/dark") {
		t.Fatalf("preview warns about layers that do not exist: %q", got)
	}
	writeTestFile(t, cfg, "[theme]\nname = \"vesper\"\n\n[theme.custom.dark]\npanel_bg = \"#000000\"\n")
	if got := DefaultRowPreview(); !strings.Contains(got, "custom.light/dark") {
		t.Fatalf("preview does not warn about the user's layer: %q", got)
	}
}

func TestResetOSCPayload(t *testing.T) {
	for _, seq := range []string{"\033]104\007", "\033]110\007", "\033]111\007"} {
		if !strings.Contains(ResetOSCPayload, seq) {
			t.Errorf("ResetOSCPayload missing %q: %q", seq, ResetOSCPayload)
		}
	}
}

func TestRestoreGhosttyFragments(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	frag := filepath.Join(home, ".config", "ghostty", "herdr-theme")
	writeTestFile(t, frag+".pre-herdr", "background = #111111\n")
	writeTestFile(t, frag, "background = #282828\n")

	if err := RestoreGhosttyFragments(); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, frag); got != "background = #111111\n" {
		t.Fatalf("fragment not restored: %q", got)
	}
	if _, err := os.Stat(frag + ".pre-herdr"); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("stash left behind after restore")
	}
}

func TestRestoreGhosttyFragmentsParksUnstashedFragment(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	frag := filepath.Join(home, ".config", "ghostty", "herdr-theme")
	writeTestFile(t, frag, "background = #282828\n")
	captureOutput(t)

	if err := RestoreGhosttyFragments(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(frag); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("a fragment with no stash kept recoloring new windows")
	}
	parked, err := filepath.Glob(frag + ".removed-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(parked) != 1 {
		t.Fatalf("expected exactly one parked fragment, got %v", parked)
	}
}

func TestRemoveCustomBlockLeavesNoGhosttyFragment(t *testing.T) {
	// A reset must not conjure a fragment the user never opted into.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	cfg := testConfig(t)
	writeTestFile(t, cfg, "[theme]\nname = \"nord\"\n")
	if err := WriteCustomBlock(cfg, map[string]string{"panel_bg": "#282828"}); err != nil {
		t.Fatal(err)
	}
	captureOutput(t)

	if err := ResetToBuiltinTheme(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "ghostty", "herdr-theme")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("reset created a Ghostty fragment")
	}
}

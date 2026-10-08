package theme_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "herdr-theme-picker/internal/theme"
)

func TestSlugify(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"Dracula Default", "dracula-default"},
		{"Rose Pine Moon", "rose-pine-moon"},
		{"Rosé_Pine 2", "ros-pine-2"},
		{"My_Awesome_Theme!!", "my-awesome-theme"},
		{"--leading-and-trailing--", "leading-and-trailing"},
		{"  Trim -- Me  ", "trim-me"},
		{"///", ""},
	}

	for _, c := range cases {
		got := Slugify(c.in)
		if got != c.want {
			t.Errorf("Slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDarkenHex(t *testing.T) {
	// #ffffff darkened 50% -> #7f7f7f
	got := DarkenHex("#ffffff", 50)
	if got != "#7f7f7f" {
		t.Errorf("DarkenHex(#ffffff, 50) = %q, want #7f7f7f", got)
	}

	// #000000 darkened 10% -> #000000
	got = DarkenHex("#000000", 10)
	if got != "#000000" {
		t.Errorf("DarkenHex(#000000, 10) = %q, want #000000", got)
	}
}

func TestParsePaletteAndTokens(t *testing.T) {
	content := `
background = #282a36
foreground = #f8f8f2
selection-background = #44475a
cursor-color = #f8f8f2
palette = 0=#21222c
palette = 1=#ff5555
palette = 2=#50fa7b
palette = 3=#f1fa8c
palette = 4=#bd93f9
palette = 5=#ff79c6
palette = 6=#8be9fd
palette = 7=#f8f8f2
palette = 8=#6272a4
palette = 9=#ff6e6e
palette = 10=#69ff94
palette = 11=#ffffa5
palette = 12=#d6acff
palette = 13=#ff92df
palette = 14=#a4ffff
palette = 15=#ffffff
# hpick-override: accent=#123456
`
	pal, err := ParsePaletteContent(content)
	if err != nil {
		t.Fatalf("ParsePaletteContent error: %v", err)
	}
	if pal.Background != "#282a36" {
		t.Errorf("Background = %q, want #282a36", pal.Background)
	}
	if pal.PaletteColors[0] != "#21222c" {
		t.Errorf("palette 0 = %q, want #21222c", pal.PaletteColors[0])
	}

	tokens, err := PaletteToTokens(pal)
	if err != nil {
		t.Fatalf("PaletteToTokens error: %v", err)
	}

	if tokens["panel_bg"] != "#282a36" {
		t.Errorf("panel_bg = %q, want #282a36", tokens["panel_bg"])
	}
	if tokens["text"] != "#f8f8f2" {
		t.Errorf("text = %q, want #f8f8f2", tokens["text"])
	}
	// Override check: accent should be #123456 instead of palette 4 (#bd93f9)
	if tokens["accent"] != "#123456" {
		t.Errorf("accent = %q, want #123456 (override)", tokens["accent"])
	}
}

func TestWriteCustomBlock(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := filepath.Join(tmpDir, "config.toml")

	initial := `
onboarding = false

[theme]
name = "catppuccin"

[ui]
sidebar_width = 30
`
	if err := os.WriteFile(cfg, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	tokens := map[string]string{
		"panel_bg": "#123456",
		"text":     "#abcdef",
	}

	if err := WriteCustomBlock(cfg, tokens); err != nil {
		t.Fatalf("WriteCustomBlock error: %v", err)
	}

	data, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)

	if !strings.Contains(out, "[theme.custom]") {
		t.Errorf("output does not contain [theme.custom]: %s", out)
	}
	if !strings.Contains(out, `panel_bg = "#123456"`) {
		t.Errorf("output missing panel_bg: %s", out)
	}
	if !strings.Contains(out, "onboarding = false") {
		t.Errorf("output lost original settings: %s", out)
	}
	if !strings.Contains(out, `name = "catppuccin"`) {
		t.Errorf("output lost theme.name: %s", out)
	}

	// Idempotent test: write again with updated tokens
	tokens2 := map[string]string{
		"panel_bg": "#999999",
		"text":     "#000000",
	}
	if err := WriteCustomBlock(cfg, tokens2); err != nil {
		t.Fatalf("WriteCustomBlock second write error: %v", err)
	}
	data2, _ := os.ReadFile(cfg)
	out2 := string(data2)
	count := strings.Count(out2, "[theme.custom]")
	if count != 1 {
		t.Errorf("expected 1 [theme.custom] block, got %d", count)
	}
	if !strings.Contains(out2, `panel_bg = "#999999"`) {
		t.Errorf("panel_bg not updated: %s", out2)
	}
}

func TestBundledThemes(t *testing.T) {
	root := PluginRoot()
	themesDir := filepath.Join(root, "themes")
	entries, err := os.ReadDir(themesDir)
	if err != nil {
		t.Fatalf("cannot read themes dir %s: %v", themesDir, err)
	}

	count := 0
	for _, e := range entries {
		if e.IsDir() || e.Name() == "index.txt" {
			continue
		}
		count++
		path := filepath.Join(themesDir, e.Name())
		pal, err := ParsePaletteFile(path)
		if err != nil {
			t.Errorf("failed to parse bundled theme %s: %v", e.Name(), err)
			continue
		}
		tokens, err := PaletteToTokens(pal)
		if err != nil {
			t.Errorf("failed to map tokens for bundled theme %s: %v", e.Name(), err)
			continue
		}
		if len(tokens) < 16 {
			t.Errorf("theme %s produced only %d tokens", e.Name(), len(tokens))
		}
	}

	if count < 15 {
		t.Errorf("expected at least 15 bundled themes, found %d", count)
	}

	indexFile := filepath.Join(themesDir, "index.txt")
	indexData, err := os.ReadFile(indexFile)
	if err != nil {
		t.Fatalf("cannot read themes/index.txt: %v", err)
	}
	if !strings.Contains(string(indexData), "dracula-default") {
		t.Errorf("themes/index.txt missing dracula-default")
	}
}

func TestResolvePaletteGuards(t *testing.T) {
	// 1. Valid bundled theme resolves
	path, err := ResolvePalette("dracula-default")
	if err != nil {
		t.Fatalf("ResolvePalette('dracula-default') failed: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("resolved path does not exist: %s", path)
	}

	// 2. Path traversal / invalid slug rejected
	invalidSlugs := []string{
		"../etc/passwd",
		"theme;rm",
		"theme with spaces",
		"UPPERCASE",
		"/root",
	}
	for _, slug := range invalidSlugs {
		if _, err := ResolvePalette(slug); err == nil {
			t.Errorf("ResolvePalette(%q) expected error, got nil", slug)
		}
	}
}

func TestRenderSwatch(t *testing.T) {
	output := RenderSwatch("dracula-default")
	if !strings.Contains(output, "dracula-default") {
		t.Errorf("RenderSwatch output does not contain theme name: %s", output)
	}
	// Verify truecolor ANSI escape sequences
	if !strings.Contains(output, "\033[48;2;") {
		t.Errorf("RenderSwatch output missing truecolor 48;2 ANSI escape")
	}
	// Verify UI labels
	for _, label := range []string{"Shell", "AGENT", "claude", "space1", "space2", "panel_bg"} {
		if !strings.Contains(output, label) {
			t.Errorf("RenderSwatch output missing expected label %q", label)
		}
	}
	// Verify box geometry characters
	for _, char := range []string{"┌", "┐", "└", "┘"} {
		if !strings.Contains(output, char) {
			t.Errorf("RenderSwatch output missing box character %q", char)
		}
	}
}

func TestUserThemeLifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HERDR_PLUGIN_STATE_DIR", tmpDir)

	slug := "test-custom-theme"
	userThemesDir := UserThemesDir()
	if err := os.MkdirAll(userThemesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Bundled theme should not be identified as user theme
	if IsUserTheme("dracula-default") {
		t.Errorf("bundled theme identified as user theme")
	}

	// Add dummy user theme
	themeContent := `
background = #123456
foreground = #abcdef
palette = 0=#123456
palette = 1=#ff0000
palette = 2=#00ff00
palette = 3=#ffff00
palette = 4=#0000ff
palette = 5=#ff00ff
palette = 6=#00ffff
palette = 7=#ffffff
palette = 8=#555555
palette = 9=#ff0000
palette = 10=#00ff00
palette = 11=#ffff00
palette = 12=#0000ff
palette = 13=#ff00ff
palette = 14=#00ffff
palette = 15=#ffffff
`
	themePath := filepath.Join(userThemesDir, slug)
	if err := os.WriteFile(themePath, []byte(themeContent), 0o644); err != nil {
		t.Fatal(err)
	}
	indexFile := UserIndexFile()
	if err := os.WriteFile(indexFile, []byte(slug+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Now IsUserTheme should be true
	if !IsUserTheme(slug) {
		t.Errorf("expected IsUserTheme(%q) to be true", slug)
	}

	// ResolvePalette should locate the user theme
	resolved, err := ResolvePalette(slug)
	if err != nil {
		t.Fatalf("ResolvePalette(%q) failed: %v", slug, err)
	}
	if resolved != themePath {
		t.Errorf("ResolvePalette(%q) = %q, want %q", slug, resolved, themePath)
	}

	// Delete user theme
	if err := deleteFixtureTheme(t, slug, "y\n"); err != nil {
		t.Fatalf("DeleteTheme(%q) failed: %v", slug, err)
	}

	// File and index should be gone
	if _, err := os.Stat(themePath); !os.IsNotExist(err) {
		t.Errorf("theme file still exists after DeleteTheme")
	}
	if IsUserTheme(slug) {
		t.Errorf("IsUserTheme(%q) is still true after DeleteTheme", slug)
	}
}

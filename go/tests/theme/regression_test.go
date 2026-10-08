package theme_test

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pelletier/go-toml/v2"

	. "herdr-theme-picker/internal/theme"
)

func TestCustomThemeOverridesBundledName(t *testing.T) {
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	body := strings.Replace(fixturePalette(t), "background = #282a36", "background = #123456", 1)
	if err := saveFixtureTheme(t, "dracula-default", body); err != nil {
		t.Fatal(err)
	}
	path, err := ResolvePalette("dracula-default")
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(UserThemesDir(), "dracula-default") {
		t.Fatal("bundled theme shadowed custom theme")
	}
	pal, err := ParsePaletteFile(path)
	if err != nil || pal.Background != "#123456" {
		t.Fatalf("wrong custom palette: %v", err)
	}
}

func TestConfigPreservesTOMLSyntax(t *testing.T) {
	cases := []string{
		"# keep\n[ui]\nmessage = '''\n[theme.custom]\nkeep this message\n'''\n",
		"onboarding = false\n[theme]\nname = 'catppuccin'\n[ 'theme' . \"custom\" ] # old tokens\naccent = '#ffffff'\n[ui]\nitems = [\n  '[theme.custom]',\n]\n",
		"theme.custom.accent = '#ffffff'\nonboarding = false\n",
		"[theme]\nname = 'catppuccin'\ncustom = { accent = '#ffffff' }\n[ui]\nwidth = 30\n",
		"theme = { name = 'catppuccin', custom = { accent = '#ffffff' } }\nonboarding = false\n[ui]\nwidth = 30\n",
		"# CRLF\r\n[theme.custom]\r\naccent = '#ffffff'\r\n[ui]\r\nwidth = 30\r\n",
	}
	for i, initial := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			cfg := filepath.Join(t.TempDir(), "config.toml")
			writeTestFile(t, cfg, initial)
			var before map[string]any
			if err := toml.Unmarshal([]byte(initial), &before); err != nil {
				t.Fatal(err)
			}
			for j := 0; j < 2; j++ {
				if err := WriteCustomBlock(cfg, map[string]string{"accent": "#123456"}); err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(cfg)
				if err != nil {
					t.Fatal(err)
				}
				var after map[string]any
				if err := toml.Unmarshal(data, &after); err != nil {
					t.Fatal(err)
				}
				if after["theme"].(map[string]any)["custom"].(map[string]any)["accent"] != "#123456" {
					t.Fatal("token did not update")
				}
				delete(after["theme"].(map[string]any), "custom")
				if theme, ok := before["theme"].(map[string]any); ok {
					delete(theme, "custom")
				} else {
					before["theme"] = map[string]any{}
				}
				if !reflect.DeepEqual(before, after) {
					t.Fatalf("unrelated TOML changed: before=%#v after=%#v", before, after)
				}
				if strings.Contains(initial, "# keep") && !bytes.Contains(data, []byte("# keep")) {
					t.Fatal("comment lost")
				}
				if strings.Contains(initial, "\r\n") && bytes.Contains(bytes.ReplaceAll(data, []byte("\r\n"), nil), []byte("\n")) {
					t.Fatal("CRLF changed")
				}
			}
			backup, err := os.ReadFile(cfg + ".bak-" + time.Now().Format("20060102"))
			if err != nil || string(backup) != initial {
				t.Fatalf("original backup not preserved: %v", err)
			}
		})
	}
}

// Herdr layers [theme.custom.light] / [theme.custom.dark] over the base tokens
// when auto_switch is on. They are user data; the picker owns only the base.
func TestConfigKeepsAutoSwitchLayers(t *testing.T) {
	want := map[string]any{
		"light": map[string]any{"panel_bg": "#eff1f5"},
		"dark":  map[string]any{"panel_bg": "#1e1e2e", "text": "#cdd6f4"},
	}
	cases := map[string]string{
		"tables":        "[theme]\nauto_switch = true\n\n[theme.custom]\naccent = '#ffffff'\n\n[theme.custom.light]\npanel_bg = '#eff1f5'\n\n[theme.custom.dark]\npanel_bg = '#1e1e2e'\ntext = '#cdd6f4'\n\n[keys]\nprefix = 'ctrl+b'\n",
		"dotted":        "[theme.custom]\naccent = '#ffffff'\nlight.panel_bg = '#eff1f5'\ndark.panel_bg = '#1e1e2e'\ndark.text = '#cdd6f4'\n[ui]\nwidth = 30\n",
		"inline-custom": "[theme]\ncustom = { accent = '#ffffff', light = { panel_bg = '#eff1f5' }, dark = { panel_bg = '#1e1e2e', text = '#cdd6f4' } }\n[ui]\nwidth = 30\n",
		"inline-theme":  "theme = { name = 'nord', custom = { light = { panel_bg = '#eff1f5' }, dark = { panel_bg = '#1e1e2e', text = '#cdd6f4' } } }\n",
		"root-dotted":   "theme.custom.light.panel_bg = '#eff1f5'\ntheme.custom.dark.panel_bg = '#1e1e2e'\ntheme.custom.dark.text = '#cdd6f4'\n",
	}
	for name, initial := range cases {
		t.Run(name, func(t *testing.T) {
			cfg := filepath.Join(t.TempDir(), "config.toml")
			writeTestFile(t, cfg, initial)
			for i := 0; i < 2; i++ {
				if err := WriteCustomBlock(cfg, map[string]string{"accent": "#123456"}); err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(cfg)
				if err != nil {
					t.Fatal(err)
				}
				var doc map[string]any
				if err := toml.Unmarshal(data, &doc); err != nil {
					t.Fatalf("invalid TOML: %v\n%s", err, data)
				}
				custom := doc["theme"].(map[string]any)["custom"].(map[string]any)
				if custom["accent"] != "#123456" {
					t.Fatalf("token not updated:\n%s", data)
				}
				if !reflect.DeepEqual(map[string]any{"light": custom["light"], "dark": custom["dark"]}, want) {
					t.Fatalf("auto_switch layers changed:\n%s", data)
				}
			}
		})
	}
}

func TestRepeatedWritesDoNotGrowConfig(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "config.toml")
	writeTestFile(t, cfg, "[theme]\nauto_switch = true\n\n[keys]\nprefix = 'ctrl+b'\n")
	var first []byte
	for i := 0; i < 3; i++ {
		if err := WriteCustomBlock(cfg, map[string]string{"accent": "#123456"}); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = data
		} else if !bytes.Equal(first, data) {
			t.Fatalf("write %d changed an already-applied config:\n%q\n%q", i+1, first, data)
		}
	}
}

func TestConfigRejectsUnsafeWrites(t *testing.T) {
	for _, kind := range []string{"invalid-toml", "invalid-token", "invalid-color", "backup-directory"} {
		t.Run(kind, func(t *testing.T) {
			cfg := filepath.Join(t.TempDir(), "config.toml")
			body := "[ui]\nwidth = 30\n"
			tokens := map[string]string{"accent": "#123456"}
			if kind == "invalid-toml" {
				body = "[ui\n"
			}
			if kind == "invalid-token" {
				tokens["unknown"] = "#123456"
			}
			if kind == "invalid-color" {
				tokens["accent"] = "#123"
			}
			writeTestFile(t, cfg, body)
			if kind == "backup-directory" {
				if err := os.Mkdir(cfg+".bak-"+time.Now().Format("20060102"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if err := WriteCustomBlock(cfg, tokens); err == nil {
				t.Fatal("expected failure")
			}
			got, err := os.ReadFile(cfg)
			if err != nil || string(got) != body {
				t.Fatal("failed operation modified config")
			}
		})
	}
}

func TestPaletteRejectsMalformedInput(t *testing.T) {
	valid := fixturePalette(t)
	cases := []string{
		"palette = 0=#123\n",
		strings.Replace(valid, "background = #282a36\n", "", 1),
		strings.Replace(valid, "foreground = #f8f8f2\n", "", 1),
		strings.Replace(valid, "palette = 15=#ffffff\n", "", 1),
		strings.Replace(valid, "#282a36", "#123xyz", 1),
		valid + "palette = 16=#123456\n",
		valid + "palette = 999999999999999999999=#123456\n",
		valid + "# hpick-override: unknown=#123456\n",
		valid + "# hpick-override: accent=#123456junk\n",
		strings.Repeat("x", (1<<20)+1),
	}
	for i, input := range cases {
		if _, err := ParsePaletteContent(input); err == nil {
			t.Errorf("accepted malformed case %d", i)
		}
	}
	pal, err := ParsePaletteContent("\ufeff" + strings.ReplaceAll(valid, "#282a36", "282A36"))
	if err != nil || pal.Background != "#282a36" {
		t.Fatalf("normalization failed: %v", err)
	}
	if _, err := PaletteToTokens(nil); err == nil {
		t.Fatal("nil palette accepted")
	}
}

func TestEditorDiscardLeavesOriginal(t *testing.T) {
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	body := fixturePalette(t)
	if err := saveFixtureTheme(t, "editable", body); err != nil {
		t.Fatal(err)
	}
	fakeFZF(t, fzfStep{Selection: "\naccent #123456\n"}, fzfStep{Selection: "esc\naccent #123456\n"})
	stdin, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}
	defer stdin.Close()
	if _, err := stdin.WriteString("#654321\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := stdin.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	originalStdin := os.Stdin
	os.Stdin = stdin
	t.Cleanup(func() { os.Stdin = originalStdin })
	if err := EditTheme("editable"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(UserThemesDir(), "editable"))
	if err != nil || string(got) != body {
		t.Fatal("discard changed original theme")
	}
}

func TestEditorSaveCommitsWithoutServerReload(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	if err := saveFixtureTheme(t, "editable", fixturePalette(t)); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(t.TempDir(), "config.toml")
	writeTestFile(t, cfg, "[ui]\nwidth = 30\n")
	t.Setenv("HERDR_CONFIG_PATH", cfg)
	t.Setenv("HERDR_THEME_CLIENT_PID", "invalid")
	fakeFZF(t, fzfStep{Selection: "\naccent #123456\n"}, fzfStep{Selection: "\n" + "✓ Save & apply" + "\n"})
	stdin, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}
	defer stdin.Close()
	if _, err := stdin.WriteString("#654321\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := stdin.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	stdout, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	defer stdout.Close()
	originalStdin, originalStdout := os.Stdin, os.Stdout
	os.Stdin, os.Stdout = stdin, stdout
	t.Cleanup(func() { os.Stdin, os.Stdout = originalStdin, originalStdout })
	if err := EditTheme("editable"); err != nil {
		t.Fatalf("local save depended on a server reload: %v", err)
	}
	pal, err := ParsePaletteFile(filepath.Join(UserThemesDir(), "editable"))
	if err != nil || pal.Overrides["accent"] != "#654321" {
		t.Fatal("Save did not commit override", err)
	}
	data, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := toml.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if document["theme"].(map[string]any)["custom"].(map[string]any)["accent"] != "#654321" {
		t.Fatal("Save did not apply updated tokens")
	}
}

func TestApplySavesWithoutServerReload(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	cfg := filepath.Join(t.TempDir(), "config.toml")
	writeTestFile(t, cfg, "[ui]\nwidth = 30\n")
	t.Setenv("HERDR_CONFIG_PATH", cfg)
	t.Setenv("HERDR_THEME_CLIENT_PID", "invalid") // Never emit into a real terminal during tests.
	fakeFZF(t)
	stdout, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	defer stdout.Close()
	originalStdout := os.Stdout
	os.Stdout = stdout
	t.Cleanup(func() { os.Stdout = originalStdout })
	err = ApplyTheme("dracula-default")
	if err != nil {
		t.Fatalf("local save depended on a server reload: %v", err)
	}
	if data, err := os.ReadFile(AppliedFile()); err != nil || string(data) != "dracula-default\n" {
		t.Fatalf("saved config and selection diverged: %q %v", data, err)
	}
}

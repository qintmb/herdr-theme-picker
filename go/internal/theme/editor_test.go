package theme

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestEditorArguments(t *testing.T) {
	for command, want := range map[string][]string{
		"code --wait": {"code", "--wait"},
		`"C:\Program Files\Editor\edit.exe" --wait`: {`C:\Program Files\Editor\edit.exe`, "--wait"},
		`editor --flag 'two words'`:                 {"editor", "--flag", "two words"},
		`editor --flag ""`:                          {"editor", "--flag", ""},
	} {
		got, err := splitEditor(command)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("%q: got %v, err=%v", command, got, err)
		}
	}
	if _, err := splitEditor(`editor "unclosed`); err == nil {
		t.Fatal("unclosed quote accepted")
	}
	t.Setenv("VISUAL", "code --wait")
	if PickEditor() != "code --wait" {
		t.Fatal("explicit editor preference ignored")
	}
}

func TestOverrideValidation(t *testing.T) {
	body := fixturePalette(t) + "# hpick-override: accent=#abcdef\n"
	updated, err := setOverride(body, "accent", "123456")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(updated, "hpick-override: accent=") != 1 {
		t.Fatal("duplicate override")
	}
	pal, err := ParsePaletteContent(updated)
	if err != nil || pal.Overrides["accent"] != "#123456" {
		t.Fatal("override did not replace", err)
	}
	for _, color := range []string{"#123", "#123456junk", "red"} {
		if _, err := setOverride(body, "accent", color); err == nil {
			t.Error("invalid color accepted", color)
		}
	}
}

func TestWindowsEditorShim(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows editor shim")
	}
	path := filepath.Join(t.TempDir(), "editor with spaces.cmd")
	writeTestFile(t, path, "@echo off\r\necho %~1\r\necho %~2\r\n")
	t.Setenv("VISUAL", fmt.Sprintf("%q --wait", path))
	file := filepath.Join(t.TempDir(), "palette with spaces.txt")
	cmd, err := editorCommand(file)
	if err != nil {
		t.Fatal(err)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("editor shim failed: %v %s", err, out)
	}
	if strings.TrimSpace(strings.ReplaceAll(string(out), "\r\n", "\n")) != "--wait\n"+file {
		t.Fatalf("editor args changed: %q", out)
	}
}

package theme_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "herdr-theme-picker/internal/theme"
	"herdr-theme-picker/tests/testutil"
)

func TestMain(m *testing.M) { testutil.Main(m) }

var writeTestFile = testutil.WriteFile
var fakeFZF = testutil.FakeFZF

type fzfStep = testutil.FZFStep

func fixturePalette(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(PluginRoot(), "themes", "dracula-default"))
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}

func saveFixtureTheme(t *testing.T, slug, body string) error {
	t.Helper()
	writeTestFile(t, filepath.Join(UserThemesDir(), slug), body)
	writeTestFile(t, UserIndexFile(), slug+"\n")
	return nil
}
func deleteFixtureTheme(t *testing.T, slug, answer string) error {
	t.Helper()
	input, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	if _, err := input.WriteString(answer); err != nil {
		t.Fatal(err)
	}
	if _, err := input.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	original := os.Stdin
	os.Stdin = input
	defer func() { os.Stdin = original }()
	return DeleteTheme(slug)
}

package theme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"herdr-theme-picker/tests/testutil"
)

func TestMain(m *testing.M) { testutil.Main(m) }

var writeTestFile = testutil.WriteFile
var fakeFZF = testutil.FakeFZF
var copyTestExecutable = testutil.CopyExecutable

type fzfStep = testutil.FZFStep

func fixturePalette(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(PluginRoot(), "themes", "dracula-default"))
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}

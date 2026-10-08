package theme

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestOutputBound(t *testing.T) {
	var b limitedBuffer
	if _, err := io.Copy(&b, strings.NewReader(strings.Repeat("x", maxPaletteBytes+1))); err == nil {
		t.Fatal("unbounded subprocess output accepted")
	}
}

func TestFZFCancellationAndFailures(t *testing.T) {
	for _, code := range []int{1, 130, 2} {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			fakeFZF(t, fzfStep{ExitCode: code})
			_, _, cancelled, err := runFZF("one", "", "", "esc", "preview")
			if code == 2 {
				if err == nil {
					t.Fatal("fzf error swallowed")
				}
			} else if err != nil || !cancelled {
				t.Fatal("cancel was treated as failure", err)
			}
		})
	}
	for _, output := range []string{"esc\none\n", "broken"} {
		t.Run(output, func(t *testing.T) {
			fakeFZF(t, fzfStep{Selection: output})
			_, _, cancelled, err := runFZF("one", "", "", "esc", "preview")
			if output == "broken" {
				if err == nil {
					t.Fatal("malformed output accepted")
				}
			} else if !cancelled || err != nil {
				t.Fatal("explicit Escape ignored")
			}
		})
	}
}

func TestPreviewCommandExecutesQuotedPaths(t *testing.T) {
	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	exe := filepath.Join(t.TempDir(), "preview 'quoted' $ path"+suffix)
	copyTestExecutable(t, exe)
	path := filepath.Join(t.TempDir(), "scratch 'quoted' $ file")
	command, shell := previewCommand(exe, "preview-file", path)
	command = strings.Replace(command, "{}", "'picked row'", 1)
	args := strings.Fields(shell)
	cmd := exec.Command(args[0], append(args[1:], command)...)
	cmd.Env = append(os.Environ(), "THEME_TEST_ARGV=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("preview did not execute: %v %s", err, out)
	}
	var got []string
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatal(err, string(out))
	}
	want := []string{"preview-file", path, "picked row"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("preview quoting changed arguments: got %q, want %q", got, want)
	}
}

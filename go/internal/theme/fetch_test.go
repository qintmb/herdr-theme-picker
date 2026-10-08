package theme

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePaletteLocalAndCache(t *testing.T) {
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	valid := fixturePalette(t)

	// 1. Resolves cached theme if valid
	cache := filepath.Join(CacheDir(), "cached-theme")
	writeTestFile(t, cache, valid)
	path, err := ResolvePalette("cached-theme")
	if err != nil || path != cache {
		t.Fatalf("failed to resolve cached theme: path=%s err=%v", path, err)
	}

	// 2. Corrupt cache falls through to not found with helpful link
	writeTestFile(t, cache, "invalid content\n")
	_, err = ResolvePalette("cached-theme")
	if err == nil || !strings.Contains(err.Error(), "terminalcolors.com") {
		t.Fatalf("expected error mentioning terminalcolors.com, got %v", err)
	}

	// 3. Unknown slug returns error mentioning terminalcolors.com
	_, err = ResolvePalette("nonexistent-theme")
	if err == nil || !strings.Contains(err.Error(), "terminalcolors.com") {
		t.Fatalf("expected error mentioning terminalcolors.com, got %v", err)
	}
}

func TestBashCacheIsReusedOffline(t *testing.T) {
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	path := filepath.Join(legacyCacheDir(), "old-download")
	writeTestFile(t, path, fixturePalette(t))

	got, err := ResolvePalette("old-download")
	if err != nil || got != path {
		t.Fatalf("ignored valid Bash cache: path=%s err=%v", got, err)
	}

	writeTestFile(t, path, "invalid")
	if _, err := ResolvePalette("old-download"); err == nil {
		t.Fatal("invalid Bash cache accepted")
	}
}

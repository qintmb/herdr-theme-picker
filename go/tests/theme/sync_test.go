package theme_test

import (
	"strings"
	"testing"

	. "herdr-theme-picker/internal/theme"
)

func TestPaletteOSCPayload(t *testing.T) {
	if payload := PaletteOSCPayload(nil); payload != "" {
		t.Fatalf("expected empty payload for nil palette, got %q", payload)
	}

	pal := &Palette{
		Background:    "#282a36",
		Foreground:    "#f8f8f2",
		PaletteColors: make(map[int]string),
	}
	pal.PaletteColors[0] = "#21222c"
	pal.PaletteColors[15] = "#ffffff"

	payload := PaletteOSCPayload(pal)
	if !strings.Contains(payload, "\033]10;#f8f8f2\007") {
		t.Errorf("missing foreground in payload: %q", payload)
	}
	if !strings.Contains(payload, "\033]11;#282a36\007") {
		t.Errorf("missing background in payload: %q", payload)
	}
	if !strings.Contains(payload, "\033]4;0;#21222c\007") {
		t.Errorf("missing color 0 in payload: %q", payload)
	}
	if !strings.Contains(payload, "\033]4;15;#ffffff\007") {
		t.Errorf("missing color 15 in payload: %q", payload)
	}
}

func TestSyncAppliedThemeMissing(t *testing.T) {
	t.Setenv("HERDR_PLUGIN_STATE_DIR", t.TempDir())
	if err := SyncAppliedTheme("invalid"); err != nil {
		t.Fatalf("expected nil when applied file does not exist, got: %v", err)
	}
	writeTestFile(t, AppliedFile(), "   \n")
	if err := SyncAppliedTheme("invalid"); err != nil {
		t.Fatalf("expected nil when applied file is empty/whitespace, got: %v", err)
	}
}

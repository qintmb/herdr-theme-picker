package terminal_test

import (
	"fmt"
	"strings"
	"testing"

	"herdr-theme-picker/internal/terminal"
)

func TestHostReadbackFollowsColorWrites(t *testing.T) {
	// All writes must finish before any query. A new client may have already
	// queried its original defaults before the daemon discovers it.
	colors := "\033]10;#abcdef\007\033]11;#123456\007\033]4;15;#ffffff\007"
	got := terminal.HostPayload(colors)
	if !strings.HasPrefix(got, colors) {
		t.Fatal("color writes were lost or reordered")
	}
	queries := strings.TrimPrefix(got, colors)
	for _, query := range []string{"\033]10;?\033\\", "\033]11;?\033\\"} {
		if strings.Count(queries, query) != 1 {
			t.Fatalf("missing or duplicate default-color readback: %q", queries)
		}
	}
	for index := 0; index < 16; index++ {
		query := fmt.Sprintf("\033]4;%d;?\033\\", index)
		if strings.Count(queries, query) != 1 {
			t.Fatalf("ANSI color %d was not queried once", index)
		}
	}
	if got := terminal.HostPayload(""); got != "" {
		t.Fatalf("no selected theme should emit nothing: %q", got)
	}
}

package theme

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"herdr-theme-picker/internal/instance"
	"herdr-theme-picker/internal/terminal"
)

// ApplyTheme writes [theme.custom] into this machine's Herdr config, records
// the selection, recolors this machine's Herdr client terminals, and reloads.
// Every Herdr client reads its own machine's config, so nothing here reaches
// a client on another machine. Terminal sync and reload are best effort: the
// config is the source of truth and a manual reload repairs either one.
func ApplyTheme(slug string) error {
	cfgPath := ConfigPath()
	palettePath, err := ResolvePalette(slug)
	if err != nil {
		return fmt.Errorf("cannot resolve theme %q: %w", slug, err)
	}
	pal, err := ParsePaletteFile(palettePath)
	if err != nil {
		return fmt.Errorf("invalid palette for %q: %w", slug, err)
	}
	tokens, err := PaletteToTokens(pal)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(StateDir(), 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	lock, err := instance.Acquire(AppliedFile() + ".lock")
	if err != nil {
		return fmt.Errorf("lock theme selection: %w", err)
	}
	defer lock.Close()
	if err := WriteCustomBlock(cfgPath, tokens); err != nil {
		return fmt.Errorf("failed to write custom theme block: %w", err)
	}
	// Record the selection with the config so a later sync or failed reload
	// can never pair the new UI theme with the previous terminal palette.
	if err := atomicWriteFile(AppliedFile(), []byte(slug+"\n"), 0o644); err != nil {
		return fmt.Errorf("theme written to %s, but selection marker failed: %w", cfgPath, err)
	}
	if err := SyncTerminalColors(pal, palettePath, os.Getenv("HERDR_THEME_CLIENT_PID")); err != nil && !errors.Is(err, terminal.ErrUnavailable) {
		fmt.Fprintf(os.Stderr, "Warning: terminal colors not fully synced: %v\n", err)
	}
	if err := ReloadHerdr(); err != nil {
		fmt.Fprintf(os.Stderr, "Theme %q saved to %s, but reload failed: %v\nUse Herdr's reload config action (default prefix+shift+r).\n", slug, cfgPath, err)
		return nil
	}
	fmt.Printf("Applied theme %q.\n", slug)
	return nil
}

// ReloadHerdr executes 'herdr server reload-config'. A machine that only
// connects with herdr --remote has no local server; the user reloads instead.
func ReloadHerdr() error {
	herdrBin := os.Getenv("HERDR_BIN_PATH")
	if herdrBin == "" {
		herdrBin = "herdr"
	}
	out, err := exec.Command(herdrBin, "server", "reload-config").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

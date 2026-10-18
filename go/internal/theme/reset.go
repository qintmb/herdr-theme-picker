package theme

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"herdr-theme-picker/internal/instance"
	"herdr-theme-picker/internal/terminal"
)

// DefaultRowMark starts the picker row that drops the plugin's colors. It can
// never be a theme slug, so the picker and the preview both recognize the row
// before any ✓/★ prefix is stripped.
const DefaultRowMark = "↩ herdr default"

// DefaultRow is the picker row for this machine. It names the built-in theme
// you land on, so the row never leaves you guessing which of Herdr's own themes
// you end up on.
func DefaultRow() string {
	return fmt.Sprintf("%s (built-in: %s)", DefaultRowMark, BaseThemeName(ConfigPath()))
}

// ResetToBuiltinTheme removes the plugin's [theme.custom] tokens so Herdr goes
// back to its built-in theme.name, hands this machine's terminals their own
// colors back, and reloads. It is the inverse of ApplyTheme, and best effort in
// the same places: the config is the source of truth and a manual reload
// repairs a missed terminal reset.
func ResetToBuiltinTheme() error {
	cfgPath := ConfigPath()
	if !HasCustomBlock(cfgPath) {
		fmt.Printf("Already on the built-in theme %q — nothing to undo.\n", BaseThemeName(cfgPath))
		return nil
	}
	if err := os.MkdirAll(StateDir(), 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	lock, err := instance.Acquire(AppliedFile() + ".lock")
	if err != nil {
		return fmt.Errorf("lock theme selection: %w", err)
	}
	defer lock.Close()
	if err := RemoveCustomBlock(cfgPath); err != nil {
		return fmt.Errorf("failed to remove custom theme block: %w", err)
	}
	// Drop the marker with the tokens: a later sync would otherwise re-send a
	// palette this config no longer carries.
	if err := os.Remove(AppliedFile()); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("custom tokens removed from %s, but selection marker failed: %w", cfgPath, err)
	}
	if err := RestoreGhosttyFragments(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Ghostty theme fragment not restored: %v\n", err)
	}
	if err := ResetTerminalColors(os.Getenv("HERDR_THEME_CLIENT_PID")); err != nil && !errors.Is(err, terminal.ErrUnavailable) {
		fmt.Fprintf(os.Stderr, "Warning: terminal colors not fully reset: %v\n", err)
	}
	if err := ReloadHerdr(); err != nil {
		fmt.Fprintf(os.Stderr, "Custom tokens removed from %s, but reload failed: %v\nUse Herdr's reload config action (default prefix+shift+r).\n", cfgPath, err)
		return nil
	}
	fmt.Printf("Custom colors removed — back on Herdr theme %q.\n", BaseThemeName(cfgPath))
	return nil
}

// DefaultRowPreview explains the row in the fzf preview pane. It names the
// theme you land on and says when a hand-written layer will keep overriding
// colors, while the popup is still readable.
func DefaultRowPreview() string {
	var sb strings.Builder
	fmt.Fprintln(&sb, "  Drops the plugin's [theme.custom] block, so Herdr")
	fmt.Fprintln(&sb, "  goes back to its own built-in theme:")
	fmt.Fprintf(&sb, "\n      %s\n\n", BaseThemeName(ConfigPath()))
	fmt.Fprintln(&sb, "  That name comes from theme.name — change it in")
	fmt.Fprintln(&sb, "  Herdr's Settings to land somewhere else.")
	if HasCustomLayers(ConfigPath()) {
		fmt.Fprintln(&sb)
		fmt.Fprintln(&sb, "  ⚠ your own [theme.custom.light/dark] section is kept and")
		fmt.Fprintln(&sb, "    keeps overriding those tokens in that appearance.")
	}
	return sb.String()
}

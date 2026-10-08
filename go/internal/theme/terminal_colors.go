package theme

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"herdr-theme-picker/internal/terminal"
)

// SyncTerminalColors recolors the host terminals of this machine's Herdr
// clients (including herdr --remote clients started here) and refreshes the
// Ghostty fragment when the user created it. Clients on other machines, and
// terminals with no Herdr client, are never touched. A non-empty clientPID
// limits the write to that client. It reports terminal.ErrUnavailable when no
// client was found.
func SyncTerminalColors(pal *Palette, palettePath, clientPID string) error {
	if err := validatePalette(pal); err != nil {
		return err
	}
	var errs []error
	// Ghostty only loads this fragment if the user includes it, so an existing
	// file is the opt-in. Never create it.
	if home, err := os.UserHomeDir(); err == nil && palettePath != "" {
		for _, fragment := range []string{
			filepath.Join(home, ".config", "ghostty", "herdr-theme"),
			filepath.Join(home, "AppData", "Roaming", "ghostty", "herdr-theme"),
		} {
			if _, err := os.Stat(fragment); errors.Is(err, os.ErrNotExist) {
				continue
			} else if err != nil {
				errs = append(errs, fmt.Errorf("Ghostty theme: %w", err))
				continue
			}
			data, err := os.ReadFile(palettePath)
			if err == nil {
				err = atomicWriteFile(fragment, data, 0o644)
			}
			if err != nil {
				errs = append(errs, fmt.Errorf("Ghostty theme: %w", err))
			}
		}
	}
	clients, err := terminal.ActiveClients(clientPID)
	if err != nil {
		return errors.Join(append(errs, err)...)
	}
	payload := PaletteOSCPayload(pal)
	synced := 0
	for _, client := range clients {
		if err := terminal.EmitClient(client, payload); err == nil {
			synced++
		} else if !errors.Is(err, terminal.ErrUnavailable) {
			errs = append(errs, fmt.Errorf("client %d: %w", client.PID, err))
		}
	}
	if synced == 0 && len(errs) == 0 {
		return terminal.ErrUnavailable
	}
	return errors.Join(errs...)
}

// PaletteOSCPayload builds the OSC 4/10/11 color sequence string for a palette.
func PaletteOSCPayload(pal *Palette) string {
	if pal == nil {
		return ""
	}
	var sb strings.Builder
	if pal.Foreground != "" {
		sb.WriteString(fmt.Sprintf("\033]10;%s\007", pal.Foreground))
	}
	if pal.Background != "" {
		sb.WriteString(fmt.Sprintf("\033]11;%s\007", pal.Background))
	}
	for idx := 0; idx < 16; idx++ {
		col := pal.PaletteColors[idx]
		if col != "" {
			sb.WriteString(fmt.Sprintf("\033]4;%d;%s\007", idx, col))
		}
	}
	return sb.String()
}

// SyncAppliedTheme re-sends the saved selection's colors, e.g. after opening a
// new terminal window. Without a saved selection nothing is written.
func SyncAppliedTheme(clientPID string) error {
	appliedPath := AppliedFile()
	data, err := os.ReadFile(appliedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read applied theme: %w", err)
	}
	slug := strings.TrimSpace(string(data))
	if slug == "" {
		return nil
	}
	palettePath, err := ResolvePalette(slug)
	if err != nil {
		return fmt.Errorf("resolve theme %q: %w", slug, err)
	}
	pal, err := ParsePaletteFile(palettePath)
	if err != nil {
		return fmt.Errorf("parse theme %q: %w", slug, err)
	}
	// A config edited independently, or a failed marker write, must never
	// resurrect stale terminal colors from the previous selection.
	data, err = os.ReadFile(ConfigPath())
	if err != nil {
		return fmt.Errorf("read config before sync: %w", err)
	}
	// Custom may also hold [theme.custom.light] and [theme.custom.dark] tables.
	var cfg struct {
		Theme struct{ Custom map[string]any }
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return err
	}
	tokens, err := PaletteToTokens(pal)
	if err != nil {
		return err
	}
	for key, value := range tokens {
		if current, _ := cfg.Theme.Custom[key].(string); !strings.EqualFold(current, value) {
			return fmt.Errorf("saved theme %q does not match %s; run apply %s again before syncing", slug, ConfigPath(), slug)
		}
	}
	return SyncTerminalColors(pal, palettePath, clientPID)
}

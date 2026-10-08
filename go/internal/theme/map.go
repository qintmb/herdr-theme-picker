package theme

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// TokenOrder is the exact ordering of Herdr theme tokens written to config.toml.
var TokenOrder = []string{
	"panel_bg", "sidebar_bg", "active_row_bg", "selection_bg", "text", "surface0",
	"surface1", "overlay0", "overlay1", "subtext0", "accent", "blue", "mauve", "green",
	"yellow", "red", "teal", "peach", "surface_dim",
}

// TokenOrigin returns the human-facing source line in a ghostty palette.
func TokenOrigin(token string) string {
	switch token {
	case "panel_bg":
		return "background"
	case "sidebar_bg":
		return "background −15%"
	case "active_row_bg":
		return "palette 0"
	case "selection_bg":
		return "selection-bg"
	case "text":
		return "foreground"
	case "surface0":
		return "palette 0"
	case "surface1", "overlay0":
		return "palette 8"
	case "overlay1", "subtext0":
		return "palette 7"
	case "accent", "blue":
		return "palette 4"
	case "mauve":
		return "palette 5"
	case "green":
		return "palette 2"
	case "yellow":
		return "palette 3"
	case "red":
		return "palette 1"
	case "teal":
		return "palette 6"
	case "peach":
		return "palette 9"
	case "surface_dim":
		return "background −8%"
	default:
		return "?"
	}
}

// TokenRole returns a short note on where the token appears in Herdr's UI.
func TokenRole(token string) string {
	switch token {
	case "panel_bg":
		return "pane background"
	case "sidebar_bg":
		return "sidebar background"
	case "active_row_bg":
		return "selected sidebar row"
	case "selection_bg":
		return "text selection bg"
	case "text":
		return "primary text"
	case "surface0":
		return "sidebar row bg"
	case "surface1":
		return "inactive pane bg"
	case "overlay0":
		return "dim text"
	case "overlay1":
		return "bright text"
	case "subtext0":
		return "secondary text"
	case "accent":
		return "active pane border/selector"
	case "blue":
		return "links/info"
	case "mauve":
		return "accent"
	case "green":
		return "success ✓"
	case "yellow":
		return "warning"
	case "red":
		return "error ✗"
	case "teal":
		return "now-playing"
	case "peach":
		return "accent"
	case "surface_dim":
		return "dividers"
	default:
		return ""
	}
}

// Palette holds the parsed entries of a ghostty palette file and any custom overrides.
type Palette struct {
	Background          string
	Foreground          string
	SelectionBackground string
	CursorColor         string
	PaletteColors       map[int]string
	Overrides           map[string]string
}

var (
	paletteEntryRe   = regexp.MustCompile(`^palette\s*=\s*(\d+)\s*=\s*(#?[A-Fa-f0-9]{6})(?:\s+#.*)?\s*$`)
	overrideEntryRe  = regexp.MustCompile(`^#\s*hpick-override:\s*([a-zA-Z0-9_]+)\s*=\s*(#?[A-Fa-f0-9]{6})(?:\s+#.*)?\s*$`)
	hexColorRe       = regexp.MustCompile(`^#?[A-Fa-f0-9]{6}$`)
	overridePrefixRe = regexp.MustCompile(`^#\s*hpick-override:`)
)

const maxPaletteBytes = 1 << 20

func normalizeColor(color string) (string, error) {
	if !hexColorRe.MatchString(color) {
		return "", fmt.Errorf("invalid color %q: expected #rrggbb", color)
	}
	return "#" + strings.ToLower(strings.TrimPrefix(color, "#")), nil
}

func isToken(key string) bool {
	for _, token := range TokenOrder {
		if token == key {
			return true
		}
	}
	return false
}

// ParsePaletteFile parses a ghostty format palette from disk.
func ParsePaletteFile(path string) (*Palette, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxPaletteBytes+1))
	if err != nil {
		return nil, err
	}
	return ParsePaletteContent(string(data))
}

// ParsePaletteContent parses a ghostty format palette from string content.
func ParsePaletteContent(content string) (*Palette, error) {
	if len(content) > maxPaletteBytes {
		return nil, fmt.Errorf("palette exceeds %d bytes", maxPaletteBytes)
	}
	pal := &Palette{
		PaletteColors: make(map[int]string),
		Overrides:     make(map[string]string),
	}

	scanner := bufio.NewScanner(strings.NewReader(strings.TrimPrefix(content, "\ufeff")))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if overridePrefixRe.MatchString(line) {
			m := overrideEntryRe.FindStringSubmatch(line)
			if len(m) != 3 || !isToken(m[1]) {
				return nil, fmt.Errorf("line %d: invalid theme token override", lineNo)
			}
			pal.Overrides[m[1]], _ = normalizeColor(m[2])
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			return nil, fmt.Errorf("line %d: expected key = value", lineNo)
		}
		key = strings.TrimSpace(key)

		if key == "palette" {
			m := paletteEntryRe.FindStringSubmatch(line)
			if len(m) != 3 {
				return nil, fmt.Errorf("line %d: invalid palette entry", lineNo)
			}
			idx, err := strconv.Atoi(m[1])
			if err != nil || idx < 0 || idx > 15 {
				return nil, fmt.Errorf("line %d: palette index must be 0..15", lineNo)
			}
			pal.PaletteColors[idx], _ = normalizeColor(m[2])
			continue
		}

		if key == "background" || key == "foreground" || key == "selection-background" || key == "cursor-color" {
			fields := strings.Fields(value)
			if len(fields) == 0 || (len(fields) > 1 && !strings.HasPrefix(fields[1], "#")) {
				return nil, fmt.Errorf("line %d: invalid color value", lineNo)
			}
			v, err := normalizeColor(fields[0])
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNo, err)
			}
			switch key {
			case "background":
				pal.Background = v
			case "foreground":
				pal.Foreground = v
			case "selection-background":
				pal.SelectionBackground = v
			case "cursor-color":
				pal.CursorColor = v
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return pal, validatePalette(pal)
}

func validatePalette(pal *Palette) error {
	if pal == nil {
		return fmt.Errorf("palette is nil")
	}
	for key, value := range map[string]string{"background": pal.Background, "foreground": pal.Foreground} {
		if _, err := normalizeColor(value); err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
	}
	for i := 0; i < 16; i++ {
		if _, err := normalizeColor(pal.PaletteColors[i]); err != nil {
			return fmt.Errorf("palette %d: %w", i, err)
		}
	}
	for i := range pal.PaletteColors {
		if i < 0 || i > 15 {
			return fmt.Errorf("palette index must be 0..15")
		}
	}
	for key, value := range pal.Overrides {
		if !isToken(key) {
			return fmt.Errorf("unknown theme token %q", key)
		}
		if _, err := normalizeColor(value); err != nil {
			return fmt.Errorf("override %s: %w", key, err)
		}
	}
	for _, value := range []string{pal.SelectionBackground, pal.CursorColor} {
		if value != "" {
			if _, err := normalizeColor(value); err != nil {
				return err
			}
		}
	}
	return nil
}

// PaletteToTokens transforms the parsed palette into the 19 Herdr theme tokens.
func PaletteToTokens(pal *Palette) (map[string]string, error) {
	if err := validatePalette(pal); err != nil {
		return nil, err
	}
	p0, ok := pal.PaletteColors[0]
	if !ok || p0 == "" {
		return nil, fmt.Errorf("palette 0 missing")
	}

	bg := pal.Background
	fg := pal.Foreground
	sel := pal.SelectionBackground
	p8 := pal.PaletteColors[8]
	if sel == "" {
		sel = p8
	}

	p1 := pal.PaletteColors[1]
	p2 := pal.PaletteColors[2]
	p3 := pal.PaletteColors[3]
	p4 := pal.PaletteColors[4]
	p5 := pal.PaletteColors[5]
	p6 := pal.PaletteColors[6]
	p7 := pal.PaletteColors[7]
	p9 := pal.PaletteColors[9]
	if p9 == "" {
		p9 = p3
	}

	tokens := map[string]string{
		"panel_bg":      bg,
		"sidebar_bg":    DarkenHex(bg, 15),
		"active_row_bg": p0,
		"selection_bg":  sel,
		"surface0":      p0,
		"surface1":      p8,
		"surface_dim":   DarkenHex(bg, 8),
		"overlay0":      p8,
		"overlay1":      p7,
		"text":          fg,
		"subtext0":      p7,
		"accent":        p4,
		"mauve":         p5,
		"green":         p2,
		"yellow":        p3,
		"red":           p1,
		"blue":          p4,
		"teal":          p6,
		"peach":         p9,
	}

	// Apply overrides
	for k, v := range pal.Overrides {
		tokens[k] = v
	}

	return tokens, nil
}

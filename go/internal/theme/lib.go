package theme

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var validSlugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var legacySlugRe = regexp.MustCompile(`^[a-z0-9-]+$`)

// Existing Bash names remain usable on Unix. New names always use the stricter
// portable rule. Never permit path separators, dot segments or shell syntax.
func isStoredUserSlug(slug string) bool {
	return IsValidSlug(slug) || (runtime.GOOS != "windows" && len(slug) <= 255 && legacySlugRe.MatchString(slug))
}

// IsValidSlug reports whether slug consists only of lowercase letters, digits, and hyphens.
func IsValidSlug(slug string) bool {
	if len(slug) > 120 || !validSlugRe.MatchString(slug) {
		return false
	}
	// These names cannot be created as normal files on Windows.
	switch slug {
	case "con", "prn", "aux", "nul", "com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9", "lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9":
		return false
	}
	return true
}

// Slugify converts an arbitrary name into a slug: lowercase, letters/numbers/hyphens only.
func Slugify(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	s = b.String()
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	return s
}

// DarkenHex darkens a hex color (#rrggbb or rrggbb) by the specified percent [0..100].
func DarkenHex(hexStr string, percent int) string {
	hex := strings.TrimPrefix(hexStr, "#")
	if len(hex) != 6 {
		return hexStr
	}
	r64, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g64, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b64, _ := strconv.ParseInt(hex[4:6], 16, 64)

	factor := int64(100 - percent)
	if factor < 0 {
		factor = 0
	}
	r := (r64 * factor) / 100
	g := (g64 * factor) / 100
	b := (b64 * factor) / 100

	if r < 0 {
		r = 0
	} else if r > 255 {
		r = 255
	}
	if g < 0 {
		g = 0
	} else if g > 255 {
		g = 255
	}
	if b < 0 {
		b = 0
	} else if b > 255 {
		b = 255
	}

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// ParseRGB parses a #rrggbb hex string into red, green, blue integer components.
func ParseRGB(hexStr string) (int, int, int) {
	hex := strings.TrimPrefix(hexStr, "#")
	if len(hex) != 6 {
		return 0, 0, 0
	}
	r, _ := strconv.ParseInt(hex[0:2], 16, 32)
	g, _ := strconv.ParseInt(hex[2:4], 16, 32)
	b, _ := strconv.ParseInt(hex[4:6], 16, 32)
	return int(r), int(g), int(b)
}

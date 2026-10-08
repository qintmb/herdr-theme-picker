package theme

import (
	"fmt"
	"os"
	"path/filepath"
)

// Hyperlink formats an OSC 8 terminal hyperlink for terminals that support it,
// while keeping the target visible for terminals that auto-detect URLs.
func Hyperlink(url, text string) string {
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}

// ResolvePalette gives custom themes precedence over bundled and cached themes.
func ResolvePalette(slug string) (string, error) {
	if !IsValidSlug(slug) {
		// Legacy Unix filenames are accepted only for indexed, regular user
		// files. They must never become arbitrary paths.
		if path, err := userThemePath(slug); err == nil {
			if _, err := ParsePaletteFile(path); err != nil {
				return "", err
			}
			return path, nil
		}
		return "", fmt.Errorf("invalid slug: %s", slug)
	}

	// 1. User-added theme. This order also repairs existing name collisions.
	userPath := filepath.Join(UserThemesDir(), slug)
	if info, err := os.Lstat(userPath); err == nil {
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("user theme is not a regular file: %s", slug)
		}
		if _, err := ParsePaletteFile(userPath); err != nil {
			return "", fmt.Errorf("invalid user theme: %w", err)
		}
		return userPath, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	// 2. Bundled theme
	bundlePath := filepath.Join(PluginRoot(), "themes", slug)
	if info, err := os.Stat(bundlePath); err == nil && !info.IsDir() {
		return bundlePath, nil
	}

	// 3. Cached theme (retained if previously downloaded or user-cached)
	cacheDir := CacheDir()
	cachePath := filepath.Join(cacheDir, slug)
	if info, err := os.Lstat(cachePath); err == nil {
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("cache is not a regular file: %s", slug)
		}
		if _, err := ParsePaletteFile(cachePath); err == nil {
			return cachePath, nil
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}

	// 4. Reuse validated Bash downloads if present
	if legacy := legacyCacheDir(); legacy != "" {
		path := filepath.Join(legacy, slug)
		if info, err := os.Lstat(path); err == nil && info.Mode().IsRegular() {
			if _, err := ParsePaletteFile(path); err == nil {
				return path, nil
			}
		}
	}

	return "", fmt.Errorf("theme %q not found (browse 400+ themes at %s and import via clipboard or editor)", slug, Hyperlink("https://terminalcolors.com", "https://terminalcolors.com"))
}

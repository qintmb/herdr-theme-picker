package theme

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ConfigPath returns the location of Herdr's config.toml across platforms.
func ConfigPath() string {
	if custom := os.Getenv("HERDR_CONFIG_PATH"); custom != "" {
		return custom
	}
	// Match Herdr's precedence on every OS, including Windows.
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "herdr", "config.toml")
	}
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "herdr", "config.toml")
		}
		if profile := os.Getenv("USERPROFILE"); profile != "" {
			return filepath.Join(profile, "AppData", "Roaming", "herdr", "config.toml")
		}
	}

	// Default fallback: XDG or ~/.config
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "herdr", "config.toml")
	}
	return "config.toml"
}

// PluginRoot returns the directory containing plugin assets (themes, herdr-plugin.toml).
func PluginRoot() string {
	if root := os.Getenv("HERDR_PLUGIN_ROOT"); root != "" {
		return root
	}
	exe, err := os.Executable()
	if err == nil {
		exe = strings.TrimPrefix(exe, `\\?\`)
		dir := filepath.Dir(exe)
		if _, err := os.Stat(filepath.Join(dir, "themes")); err == nil {
			return dir
		}
	}
	// Development runs (go run, go test) start inside go/; the assets live at
	// the repository root next to herdr-plugin.toml.
	if dir, err := os.Getwd(); err == nil {
		for range 5 {
			if _, err := os.Stat(filepath.Join(dir, "themes")); err == nil {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return "."
}

// StateDir returns the directory where plugin state (applied theme, user themes) is persisted.
func StateDir() string {
	if state := os.Getenv("HERDR_PLUGIN_STATE_DIR"); state != "" {
		return state
	}
	// Herdr supplies this path to plugin commands. Standalone picker/sync runs
	// must use the same state when the plugin is installed.
	if managed := managedStateDir(); managed != "" {
		if info, err := os.Stat(managed); err == nil && info.IsDir() {
			return managed
		}
	}
	// Preserve the standalone location for users who have not installed or
	// linked the plugin. Explicit HERDR_PLUGIN_STATE_DIR always wins.
	if runtime.GOOS == "windows" {
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, "herdr-theme-picker")
		}
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "herdr", "plugins", "state", "herdr-theme-picker")
		}
	}
	if legacy := legacyCacheDir(); legacy != "" {
		return legacy
	}
	return filepath.Join(".", ".state")
}

func managedStateDir() string {
	if state := os.Getenv("XDG_STATE_HOME"); state != "" {
		return filepath.Join(state, "herdr", "plugins", "herdr-theme-picker")
	}
	if runtime.GOOS == "windows" {
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, "herdr", "plugins", "herdr-theme-picker")
		}
		if profile := os.Getenv("USERPROFILE"); profile != "" {
			return filepath.Join(profile, "AppData", "Local", "herdr", "plugins", "herdr-theme-picker")
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".local", "state", "herdr", "plugins", "herdr-theme-picker")
	}
	return ""
}

// CacheDir returns the download cache directory for remote themes.
func CacheDir() string {
	return filepath.Join(StateDir(), "cache")
}

// legacyCacheDir is Bash's cache: its downloads, and its state when the plugin
// was not installed. Bash stored downloads here even for managed installs.
func legacyCacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "herdr-theme-picker")
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".cache", "herdr-theme-picker")
	}
	return ""
}

// UserThemesDir returns the directory storing user-added themes.
func UserThemesDir() string {
	return filepath.Join(StateDir(), "themes")
}

// UserIndexFile returns the path to the user themes index list.
func UserIndexFile() string {
	return filepath.Join(StateDir(), "index.txt")
}

// AppliedFile returns the path to the file storing the active theme slug.
// It is per machine, like the config it describes, and matches Bash's name.
func AppliedFile() string {
	return filepath.Join(StateDir(), "applied")
}

// IsUserTheme reports whether slug is an existing user-added theme.
func IsUserTheme(slug string) bool {
	_, err := userThemePath(slug)
	return err == nil
}

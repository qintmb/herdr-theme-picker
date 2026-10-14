package theme

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

// DefaultThemeName is the built-in theme Herdr uses when theme.name is unset or
// names an unknown theme. Mirrored from Herdr's own fallback so the picker's
// "herdr default" row can label the theme it lands on.
const DefaultThemeName = "catppuccin"

// WriteCustomBlock replaces or appends the [theme.custom] section in Herdr's config.toml.
func WriteCustomBlock(cfgPath string, tokens map[string]string) error {
	return writeConfig(cfgPath, func(data []byte) ([]byte, error) {
		return replaceCustomTokens(data, tokens)
	})
}

// RemoveCustomBlock drops the plugin's [theme.custom] tokens so Herdr falls back
// to its built-in theme. A [theme.custom.light] / [theme.custom.dark] table
// survives: the plugin never writes one, so it belongs to the user, and the
// picker warns about it before resetting.
func RemoveCustomBlock(cfgPath string) error {
	if _, err := os.Stat(cfgPath); errors.Is(err, os.ErrNotExist) {
		// Nothing to remove, and a reset must not create a config Herdr lacks.
		return nil
	}
	return writeConfig(cfgPath, func(data []byte) ([]byte, error) {
		return replaceCustomTokens(data, nil)
	})
}

// BaseThemeName reports the built-in theme Herdr uses once the plugin's tokens
// are gone, so a "herdr default" row can name it instead of leaving the user to
// guess. Unreadable or invalid config falls back to Herdr's own default.
func BaseThemeName(cfgPath string) string {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return DefaultThemeName
	}
	var cfg struct {
		Theme struct {
			Name string
		}
	}
	if err := toml.Unmarshal(data, &cfg); err != nil || cfg.Theme.Name == "" {
		return DefaultThemeName
	}
	return cfg.Theme.Name
}

// customTable reads theme.custom, or nil when the config has none. The same
// parser entry as the writer, so an inline `theme = { custom = … }` counts too.
func customTable(cfgPath string) map[string]any {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil
	}
	var cfg struct {
		Theme struct {
			Custom map[string]any
		}
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	return cfg.Theme.Custom
}

// HasCustomBlock reports whether the plugin's tokens are in the config. This —
// not the applied marker — decides which picker row is marked ✓, because a
// block written before the marker existed still means custom colors are live.
func HasCustomBlock(cfgPath string) bool {
	for key := range customTable(cfgPath) {
		if isToken(key) {
			return true
		}
	}
	return false
}

// HasCustomLayers reports a [theme.custom.light] / [theme.custom.dark] table.
// Those survive a reset, so the picker warns that the UI stays overridden in
// that appearance.
func HasCustomLayers(cfgPath string) bool {
	for key := range customTable(cfgPath) {
		if !isToken(key) {
			return true
		}
	}
	return false
}

// writeConfig keeps one backup of the original config each day, then replaces
// it. A config that does not exist yet is created without inventing a backup.
func writeConfig(cfgPath string, rewrite func([]byte) ([]byte, error)) error {
	data, err := os.ReadFile(cfgPath)
	missing := errors.Is(err, os.ErrNotExist)
	if err != nil && !missing {
		return err
	}

	updated, err := rewrite(data)
	if err != nil {
		return err
	}
	if missing {
		// Herdr works without a config; create one without inventing a backup.
		if err := os.MkdirAll(filepath.Dir(cfgPath), 0o700); err != nil {
			return err
		}
		return atomicWriteFile(cfgPath, updated, 0o600)
	}
	// Keep one backup of the original config each day. Never overwrite it.
	bak := cfgPath + ".bak-" + time.Now().Format("20060102")
	f, err := os.OpenFile(bak, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err == nil {
		_, writeErr := f.Write(data)
		if writeErr == nil {
			writeErr = f.Sync()
		}
		closeErr := f.Close()
		if err := errors.Join(writeErr, closeErr); err != nil {
			cleanupErr := os.Remove(bak)
			return errors.Join(fmt.Errorf("config backup failed: %w", err), cleanupErr)
		}
	} else if !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("config backup failed: %w", err)
	} else {
		info, statErr := os.Lstat(bak)
		if statErr != nil {
			return fmt.Errorf("check config backup: %w", statErr)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("config backup is not a regular file: %s", bak)
		}
	}
	return atomicWriteFile(cfgPath, updated, 0o600)
}

// Use parser expression ranges rather than matching lines, which can be part of
// multiline strings or arrays. All unrelated bytes (including comments) survive.
func replaceCustomTokens(data []byte, tokens map[string]string) ([]byte, error) {
	var document map[string]any
	if err := toml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("invalid existing TOML: %w", err)
	}
	normalized := make(map[string]string, len(tokens))
	for key, value := range tokens {
		if !isToken(key) {
			return nil, fmt.Errorf("unknown theme token %q", key)
		}
		color, err := normalizeColor(value)
		if err != nil {
			return nil, err
		}
		normalized[key] = color
	}
	var parser unstable.Parser
	parser.Reset(data)
	var table []string
	var out bytes.Buffer
	last := 0
	inlineTheme := false
	keptLayers := make(map[string]bool)
	for parser.NextExpression() {
		node := parser.Expression()
		it := node.Key()
		var key []string
		start := -1
		for it.Next() {
			if start < 0 {
				start = int(it.Node().Raw.Offset)
			}
			key = append(key, string(it.Node().Data))
		}
		path := key
		end := start
		if node.Kind == unstable.Table || node.Kind == unstable.ArrayTable {
			table = key
		} else {
			path = append(append([]string{}, table...), key...)
			end = int(node.Raw.Offset + node.Raw.Length)
			if len(path) == 1 && path[0] == "theme" {
				if _, ok := document["theme"].(map[string]any); !ok {
					return nil, fmt.Errorf("theme must be a TOML table")
				}
				inlineTheme = true
			}
		}
		owned := len(path) == 1 && path[0] == "theme" && inlineTheme
		if len(path) >= 2 && path[0] == "theme" && path[1] == "custom" {
			// Keep auto_switch layers written as [theme.custom.light|dark]
			// tables. Any other spelling (dotted or inline) also defines
			// theme.custom, which would clash with the appended header, so it
			// is removed and re-emitted as a table below.
			name := layerName(table)
			owned = name == ""
			if !owned {
				keptLayers[name] = true
			}
		}
		if !owned {
			continue
		}
		start = bytes.LastIndexByte(data[:start], '\n') + 1
		if n := bytes.IndexByte(data[end:], '\n'); n >= 0 {
			end += n + 1
		} else {
			end = len(data)
		}
		out.Write(data[last:start])
		last = end
	}
	if err := parser.Error(); err != nil {
		return nil, err
	}
	out.Write(data[last:])
	nl := "\n"
	if bytes.Contains(data, []byte("\r\n")) {
		nl = "\r\n"
	}
	// Drop trailing blank lines so repeated applies do not grow the file.
	out.Truncate(len(bytes.TrimRight(out.Bytes(), "\r\n")))
	if out.Len() > 0 {
		out.WriteString(nl)
	}
	layers := customLayers(document)
	if inlineTheme {
		custom := make(map[string]any)
		for key, color := range normalized {
			custom[key] = color
		}
		for name, layer := range layers {
			custom[name] = layer
		}
		theme := document["theme"].(map[string]any)
		// A reset has no tokens to write: drop the inline table rather than
		// leave `custom = {}` behind.
		if len(custom) == 0 {
			delete(theme, "custom")
		} else {
			theme["custom"] = custom
		}
		encoded, err := toml.Marshal(map[string]any{"theme": theme})
		if err != nil {
			return nil, err
		}
		out.WriteString(nl)
		out.WriteString(strings.ReplaceAll(string(encoded), "\n", nl))
	} else {
		// No tokens means a reset: skip the header instead of leaving an empty
		// [theme.custom] table, which would still read as "the plugin wrote here".
		if len(normalized) > 0 {
			out.WriteString(nl + "[theme.custom]" + nl)
			for _, k := range TokenOrder {
				if color, ok := normalized[k]; ok && color != "" {
					fmt.Fprintf(&out, "%s = %q%s", k, color, nl)
				}
			}
		}
		for _, name := range layerNames {
			layer, ok := layers[name]
			if !ok || keptLayers[name] {
				continue
			}
			encoded, err := toml.Marshal(layer)
			if err != nil {
				return nil, err
			}
			out.WriteString(nl + "[theme.custom." + name + "]" + nl)
			out.WriteString(strings.ReplaceAll(string(encoded), "\n", nl))
		}
	}
	var updated map[string]any
	if err := toml.Unmarshal(out.Bytes(), &updated); err != nil {
		return nil, fmt.Errorf("invalid generated TOML: %w", err)
	}
	if !reflect.DeepEqual(customLayers(updated), layers) {
		return nil, fmt.Errorf("could not preserve [theme.custom.light] / [theme.custom.dark]")
	}
	return out.Bytes(), nil
}

// layerNames are the auto_switch override tables inside [theme.custom].
var layerNames = []string{"light", "dark"}

// layerName returns "light" or "dark" for a path inside an auto_switch layer.
func layerName(path []string) string {
	if len(path) >= 3 && path[0] == "theme" && path[1] == "custom" && slices.Contains(layerNames, path[2]) {
		return path[2]
	}
	return ""
}

// customLayers returns the user's [theme.custom.light] / [theme.custom.dark]
// overrides. Herdr applies them on top of [theme.custom] when auto_switch is
// enabled; the picker owns only the base tokens.
func customLayers(document map[string]any) map[string]map[string]any {
	layers := make(map[string]map[string]any)
	theme, _ := document["theme"].(map[string]any)
	custom, _ := theme["custom"].(map[string]any)
	for _, name := range layerNames {
		if layer, ok := custom[name].(map[string]any); ok {
			layers[name] = layer
		}
	}
	return layers
}

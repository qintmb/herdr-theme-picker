package theme

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

// PickEditor honors explicit preferences before choosing an installed fallback.
func PickEditor() string {
	for _, key := range []string{"VISUAL", "EDITOR"} {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	for _, name := range []string{"nvim", "nano"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	if runtime.GOOS == "windows" {
		return "notepad.exe"
	}
	return "vi"
}

// Parse arguments without evaluating shell operators, expansions or substitutions.
// Backslashes in Windows paths stay literal; quoted executable paths are supported.
func splitEditor(command string) ([]string, error) {
	var args []string
	var word strings.Builder
	var quote rune
	started := false
	runes := []rune(command)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\\' && i+1 < len(runes) && ((quote == '"' && runes[i+1] == '"') || (quote == 0 && unicode.IsSpace(runes[i+1]))) {
			i++
			word.WriteRune(runes[i])
			started = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				word.WriteRune(r)
			}
			started = true
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			started = true
			continue
		}
		if unicode.IsSpace(r) {
			if started {
				args = append(args, word.String())
				word.Reset()
				started = false
			}
			continue
		}
		word.WriteRune(r)
		started = true
	}
	if quote != 0 {
		return nil, fmt.Errorf("unclosed quote in editor command")
	}
	if started {
		args = append(args, word.String())
	}
	if len(args) == 0 || args[0] == "" {
		return nil, fmt.Errorf("editor command is empty")
	}
	return args, nil
}

func editorCommand(path string) (*exec.Cmd, error) {
	editor := PickEditor()
	// A preference may be an unquoted executable path containing spaces.
	var args []string
	if resolved, err := exec.LookPath(editor); err == nil {
		args = []string{resolved}
	} else {
		var err error
		args, err = splitEditor(editor)
		if err != nil {
			return nil, err
		}
	}
	executable, err := exec.LookPath(args[0])
	if err != nil {
		return nil, err
	}
	args = append(args[1:], path)
	if runtime.GOOS == "windows" && (strings.EqualFold(filepath.Ext(executable), ".cmd") || strings.EqualFold(filepath.Ext(executable), ".bat")) {
		// Editors such as VS Code install a .cmd shim rather than a PE executable.
		// PowerShell's literal argument quoting avoids cmd.exe expansion of paths.
		command, shell := previewCommand(executable, args...)
		shellArgs := strings.Fields(shell)
		return exec.Command(shellArgs[0], append(shellArgs[1:], strings.TrimSuffix(command, " {}"))...), nil
	}
	return exec.Command(executable, args...), nil
}

const saveRow = "✓ Save & apply"

func setOverride(body, token, color string) (string, error) {
	if !isToken(token) {
		return "", fmt.Errorf("unknown token: %s", token)
	}
	color, err := normalizeColor(color)
	if err != nil {
		return "", err
	}
	var lines []string
	for _, line := range strings.Split(strings.TrimRight(body, "\r\n"), "\n") {
		m := overrideEntryRe.FindStringSubmatch(strings.TrimSpace(line))
		if len(m) == 3 && m[1] == token {
			continue
		}
		lines = append(lines, line)
	}
	lines = append(lines, fmt.Sprintf("# hpick-override: %s=%s", token, color))
	updated := strings.Join(lines, "\n") + "\n"
	if _, err := ParsePaletteContent(updated); err != nil {
		return "", err
	}
	return updated, nil
}

// EditTheme previews and edits a private scratch file; only Save & apply commits.
func EditTheme(slug string) error {
	dest, err := userThemePath(slug)
	if err != nil {
		return err
	}
	if _, err := ParsePaletteFile(dest); err != nil {
		return err
	}
	body, err := os.ReadFile(dest)
	if err != nil {
		return err
	}
	scratch, err := os.CreateTemp("", "herdr-theme-edit-*")
	if err != nil {
		return err
	}
	path := scratch.Name()
	defer os.Remove(path)
	if err := scratch.Close(); err != nil {
		return err
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		if err := atomicWriteFile(path, body, 0o600); err != nil {
			return err
		}
		pal, err := ParsePaletteContent(string(body))
		if err != nil {
			return err
		}
		tokens, err := PaletteToTokens(pal)
		if err != nil {
			return err
		}
		var rows []string
		for _, token := range TokenOrder {
			origin := TokenOrigin(token)
			if _, overridden := pal.Overrides[token]; overridden {
				origin = "custom override"
			}
			rows = append(rows, fmt.Sprintf("%-14s %s ← %-16s %s", token, tokens[token], origin, TokenRole(token)))
		}
		rows = append(rows, saveRow)
		_, selected, cancelled, err := runFZF(strings.Join(rows, "\n"), "Edit "+slug+": ",
			"↵ change hex · Save & apply commits · esc discards", "esc", "preview-file", path)
		if err != nil || cancelled {
			return err
		}
		if selected == saveRow {
			// Recheck ownership before committing; never follow a replaced symlink.
			if _, err := userThemePath(slug); err != nil {
				return err
			}
			if err := atomicWriteFile(dest, body, 0o644); err != nil {
				return err
			}
			return ApplyTheme(slug)
		}
		fields := strings.Fields(selected)
		if len(fields) == 0 || !isToken(fields[0]) {
			return fmt.Errorf("invalid token selection: %q", selected)
		}
		color, err := readAnswer(reader, "New hex for "+fields[0]+" (#rrggbb; empty cancels): ")
		if err != nil {
			return err
		}
		if color == "" {
			continue
		}
		updated, err := setOverride(string(body), fields[0], color)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		body = []byte(updated)
	}
}

package theme

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"herdr-theme-picker/internal/terminal"
)

const (
	addRow     = "+ Add new theme…"
	addClipRow = "+ Add from clipboard…"
	template   = `# Paste a ghostty-format theme below, then save and close this editor.
# Required: background, foreground, and palette 0..15 as #rrggbb colors.
# Get one from https://terminalcolors.com → Download → Ghostty.
#
# background = #1e1e2e
# foreground = #cdd6f4
# cursor-color = #f5e0dc
# palette = 0=#45475a
# palette = 1=#f38ba8
# ... through ...
# palette = 15=#a6adc8
`
)

// RunPicker uses a loop so repeated deletion/editing never grows the call stack.
// Inside Herdr with only herdr --remote clients attached, whoever opened it is
// on another machine whose theme comes from its own config, so changing this
// machine's config would not affect them; it explains where to run it instead.
func RunPicker() error {
	if os.Getenv("HERDR_ENV") != "" {
		if processes, err := terminal.Processes(); err == nil && terminal.OnlyRemoteClients(processes) {
			ShowClientSetup()
			return nil
		}
	}
	for {
		input, err := pickerItems()
		if err != nil {
			return err
		}
		header := fmt.Sprintf("↵ apply · tab/+ Add: new · ctrl-e edit ★ · ctrl-d delete ★ · esc cancel\nBrowse 400+ themes: %s", Hyperlink("https://terminalcolors.com", "https://terminalcolors.com"))
		key, selection, cancelled, err := runFZF(input, "Search themes: ",
			header,
			"tab,ctrl-e,ctrl-d", "preview")
		if err != nil || cancelled {
			return err
		}
		if key == "tab" || (key == "" && (selection == addRow || selection == addClipRow)) {
			var slug string
			if key != "tab" && selection == addClipRow {
				slug, err = AddThemeClipboard()
			} else {
				slug, err = AddThemeEditor()
			}
			if err != nil {
				return err
			}
			if slug == "" {
				return nil
			}
			return ApplyTheme(slug)
		}
		slug := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(selection, "✓ "), "★ "))
		if key == "ctrl-e" || key == "ctrl-d" {
			if !IsUserTheme(slug) {
				fmt.Fprintf(os.Stderr, "Only ★ user-added themes can be edited or deleted: %q.\n", slug)
				continue
			}
			if key == "ctrl-e" {
				return EditTheme(slug)
			}
			if err := DeleteTheme(slug); err != nil {
				return err
			}
			continue
		}
		if slug == "" {
			return nil
		}
		return ApplyTheme(slug)
	}
}

// ShowClientSetup explains how a remote client picks its own theme.
func ShowClientSetup() {
	fmt.Println("Themes belong to your client machine.\n\nOpen a local terminal outside Herdr and run:\n\n  herdr-theme-picker picker\n\nThen use Reload config in the Herdr client you want to update.\nInstall the picker locally if you are connected to a remote server.\nThis server's theme and terminal colors are left unchanged.\n\nPress Enter to close.")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

func pickerItems() (string, error) {
	users, err := readUserIndex()
	if err != nil {
		return "", err
	}
	bundledIndex := filepath.Join(PluginRoot(), "themes", "index.txt")
	bundled, invalid, err := readIndex(bundledIndex, IsValidSlug)
	if err != nil {
		return "", err
	}
	if len(invalid) > 0 {
		return "", fmt.Errorf("invalid slug %q in %s", invalid[0], bundledIndex)
	}
	applied := ""
	if data, err := os.ReadFile(AppliedFile()); err == nil {
		applied = strings.TrimSpace(string(data))
		if applied != "" && !IsValidSlug(applied) && !IsUserTheme(applied) {
			fmt.Fprintf(os.Stderr, "Ignoring unsupported saved theme %q; choose a theme to replace it.\n", applied)
			applied = ""
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	var rows []string
	seen := make(map[string]bool)
	if applied != "" {
		rows = append(rows, "✓ "+applied)
		seen[applied] = true
	}
	for _, slug := range users {
		// users came from the index already; only the file needs checking.
		if _, err := userThemeFile(slug); err != nil {
			fmt.Fprintf(os.Stderr, "Skipping unavailable user theme %q: %v\n", slug, err)
			continue
		}
		if !seen[slug] {
			rows = append(rows, "★ "+slug)
			seen[slug] = true
		}
	}
	for _, slug := range bundled {
		if !seen[slug] {
			rows = append(rows, "  "+slug)
			seen[slug] = true
		}
	}
	rows = append(rows, addRow, addClipRow)
	return strings.Join(rows, "\n"), nil
}

func readAnswer(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	answer, err := reader.ReadString('\n')
	if err != nil && !(errors.Is(err, io.EOF) && answer != "") {
		return "", err
	}
	return strings.TrimSpace(answer), nil
}

func finalizeTheme(body string) (string, error) {
	if _, err := ParsePaletteContent(body); err != nil {
		return "", fmt.Errorf("invalid palette: %w", err)
	}
	reader := bufio.NewReader(os.Stdin)
	name, err := readAnswer(reader, "Theme name: ")
	if err != nil {
		return "", err
	}
	slug := Slugify(name)
	if !IsValidSlug(slug) {
		return "", fmt.Errorf("invalid or reserved theme name: %q", name)
	}
	dest := filepath.Join(UserThemesDir(), slug)
	if info, err := os.Lstat(dest); err == nil {
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("theme is not a regular file: %s", slug)
		}
		yn, err := readAnswer(reader, fmt.Sprintf("%q already exists. Overwrite? [y/N] ", slug))
		if err != nil {
			return "", err
		}
		if !strings.EqualFold(yn, "y") && !strings.EqualFold(yn, "yes") {
			return "", nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err := saveUserTheme(slug, body); err != nil {
		return "", err
	}
	fmt.Fprintf(os.Stderr, "Saved theme %q.\n", slug)
	return slug, nil
}

func saveUserTheme(slug, body string) error {
	if !IsValidSlug(slug) {
		return fmt.Errorf("invalid slug: %q", slug)
	}
	if _, err := ParsePaletteContent(body); err != nil {
		return err
	}
	if err := os.MkdirAll(UserThemesDir(), 0o755); err != nil {
		return err
	}
	dest := filepath.Join(UserThemesDir(), slug)
	if info, err := os.Lstat(dest); err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("theme is not a regular file: %s", slug)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := atomicWriteFile(dest, []byte(body), 0o644); err != nil {
		return err
	}
	if err := updateUserIndex(slug, false); err != nil {
		return fmt.Errorf("theme saved at %s, but index update failed: %w", dest, err)
	}
	return nil
}

func AddThemeClipboard() (string, error) {
	clip, err := ReadClipboard()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(clip) == "" {
		return "", fmt.Errorf("clipboard is empty (copy a Ghostty theme from %s first)", Hyperlink("https://terminalcolors.com", "https://terminalcolors.com"))
	}
	return finalizeTheme(clip)
}

func AddThemeEditor() (string, error) {
	f, err := os.CreateTemp("", "herdr-theme-*.txt")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(template); err != nil {
		f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	cmd, err := editorCommand(f.Name())
	if err != nil {
		return "", err
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor failed: %w", err)
	}
	pal, err := ParsePaletteFile(f.Name())
	if err != nil {
		return "", err
	}
	if err := validatePalette(pal); err != nil {
		return "", err
	}
	data, err := os.ReadFile(f.Name())
	if err != nil {
		return "", err
	}
	return finalizeTheme(string(data))
}

// DeleteTheme confirms on stdin; guards apply equally to the picker and CLI.
func DeleteTheme(slug string) error { return deleteTheme(slug, os.Stdin) }

func deleteTheme(slug string, input io.Reader) error {
	dest, err := userThemePath(slug)
	if err != nil {
		return err
	}
	answer, err := readAnswer(bufio.NewReader(input), fmt.Sprintf("Delete %q? [y/N] ", slug))
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return err
	}
	if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
		return nil
	}
	marker, err := os.ReadFile(AppliedFile())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Remove(dest); err != nil {
		return fmt.Errorf("delete theme: %w", err)
	}
	indexErr := updateUserIndex(slug, true)
	var markerErr error
	if strings.TrimSpace(string(marker)) == slug {
		markerErr = os.Remove(AppliedFile())
	}
	if err := errors.Join(indexErr, markerErr); err != nil {
		return fmt.Errorf("theme deleted, but state cleanup failed: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Deleted theme %q.\n", slug)
	return nil
}

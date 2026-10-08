package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"herdr-theme-picker/internal/terminal"
	"herdr-theme-picker/internal/theme"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"picker"}
	}

	switch args[0] {
	case "help", "--help", "-h":
		fmt.Println("Usage: herdr-theme-picker [picker | apply <slug> | sync [--client <pid>] | preview <slug> | add [--clipboard] | edit <slug> | delete <slug>]")
		fmt.Println("Themes are written to this machine's Herdr config. Clients attached with herdr --remote use their own machine's config: run the picker there.")
		fmt.Printf("Browse 400+ additional themes at: %s\n", theme.Hyperlink("https://terminalcolors.com", "https://terminalcolors.com"))
		return
	case "client-setup":
		theme.ShowClientSetup()
		return
	case "startup":
		// Old manifests registered a startup daemon. Terminal colors are now
		// synced only when a theme is applied, so this is a no-op.
		return

	case "sync":
		client := os.Getenv("HERDR_THEME_CLIENT_PID")
		if len(args) == 3 && args[1] == "--client" {
			client = args[2]
		} else if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "usage: herdr-theme-picker sync [--client <pid>]")
			os.Exit(1)
		}
		if err := theme.SyncAppliedTheme(client); err != nil {
			if errors.Is(err, terminal.ErrUnavailable) {
				fmt.Fprintln(os.Stderr, "No Herdr client terminal found on this machine.")
			} else {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			os.Exit(1)
		}

	case "terminal-sync":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "usage: herdr-theme-picker terminal-sync <client-pid>")
			os.Exit(1)
		}
		if err := terminal.RunHelper(args[1]); err != nil {
			if errors.Is(err, terminal.ErrUnavailable) {
				os.Exit(3)
			}
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "preview-file":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: herdr-theme-picker preview-file <palette-path>")
			os.Exit(1)
		}
		fmt.Print(theme.RenderSwatchFile(args[1], "editing"))
	case "picker":
		if err := theme.RunPicker(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "open":
		herdrBin := os.Getenv("HERDR_BIN_PATH")
		if herdrBin == "" {
			herdrBin = "herdr"
		}
		cmd := exec.Command(herdrBin, "plugin", "pane", "open",
			"--plugin", "herdr-theme-picker",
			// Herdr spawns panes without resolving ".exe", so Windows has its
			// own pane id; the Bash implementation owns "picker".
			"--entrypoint", "picker-windows",
			"--placement", "popup",
			"--focus",
		)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open theme picker pane: %v\n", err)
			os.Exit(1)
		}

	case "preview", "--preview":
		if len(args) < 2 {
			fmt.Println("usage: herdr-theme-picker preview <slug>")
			return
		}
		row := strings.Join(args[1:], " ")
		if strings.HasPrefix(row, "+ Add from clipboard") {
			fmt.Println("  Reads a ghostty-format palette from your clipboard,")
			fmt.Println("  asks a name, saves and applies — no editor.")
			fmt.Println()
			fmt.Println("  Browse 400+ themes:")
			fmt.Printf("  %s (Download → Ghostty)\n", theme.Hyperlink("https://terminalcolors.com", "https://terminalcolors.com"))
			return
		}
		if strings.HasPrefix(row, "+ Add new theme") {
			fmt.Println("  Opens an editor to paste a")
			fmt.Println("  ghostty-format theme; name it, saved and applied.")
			fmt.Println()
			fmt.Println("  Browse 400+ themes:")
			fmt.Printf("  %s (Download → Ghostty)\n", theme.Hyperlink("https://terminalcolors.com", "https://terminalcolors.com"))
			return
		}

		bare := strings.TrimPrefix(row, "✓ ")
		bare = strings.TrimPrefix(bare, "★ ")
		bare = strings.TrimSpace(bare)
		fmt.Print(theme.RenderSwatch(bare))

	case "apply":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: herdr-theme-picker apply <slug>")
			os.Exit(1)
		}
		if err := theme.ApplyTheme(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "add":
		var slug string
		var err error
		if len(args) > 1 && args[1] == "--clipboard" {
			slug, err = theme.AddThemeClipboard()
		} else {
			slug, err = theme.AddThemeEditor()
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add theme: %v\n", err)
			os.Exit(1)
		}
		if slug != "" {
			if err := theme.ApplyTheme(slug); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to apply added theme: %v\n", err)
				os.Exit(1)
			}
		}

	case "edit":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: herdr-theme-picker edit <slug>")
			os.Exit(1)
		}
		if err := theme.EditTheme(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "delete":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: herdr-theme-picker delete <slug>")
			os.Exit(1)
		}
		if err := theme.DeleteTheme(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		// Fallback: if argument is a slug, treat as apply
		if theme.IsValidSlug(args[0]) {
			if err := theme.ApplyTheme(args[0]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		os.Exit(1)
	}
}

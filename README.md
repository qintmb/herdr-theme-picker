# herdr-theme-picker

<p align="center">
  <img alt="Screenshot 2026-09-01 at 07 13 31" src="https://github.com/user-attachments/assets/5100a157-1ed1-4659-be98-b846fb98681d" width=70%/>
</p>

Pick any color theme from [terminalcolors.com](https://terminalcolors.com)
and apply it to your [Herdr](https://herdr.dev) UI — with a single keybind,
`prefix+t`.

Herdr ships only a handful of built-in themes (`catppuccin`, `tokyo-night`,
`gruvbox`, `solarized`, `terminal`, …). This plugin unlocks hundreds of
palettes from terminalcolors.com: pick one from a popup, and its colors are
mapped into `[theme.custom]` in your `config.toml` and reloaded automatically.

> **Scope:** the plugin recolors **Herdr's UI chrome** (panels, sidebar,
> accents) through `[theme.custom]` **and** syncs the 16 ANSI terminal cell
> colors to your host terminal emulator live via OSC sequences. See
> [Terminal emulator sync](#terminal-emulator-sync) for optional per-emulator
> setup (e.g. Ghostty config persistence for new windows).

---
## Preview
<p align="center">
<video
  src="https://github.com/user-attachments/assets/866d7f05-216d-443e-93be-5edfd65a1312"
  autoplay
  muted
  loop
  playsinline
  width="80%">
</video>
</p>

---
## Features

- **`prefix+t`** → an fzf popup with a live truecolor swatch preview.
- **19 popular themes bundled offline** (Dracula, Gruvbox, Nord, Tokyo Night,
  Catppuccin, Solarized, One Dark, Everforest, Rosé Pine, Kanagawa, Ayu,
  GitHub, …).
- **Live-fetch** any other theme from terminalcolors.com on selection
  (needs internet), cached locally.
- **Live terminal color sync** — OSC 4/10/11 sequences update your terminal
  emulator's ANSI palette instantly, no restart needed. Works with any
  emulator that supports OSC color setting (Ghostty, iTerm2, WezTerm, …).
- **Automatic backup** of `config.toml` before writing.
- **Idempotent & reversible.**

---

## Compatibility

| | |
|---|---|
| Herdr | ≥ 0.8.0 |
| OS | macOS, Linux |
| Dependencies | `bash`, `curl`, [`fzf`](https://github.com/junegunn/fzf) |

`fzf` is usually already present if you use other Herdr plugins (file-picker,
termscope). If not: `brew install fzf` (macOS) or your distro's package.

---

## Install

### From GitHub (recommended)

Herdr installs plugins directly from a repository:

```bash
herdr plugin install qintmb/herdr-theme-picker
```

Herdr shows an install preview for review before proceeding (use `--yes` for
non-interactive install).

⚠️ **Important:** Herdr 0.8.2 does not automatically bind keys from plugin
manifests. After installing, add this to your `config.toml`:

```toml
[[keys.command]]
key = "prefix+t"
type = "plugin_action"
command = "herdr-theme-picker.open"
```

Then reload the running server:

```bash
herdr server reload-config
```

**This is the only path that receives updates.** Herdr tracks the plugin's
source repo and compares the manifest `version`, so `herdr plugin install`
is how you get future fixes and new themes (see [Updating](#updating)).

### Local checkout (development only)

Use this only when hacking on the plugin itself. A linked local checkout is
**not** tracked for updates — pull changes with `git` yourself.

```bash
git clone https://github.com/qintmb/herdr-theme-picker.git
cd herdr-theme-picker
herdr plugin link "$PWD"        # or: bash bin/install.sh
herdr server reload-config
```

`bin/install.sh` is a thin convenience wrapper around `herdr plugin link` +
`reload-config` for a local checkout.

---

## Updating

If you installed from GitHub, reinstall to pull the latest release:

```bash
herdr plugin install qintmb/herdr-theme-picker
herdr server reload-config
```

Herdr resolves the repo's default branch (`main`) and detects a new release
by a bumped manifest `version`. To pin a specific commit, tag, or branch,
pass `--ref`:

```bash
herdr plugin install qintmb/herdr-theme-picker --ref v0.1.0
```

A local linked checkout does not auto-update — `git pull` in that directory
and `herdr server reload-config`.

---

## Usage

Press **`prefix+t`** (Herdr's default prefix is `cmd+b`, so `cmd+b` then `t`).

- The fzf popup opens. The **search field is at the top** — just type to
  filter themes.
- The right pane previews the theme applied to a mock Herdr layout — `spaces`
  sidebar, a two-pane workspace with the active pane highlighted, tab bar, and
  a bottom agent pane with state colors.
- **Enter** → the theme is applied, Herdr reloads, and the popup closes.
- **Esc** → cancel.

The currently applied theme is marked with `✓` and floated to the top of the
list; the rest stay browsable so you can switch freely.

### Add your own theme (from inside the picker)

Two rows are pinned at the bottom of the list:

**`+ Add new theme…`** — press **Tab** (from anywhere) or **Enter** on that row to:

1. Open an editor with a ghostty-format template — **nvim** is used when installed, otherwise `$VISUAL`/`$EDITOR`, then `nano`/`vi`.
2. Paste a palette (e.g. terminalcolors.com → Download → **Ghostty**), save, and close.
3. Type a name — it's slugified, validated, saved, appended to the list, and applied immediately.

**`+ Add from clipboard…`** — Enter on this row skips the editor entirely: it reads a ghostty-format palette straight from your system clipboard (`pbpaste`/`wl-paste`/`xclip`/`xsel`), asks only for a name, then saves and applies. Copy a palette, open the picker, pick this row, type a name — done.

User-added themes are stored in `HERDR_PLUGIN_STATE_DIR/themes` (with their own
`index.txt`), so plugin updates never overwrite them. They appear in the list
marked `★`.

### Which palette line drives which part of Herdr

A ghostty palette has 18+ colors, but Herdr's theme uses a fixed set of tokens.
The preview pane shows a **legend** mapping each Herdr token to the palette line
it comes from — so if a scheme has a dominant bright red in `palette 13–15`,
you can see at a glance that the *selector/active-pane border* is `accent ←
palette 4`, not that red. `palette 10–15` are not used by the default mapping.

| Herdr token | from | Herdr role |
|---|---|---|
| accent / blue | palette 4 | active pane border / selector |
| red | palette 1 | error ✗ |
| green | palette 2 | success ✓ |
| yellow | palette 3 | warning |
| teal | palette 6 | now-playing |
| mauve / peach | palette 5 / 9 | accents |
| panel_bg / text | background / foreground | pane bg / text |
| surface0/1, overlay0/1, subtext0 | palette 0/7/8 | surfaces & dim text |

### Edit a custom theme (`ctrl-e`)

Select a `★` user theme and press **ctrl-e** to open the built-in token editor
(no external editor). It lists the editable Herdr tokens with their current hex
and origin; the preview pane shows the Herdr mockup **rebuilt live** as you
change colors. Enter on a token → type a new `#rrggbb` → the preview updates.
`✓ Save & apply` commits and applies; `esc` discards. Edits are stored as
per-theme override lines, so the derived mapping stays intact for everything you
didn't touch. (Bundled themes are read-only.)

### Delete a custom theme (`ctrl-d`)

Select a `★` user theme and press **ctrl-d**; confirm `y` at the prompt. Only
user-added themes can be deleted — bundled Herdr themes are refused.

Available themes are listed in `themes/index.txt`. Bundled ones work offline;
the rest are fetched on selection.

You can also trigger it without the keybind:

```bash
herdr plugin pane open --plugin herdr-theme-picker --entrypoint picker --placement popup --focus
```

---

## How it works

```
prefix+t → picker.sh (fzf)
             │  pick a slug
             ▼
          apply.sh <slug>
             │  1. resolve_palette      → themes/<slug>  or  fetch terminalcolors.com (cached)
             │  2. palette_to_tokens    → 16 [theme.custom] tokens
             │  3. write config.toml   (backup: config.toml.bak-YYYYMMDD)
             │  4. sync_terminal_colors → OSC 4/10/11 to outer PTY (live ANSI update)
             │                         → copy palette to ~/.config/ghostty/herdr-theme (if present)
             │  5. herdr server reload-config
             ▼
          Herdr UI + terminal emulator update simultaneously
```

Palettes are read in **ghostty format** (`background`, `foreground`,
`palette N=#hex`), then mapped to Herdr's UI tokens:

| Herdr token | palette source | | Herdr token | palette source |
|---|---|---|---|---|
| `panel_bg` | `background` | | `text` | `foreground` |
| `surface0` | `palette 0` | | `subtext0` | `palette 7` |
| `surface1` | `palette 8` | | `accent` / `blue` | `palette 4` |
| `surface_dim` | `background` −8% | | `red` | `palette 1` |
| `overlay0` | `palette 8` | | `green` | `palette 2` |
| `overlay1` | `palette 7` | | `yellow` | `palette 3` |
| `mauve` | `palette 5` | | `teal` | `palette 6` |
| `peach` | `palette 9` | | | |

---

## Terminal emulator sync

When you pick a theme, `apply.sh` automatically emits **OSC 4** (palette),
**OSC 10** (foreground), and **OSC 11** (background) sequences to the
terminal emulator that is hosting Herdr. This updates the live session's
ANSI colors instantly — no restart required.

Because Herdr is a terminal multiplexer, its internal panes intercept
`/dev/tty` writes before they reach the outer emulator. The plugin works
around this by locating the PTY slave that the Herdr client renders into
and writing there directly.

Any terminal emulator that supports OSC 4/10/11 benefits automatically —
Ghostty, iTerm2, WezTerm, Alacritty, and others. **No configuration is
needed** for the live sync to work.

### Ghostty: persist colors for new windows

The OSC approach updates the **current** window. New Ghostty windows will
still open with whatever colors your `~/.config/ghostty/config` defines.
To keep new windows in sync too, set up a one-time Ghostty config include:

**1.** Create the theme fragment file:

```bash
touch ~/.config/ghostty/herdr-theme
```

**2.** Add it to your Ghostty config (`~/.config/ghostty/config`):

```
config-file = herdr-theme
```

**3.** Restart Ghostty once.

From then on, every theme pick writes the palette to `herdr-theme` and
Ghostty picks it up on next launch. Combined with the live OSC sync, the
current window and all future windows stay consistent.

---

## Manifest

`herdr-plugin.toml` declares everything Herdr needs — the pane, the action,
and the keybind:

```toml
id = "herdr-theme-picker"
name = "Theme Picker"
version = "0.1.0"
min_herdr_version = "0.8.0"
platforms = ["linux", "macos"]

[[panes]]
id = "picker"
placement = "popup"
command = ["bash", "-c", "exec bash \"$HERDR_PLUGIN_ROOT/bin/picker.sh\""]

[[actions]]
id = "open"
contexts = ["workspace"]
command = ["bash", "-c", "exec \"${HERDR_BIN_PATH:-herdr}\" plugin pane open --plugin herdr-theme-picker --entrypoint picker --placement popup --focus"]

[[keys.command]]
key = "prefix+t"
type = "plugin_action"
command = "herdr-theme-picker.open"
description = "pick a color theme"
```

Herdr does not run commands through a shell, so each `command` is an explicit
argv array. Plugins reach Herdr via `HERDR_BIN_PATH` (or the socket API).

> **Note:** Herdr 0.8.2 does not bind keys declared in the manifest's
> `[[keys.command]]` block. The block above is kept as a declaration of intent,
> but you must also add the same block to your `config.toml` for the keybind to
> actually trigger.

---

## Customization

### Add an offline theme

Save a ghostty-format file at `themes/<slug>`, then register the slug in
`themes/index.txt`:

```
background = #1e1e2e
foreground = #cdd6f4
selection-background = #585b70
cursor-color = #f5e0dc
palette = 0=#45475a
palette = 1=#f38ba8
...
palette = 15=#a6adc8
```

Quick download from terminalcolors.com:

```bash
curl -fsSL "https://terminalcolors.com/downloads/ghostty/<slug>" -o "themes/<slug>"
echo "<slug>" >> themes/index.txt
```

Slugs may only contain `[a-z0-9-]` (validated to prevent path/URL injection).

### Change the color mapping

Edit `bin/map.sh` → `palette_to_tokens`. Each `printf` line maps one Herdr
token to a palette source. `bin/lib.sh` provides `darken_hex <#hex> <percent>`
for derived shades.

### Change the keybind

Edit `key = "prefix+t"` in your `config.toml` to another combination
(e.g. `prefix+shift+t`), then `herdr server reload-config`. The manifest's
`[[keys.command]]` block is kept as a declaration of intent, but Herdr 0.8.2
ignores it.

---

## Test

```bash
bash tests/run.sh
```

Runs assert-based self-checks (no framework): palette mapping, `darken_hex`,
slug validation, idempotent `[theme.custom]` writing, and swatch rendering.

---

## Uninstall

```bash
herdr plugin uninstall herdr-theme-picker     # installed from GitHub
herdr plugin unlink    herdr-theme-picker     # linked local checkout (or: bash bin/uninstall.sh)
herdr server reload-config
```

If you added the `prefix+t` keybind block to your `config.toml` manually (see
[Install](#install)), remove it there too — Herdr 0.8.2 does not bind keys from
plugin manifests. Your `[theme.custom]` block and its backups remain in
`config.toml` until you edit them out.

---

## License

MIT — see [LICENSE](LICENSE).

`herdr` · `herdr-plugin`

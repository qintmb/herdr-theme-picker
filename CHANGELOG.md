# Changelog

All notable changes to this project are documented here. Format follows
[Keep a Changelog](https://keepachangelog.com); versions follow semver and
match the `version` field in `herdr-plugin.toml`.

## [Unreleased]

### Added

- **Back to Herdr's own theme.** The picker now opens with an
  `↩ herdr default` row, which drops the plugin's `[theme.custom]` tokens so
  Herdr falls back to its built-in `theme.name` again. The label names the
  theme you land on (`↩ herdr default (built-in: vesper)`), and `catppuccin`
  is Herdr's own default when the key is unset. It also resets the host
  terminal's palette with OSC 104/110/111 (the counterpart of the OSC 4/10/11
  that applying sends), clears the applied marker, restores
  `~/.config/ghostty/herdr-theme` from the pre-plugin copy the first apply now
  stashes, and reloads. Enter-only, so Tab still means "add new theme" from any
  row; `herdr-theme-picker reset` does the same from the command line. The row
  hands control back to Herdr's own settings rather than changing `theme.name`,
  and it keeps a hand-written `[theme.custom.light]` / `[theme.custom.dark]`
  table — those are the user's own config, so the preview warns about them
  instead. Implemented in both the Bash picker and the Go port.

### Fixed

- **`config.toml` no longer accumulates blank lines.** Every apply appended a
  blank line in front of its `[theme.custom]` block and never removed the
  previous run's, so a config cycled through the picker grew an ever-longer gap
  above the block. The Bash writer now collapses the trailing blank run before
  appending, matching what the Go port already did.

---

## [0.9.0] — 2026-10-13

### Added

- **Windows support (beta)** via a Go port in `go/`, contributed by
  [@wapenshaw](https://github.com/wapenshaw) (#7, #8). It sits alongside the
  Bash implementation: Linux/macOS keep the `bin/` picker, and Windows runs
  `herdr-theme-picker.exe`, built from `go/` on install. The `open` action runs
  `./herdr-theme-picker` (a small Bash dispatcher on Linux/macOS, the `.exe` on
  Windows), so the `herdr-theme-picker.open` keybind works on every platform.
  The bundled theme set grows from 19 to 112; the Bash picker sees the new
  themes too. Note: Herdr ships Windows plugin manifests on its preview channel
  and describes Windows plugin exec as best-effort, so the Go path is beta. The
  Go port does not download themes from terminalcolors.com; Windows uses the
  bundled set plus clipboard/editor imports.

## [0.8.2] — 2026-10-05

### Fixed

- **Symlinked `config.toml` preserved** (#3). Applying a theme used to `mv`
  the temp file over the path, replacing a symlink with a regular file and
  disconnecting dotfile setups that keep `config.toml` in a repo. The target
  is now resolved (via `readlink -f`) and the atomic rename lands on the
  target, so the link survives and receives the `[theme.custom]` block.
  A chain longer than 40 hops is rejected rather than followed forever.
- **Outer PTY detection on Linux** (#5). `sync_terminal_colors` only skipped
  `??` (macOS/BSD) when filtering `ps -eo tty,comm`, but Linux reports `?`
  for daemons without a controlling TTY — so it matched the `herdr server`
  daemon first, resolved to the nonexistent `/dev/?`, and silently emitted no
  OSC 4/10/11 sequences on any Linux system. The filter now skips both.

## [0.8.1] — 2026-09-01

- **Precise pane borders.** The preview's box-drawing rules no longer get
  ragged: `_padr` was padding by *bytes* (`printf %-*. Ns`), so rows containing
  `─ │ ○ ✓` came out 2 columns short at any font size. Padding now counts
  characters; the UTF-8 locale is forced when the inherited one is byte-based.
- **Closed inactive-pane box.** The dim pane is now a full rectangle with a
  gray (`overlay0`) border matching the accent border of the focused pane.
- **Sidebar spaces renamed** `AGENT/LINUX → space1/space2`, each showing its
  `main` workspace — no theme-specific workspace names in the mock.
- **Gray footer text.** `new/menu/agents/priority` now render in `overlay0`
  (they were one notch too bright at `subtext0`).
- **Selected agent row.** The AGENT panel shows a highlighted agent (accent
  name on `active_row_bg`) beside a dim unselected one; legend adds
  `overlay0`/`surface_dim` rows and no longer truncates token names.

## [0.8.0] — 2026-08-31

- **Workspace-faithful theme preview.** The fzf preview pane now renders a mock
  of a real Herdr workspace in the candidate theme's colors: focused pane with
  its full accent border, active vs inactive tabs, spaces sidebar (selected
  workspace row, `new/menu`, `agents/priority` footer), powerline prompt,
  a text-selection run, and the AGENT panel with a state dot and agent name.
  The preview now answers "what will my workspace look like with this theme
  applied" instead of showing an abstract swatch.

## [0.7.0] — 2026-08-30

- **Live terminal color sync** (contributed by @neospeed83 in #2). `apply.sh`
  now emits OSC 4/10/11 sequences to the terminal emulator hosting Herdr,
  updating the 16 ANSI palette colors instantly on every theme pick — no
  terminal restart required. Works with any emulator that supports OSC color
  setting (Ghostty, iTerm2, WezTerm, Alacritty, …). The plugin locates the
  outer PTY slave via `ps` because Herdr's multiplexer intercepts
  `/dev/tty` writes.
- **Optional Ghostty new-window persistence.** When `~/.config/ghostty/herdr-theme`
  exists and is `config-file`d from your Ghostty config, the plugin mirrors
  every picked palette there so new Ghostty windows open with the same theme.
- **README fix.** Herdr 0.8.2 does not bind keys declared in a plugin's
  manifest `[[keys.command]]` block. README now instructs users to add the
  `prefix+t` block to their `config.toml` manually (Install, Uninstall,
  Change the keybind, Manifest sections updated).

## [0.6.0] — 2026-08-21

- **Legend in the preview.** The preview pane now shows which Herdr token each
  palette line drives (`accent ← palette 4`, `red ← palette 1`, …) with a color
  chip, so it's clear why a scheme's dominant color isn't necessarily the
  selector. Documents that `palette 10–15` are unused by the mapping.
- **Edit a custom theme (`ctrl-e`).** Built-in fzf token editor (no external
  editor): pick a Herdr token, type a new hex, and the preview rebuilds live to
  show what changes. `✓ Save & apply` commits; `esc` discards. Edits persist as
  per-theme `# hpick-override:` lines layered over the derived mapping.
- **Delete a custom theme (`ctrl-d`).** Confirm `y`/`N`; only ★ user themes can
  be deleted — bundled themes are refused.
- `map.sh` gains `apply_overrides` + token origin/role metadata; new
  `bin/edit.sh`, `bin/delete.sh`, `lib.sh:is_user_theme`; new
  `tests/test_edit.sh`, `tests/test_delete.sh`, override cases in
  `tests/test_map.sh`.

## [0.5.0] — 2026-08-21

- **Editor prefers nvim.** The `+ Add new theme…` flow now opens **nvim** when
  it's installed (falling back to `$VISUAL`/`$EDITOR`, then `nano`/`vi`), so
  pasting/editing a palette is comfortable instead of dropping into raw `vi`.
- **`+ Add from clipboard…` row.** A second add row skips the editor entirely:
  reads a ghostty-format palette from the system clipboard
  (`pbpaste`/`wl-paste`/`xclip`/`xsel`), prompts only for a name, then saves,
  indexes, and applies. Copy → open picker → pick → name → done.
- Refactored `add.sh` around a shared `_finalize` (validate → name → save →
  index) with editor and clipboard entry points; prompts are testable via
  `TTY_IN`. Added `_finalize`/`pick_editor` self-checks to `test_add.sh`.

## [0.4.0] — 2026-08-19

- **Add themes from inside the picker.** A `+ Add new theme…` row is pinned at
  the bottom; press **Tab** (from anywhere) or **Enter** on that row to open
  your `$EDITOR` with a ghostty-format template. Paste a palette, save, name
  it, and it's validated, saved, appended to the list, and applied.
- User-added themes live in `HERDR_PLUGIN_STATE_DIR/themes` + `index.txt`, so
  plugin updates never overwrite them; they show in the list marked `★`.
- `apply.sh` now fails safely (config untouched) on an unresolved or invalid
  palette instead of writing an empty block.
- All user-facing strings are English.

## [0.3.0] — 2026-08-19

- **Fix:** reopening the picker no longer hides the other themes. The applied
  theme is floated to the top with `✓` instead of being used as a search
  query, so the full list stays browsable.
- Search field moved to the **top** (`--layout=reverse`) with a clear
  `Search themes:` prompt and a "type to search" hint, so it's obvious you
  can type to filter.
- Preview redesigned into a fuller Herdr layout mock: left `spaces` sidebar
  (space header, cpu·ram, now-playing, second space), a two-pane workspace
  with the active pane highlighted vs. a dim inactive pane, tab bar
  (`WORKSPACE`/`Files`), and a bottom agent pane with state dots.

## [0.2.0] — 2026-08-19

- Preview now mocks the actual Herdr UI in the theme's colors — workspace
  header, agent panes (working/done/blocked/idle state colors), active/inactive
  tab bar, and pane body — instead of a bare color list.
- All picker UI text is English; compact layout with an inline header hint.
- Popup closes automatically after applying (no keypress needed).
- Applied theme is remembered (`HERDR_PLUGIN_STATE_DIR/applied`), marked with
  `✓`, and the cursor lands on it when the picker reopens.

## [0.1.0] — 2026-08-19

Initial release.

- `prefix+t` fzf theme picker with truecolor swatch preview.
- 19 bundled offline themes + live-fetch from terminalcolors.com (cached).
- Ghostty-format palette → 16 Herdr `[theme.custom]` tokens.
- Idempotent config write with dated backup + `herdr server reload-config`.
- Keybind declared in manifest (`[[keys.command]]`).

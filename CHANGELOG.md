# Changelog

All notable changes to this project are documented here. Format follows
[Keep a Changelog](https://keepachangelog.com); versions follow semver and
match the `version` field in `herdr-plugin.toml`.

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

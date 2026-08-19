# Changelog

All notable changes to this project are documented here. Format follows
[Keep a Changelog](https://keepachangelog.com); versions follow semver and
match the `version` field in `herdr-plugin.toml`.

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

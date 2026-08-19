# Changelog

All notable changes to this project are documented here. Format follows
[Keep a Changelog](https://keepachangelog.com); versions follow semver and
match the `version` field in `herdr-plugin.toml`.

## [0.1.0] — 2026-08-19

Initial release.

- `prefix+t` fzf theme picker with truecolor swatch preview.
- 19 bundled offline themes + live-fetch from terminalcolors.com (cached).
- Ghostty-format palette → 16 Herdr `[theme.custom]` tokens.
- Idempotent config write with dated backup + `herdr server reload-config`.
- Keybind declared in manifest (`[[keys.command]]`).

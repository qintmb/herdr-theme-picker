#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

tmp="$(mktemp -d)"
export HERDR_PLUGIN_STATE_DIR="$tmp"
source bin/add.sh

fail=0
ck() { if [ "$1" != "$2" ]; then echo "FAIL $3: got '$1' want '$2'"; fail=1; fi; }

# slugify
ck "$(slugify 'My Cool Theme')" "my-cool-theme" "spaces -> dashes"
ck "$(slugify 'Rosé_Pine 2')" "ros-pine-2" "underscore + strip accents/keep digits"
ck "$(slugify '  Trim -- Me  ')" "trim-me" "collapse + trim dashes"
ck "$(slugify '///')" "" "punctuation only -> empty"

# pick_editor prefers nvim when present, else falls back.
if command -v nvim >/dev/null 2>&1; then
  ck "$(pick_editor)" "nvim" "prefers nvim"
else
  # with no nvim, VISUAL wins
  ck "$(VISUAL=myed EDITOR=other pick_editor)" "myed" "VISUAL fallback"
fi

# _finalize: validate + name + save + index, driving prompts over stdin.
export TTY_IN=/dev/stdin
mkdir -p "$USER_THEMES_DIR"
palette="$(mktemp)"; cp tests/fixtures/dracula-default "$palette"
slug="$(printf 'Paste Me\n' | _finalize "$palette")"
ck "$slug" "paste-me" "finalize returns slug"
[ -f "$USER_THEMES_DIR/paste-me" ] || { echo "FAIL finalize saved file"; fail=1; }
grep -qx "paste-me" "$USER_INDEX" || { echo "FAIL finalize indexed slug"; fail=1; }

# _finalize rejects a non-palette body.
bad="$(mktemp)"; printf 'not a palette\n' > "$bad"
if printf 'X\n' | _finalize "$bad" >/dev/null 2>&1; then
  echo "FAIL finalize should reject non-palette"; fail=1
fi

# save flow: seed user index + a theme file the way add_theme would, then
# confirm fetch.sh resolves it and it lands in the user index.
cp tests/fixtures/dracula-default "$USER_THEMES_DIR/my-theme"
touch "$USER_INDEX"; echo "my-theme" >> "$USER_INDEX"

source bin/fetch.sh
p="$(resolve_palette my-theme)"
[ "$p" = "$USER_THEMES_DIR/my-theme" ] || { echo "FAIL fetch resolves user theme: $p"; fail=1; }
grep -qx "my-theme" "$USER_INDEX" || { echo "FAIL index has slug"; fail=1; }

rm -rf "$tmp"
[ "$fail" = 0 ] && echo "PASS test_add" || exit 1

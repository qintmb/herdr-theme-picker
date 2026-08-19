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

# save flow: seed user index + a theme file the way add_theme would, then
# confirm fetch.sh resolves it and it lands in the user index.
mkdir -p "$USER_THEMES_DIR"
cp tests/fixtures/dracula-default "$USER_THEMES_DIR/my-theme"
touch "$USER_INDEX"; echo "my-theme" >> "$USER_INDEX"

source bin/fetch.sh
p="$(resolve_palette my-theme)"
[ "$p" = "$USER_THEMES_DIR/my-theme" ] || { echo "FAIL fetch resolves user theme: $p"; fail=1; }
grep -qx "my-theme" "$USER_INDEX" || { echo "FAIL index has slug"; fail=1; }

rm -rf "$tmp"
[ "$fail" = 0 ] && echo "PASS test_add" || exit 1

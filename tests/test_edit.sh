#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

tmp="$(mktemp -d)"
export HERDR_PLUGIN_STATE_DIR="$tmp"
source bin/edit.sh   # build_list, set_override, EDIT_TOKENS, map/lib

fail=0
mkdir -p "$USER_THEMES_DIR"
cp tests/fixtures/dracula-default "$USER_THEMES_DIR/geohot"
echo "geohot" >> "$USER_INDEX"
src="$USER_THEMES_DIR/geohot"

# build_list yields one row per editable token + the Save row
rows="$(build_list "$src")"
printf '%s\n' "$rows" | grep -q '^accent ' || { echo "FAIL build_list missing accent"; fail=1; }
printf '%s\n' "$rows" | grep -q '✓ Save & apply' || { echo "FAIL build_list missing save row"; fail=1; }
# accent origin shown
printf '%s\n' "$rows" | grep '^accent ' | grep -q 'palette 4' || { echo "FAIL accent origin"; fail=1; }

# set_override writes a hpick-override line, and map.sh honors it
set_override "$src" accent "#fa390f"
grep -qE '^#[[:space:]]*hpick-override:[[:space:]]*accent=#fa390f' "$src" \
  || { echo "FAIL override line not written"; fail=1; }
newaccent="$(palette_to_tokens "$src" | grep '^accent=' | cut -d= -f2-)"
[ "$newaccent" = "#fa390f" ] || { echo "FAIL override not applied: $newaccent"; fail=1; }

# editing the same token again replaces (no duplicate)
set_override "$src" accent "#123456"
c="$(grep -cE 'hpick-override:[[:space:]]*accent=' "$src")"
[ "$c" = 1 ] || { echo "FAIL duplicate override lines: $c"; fail=1; }

# is_user_theme guard: bundled refused
if is_user_theme "dracula-default"; then echo "FAIL bundled treated as user"; fail=1; fi

rm -rf "$tmp"
[ "$fail" = 0 ] && echo "PASS test_edit" || exit 1

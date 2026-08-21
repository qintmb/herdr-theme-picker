#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

tmp="$(mktemp -d)"
export HERDR_PLUGIN_STATE_DIR="$tmp"
source bin/delete.sh   # brings in lib.sh (dirs, is_user_theme)

fail=0
mkdir -p "$USER_THEMES_DIR"
cp tests/fixtures/dracula-default "$USER_THEMES_DIR/geohot"
echo "geohot" >> "$USER_INDEX"

# guard: bundled theme is refused (dracula-default exists in themes/index.txt)
if delete_theme "dracula-default" >/dev/null 2>&1; then
  echo "FAIL should refuse bundled theme"; fail=1
fi
[ -f "$USER_THEMES_DIR/geohot" ] || { echo "FAIL user theme vanished early"; fail=1; }

# 'N' keeps the theme
export TTY_IN=/dev/stdin
if printf 'n\n' | delete_theme "geohot" >/dev/null 2>&1; then
  echo "FAIL N should return non-zero (kept)"; fail=1
fi
[ -f "$USER_THEMES_DIR/geohot" ] || { echo "FAIL N deleted the theme"; fail=1; }
grep -qx "geohot" "$USER_INDEX" || { echo "FAIL N removed from index"; fail=1; }

# 'y' deletes file + index line
printf 'y\n' | delete_theme "geohot" >/dev/null 2>&1 || true
[ -f "$USER_THEMES_DIR/geohot" ] && { echo "FAIL Y did not delete file"; fail=1; }
grep -qx "geohot" "$USER_INDEX" 2>/dev/null && { echo "FAIL Y left index line"; fail=1; }

rm -rf "$tmp"
[ "$fail" = 0 ] && echo "PASS test_delete" || exit 1

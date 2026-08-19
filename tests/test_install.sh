#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source bin/install.sh

tmp="$(mktemp -d)"; cfg="$tmp/config.toml"
printf '[keys]\nprefix = "cmd+b"\n' > "$cfg"

fail=0
add_keybind "$cfg"
grep -q 'command = "herdr-theme-picker.open"' "$cfg" || { echo "FAIL keybind added"; fail=1; }
grep -q 'key = "prefix+t"' "$cfg" || { echo "FAIL key line"; fail=1; }

# idempoten
add_keybind "$cfg"
n="$(grep -c 'command = "herdr-theme-picker.open"' "$cfg")"
[ "$n" = 1 ] || { echo "FAIL dup keybind: $n"; fail=1; }

# remove
remove_keybind "$cfg"
grep -q 'herdr-theme-picker.open' "$cfg" && { echo "FAIL keybind not removed"; fail=1; } || true

[ "$fail" = 0 ] && echo "PASS test_install" || exit 1

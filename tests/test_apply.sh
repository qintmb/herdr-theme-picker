#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source bin/apply.sh

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cfg="$tmp/config.toml"
printf '[keys]\nprefix = "cmd+b"\n\n[theme]\nname = "solarized"\n' > "$cfg"

tokens=$'panel_bg=#282a36\ntext=#f8f8f2\nred=#ff5555'
write_custom_block "$cfg" "$tokens"

fail=0
grep -q '^\[theme.custom\]' "$cfg" || { echo "FAIL block added"; fail=1; }
grep -q '^panel_bg = "#282a36"' "$cfg" || { echo "FAIL panel_bg quoted"; fail=1; }
grep -q '^name = "solarized"' "$cfg" || { echo "FAIL preserved theme.name"; fail=1; }

# idempoten: tulis lagi, tak duplikat blok
write_custom_block "$cfg" "$tokens"
n="$(grep -c '^\[theme.custom\]' "$cfg")"
[ "$n" = 1 ] || { echo "FAIL duplicate block: $n"; fail=1; }

# Atomic writes must update a symlink's target without replacing the link.
absolute_target="$tmp/absolute-target.toml"
absolute_link="$tmp/absolute-link.toml"
printf '[theme]\nname = "solarized"\n' > "$absolute_target"
ln -s "$absolute_target" "$absolute_link"
write_custom_block "$absolute_link" "$tokens"
[ -L "$absolute_link" ] || { echo "FAIL absolute symlink replaced"; fail=1; }
grep -q '^panel_bg = "#282a36"' "$absolute_target" \
  || { echo "FAIL absolute symlink target not updated"; fail=1; }

mkdir "$tmp/relative"
relative_target="$tmp/relative/target.toml"
relative_link="$tmp/relative/link.toml"
printf '[theme]\nname = "solarized"\n' > "$relative_target"
ln -s target.toml "$relative_link"
write_custom_block "$relative_link" "$tokens"
[ -L "$relative_link" ] || { echo "FAIL relative symlink replaced"; fail=1; }
grep -q '^panel_bg = "#282a36"' "$relative_target" \
  || { echo "FAIL relative symlink target not updated"; fail=1; }

[ "$fail" = 0 ] && echo "PASS test_apply" || exit 1

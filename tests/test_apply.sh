#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source bin/apply.sh

tmp="$(mktemp -d)"
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

# verify outer_tty awk filter skips both '?' (Linux) and '??' (macOS/BSD)
mock_ps="? herdr
?? herdr
pts/1 herdr"
detected_tty=$(printf "%s\n" "$mock_ps" | awk '$2=="herdr" && $1!="?" && $1!="??" {print "/dev/"$1; exit}')
[ "$detected_tty" = "/dev/pts/1" ] || { echo "FAIL outer_tty detection: $detected_tty"; fail=1; }

[ "$fail" = 0 ] && echo "PASS test_apply" || exit 1

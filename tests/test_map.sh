#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source bin/map.sh

out="$(palette_to_tokens tests/fixtures/dracula-default)"
get() { printf '%s\n' "$out" | grep "^$1=" | cut -d= -f2-; }
fail=0
check() { if [ "$(get "$1")" != "$2" ]; then echo "FAIL $1: got '$(get "$1")' want '$2'"; fail=1; fi; }

check panel_bg "#282a36"
check text "#f8f8f2"
check red "#ff5555"
check green "#50fa7b"
check yellow "#f1fa8c"
check blue "#bd93f9"
check accent "#bd93f9"
check teal "#8be9fd"
check mauve "#ff79c6"
check surface0 "#21222c"
check surface1 "#6272a4"
check overlay0 "#6272a4"
check overlay1 "#f8f8f2"
check subtext0 "#f8f8f2"
check peach "#ff6e6e"

# new herdr tokens
check selection_bg "#44475a"
check active_row_bg "#21222c"
if [ "$(get sidebar_bg)" = "#282a36" ]; then echo "FAIL sidebar_bg not darkened"; fail=1; fi

if [ "$(get surface_dim)" = "#282a36" ]; then echo "FAIL surface_dim not darkened"; fail=1; fi

n="$(printf '%s\n' "$out" | grep -c '=')"
[ "$n" = 19 ] || { echo "FAIL token count: $n"; fail=1; }

# hpick-override lines win over the derived mapping.
ov="$(mktemp)"; cat tests/fixtures/dracula-default > "$ov"
printf '# hpick-override: accent=#fa390f\n# hpick-override: red=#ed4b19\n' >> "$ov"
oout="$(palette_to_tokens "$ov")"
oget() { printf '%s\n' "$oout" | grep "^$1=" | cut -d= -f2-; }
[ "$(oget accent)" = "#fa390f" ] || { echo "FAIL override accent: $(oget accent)"; fail=1; }
[ "$(oget red)" = "#ed4b19" ] || { echo "FAIL override red: $(oget red)"; fail=1; }
# non-overridden token unchanged
[ "$(oget blue)" = "#bd93f9" ] || { echo "FAIL override leaked to blue: $(oget blue)"; fail=1; }
# still 19 tokens after overrides
on="$(printf '%s\n' "$oout" | grep -c '=')"
[ "$on" = 19 ] || { echo "FAIL override token count: $on"; fail=1; }
rm -f "$ov"

[ "$fail" = 0 ] && echo "PASS test_map" || exit 1

#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source bin/lib.sh

fail=0
check() { if [ "$1" != "$2" ]; then echo "FAIL $3: got '$1' want '$2'"; fail=1; fi; }

# darken: setiap kanal turun ~10%
check "$(darken_hex '#282a36' 10)" "#242530" "darken dracula bg"
check "$(darken_hex '#000000' 50)" "#000000" "darken black clamps"
check "$(darken_hex '#ffffff' 100)" "#000000" "darken 100 = black"

# slug valid
if is_valid_slug "dracula-default"; then :; else echo "FAIL slug ok"; fail=1; fi
if is_valid_slug "../../etc/passwd"; then echo "FAIL slug reject path"; fail=1; fi
if is_valid_slug "a b"; then echo "FAIL slug reject space"; fail=1; fi

[ "$fail" = 0 ] && echo "PASS test_lib" || exit 1
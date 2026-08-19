#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source bin/fetch.sh

fail=0
# bundle resolve tanpa network
p="$(resolve_palette dracula-default)"
[ -f "$p" ] && grep -q '^palette = 0=' "$p" || { echo "FAIL bundle resolve"; fail=1; }

# slug invalid ditolak (subshell agar exit dari die tidak membunuh test)
if (resolve_palette "../etc") 2>/dev/null; then echo "FAIL slug guard"; fail=1; fi

[ "$fail" = 0 ] && echo "PASS test_fetch" || exit 1

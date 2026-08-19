#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source bin/map.sh

fail=0
n=0
for f in themes/*; do
  [ "$(basename "$f")" = "index.txt" ] && continue
  if palette_to_tokens "$f" >/dev/null 2>&1; then n=$((n+1)); else echo "FAIL invalid bundle: $f"; fail=1; fi
done
[ "$n" -ge 15 ] || { echo "FAIL bundle count: $n (<15)"; fail=1; }
[ -f themes/index.txt ] && grep -q '^dracula-default$' themes/index.txt || { echo "FAIL index.txt"; fail=1; }

[ "$fail" = 0 ] && echo "PASS test_bundle ($n themes)" || exit 1

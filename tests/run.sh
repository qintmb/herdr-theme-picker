#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "$0")/.."
rc=0
for t in tests/test_*.sh; do
  echo "== $t =="
  bash "$t" || rc=1
done
[ "$rc" = 0 ] && echo "ALL PASS" || echo "SOME FAILED"
exit "$rc"

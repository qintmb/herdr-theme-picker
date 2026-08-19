#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "$0")/.."
# run under a temp state dir so tests never touch the user's real state
export HERDR_PLUGIN_STATE_DIR="$(mktemp -d)"
trap 'rm -rf "$HERDR_PLUGIN_STATE_DIR"' EXIT
rc=0
for t in tests/test_*.sh; do
  echo "== $t =="
  bash "$t" || rc=1
done
[ "$rc" = 0 ] && echo "ALL PASS" || echo "SOME FAILED"
exit "$rc"

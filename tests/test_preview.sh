#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
export HERDR_PLUGIN_ROOT="$PWD"
source bin/picker.sh

out="$(swatch dracula-default)"
fail=0
# truecolor bg escape present (48;2)
printf '%s' "$out" | grep -q $'\033\[48;2;' || { echo "FAIL no truecolor bg"; fail=1; }
# dracula panel_bg #282a36 -> rgb 40;42;54 rendered as a cell bg
printf '%s' "$out" | grep -q '48;2;40;42;54' || { echo "FAIL panel_bg rgb missing"; fail=1; }
# UI mock labels present
printf '%s' "$out" | grep -q 'Shell' || { echo "FAIL no tab row"; fail=1; }
printf '%s' "$out" | grep -q 'AGENT' || { echo "FAIL no agent header"; fail=1; }
printf '%s' "$out" | grep -q 'claude' || { echo "FAIL no agent name"; fail=1; }

[ "$fail" = 0 ] && echo "PASS test_preview" || exit 1

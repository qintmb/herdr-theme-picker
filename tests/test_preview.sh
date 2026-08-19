#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
export HERDR_PLUGIN_ROOT="$PWD"
source bin/picker.sh

out="$(swatch dracula-default)"
fail=0
# memuat kode escape truecolor (48;2)
printf '%s' "$out" | grep -q $'\033\[48;2;' || { echo "FAIL no truecolor bg in swatch"; fail=1; }
# menyebut beberapa hex
printf '%s' "$out" | grep -q '#282a36' || { echo "FAIL bg hex missing"; fail=1; }

[ "$fail" = 0 ] && echo "PASS test_preview" || exit 1

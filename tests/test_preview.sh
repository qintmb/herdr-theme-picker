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
# spaces sidebar lists space1/space2, each with its 'main' workspace
printf '%s' "$out" | grep -q 'space1' || { echo "FAIL no space1 row"; fail=1; }
printf '%s' "$out" | grep -q 'space2' || { echo "FAIL no space2 row"; fail=1; }

# Both pane boxes are closed rectangles: every ┌ has a matching ┐/└/┘, and the
# top/bottom rules are the same width as the body rows they enclose.
plain="$(printf '%s' "$out" | LC_ALL=C sed 's/\x1b\[[0-9;]*m//g')"
for g in '┌' '┐' '└' '┘'; do
  n="$(printf '%s' "$plain" | grep -o "$g" | wc -l | tr -d ' ')"
  [ "$n" = 2 ] || { echo "FAIL expected 2 '$g' (active+inactive pane), got $n"; fail=1; }
done
# Crop past the sidebar (col 20 onward — the mock's sidebar is 18 cols + space)
# so box widths compare across only the pane area. Character-aware (perl -CSD):
# awk length counts bytes, which miscounts ─ ○ │ as 3.
crop() { perl -CSD -pe 'chomp; $_ = substr($_, 19) . "\n"'; }
clen() { perl -CSD -ne 'chomp; print length($_), "\n"'; }
topw="$(printf '%s\n' "$plain" | grep -n '┌' | head -1 | cut -d: -f1)"
botw="$(printf '%s\n' "$plain" | grep -n '└' | head -1 | cut -d: -f1)"
lt="$(printf '%s\n' "$plain" | sed -n "${topw}p" | crop | clen)"
lb="$(printf '%s\n' "$plain" | sed -n "${botw}p" | crop | clen)"
[ "$lt" = "$lb" ] || { echo "FAIL box top ($lt cols) != bottom ($lb cols)"; fail=1; }
# body rows between the rules must match that width too (no ragged right edge)
for ((r=topw+1; r<botw; r++)); do
  lr="$(printf '%s\n' "$plain" | sed -n "${r}p" | crop | clen)"
  [ "$lr" = "$lt" ] || { echo "FAIL row $r width $lr != box width $lt"; fail=1; }
done

# Inactive pane border uses overlay0 (dracula palette 8 #6272a4 -> 98;114;164),
# distinct from the accent border on the focused pane.
printf '%s' "$out" | grep -q '38;2;98;114;164m│' || { echo "FAIL inactive pane has no gray border"; fail=1; }

[ "$fail" = 0 ] && echo "PASS test_preview" || exit 1

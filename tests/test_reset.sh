#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source bin/lib.sh

fail=0
check() { if [ "$1" != "$2" ]; then echo "FAIL $3: got '$1' want '$2'"; fail=1; fi; }

tmp="$(mktemp -d)"

# --- base_theme: reads [theme] name, not the [theme.custom] block -----------
cfg="$tmp/base.toml"
printf '[keys]\nprefix = "cmd+b"\n\n[theme]\nname = "vesper"\nauto_switch = false\n\n[theme.custom]\npanel_bg = "#282828"\n' > "$cfg"
check "$(base_theme "$cfg")" "vesper" "base_theme reads name"
# a name belonging to another section must not leak in
printf '[other]\nname = "nope"\n\n[theme]\nname = "nord"\n' > "$cfg"
check "$(base_theme "$cfg")" "nord" "base_theme ignores other sections"
# unset / empty name falls back to herdr's own default
printf '[theme]\nauto_switch = false\n' > "$cfg"
check "$(base_theme "$cfg")" "catppuccin" "base_theme fallback"
: > "$cfg"
check "$(base_theme "$cfg")" "catppuccin" "base_theme empty config"

# --- remove_custom_block: exact inverse of write_custom_block ---------------
full="$tmp/full.toml"
printf '[theme]\nname = "nord"\n\n[theme.custom]\npanel_bg = "#111111"\ntext = "#eeeeee"\n\n[ui]\nstatus_indicators = "dots"\n' > "$full"
if ! has_custom_block "$full"; then echo "FAIL has_custom_block false before removal"; fail=1; fi
remove_custom_block "$full"
grep -q '^\[theme\.custom\]' "$full" && { echo "FAIL block survived"; fail=1; }
grep -q 'panel_bg' "$full" && { echo "FAIL token line survived"; fail=1; }
grep -q '^name = "nord"' "$full" || { echo "FAIL theme.name eaten"; fail=1; }
grep -q '^\[ui\]' "$full" || { echo "FAIL following section eaten"; fail=1; }
grep -q 'status_indicators' "$full" || { echo "FAIL following section body eaten"; fail=1; }
if has_custom_block "$full"; then echo "FAIL has_custom_block true after removal"; fail=1; fi
# second run must not eat the rest of the file
remove_custom_block "$full"
grep -q '^\[theme\]' "$full" || { echo "FAIL theme section eaten on re-run"; fail=1; }
grep -q '^\[ui\]' "$full" || { echo "FAIL ui section eaten on re-run"; fail=1; }

# block at EOF with a trailing blank run, the shape apply.sh leaves behind
tail_run="$tmp/tail.toml"
printf '[theme]\nname = "nord"\n\n\n\n[theme.custom]\npanel_bg = "#111111"\n' > "$tail_run"
remove_custom_block "$tail_run"; tidy_trailing_blanks "$tail_run"
check "$(tail -1 "$tail_run")" 'name = "nord"' "block at EOF leaves theme.name last"
check "$(wc -l < "$tail_run" | tr -d ' ')" "2" "trailing blank run collapsed"

# a hand-written light/dark subsection is the user's own config, not the
# plugin's — it survives, and the picker warns about it before you reset
sub="$tmp/sub.toml"
printf '[theme]\nname = "nord"\n\n[theme.custom]\npanel_bg = "#111111"\n\n[theme.custom.dark]\npanel_bg = "#000000"\n\n[ui]\nstatus_indicators = "dots"\n' > "$sub"
remove_custom_block "$sub"
grep -q '^\[theme\.custom\]' "$sub" && { echo "FAIL block survived"; fail=1; }
grep -q '"#111111"' "$sub" && { echo "FAIL plugin token survived"; fail=1; }
grep -q '^\[theme\.custom\.dark\]' "$sub" || { echo "FAIL user subsection removed"; fail=1; }
grep -q '"#000000"' "$sub" || { echo "FAIL user subsection body removed"; fail=1; }
grep -q '^\[ui\]' "$sub" || { echo "FAIL ui eaten with subsections"; fail=1; }

# --- tidy_trailing_blanks ---------------------------------------------------
tidy="$tmp/tidy.toml"
printf '[theme]\nname = "nord"\n\n\n\n\n' > "$tidy"
tidy_trailing_blanks "$tidy"
check "$(wc -l < "$tidy" | tr -d ' ')" "2" "tidy drops trailing blanks"
check "$(tail -1 "$tidy")" 'name = "nord"' "tidy keeps last real line"
# no trailing blanks → untouched
printf '[a]\nx = 1\n' > "$tidy"
tidy_trailing_blanks "$tidy"
check "$(wc -l < "$tidy" | tr -d ' ')" "2" "tidy leaves clean file alone"

# no stray tmp files next to any of the fixtures
leftovers="$(find "$tmp" -name '*.tmp' 2>/dev/null)"
[ -z "$leftovers" ] || { echo "FAIL stray tmp files: $leftovers"; fail=1; }

# --- apply → reset round-trip ----------------------------------------------
source bin/apply.sh
round="$tmp/round.toml"
printf '[theme]\nname = "kanagawa"\n\n[ui]\nstatus_indicators = "dots"\n' > "$round"
orig="$tmp/round.orig"; cp "$round" "$orig"
tokens=$'panel_bg=#282828\ntext=#ebdbb2\naccent=#458588'
write_custom_block "$round" "$tokens"
if ! has_custom_block "$round"; then echo "FAIL apply did not write block"; fail=1; fi
# re-applying must not stack blank lines in front of the block
write_custom_block "$round" "$tokens"
write_custom_block "$round" "$tokens"
blanks="$(awk '/^\[theme\.custom\]/{print n; exit} /^[[:space:]]*$/{n++; next} {n=0}' "$round")"
check "$blanks" "1" "apply keeps a single blank line before the block"
remove_custom_block "$round"; tidy_trailing_blanks "$round"
if diff -q "$orig" "$round" >/dev/null; then :; else
  echo "FAIL round-trip changed the file:"; diff "$orig" "$round" || true; fail=1
fi

[ "$fail" = 0 ] && echo "PASS test_reset" || exit 1

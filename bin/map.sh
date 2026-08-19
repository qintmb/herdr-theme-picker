#!/usr/bin/env bash
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

# Baca satu key dari palette ghostty. Untuk 'palette N', key = "pN".
_pget() {
  local file="$1" key="$2"
  if [[ "$key" == p* ]]; then
    grep -E "^palette *= *${key#p}=" "$file" | head -1 | sed 's/.*=//'
  else
    grep -E "^${key} *=" "$file" | head -1 | sed 's/^[^=]*= *//'
  fi
}

# palette_to_tokens <file> -> 16 baris "token=#hex"
palette_to_tokens() {
  local f="$1" bg fg p0 p1 p2 p3 p4 p5 p6 p7 p8 p9
  p0="$(_pget "$f" p0)"
  [ -n "$p0" ] || die "not a valid ghostty palette: $f"
  bg="$(_pget "$f" background)"; fg="$(_pget "$f" foreground)"
  p1="$(_pget "$f" p1)"; p2="$(_pget "$f" p2)"; p3="$(_pget "$f" p3)"
  p4="$(_pget "$f" p4)"; p5="$(_pget "$f" p5)"; p6="$(_pget "$f" p6)"
  p7="$(_pget "$f" p7)"; p8="$(_pget "$f" p8)"; p9="$(_pget "$f" p9)"
  : "${p9:=$p3}"  # fallback peach
  printf 'panel_bg=%s\n' "$bg"
  printf 'surface0=%s\n' "$p0"
  printf 'surface1=%s\n' "$p8"
  printf 'surface_dim=%s\n' "$(darken_hex "$bg" 8)"
  printf 'overlay0=%s\n' "$p8"
  printf 'overlay1=%s\n' "$p7"
  printf 'text=%s\n' "$fg"
  printf 'subtext0=%s\n' "$p7"
  printf 'accent=%s\n' "$p4"
  printf 'mauve=%s\n' "$p5"
  printf 'green=%s\n' "$p2"
  printf 'yellow=%s\n' "$p3"
  printf 'red=%s\n' "$p1"
  printf 'blue=%s\n' "$p4"
  printf 'teal=%s\n' "$p6"
  printf 'peach=%s\n' "$p9"
}

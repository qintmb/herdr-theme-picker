#!/usr/bin/env bash
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

# The 13 Herdr theme tokens this plugin writes, and the ghostty palette line
# each is derived from — surfaced to the user as a legend and as the editable
# set for per-token overrides. Keep in sync with palette_to_tokens below.
TOKEN_ORDER=(panel_bg text surface0 surface1 overlay0 overlay1 subtext0 \
  accent blue mauve green yellow red teal peach surface_dim)
# human-facing origin label per token (what palette line feeds it)
token_origin() {
  case "$1" in
    panel_bg) echo "background" ;;
    text) echo "foreground" ;;
    surface0) echo "palette 0" ;;
    surface1|overlay0) echo "palette 8" ;;
    overlay1|subtext0) echo "palette 7" ;;
    accent|blue) echo "palette 4" ;;
    mauve) echo "palette 5" ;;
    green) echo "palette 2" ;;
    yellow) echo "palette 3" ;;
    red) echo "palette 1" ;;
    teal) echo "palette 6" ;;
    peach) echo "palette 9" ;;
    surface_dim) echo "background −8%" ;;
    *) echo "?" ;;
  esac
}
# short note on where the token shows up in Herdr's UI
token_role() {
  case "$1" in
    panel_bg) echo "pane background" ;;
    text) echo "primary text" ;;
    surface0) echo "sidebar row bg" ;;
    surface1) echo "inactive pane bg" ;;
    overlay0) echo "dim text" ;;
    overlay1) echo "bright text" ;;
    subtext0) echo "secondary text" ;;
    accent) echo "active pane border/selector" ;;
    blue) echo "links/info" ;;
    mauve) echo "accent" ;;
    green) echo "success ✓" ;;
    yellow) echo "warning" ;;
    red) echo "error ✗" ;;
    teal) echo "now-playing" ;;
    peach) echo "accent" ;;
    surface_dim) echo "dividers" ;;
    *) echo "" ;;
  esac
}

# Baca satu key dari palette ghostty. Untuk 'palette N', key = "pN".
_pget() {
  local file="$1" key="$2"
  if [[ "$key" == p* ]]; then
    grep -E "^palette *= *${key#p}=" "$file" | head -1 | sed 's/.*=//'
  else
    grep -E "^${key} *=" "$file" | head -1 | sed 's/^[^=]*= *//'
  fi
}

# apply_overrides <file> <tokens-text> -> tokens-text with any
# "# hpick-override: token=#hex" lines from <file> taking precedence.
# These override lines are comments, so ghostty/palette parsers ignore them;
# only this plugin reads them. The editor (edit.sh) writes them.
apply_overrides() {
  local file="$1" tokens="$2" line k v
  # collect overrides into an assoc-free stream, then rewrite matching tokens
  local ov; ov="$(grep -E '^#[[:space:]]*hpick-override:' "$file" 2>/dev/null \
    | sed -E 's/^#[[:space:]]*hpick-override:[[:space:]]*//')" || true
  [ -n "$ov" ] || { printf '%s\n' "$tokens"; return 0; }
  # For each token line, if an override exists for that key, use it.
  printf '%s\n' "$tokens" | while IFS='=' read -r k _; do
    [ -n "$k" ] || continue
    v="$(printf '%s\n' "$ov" | grep -E "^${k}=" | tail -1 | cut -d= -f2-)"
    if [ -n "$v" ]; then printf '%s=%s\n' "$k" "$v"; else grep -E "^${k}=" <<<"$tokens" | head -1; fi
  done
  # Also allow overrides for tokens not present as derived lines (defensive).
}

# palette_to_tokens <file> -> 16 baris "token=#hex" (overrides applied)
palette_to_tokens() {
  local f="$1" bg fg p0 p1 p2 p3 p4 p5 p6 p7 p8 p9
  p0="$(_pget "$f" p0)"
  [ -n "$p0" ] || die "not a valid ghostty palette: $f"
  bg="$(_pget "$f" background)"; fg="$(_pget "$f" foreground)"
  p1="$(_pget "$f" p1)"; p2="$(_pget "$f" p2)"; p3="$(_pget "$f" p3)"
  p4="$(_pget "$f" p4)"; p5="$(_pget "$f" p5)"; p6="$(_pget "$f" p6)"
  p7="$(_pget "$f" p7)"; p8="$(_pget "$f" p8)"; p9="$(_pget "$f" p9)"
  : "${p9:=$p3}"  # fallback peach
  local derived
  derived="$(printf 'panel_bg=%s\nsurface0=%s\nsurface1=%s\nsurface_dim=%s\noverlay0=%s\noverlay1=%s\ntext=%s\nsubtext0=%s\naccent=%s\nmauve=%s\ngreen=%s\nyellow=%s\nred=%s\nblue=%s\nteal=%s\npeach=%s\n' \
    "$bg" "$p0" "$p8" "$(darken_hex "$bg" 8)" "$p8" "$p7" "$fg" "$p7" \
    "$p4" "$p5" "$p2" "$p3" "$p1" "$p4" "$p6" "$p9")"
  apply_overrides "$f" "$derived"
}


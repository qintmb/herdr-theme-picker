#!/usr/bin/env bash
# Fungsi bersama untuk herdr-theme-picker.
CONFIG_PATH="${HERDR_CONFIG_PATH:-$HOME/.config/herdr/config.toml}"
CACHE_DIR="${XDG_CACHE_HOME:-$HOME/.cache}/herdr-theme-picker"
# Herdr provides HERDR_PLUGIN_STATE_DIR; fall back to cache dir when run standalone.
STATE_DIR="${HERDR_PLUGIN_STATE_DIR:-$CACHE_DIR}"
APPLIED_FILE="$STATE_DIR/applied"
# User-added themes live in state (survive plugin updates), separate from the
# bundled themes/ dir that a reinstall overwrites.
USER_THEMES_DIR="$STATE_DIR/themes"
USER_INDEX="$STATE_DIR/index.txt"

die() { printf '%s\n' "$*" >&2; exit 1; }

is_valid_slug() {
  [[ "$1" =~ ^[a-z0-9-]+$ ]]
}

# is_user_theme <slug>: true only for themes the user added (in USER_INDEX with
# a file under USER_THEMES_DIR). Bundled/remote themes are never user themes,
# so edit/delete can refuse to touch them.
is_user_theme() {
  local slug="$1"
  is_valid_slug "$slug" || return 1
  [ -f "$USER_INDEX" ] || return 1
  grep -qxF "$slug" "$USER_INDEX" || return 1
  [ -f "$USER_THEMES_DIR/$slug" ]
}

# resolve_symlink <path> -> absolute path of the final target.
# Portable: BSD/macOS `readlink -f` exits non-zero on a dangling link and
# `realpath` refuses to resolve one at all, so walk the chain by hand.
resolve_symlink() {
  local p="$1" dir link n=0
  # make absolute so relative link targets resolve against the right dir
  case "$p" in
    /*) : ;;
    *)  p="$PWD/$p" ;;
  esac
  while [ -L "$p" ]; do
    n=$((n + 1))
    [ "$n" -le 40 ] || { echo "too many symlinks: $1" >&2; return 1; }
    dir="$(cd "$(dirname "$p")" && pwd)"
    link="$(readlink "$p")"
    case "$link" in
      /*) p="$link" ;;
      *)  p="$dir/$link" ;;
    esac
  done
  dir="$(cd "$(dirname "$p")" 2>/dev/null && pwd)" || dir="$(dirname "$p")"
  printf '%s/%s\n' "$dir" "$(basename "$p")"
}

# has_custom_block <config>: true when the plugin's [theme.custom] block is in
# the file. The block — not the `applied` state file — is what overrides the
# built-in theme, so the picker uses this to decide who wears the ✓.
has_custom_block() {
  grep -q '^\[theme\.custom\]' "$1" 2>/dev/null
}

# remove_custom_block <target>: drop the [theme.custom] section — the exact span
# write_custom_block replaces, so reset is the inverse of apply. A hand-written
# [theme.custom.light]/[theme.custom.dark] section is left alone: the plugin
# never writes one, so it belongs to the user. Expects an already-resolved
# (non-symlink) path.
remove_custom_block() {
  local target="$1"
  _strip_custom_block < "$target" > "$target.tmp" && mv "$target.tmp" "$target"
}

# Stream filters behind the in-place helpers, so write_custom_block can chain
# them into a single temp file and still replace the config in one move.
_strip_custom_block() {
  awk '
    /^\[theme\.custom\]/ { skip=1; next }
    skip && /^\[/ { skip=0 }
    !skip { print }
  '
}

# tidy_trailing_blanks <file>: drop the trailing blank run. Every apply appends
# a blank line in front of its block, so a config cycled through the picker
# otherwise grows a longer gap each time.
tidy_trailing_blanks() {
  local target="$1"
  _strip_trailing_blanks < "$target" > "$target.tmp" && mv "$target.tmp" "$target"
}

_strip_trailing_blanks() {
  awk '
    BEGIN { blanks = 0 }
    { if ($0 ~ /^[[:space:]]*$/) blanks++; else blanks = 0; line[NR] = $0; tail = blanks }
    END { for (i = 1; i <= NR - tail; i++) print line[i] }
  '
}

# base_theme <config>: the built-in theme named by [theme] name, or herdr's own
# fallback when the key is unset. Lets "herdr default" rows state the theme you
# actually land on instead of leaving it to guesswork.
base_theme() {
  local name
  name="$(awk '
    /^\[/ { intheme = ($0 == "[theme]"); next }
    intheme && /^[[:space:]]*name[[:space:]]*=/ {
      sub(/^[^=]*=[[:space:]]*/, "")
      gsub(/["[:space:]]/, "")
      print
      exit
    }
  ' "$1" 2>/dev/null)"
  printf '%s\n' "${name:-catppuccin}"
}

# darken_hex #rrggbb percent -> #rrggbb (tiap kanal * (100-p)/100, clamp 0)
darken_hex() {
  local hex="${1#\#}" p="$2" r g b
  r=$((16#${hex:0:2})); g=$((16#${hex:2:2})); b=$((16#${hex:4:2}))
  r=$(( r * (100 - p) / 100 )); g=$(( g * (100 - p) / 100 )); b=$(( b * (100 - p) / 100 ))
  (( r < 0 )) && r=0; (( g < 0 )) && g=0; (( b < 0 )) && b=0
  printf '#%02x%02x%02x\n' "$r" "$g" "$b"
}
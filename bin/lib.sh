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

# darken_hex #rrggbb percent -> #rrggbb (tiap kanal * (100-p)/100, clamp 0)
darken_hex() {
  local hex="${1#\#}" p="$2" r g b
  r=$((16#${hex:0:2})); g=$((16#${hex:2:2})); b=$((16#${hex:4:2}))
  r=$(( r * (100 - p) / 100 )); g=$(( g * (100 - p) / 100 )); b=$(( b * (100 - p) / 100 ))
  (( r < 0 )) && r=0; (( g < 0 )) && g=0; (( b < 0 )) && b=0
  printf '#%02x%02x%02x\n' "$r" "$g" "$b"
}
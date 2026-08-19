#!/usr/bin/env bash
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
PLUGIN_ROOT="${HERDR_PLUGIN_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
BASE_URL="https://terminalcolors.com/downloads/ghostty"

# resolve_palette <slug> -> path file palette (stdout)
resolve_palette() {
  local slug="$1"
  is_valid_slug "$slug" || die "invalid slug: $slug"

  local bundle="$PLUGIN_ROOT/themes/$slug"
  [ -f "$bundle" ] && { printf '%s\n' "$bundle"; return 0; }

  # User-added themes (survive plugin updates).
  local user="$USER_THEMES_DIR/$slug"
  [ -f "$user" ] && { printf '%s\n' "$user"; return 0; }

  local cache="$CACHE_DIR/$slug"
  [ -f "$cache" ] && { printf '%s\n' "$cache"; return 0; }

  mkdir -p "$CACHE_DIR"
  local tmp; tmp="$(mktemp)"
  if ! curl -fsSL -m 15 "$BASE_URL/$slug" -o "$tmp" 2>/dev/null; then
    rm -f "$tmp"; die "failed to fetch '$slug' (offline or not found). Needs internet."
  fi
  grep -q '^palette = 0=' "$tmp" || { rm -f "$tmp"; die "response is not a valid palette for '$slug'"; }
  mv "$tmp" "$cache"
  printf '%s\n' "$cache"
}

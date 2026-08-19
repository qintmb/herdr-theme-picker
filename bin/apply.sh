#!/usr/bin/env bash
_here="$(dirname "${BASH_SOURCE[0]}")"
source "$_here/map.sh"
source "$_here/fetch.sh"

# Hapus blok [theme.custom] lama (sampai header berikutnya / EOF), lalu append baru.
write_custom_block() {
  local cfg="$1" tokens="$2"
  awk '
    /^\[theme\.custom\]/ { skip=1; next }
    skip && /^\[/ { skip=0 }
    !skip { print }
  ' "$cfg" > "$cfg.tmp"
  {
    printf '\n[theme.custom]\n'
    while IFS='=' read -r k v; do
      [ -n "$k" ] && printf '%s = "%s"\n' "$k" "$v"
    done <<< "$tokens"
  } >> "$cfg.tmp"
  mv "$cfg.tmp" "$cfg"
}

main() {
  local slug="${1:-}"
  [ -n "$slug" ] || die "usage: apply.sh <slug>"
  [ -f "$CONFIG_PATH" ] || die "config not found: $CONFIG_PATH"

  local palette tokens
  palette="$(resolve_palette "$slug")" || die "cannot resolve theme: $slug"
  [ -n "$palette" ] && [ -f "$palette" ] || die "cannot resolve theme: $slug"
  tokens="$(palette_to_tokens "$palette")" || die "invalid palette for: $slug"
  [ -n "$tokens" ] || die "invalid palette for: $slug"

  local stamp; stamp="$(date +%Y%m%d)"
  local bak="$CONFIG_PATH.bak-$stamp"
  [ -f "$bak" ] || cp "$CONFIG_PATH" "$bak"

  write_custom_block "$CONFIG_PATH" "$tokens"

  # Record applied theme so the picker can mark it next time.
  mkdir -p "$STATE_DIR"
  printf '%s\n' "$slug" > "$APPLIED_FILE"

  if herdr server reload-config >/dev/null 2>&1; then
    printf 'Applied theme "%s".\n' "$slug"
  else
    printf 'Theme "%s" written, but reload failed. Try: herdr server reload-config\n' "$slug" >&2
  fi
}

# Run main only when executed directly (not when sourced by a test).
if [ "${BASH_SOURCE[0]}" = "$0" ]; then main "$@"; fi

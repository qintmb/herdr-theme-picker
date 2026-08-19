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
  [ -n "$slug" ] || die "pemakaian: apply.sh <slug>"
  [ -f "$CONFIG_PATH" ] || die "config tak ditemukan: $CONFIG_PATH"

  local palette tokens
  palette="$(resolve_palette "$slug")"
  tokens="$(palette_to_tokens "$palette")"

  local stamp; stamp="$(date +%Y%m%d)"
  local bak="$CONFIG_PATH.bak-$stamp"
  [ -f "$bak" ] || cp "$CONFIG_PATH" "$bak"

  write_custom_block "$CONFIG_PATH" "$tokens"

  if herdr server reload-config >/dev/null 2>&1; then
    printf 'Theme "%s" diterapkan.\n' "$slug"
  else
    printf 'Theme "%s" ditulis, tapi reload gagal. Coba: herdr server reload-config\n' "$slug" >&2
  fi
}

# Jalankan main hanya bila dieksekusi langsung (bukan di-source oleh test).
if [ "${BASH_SOURCE[0]}" = "$0" ]; then main "$@"; fi

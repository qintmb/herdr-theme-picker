#!/usr/bin/env bash
set -euo pipefail
_here="$(dirname "${BASH_SOURCE[0]}")"
source "$_here/fetch.sh"
source "$_here/map.sh"
PLUGIN_ROOT="${HERDR_PLUGIN_ROOT:-$(cd "$_here/.." && pwd)}"

# Blok warna truecolor: bg=hex, teks hex.
_block() { local hex="${1#\#}"; printf '\033[48;2;%d;%d;%dm  %s  \033[0m' \
  "$((16#${hex:0:2}))" "$((16#${hex:2:2}))" "$((16#${hex:4:2}))" "$1"; }

# swatch <slug>: preview token utama (dipakai fzf --preview). Fetch bila perlu.
swatch() {
  local slug="$1" p tok
  p="$(resolve_palette "$slug" 2>/dev/null)" || { echo "(gagal memuat $slug — butuh internet)"; return 0; }
  tok="$(palette_to_tokens "$p")"
  printf '%s\n\n' "$slug"
  local key
  for key in panel_bg text accent red green yellow blue teal mauve peach; do
    printf '%-10s ' "$key"
    _block "$(printf '%s\n' "$tok" | grep "^$key=" | cut -d= -f2-)"
    printf '\n'
  done
}

# Bila dipanggil fzf untuk preview satu baris.
if [ "${1:-}" = "--preview" ]; then swatch "$2"; exit 0; fi

main() {
  command -v fzf >/dev/null || { echo "fzf tidak terpasang"; read -r; exit 1; }
  local list; list="$(cat "$PLUGIN_ROOT/themes/index.txt")"
  local sel
  sel="$(printf '%s\n' "$list" | fzf \
    --prompt="theme> " \
    --preview="bash '$PLUGIN_ROOT/bin/picker.sh' --preview {}" \
    --preview-window=right,50%)" || exit 0
  [ -n "$sel" ] || exit 0
  bash "$PLUGIN_ROOT/bin/apply.sh" "$sel"
  echo; echo "Tekan Enter untuk tutup."; read -r
}

if [ "${BASH_SOURCE[0]}" = "$0" ] && [ "${1:-}" != "--preview" ]; then main "$@"; fi

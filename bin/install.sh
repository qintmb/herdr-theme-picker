#!/usr/bin/env bash
# Local development install: link this checkout into Herdr.
# For normal use, prefer: herdr plugin install qintmb/herdr-theme-picker
_here="$(dirname "${BASH_SOURCE[0]}")"
source "$_here/lib.sh"
PLUGIN_ROOT="$(cd "$_here/.." && pwd)"

main() {
  command -v herdr >/dev/null || die "herdr tidak ditemukan di PATH."
  herdr plugin link "$PLUGIN_ROOT" || die "herdr plugin link gagal"
  herdr server reload-config >/dev/null 2>&1 || true
  echo "Linked. Keybind prefix+t sudah didaftarkan lewat manifest."
  echo "Tekan prefix+t (prefix default cmd+b) untuk buka picker."
}
if [ "${BASH_SOURCE[0]}" = "$0" ]; then main "$@"; fi

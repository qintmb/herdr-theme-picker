#!/usr/bin/env bash
# Remove the plugin from Herdr (mirror of install.sh).
_here="$(dirname "${BASH_SOURCE[0]}")"
source "$_here/lib.sh"

main() {
  command -v herdr >/dev/null || die "herdr not found in PATH."
  herdr plugin unlink herdr-theme-picker 2>/dev/null || true
  herdr server reload-config >/dev/null 2>&1 || true
  echo "Unlinked."
}
if [ "${BASH_SOURCE[0]}" = "$0" ]; then main "$@"; fi

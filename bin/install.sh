#!/usr/bin/env bash
# Local development install: link this checkout into Herdr.
# For normal use, prefer: herdr plugin install qintmb/herdr-theme-picker
_here="$(dirname "${BASH_SOURCE[0]}")"
source "$_here/lib.sh"
PLUGIN_ROOT="$(cd "$_here/.." && pwd)"

main() {
  command -v herdr >/dev/null || die "herdr not found in PATH."
  herdr plugin link "$PLUGIN_ROOT" || die "herdr plugin link failed"
  herdr server reload-config >/dev/null 2>&1 || true
  echo "Linked. The prefix+t keybind is registered via the manifest."
  echo "Press prefix+t (default prefix cmd+b) to open the picker."
}
if [ "${BASH_SOURCE[0]}" = "$0" ]; then main "$@"; fi

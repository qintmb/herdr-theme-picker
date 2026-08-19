#!/usr/bin/env bash
_here="$(dirname "${BASH_SOURCE[0]}")"
source "$_here/install.sh"  # pakai remove_keybind + CONFIG_PATH
main_uninstall() {
  [ -f "$CONFIG_PATH" ] && remove_keybind "$CONFIG_PATH"
  herdr plugin unlink herdr-theme-picker 2>/dev/null || true
  herdr server reload-config >/dev/null 2>&1 || true
  echo "Dicopot."
}
main_uninstall "$@"

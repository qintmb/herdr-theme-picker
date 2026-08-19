#!/usr/bin/env bash
_here="$(dirname "${BASH_SOURCE[0]}")"
source "$_here/lib.sh"
PLUGIN_ROOT="$(cd "$_here/.." && pwd)"
MARKER="command = \"herdr-theme-picker.open\""

add_keybind() {
  local cfg="$1"
  grep -qF "$MARKER" "$cfg" && return 0
  cp "$cfg" "$cfg.bak-$(date +%Y%m%d)" 2>/dev/null || true
  cat >> "$cfg" <<'EOF'

[[keys.command]]
key = "prefix+t"
type = "plugin_action"
command = "herdr-theme-picker.open"
description = "pick a color theme"
EOF
}

remove_keybind() {
  local cfg="$1"
  awk '
    /^\[\[keys\.command\]\]/ { buf=$0"\n"; inblk=1; hit=0; next }
    inblk {
      buf=buf $0 "\n"
      if ($0 ~ /herdr-theme-picker\.open/) hit=1
      if ($0 ~ /^$/ || $0 ~ /^\[/) { if (!hit) printf "%s", buf; inblk=0; buf=""; if ($0 ~ /^\[/) print }
      next
    }
    { print }
    END { if (inblk && !hit) printf "%s", buf }
  ' "$cfg" > "$cfg.tmp" && mv "$cfg.tmp" "$cfg"
}

main() {
  [ -f "$CONFIG_PATH" ] || die "config tak ditemukan: $CONFIG_PATH"
  herdr plugin link "$PLUGIN_ROOT" || die "herdr plugin link gagal"
  add_keybind "$CONFIG_PATH"
  herdr server reload-config >/dev/null 2>&1 || true
  echo "Terpasang. Tekan prefix+t (prefix default cmd+b) untuk buka picker."
}
if [ "${BASH_SOURCE[0]}" = "$0" ]; then main "$@"; fi

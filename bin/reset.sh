#!/usr/bin/env bash
# Undo apply.sh: drop the plugin's [theme.custom] block so Herdr's own built-in
# theme takes over again, and put the host terminal's ANSI palette back to its
# own defaults. Reached from the picker's "↩ herdr default" row.
set -euo pipefail
_here="$(dirname "${BASH_SOURCE[0]}")"
source "$_here/lib.sh"

# Emit the OSC reset counterparts (104 = palette, 110/111 = fg/bg) of the
# OSC 4/10/11 sequences apply.sh sends, to the same outer PTY. Herdr's
# multiplexer intercepts /dev/tty writes, so target the terminal emulator
# hosting the herdr client directly.
sync_terminal_reset() {
  local outer_tty
  outer_tty=$(ps -eo tty,comm 2>/dev/null \
    | awk '$2=="herdr" && $1!="?" && $1!="??" {print "/dev/"$1; exit}')
  [ -n "$outer_tty" ] && [ -w "$outer_tty" ] || return 0
  printf '\033]104\007\033]110\007\033]111\007' >> "$outer_tty"
}

# apply.sh copies every palette into ~/.config/ghostty/herdr-theme when that
# opt-in fragment exists, so new Ghostty windows would keep the plugin's colors.
# Put the pre-plugin content back, or park the stale palette when there was
# nothing to restore.
restore_ghostty_fragment() {
  local frag="$HOME/.config/ghostty/herdr-theme" bak="$HOME/.config/ghostty/herdr-theme.pre-herdr"
  [ -f "$frag" ] || return 0
  if [ -f "$bak" ]; then
    mv "$bak" "$frag"
    printf 'Ghostty fragment restored to its pre-plugin colors.\n'
  else
    mv "$frag" "$frag.removed-$(date +%Y%m%d%H%M%S)"
    printf 'Ghostty fragment parked (no pre-plugin copy to restore).\n'
  fi
}

main() {
  [ -f "$CONFIG_PATH" ] || die "config not found: $CONFIG_PATH"

  local target
  target="$(resolve_symlink "$CONFIG_PATH")" || die "cannot resolve config path: $CONFIG_PATH"

  if ! has_custom_block "$target"; then
    printf 'Already on the built-in theme "%s" — nothing to undo.\n' "$(base_theme "$target")"
    return 0
  fi

  local stamp; stamp="$(date +%Y%m%d)"
  local bak="$CONFIG_PATH.bak-$stamp"
  [ -f "$bak" ] || cp "$CONFIG_PATH" "$bak"

  remove_custom_block "$target"
  tidy_trailing_blanks "$target"

  # Forget the applied slug so the picker stops marking it.
  rm -f "$APPLIED_FILE"

  sync_terminal_reset
  restore_ghostty_fragment

  if herdr server reload-config >/dev/null 2>&1; then
    printf 'Custom colors removed — back on Herdr theme "%s".\n' "$(base_theme "$target")"
  else
    printf 'Custom colors removed, but reload failed. Try: herdr server reload-config\n' >&2
  fi
}

# Run main only when executed directly (not when sourced by a test).
if [ "${BASH_SOURCE[0]}" = "$0" ]; then main "$@"; fi

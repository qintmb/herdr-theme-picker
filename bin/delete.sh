#!/usr/bin/env bash
# Delete a user-added theme. Refuses bundled/remote themes. Prompts a Y/N
# confirmation on the tty (overridable via TTY_IN for tests).
_here="$(dirname "${BASH_SOURCE[0]}")"
source "$_here/lib.sh"

delete_theme() {
  local slug="${1:-}"
  local tin="${TTY_IN:-/dev/tty}"
  [ -n "$slug" ] || { echo "usage: delete.sh <slug>" >&2; return 2; }

  if ! is_user_theme "$slug"; then
    echo "'$slug' is not a user-added theme — only ★ themes can be deleted." >&2
    return 1
  fi

  printf "Delete '%s'? [y/N] " "$slug" >&2
  local yn; IFS= read -r yn <"$tin" || return 1
  case "$yn" in
    y|Y) ;;
    *) echo "Kept '$slug'." >&2; return 1 ;;
  esac

  rm -f "$USER_THEMES_DIR/$slug"
  # Drop the slug line from the user index.
  if [ -f "$USER_INDEX" ]; then
    grep -vxF "$slug" "$USER_INDEX" > "$USER_INDEX.tmp" 2>/dev/null || true
    mv "$USER_INDEX.tmp" "$USER_INDEX"
  fi
  # If it was the applied theme, forget the marker (don't rewrite config).
  if [ -f "$APPLIED_FILE" ] && [ "$(head -1 "$APPLIED_FILE" 2>/dev/null)" = "$slug" ]; then
    rm -f "$APPLIED_FILE"
  fi

  echo "Deleted '$slug'." >&2
  return 0
}

if [ "${BASH_SOURCE[0]}" = "$0" ]; then delete_theme "$@"; fi

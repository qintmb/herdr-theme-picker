#!/usr/bin/env bash
# Add a user theme: paste a ghostty-format palette in $EDITOR, name it,
# validate, save to the user themes dir, and append the slug to the user index.
# User themes live in HERDR_PLUGIN_STATE_DIR so plugin updates never wipe them.
_here="$(dirname "${BASH_SOURCE[0]}")"
source "$_here/map.sh"   # palette_to_tokens, and lib.sh (slug/dirs) transitively

TEMPLATE='# Paste a ghostty-format theme below, then save and close this editor.
# Lines starting with # are ignored. Required: background, palette 0..15.
# Get one from https://terminalcolors.com → Download → Ghostty.
#
# background = #1e1e2e
# foreground = #cdd6f4
# cursor-color = #f5e0dc
# palette = 0=#45475a
# palette = 1=#f38ba8
# ... through ...
# palette = 15=#a6adc8
'

# slugify <name> -> lowercase, spaces/underscores to -, keep [a-z0-9-]
slugify() {
  printf '%s' "$1" \
    | tr '[:upper:]' '[:lower:]' \
    | tr ' _' '--' \
    | tr -cd 'a-z0-9-' \
    | sed 's/-\{2,\}/-/g; s/^-//; s/-$//'
}

add_theme() {
  mkdir -p "$USER_THEMES_DIR"
  local editor="${VISUAL:-${EDITOR:-}}"
  [ -n "$editor" ] || { command -v nano >/dev/null && editor=nano || editor=vi; }

  local draft; draft="$(mktemp)"
  printf '%s' "$TEMPLATE" > "$draft"
  "$editor" "$draft" </dev/tty >/dev/tty 2>&1 || { rm -f "$draft"; echo "Cancelled."; return 1; }

  # Strip comments/blank-only; keep real content.
  local body; body="$(grep -vE '^[[:space:]]*(#|$)' "$draft" || true)"
  if ! printf '%s\n' "$body" | grep -q '^palette *= *0='; then
    rm -f "$draft"
    echo "No valid palette found (need 'palette = 0=...'). Nothing saved."
    return 1
  fi

  # Validate it maps cleanly.
  local clean; clean="$(mktemp)"
  printf '%s\n' "$body" > "$clean"
  if ! palette_to_tokens "$clean" >/dev/null 2>&1; then
    rm -f "$draft" "$clean"
    echo "Palette could not be parsed. Nothing saved."
    return 1
  fi

  # Name it.
  local name slug
  printf 'Theme name: ' >/dev/tty
  IFS= read -r name </dev/tty || { rm -f "$draft" "$clean"; return 1; }
  slug="$(slugify "$name")"
  [ -n "$slug" ] || { rm -f "$draft" "$clean"; echo "Empty name. Nothing saved."; return 1; }

  local dest="$USER_THEMES_DIR/$slug"
  if [ -e "$dest" ]; then
    printf "'%s' already exists. Overwrite? [y/N] " "$slug" >/dev/tty
    local yn; IFS= read -r yn </dev/tty
    case "$yn" in y|Y) ;; *) rm -f "$draft" "$clean"; echo "Kept existing. Nothing saved."; return 1;; esac
  fi

  mv "$clean" "$dest"
  rm -f "$draft"

  # Append to user index if not already present.
  touch "$USER_INDEX"
  grep -qxF "$slug" "$USER_INDEX" || printf '%s\n' "$slug" >> "$USER_INDEX"

  echo "Saved theme '$slug'."
  printf '%s\n' "$slug"   # last line = slug, for callers
}

if [ "${BASH_SOURCE[0]}" = "$0" ]; then add_theme "$@"; fi

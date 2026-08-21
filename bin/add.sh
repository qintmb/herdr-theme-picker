#!/usr/bin/env bash
# Add a user theme two ways:
#   add.sh              → open the palette in an editor (prefers nvim), then name+save
#   add.sh --clipboard  → read the palette straight from the system clipboard,
#                         then just name+save (no editor)
# Validate, save to the user themes dir, and append the slug to the user index.
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

# pick_editor: prefer nvim (user request), then $VISUAL/$EDITOR, then nano/vi.
pick_editor() {
  if command -v nvim >/dev/null 2>&1; then printf 'nvim'; return; fi
  local e="${VISUAL:-${EDITOR:-}}"
  if [ -n "$e" ]; then printf '%s' "$e"; return; fi
  command -v nano >/dev/null 2>&1 && { printf 'nano'; return; }
  printf 'vi'
}

# read_clipboard: echo clipboard contents, or nothing if no tool/empty.
read_clipboard() {
  if command -v pbpaste >/dev/null 2>&1; then pbpaste
  elif command -v wl-paste >/dev/null 2>&1; then wl-paste --no-newline
  elif command -v xclip >/dev/null 2>&1; then xclip -selection clipboard -o
  elif command -v xsel >/dev/null 2>&1; then xsel --clipboard --output
  fi
}

# _finalize <body-file>: validate palette, prompt name, save, index. Echoes slug.
# Prompts read from $TTY_IN (default /dev/tty) so tests can drive it via stdin.
_finalize() {
  local body_file="$1"
  local tin="${TTY_IN:-/dev/tty}"
  # Require a real palette line.
  if ! grep -q '^[[:space:]]*palette *= *0=' "$body_file"; then
    rm -f "$body_file"
    echo "No valid palette found (need 'palette = 0=...'). Nothing saved." >&2
    return 1
  fi
  # Must map cleanly.
  if ! palette_to_tokens "$body_file" >/dev/null 2>&1; then
    rm -f "$body_file"
    echo "Palette could not be parsed. Nothing saved." >&2
    return 1
  fi

  local name slug
  printf 'Theme name: ' >&2
  IFS= read -r name <"$tin" || { rm -f "$body_file"; return 1; }
  slug="$(slugify "$name")"
  [ -n "$slug" ] || { rm -f "$body_file"; echo "Empty name. Nothing saved." >&2; return 1; }

  mkdir -p "$USER_THEMES_DIR"
  local dest="$USER_THEMES_DIR/$slug"
  if [ -e "$dest" ]; then
    printf "'%s' already exists. Overwrite? [y/N] " "$slug" >&2
    local yn; IFS= read -r yn <"$tin"
    case "$yn" in y|Y) ;; *) rm -f "$body_file"; echo "Kept existing. Nothing saved." >&2; return 1;; esac
  fi

  mv "$body_file" "$dest"

  touch "$USER_INDEX"
  grep -qxF "$slug" "$USER_INDEX" || printf '%s\n' "$slug" >> "$USER_INDEX"

  echo "Saved theme '$slug'." >&2
  printf '%s\n' "$slug"   # last stdout line = slug, for callers
}

# strip comments/blank lines into a clean palette file.
_strip_to() { grep -vE '^[[:space:]]*(#|$)' "$1" > "$2" || true; }

add_theme_editor() {
  local editor; editor="$(pick_editor)"
  local draft; draft="$(mktemp)"
  printf '%s' "$TEMPLATE" > "$draft"
  "$editor" "$draft" </dev/tty >/dev/tty 2>&1 || { rm -f "$draft"; echo "Cancelled." >&2; return 1; }
  local clean; clean="$(mktemp)"
  _strip_to "$draft" "$clean"
  rm -f "$draft"
  _finalize "$clean"
}

add_theme_clipboard() {
  local clip; clip="$(read_clipboard)"
  if [ -z "$clip" ]; then
    echo "Clipboard empty or no clipboard tool (pbpaste/wl-paste/xclip/xsel)." >&2
    return 1
  fi
  local clean; clean="$(mktemp)"
  printf '%s\n' "$clip" | grep -vE '^[[:space:]]*(#|$)' > "$clean" || true
  echo "Read palette from clipboard." >&2
  _finalize "$clean"
}

add_theme() {
  case "${1:-}" in
    --clipboard) add_theme_clipboard ;;
    *) add_theme_editor ;;
  esac
}

if [ "${BASH_SOURCE[0]}" = "$0" ]; then add_theme "$@"; fi

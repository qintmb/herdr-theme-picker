#!/usr/bin/env bash
# In-plugin token editor for a user-added theme. No external editor: an fzf
# loop lists the 13 Herdr tokens (current hex + which palette line feeds them),
# the preview pane shows the live Herdr mockup rebuilt from a scratch copy of
# the theme, and editing a token writes a "# hpick-override:" line. Save & apply
# commits the scratch file over the user theme and applies it.
_here="$(dirname "${BASH_SOURCE[0]}")"
source "$_here/map.sh"    # TOKEN_ORDER, token_origin/role, palette_to_tokens, lib
PLUGIN_ROOT="${HERDR_PLUGIN_ROOT:-$(cd "$_here/.." && pwd)}"

# Editable tokens (skip surface_dim which is derived from bg).
EDIT_TOKENS=(accent red green yellow blue teal mauve peach text panel_bg \
  surface0 surface1 subtext0)

# build_list <scratch-file>: one row per token = "token  #hex  ← origin  role"
build_list() {
  local file="$1" tok t val
  tok="$(palette_to_tokens "$file" 2>/dev/null)" || return 1
  for t in "${EDIT_TOKENS[@]}"; do
    val="$(printf '%s\n' "$tok" | grep "^$t=" | cut -d= -f2-)"
    printf '%-10s %-8s ← %-13s %s\n' "$t" "${val:-—}" "$(token_origin "$t")" "$(token_role "$t")"
  done
  printf '%s\n' "✓ Save & apply"
}

# set_override <file> <token> <hex>: replace/append a hpick-override line.
set_override() {
  local file="$1" tok="$2" hex="$3"
  grep -vE "^#[[:space:]]*hpick-override:[[:space:]]*${tok}=" "$file" > "$file.tmp" 2>/dev/null || true
  mv "$file.tmp" "$file"
  printf '# hpick-override: %s=%s\n' "$tok" "$hex" >> "$file"
}

# Preview hook: render the scratch file's mockup (called by fzf).
if [ "${1:-}" = "--preview" ]; then
  scratch="${2:-}"; row="${3:-}"
  # Source picker.sh for swatch_file/_cell; pass a dummy arg so picker.sh's own
  # "--preview" guard doesn't fire off our inherited positional params.
  source "$PLUGIN_ROOT/bin/picker.sh" --sourced >/dev/null 2>&1 || true
  case "$row" in
    "✓ Save"*) echo "  Commit these colors to the theme and apply."; echo ;;
  esac
  swatch_file "$scratch" "editing"
  exit 0
fi

edit_theme() {
  local slug="${1:-}"
  local tin="${TTY_IN:-/dev/tty}"
  [ -n "$slug" ] || { echo "usage: edit.sh <slug>" >&2; return 2; }
  command -v fzf >/dev/null || { echo "fzf is not installed" >&2; return 1; }

  if ! is_user_theme "$slug"; then
    echo "'$slug' is not a user-added theme — only ★ themes can be edited." >&2
    return 1
  fi

  local src="$USER_THEMES_DIR/$slug"
  local scratch; scratch="$(mktemp)"
  cp "$src" "$scratch"

  while true; do
    local out key sel
    out="$(build_list "$scratch" | fzf \
      --layout=reverse \
      --prompt="Edit $slug — pick a token: " \
      --header="↵ change hex · ↵ on 'Save & apply' to commit · esc discard" \
      --info=inline --height=100% \
      --expect=esc \
      --preview="bash '$PLUGIN_ROOT/bin/edit.sh' --preview '$scratch' {}" \
      --preview-window=right,62%,border-left)" || { rm -f "$scratch"; return 0; }

    key="$(printf '%s\n' "$out" | sed -n '1p')"
    sel="$(printf '%s\n' "$out" | sed -n '2p')"
    [ "$key" = "esc" ] && { rm -f "$scratch"; echo "Discarded edits." >&2; return 0; }
    [ -n "$sel" ] || { rm -f "$scratch"; return 0; }

    if printf '%s' "$sel" | grep -q '^✓ Save'; then
      cp "$scratch" "$src"; rm -f "$scratch"
      bash "$PLUGIN_ROOT/bin/apply.sh" "$slug"
      return 0
    fi

    local tok; tok="$(printf '%s' "$sel" | awk '{print $1}')"
    printf 'New hex for %s (e.g. #fa390f), empty to cancel: ' "$tok" >&2
    local hex; IFS= read -r hex <"$tin" || { rm -f "$scratch"; return 0; }
    [ -n "$hex" ] || continue
    if ! printf '%s' "$hex" | grep -qiE '^#[0-9a-f]{6}$'; then
      echo "Not a #rrggbb hex: $hex" >&2; continue
    fi
    set_override "$scratch" "$tok" "$hex"
  done
}

if [ "${BASH_SOURCE[0]}" = "$0" ] && [ "${1:-}" != "--preview" ]; then edit_theme "$@"; fi

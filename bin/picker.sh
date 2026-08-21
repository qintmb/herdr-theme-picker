#!/usr/bin/env bash
set -euo pipefail
_here="$(dirname "${BASH_SOURCE[0]}")"
source "$_here/fetch.sh"
source "$_here/map.sh"
PLUGIN_ROOT="${HERDR_PLUGIN_ROOT:-$(cd "$_here/.." && pwd)}"

# "#rrggbb" -> "r;g;b" (for ANSI truecolor).
_rgb() { local h="${1#\#}"; printf '%d;%d;%d' "$((16#${h:0:2}))" "$((16#${h:2:2}))" "$((16#${h:4:2}))"; }

# _cell <bg#hex> <fg#hex> <text> : text on a colored cell (no newline).
_cell() { printf '\033[48;2;%s;38;2;%sm%s\033[0m' "$(_rgb "$1")" "$(_rgb "$2")" "$3"; }
# _fg <fg#hex> <text> : colored fg only (transparent bg).
_fg() { printf '\033[38;2;%sm%s\033[0m' "$(_rgb "$1")" "$2"; }
# pad text to width n (plain, before coloring)
_padr() { printf '%-*.*s' "$1" "$1" "$2"; }

# swatch <slug>: full mock of the Herdr layout in the theme's colors —
# left "spaces" sidebar, a two-pane workspace (active pane highlighted),
# tab bar, and an agent pane. Used as the fzf --preview; fetches on demand.
swatch() {
  local slug="$1" p tok
  p="$(resolve_palette "$slug" 2>/dev/null)" || { echo "(cannot load $slug — needs internet)"; return 0; }
  tok="$(palette_to_tokens "$p")"
  tget() { printf '%s\n' "$tok" | grep "^$1=" | cut -d= -f2-; }
  local panel_bg text accent blue green red yellow mauve teal peach
  local subtext0 surface0 surface1 surface_dim overlay0 overlay1
  panel_bg="$(tget panel_bg)"; text="$(tget text)"; accent="$(tget accent)"
  blue="$(tget blue)"; green="$(tget green)"; red="$(tget red)"
  yellow="$(tget yellow)"; mauve="$(tget mauve)"; teal="$(tget teal)"
  peach="$(tget peach)"; subtext0="$(tget subtext0)"
  surface0="$(tget surface0)"; surface1="$(tget surface1)"
  surface_dim="$(tget surface_dim)"; overlay0="$(tget overlay0)"
  overlay1="$(tget overlay1)"

  local SB=13   # sidebar width
  local PA=13   # left (active) pane width
  local PB=13   # right pane width

  # Column helpers bound to this theme's colors.
  # sidebar row: bg=surface0, fg passed
  sb()  { _cell "$surface0" "$1" "$(_padr "$SB" "$2")"; }
  # active pane row: bg=panel_bg (border accent shown via left bar)
  pane(){ local bg="$1" fg="$2" w="$3" t="$4"; _cell "$bg" "$fg" "$(_padr "$w" "$t")"; }

  printf '  %s\n\n' "$slug"

  # ── Tab bar (spans the workspace area) ────────────────────────────
  printf '%s' "$(sb "$subtext0" ' spaces')"
  _cell "$accent"   "$panel_bg" ' WORKSPACE '
  _cell "$surface1" "$subtext0" ' Files '
  printf '\n'

  # ── Row 1: sidebar space header + two pane title bars ─────────────
  printf '%s' "$(sb "$accent" ' ■ AGENT')"
  _cell "$surface1" "$text"     " ~/proj  pi "
  _cell "$panel_bg" "$overlay0" " ~/proj     "
  printf '\n'

  # ── Row 2: cpu/ram + pane bodies (active pane = brighter bg) ──────
  printf '%s' "$(sb "$subtext0" ' cpu·ram 3%')"
  pane "$panel_bg" "$text"    "$PA" ' fox jumps'
  pane "$surface_dim" "$overlay0" "$PB" ' idle'
  printf '\n'

  # ── Row 3: now-playing + accent line in active pane ───────────────
  printf '%s' "$(sb "$teal" ' ▶ Runaway')"
  pane "$panel_bg" "$accent"  "$PA" ' accent >'
  pane "$surface_dim" "$overlay1" "$PB" ''
  printf '\n'

  # ── Row 4: second space (dim) + pane status colors ────────────────
  printf '%s' "$(sb "$overlay0" ' · LINUX')"
  pane "$panel_bg" "$green"   "$PA" ' ✓ done'
  pane "$surface_dim" "$red"      "$PB" ' ✗ err'
  printf '\n'

  # ── Divider ───────────────────────────────────────────────────────
  _cell "$surface_dim" "$overlay0" "$(_padr $((SB+PA+PB)) '')"; printf '\n'

  # ── Agent pane (bottom): state dots ───────────────────────────────
  _cell "$panel_bg" "$subtext0" ' AGENTS  '
  _fg   "$blue"  '● '; _fg "$green" '● '; _fg "$red" '● '; _fg "$overlay0" '● '
  printf '\n '
  _fg "$blue" 'claude '; _fg "$green" 'codex '; _fg "$red" 'hermes '; _fg "$overlay0" 'idle'
  printf '\n\n'

  # ── ANSI palette strip ────────────────────────────────────────────
  printf ' '
  local col
  for col in "$red" "$green" "$yellow" "$blue" "$mauve" "$teal" "$peach" "$accent"; do
    _cell "$col" "$panel_bg" '  '
  done
  printf '\n'
}

# Called by fzf to render one row's preview. Arg may carry the list prefix.
if [ "${1:-}" = "--preview" ]; then
  row="${2:-}"
  case "$row" in
    "+ Add from clipboard"*) echo "  Reads a ghostty-format palette from your clipboard,"; echo "  asks a name, saves and applies — no editor."; exit 0 ;;
    "+ Add new theme"*) echo "  Opens an editor (nvim if present) to paste a"; echo "  ghostty-format theme; name it, saved and applied."; exit 0 ;;
  esac
  row="${row#✓ }"; row="${row#★ }"; row="${row#  }"
  swatch "$row"; exit 0
fi

main() {
  command -v fzf >/dev/null || { echo "fzf is not installed"; read -r; exit 1; }

  local applied=""
  [ -f "$APPLIED_FILE" ] && applied="$(head -1 "$APPLIED_FILE" 2>/dev/null)"

  # Merge bundled index + user-added index (deduped, user themes marked ★).
  local bundled user_slugs=""
  bundled="$(grep -v '^[[:space:]]*$' "$PLUGIN_ROOT/themes/index.txt")"
  [ -f "$USER_INDEX" ] && user_slugs="$(grep -v '^[[:space:]]*$' "$USER_INDEX" || true)"

  local ADD_ROW="+ Add new theme…"
  local ADD_CLIP_ROW="+ Add from clipboard…"

  # Compose display list: applied (✓) first, then user themes (★), then bundled.
  build_display() {
    local applied="$1" bundled="$2" user_slugs="$3"
    # user themes not equal to applied
    printf '%s\n' "$user_slugs" | while IFS= read -r s; do
      [ -z "$s" ] && continue
      [ "$s" = "$applied" ] && continue
      printf '★ %s\n' "$s"
    done
    # bundled not equal to applied and not already a user slug
    printf '%s\n' "$bundled" | while IFS= read -r s; do
      [ -z "$s" ] && continue
      [ "$s" = "$applied" ] && continue
      printf '%s\n' "$user_slugs" | grep -qxF "$s" && continue
      printf '  %s\n' "$s"
    done
  }

  local rest; rest="$(build_display "$applied" "$bundled" "$user_slugs")"
  local display=""
  if [ -n "$applied" ]; then
    display="$(printf '✓ %s\n' "$applied")"$'\n'"$rest"
  else
    display="$rest"
  fi
  # Add-theme entries pinned at the bottom.
  display="$display"$'\n'"$ADD_ROW"$'\n'"$ADD_CLIP_ROW"

  # --expect=tab lets Tab trigger add-theme on the highlighted-or-not row too.
  local out key sel
  out="$(printf '%s\n' "$display" | fzf \
    --layout=reverse \
    --prompt="Search themes: " \
    --header="type to search · ↵ on '+ Add' to paste in editor · '+ Add from clipboard' to skip editor · esc cancel" \
    --info=inline \
    --height=100% \
    --expect=tab \
    --preview="bash '$PLUGIN_ROOT/bin/picker.sh' --preview {}" \
    --preview-window=right,62%,border-left)" || exit 0

  key="$(printf '%s\n' "$out" | sed -n '1p')"
  sel="$(printf '%s\n' "$out" | sed -n '2p')"

  # Add from clipboard → no editor, straight to name+save.
  if [ "$sel" = "$ADD_CLIP_ROW" ]; then
    local new
    new="$(bash "$PLUGIN_ROOT/bin/add.sh" --clipboard | tail -1)" || exit 0
    if [ -n "$new" ] && printf '%s\n' "$new" | grep -qxE '[a-z0-9-]+'; then
      bash "$PLUGIN_ROOT/bin/apply.sh" "$new"
    fi
    exit 0
  fi

  # Strip row prefixes back to the bare slug.
  sel="${sel#✓ }"; sel="${sel#★ }"; sel="${sel#  }"

  # Tab, or choosing the add row → open the theme editor, then apply on save.
  if [ "$key" = "tab" ] || [ "$sel" = "$ADD_ROW" ]; then
    local new
    new="$(bash "$PLUGIN_ROOT/bin/add.sh" | tail -1)" || exit 0
    # add.sh prints the slug on its last line only on success.
    if [ -n "$new" ] && printf '%s\n' "$new" | grep -qxE '[a-z0-9-]+'; then
      bash "$PLUGIN_ROOT/bin/apply.sh" "$new"
    fi
    exit 0
  fi

  [ -n "$sel" ] || exit 0
  bash "$PLUGIN_ROOT/bin/apply.sh" "$sel"
  # Popup closes automatically when this process exits — no keypress needed.
}

if [ "${BASH_SOURCE[0]}" = "$0" ] && [ "${1:-}" != "--preview" ]; then main "$@"; fi

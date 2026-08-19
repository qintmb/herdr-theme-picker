#!/usr/bin/env bash
set -euo pipefail
_here="$(dirname "${BASH_SOURCE[0]}")"
source "$_here/fetch.sh"
source "$_here/map.sh"
PLUGIN_ROOT="${HERDR_PLUGIN_ROOT:-$(cd "$_here/.." && pwd)}"

# "#rrggbb" -> "r;g;b" (for ANSI truecolor).
_rgb() { local h="${1#\#}"; printf '%d;%d;%d' "$((16#${h:0:2}))" "$((16#${h:2:2}))" "$((16#${h:4:2}))"; }

# paint <bg#hex> <fg#hex> <text>  -> text on colored cell, reset after.
_paint() { printf '\033[48;2;%s;38;2;%sm%s\033[0m' "$(_rgb "$1")" "$(_rgb "$2")" "$3"; }

# swatch <slug>: mock of Herdr UI (sidebar/space, agent panes, tabs) in the
# theme's colors. Used as the fzf --preview. Fetches on demand.
swatch() {
  local slug="$1" p tok
  p="$(resolve_palette "$slug" 2>/dev/null)" || { echo "(cannot load $slug — needs internet)"; return 0; }
  tok="$(palette_to_tokens "$p")"
  # tok is "token=#hex" lines; look one up (bash 3.2, no assoc arrays).
  tget() { printf '%s\n' "$tok" | grep "^$1=" | cut -d= -f2-; }
  local panel_bg text accent blue green red yellow mauve teal peach
  local subtext0 surface0 surface1 surface_dim overlay0
  panel_bg="$(tget panel_bg)"; text="$(tget text)"; accent="$(tget accent)"
  blue="$(tget blue)"; green="$(tget green)"; red="$(tget red)"
  yellow="$(tget yellow)"; mauve="$(tget mauve)"; teal="$(tget teal)"
  peach="$(tget peach)"; subtext0="$(tget subtext0)"
  surface0="$(tget surface0)"; surface1="$(tget surface1)"
  surface_dim="$(tget surface_dim)"; overlay0="$(tget overlay0)"

  local w=30
  local pad; pad="$(printf '%*s' "$w" '')"

  printf '  %s\n\n' "$slug"

  # Sidebar / space header
  _paint "$panel_bg" "$accent" "$(printf ' ■ Workspace%*s' $((w-12)) '')"; printf '\n'
  # Agent panes: dot = state color, name in text, status in subtext
  _paint "$surface0" "$blue"     "$(printf ' ● claude   %*s' $((w-12)) 'working')"; printf '\n'
  _paint "$panel_bg" "$green"    "$(printf ' ● codex    %*s' $((w-12)) 'done')"; printf '\n'
  _paint "$surface0" "$red"      "$(printf ' ● hermes   %*s' $((w-12)) 'blocked')"; printf '\n'
  _paint "$panel_bg" "$subtext0" "$(printf ' ● opencode %*s' $((w-12)) 'idle')"; printf '\n'

  # Divider
  _paint "$surface_dim" "$overlay0" "$pad"; printf '\n'

  # Tab bar: active tab uses accent bg, inactive use surface1
  printf ' '
  _paint "$accent"   "$panel_bg" ' main '
  _paint "$surface1" "$subtext0" ' logs '
  _paint "$surface1" "$subtext0" ' test '
  printf '\n'

  # Pane body: text + accent samples
  _paint "$panel_bg" "$text"   "$(printf ' The quick brown fox %*s' $((w-20)) '')"; printf '\n'
  _paint "$panel_bg" "$accent" "$(printf ' accent link  %*s' $((w-13)) '')"; printf '\n'

  # ANSI palette strip
  printf '\n '
  local col
  for col in "$red" "$green" "$yellow" "$blue" "$mauve" "$teal" "$peach" "$accent"; do
    _paint "$col" "$panel_bg" '  '
  done
  printf '\n'
}

# Called by fzf to render one row's preview. Arg may carry the list prefix.
if [ "${1:-}" = "--preview" ]; then
  row="${2:-}"; row="${row#✓ }"; row="${row#  }"
  swatch "$row"; exit 0
fi

main() {
  command -v fzf >/dev/null || { echo "fzf is not installed"; read -r; exit 1; }

  local applied=""
  [ -f "$APPLIED_FILE" ] && applied="$(head -1 "$APPLIED_FILE" 2>/dev/null)"

  # Build the list; prefix the active theme with a check mark. fzf shows the
  # marker but we strip it back to the bare slug on selection.
  local list; list="$(cat "$PLUGIN_ROOT/themes/index.txt")"
  local display
  display="$(printf '%s\n' "$list" | while IFS= read -r slug; do
    [ -z "$slug" ] && continue
    if [ "$slug" = "$applied" ]; then printf '✓ %s\n' "$slug"; else printf '  %s\n' "$slug"; fi
  done)"

  local sel
  if [ -n "$applied" ]; then
    sel="$(printf '%s\n' "$display" | fzf \
      --prompt="theme> " \
      --header="↑↓ browse · enter apply · esc cancel" \
      --info=inline --height=100% \
      --query "$applied" \
      --preview="bash '$PLUGIN_ROOT/bin/picker.sh' --preview {}" \
      --preview-window=right,60%,border-left)" || exit 0
  else
    sel="$(printf '%s\n' "$display" | fzf \
      --prompt="theme> " \
      --header="↑↓ browse · enter apply · esc cancel" \
      --info=inline --height=100% \
      --preview="bash '$PLUGIN_ROOT/bin/picker.sh' --preview {}" \
      --preview-window=right,60%,border-left)" || exit 0
  fi

  # Strip the "✓ " / "  " prefix back to the bare slug.
  sel="${sel#✓ }"; sel="${sel#  }"
  [ -n "$sel" ] || exit 0

  bash "$PLUGIN_ROOT/bin/apply.sh" "$sel"
  # Popup closes automatically when this process exits — no keypress needed.
}

if [ "${BASH_SOURCE[0]}" = "$0" ] && [ "${1:-}" != "--preview" ]; then main "$@"; fi

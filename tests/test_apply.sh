#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source bin/apply.sh

tmp="$(mktemp -d)"
cfg="$tmp/config.toml"
printf '[keys]\nprefix = "cmd+b"\n\n[theme]\nname = "solarized"\n' > "$cfg"

tokens=$'panel_bg=#282a36\ntext=#f8f8f2\nred=#ff5555'
write_custom_block "$cfg" "$tokens"

fail=0
grep -q '^\[theme.custom\]' "$cfg" || { echo "FAIL block added"; fail=1; }
grep -q '^panel_bg = "#282a36"' "$cfg" || { echo "FAIL panel_bg quoted"; fail=1; }
grep -q '^name = "solarized"' "$cfg" || { echo "FAIL preserved theme.name"; fail=1; }

# idempoten: tulis lagi, tak duplikat blok
write_custom_block "$cfg" "$tokens"
n="$(grep -c '^\[theme.custom\]' "$cfg")"
[ "$n" = 1 ] || { echo "FAIL duplicate block: $n"; fail=1; }

# symlink utuh: config.toml yang di-symlink tetap symlink, target dapat blok (issue #3)
linktmp="$(mktemp -d)"
mkdir -p "$linktmp/dotfiles" "$linktmp/cfgdir"
printf '[theme]\nname = "nord"\n' > "$linktmp/dotfiles/config.toml"
ln -s "$linktmp/dotfiles/config.toml" "$linktmp/cfgdir/config.toml"
write_custom_block "$linktmp/cfgdir/config.toml" "$tokens"
[ -L "$linktmp/cfgdir/config.toml" ] || { echo "FAIL symlink replaced by file"; fail=1; }
grep -q '^\[theme\.custom\]' "$linktmp/dotfiles/config.toml" \
  || { echo "FAIL symlink target not updated"; fail=1; }
[ -z "$(ls "$linktmp/dotfiles"/*.tmp 2>/dev/null)" ] \
  || { echo "FAIL stray tmp file left behind"; fail=1; }

# file biasa (bukan symlink) tetap ditulis di tempat, tanpa tmp nyangkut
plaintmp="$(mktemp -d)"
plain="$plaintmp/config.toml"
printf '[theme]\nname = "gruvbox"\n' > "$plain"
write_custom_block "$plain" "$tokens"
[ -L "$plain" ] && { echo "FAIL plain file became a symlink"; fail=1; }
grep -q '^\[theme\.custom\]' "$plain" || { echo "FAIL plain file not updated"; fail=1; }
[ -z "$(ls "$plaintmp"/*.tmp 2>/dev/null)" ] \
  || { echo "FAIL stray tmp next to plain file"; fail=1; }

# dangling symlink: target belum ada tapi link harus tetap link setelah apply (issue #3)
dang="$(mktemp -d)"
mkdir -p "$dang/dotfiles"
ln -s "$dang/dotfiles/config.toml" "$dang/link.toml"
printf '[theme]\nname = "nord"\n' > "$dang/dotfiles/config.toml"
write_custom_block "$dang/link.toml" "$tokens"
[ -L "$dang/link.toml" ] || { echo "FAIL dangling symlink replaced"; fail=1; }
grep -q '^\[theme\.custom\]' "$dang/dotfiles/config.toml" \
  || { echo "FAIL dangling target not written"; fail=1; }

# chain symlink: link -> link -> file, target akhir yang harus menerima blok
chain="$(mktemp -d)"
mkdir -p "$chain/real"
printf '[theme]\nname = "nord"\n' > "$chain/real/config.toml"
ln -s "$chain/real/config.toml" "$chain/step1.toml"
ln -s "$chain/step1.toml" "$chain/step2.toml"
write_custom_block "$chain/step2.toml" "$tokens"
[ -L "$chain/step2.toml" ] || { echo "FAIL chain link2 replaced"; fail=1; }
[ -L "$chain/step1.toml" ] || { echo "FAIL chain link1 replaced"; fail=1; }
grep -q '^\[theme\.custom\]' "$chain/real/config.toml" \
  || { echo "FAIL chain final target not written"; fail=1; }
[ -z "$(ls "$chain/real"/*.tmp 2>/dev/null)" ] \
  || { echo "FAIL chain stray tmp"; fail=1; }

# resolve_symlink langsung: path non-symlink kembali apa adanya (absolut)
rs="$(mktemp -d)"; : > "$rs/plain.toml"
out="$(resolve_symlink "$rs/plain.toml")"
[ "$out" = "$(cd "$rs" && pwd)/plain.toml" ] \
  || { echo "FAIL resolve_symlink plain: got '$out'"; fail=1; }

# circular symlink: resolver must bail out instead of looping forever
circ="$(mktemp -d)"
ln -s "$circ/b.toml" "$circ/a.toml"
ln -s "$circ/a.toml" "$circ/b.toml"
if resolve_symlink "$circ/a.toml" >/dev/null 2>&1; then
  echo "FAIL circular symlink not rejected"; fail=1
fi

[ "$fail" = 0 ] && echo "PASS test_apply" || exit 1

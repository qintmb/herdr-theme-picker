# herdr-theme-picker Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Plugin herdr lokal yang lewat `prefix+t` membuka picker theme (bundle offline + fetch terminalcolors.com) lalu menulis `[theme.custom]` ke config.toml herdr dan reload.

**Architecture:** Semua bash + fzf + curl. Palette diambil format ghostty (`background`/`palette N=#hex`), dipetakan ke 16 token UI herdr, ditulis sebagai blok `[theme.custom]` (dengan backup), lalu `herdr server reload-config`.

**Tech Stack:** bash, fzf, curl, herdr CLI. Tanpa runtime tambahan.

**Spec:** `docs/superpowers/specs/2026-08-19-herdr-theme-picker-design.md`

## Global Constraints

- Plugin id: `herdr-theme-picker`. `min_herdr_version = "0.8.0"`. platforms `["linux","macos"]`.
- Config herdr: `${HERDR_CONFIG_PATH:-$HOME/.config/herdr/config.toml}`. JANGAN buat file baru bila tak ada — error.
- Token `[theme.custom]` herdr (16): `accent, panel_bg, surface0, surface1, surface_dim, overlay0, overlay1, text, subtext0, mauve, green, yellow, red, blue, teal, peach`.
- Sumber palette live: `https://terminalcolors.com/downloads/ghostty/<slug>` (plaintext). Slug hanya `[a-z0-9-]+`; tolak selain itu (anti path/URL injection).
- Reload: `herdr server reload-config`.
- Keybind: `key="prefix+t"`, `type="plugin_action"`, `command="herdr-theme-picker.open"`.
- Semua skrip `set -euo pipefail`. Cache: `${XDG_CACHE_HOME:-$HOME/.cache}/herdr-theme-picker`.
- Test = skrip bash assert (`bash tests/xxx.sh`), tanpa framework.

---

### Task 1: Scaffold + manifest + darken helper (dengan test)

**Files:**
- Create: `herdr-plugin.toml`
- Create: `bin/lib.sh`
- Test: `tests/test_lib.sh`
- Create: `.gitignore`

**Interfaces:**
- Produces: `bin/lib.sh` mengekspor fungsi `is_valid_slug <s>` (return 0/1), `darken_hex <#rrggbb> <percent>` (echo `#rrggbb`), `die <msg>` (stderr + exit 1), `CONFIG_PATH` (var), `CACHE_DIR` (var).

- [ ] **Step 1: Tulis test yang gagal**

`tests/test_lib.sh`:
```bash
#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source bin/lib.sh

fail=0
check() { if [ "$1" != "$2" ]; then echo "FAIL $3: got '$1' want '$2'"; fail=1; fi; }

# darken: setiap kanal turun ~10%
check "$(darken_hex '#282a36' 10)" "#242530" "darken dracula bg"
check "$(darken_hex '#000000' 50)" "#000000" "darken black clamps"
check "$(darken_hex '#ffffff' 100)" "#000000" "darken 100 = black"

# slug valid
if is_valid_slug "dracula-default"; then :; else echo "FAIL slug ok"; fail=1; fi
if is_valid_slug "../../etc/passwd"; then echo "FAIL slug reject path"; fail=1; fi
if is_valid_slug "a b"; then echo "FAIL slug reject space"; fail=1; fi

[ "$fail" = 0 ] && echo "PASS test_lib" || exit 1
```

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `bash tests/test_lib.sh`
Expected: FAIL — `bin/lib.sh` belum ada (source error).

- [ ] **Step 3: Tulis `bin/lib.sh`**

```bash
#!/usr/bin/env bash
# Fungsi bersama untuk herdr-theme-picker.
CONFIG_PATH="${HERDR_CONFIG_PATH:-$HOME/.config/herdr/config.toml}"
CACHE_DIR="${XDG_CACHE_HOME:-$HOME/.cache}/herdr-theme-picker"

die() { printf '%s\n' "$*" >&2; exit 1; }

is_valid_slug() {
  [[ "$1" =~ ^[a-z0-9-]+$ ]]
}

# darken_hex #rrggbb percent -> #rrggbb (tiap kanal * (100-p)/100, clamp 0)
darken_hex() {
  local hex="${1#\#}" p="$2" r g b
  r=$((16#${hex:0:2})); g=$((16#${hex:2:2})); b=$((16#${hex:4:2}))
  r=$(( r * (100 - p) / 100 )); g=$(( g * (100 - p) / 100 )); b=$(( b * (100 - p) / 100 ))
  (( r < 0 )) && r=0; (( g < 0 )) && g=0; (( b < 0 )) && b=0
  printf '#%02x%02x%02x\n' "$r" "$g" "$b"
}
```

- [ ] **Step 4: Jalankan, pastikan lulus**

Run: `bash tests/test_lib.sh`
Expected: `PASS test_lib`

(Jika `darken_hex '#282a36' 10` ≠ `#242530`, hitung ulang nilai harapan di test dari rumus — bukan ubah rumus. `0x28*90/100=36=0x24`, `0x2a*90/100=38=0x26`… sesuaikan angka test agar cocok dengan aritmetika integer.)

- [ ] **Step 5: Tulis `herdr-plugin.toml`**

```toml
id = "herdr-theme-picker"
name = "Theme Picker"
version = "0.1.0"
min_herdr_version = "0.8.0"
description = "Pick any terminalcolors.com theme and apply it to Herdr's UI via [theme.custom]."
platforms = ["linux", "macos"]

[[panes]]
id = "picker"
title = "Theme Picker"
placement = "popup"
width = "80%"
height = "70%"
command = ["bash", "-c", "exec bash \"$HERDR_PLUGIN_ROOT/bin/picker.sh\""]

[[actions]]
id = "open"
title = "Theme picker: open"
contexts = ["workspace"]
command = ["bash", "-c", "exec \"${HERDR_BIN_PATH:-herdr}\" plugin pane open --plugin herdr-theme-picker --entrypoint picker --placement popup --focus"]
```

- [ ] **Step 6: Tulis `.gitignore`**

```
*.bak-*
.DS_Store
```

- [ ] **Step 7: Commit**

```bash
git add herdr-plugin.toml bin/lib.sh tests/test_lib.sh .gitignore
git commit -m "feat: scaffold plugin manifest + lib helpers with tests"
```

---
### Task 2: Parse palette ghostty → 16 token herdr

**Files:**
- Create: `bin/map.sh`
- Test: `tests/test_map.sh`
- Create: `tests/fixtures/dracula-default` (palette ghostty contoh)

**Interfaces:**
- Consumes: `bin/lib.sh` (`darken_hex`).
- Produces: `bin/map.sh` mengekspor `palette_to_tokens <file>` yang membaca palette format ghostty dari `<file>` dan mencetak 16 baris `token=#hex` (urutan tabel spec). Exit 1 bila `palette 0=` tak ditemukan.

- [ ] **Step 1: Tulis fixture `tests/fixtures/dracula-default`**

```
background = #282a36
foreground = #f8f8f2
selection-background = #44475a
cursor-color = #f8f8f2
palette = 0=#21222c
palette = 1=#ff5555
palette = 2=#50fa7b
palette = 3=#f1fa8c
palette = 4=#bd93f9
palette = 5=#ff79c6
palette = 6=#8be9fd
palette = 7=#f8f8f2
palette = 8=#6272a4
palette = 9=#ff6e6e
palette = 10=#69ff94
palette = 11=#ffffa5
palette = 12=#d6acff
palette = 13=#ff92df
palette = 14=#a4ffff
palette = 15=#ffffff
```

- [ ] **Step 2: Tulis test yang gagal**

`tests/test_map.sh`:
```bash
#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source bin/map.sh

out="$(palette_to_tokens tests/fixtures/dracula-default)"
get() { printf '%s\n' "$out" | grep "^$1=" | cut -d= -f2-; }
fail=0
check() { if [ "$(get "$1")" != "$2" ]; then echo "FAIL $1: got '$(get "$1")' want '$2'"; fail=1; fi; }

check panel_bg "#282a36"
check text "#f8f8f2"
check red "#ff5555"
check green "#50fa7b"
check yellow "#f1fa8c"
check blue "#bd93f9"
check accent "#bd93f9"
check teal "#8be9fd"
check mauve "#ff79c6"
check surface0 "#21222c"
check surface1 "#6272a4"
check overlay0 "#6272a4"
check overlay1 "#f8f8f2"
check subtext0 "#f8f8f2"
check peach "#ff6e6e"

# surface_dim harus lebih gelap dari panel_bg (bukan sama)
if [ "$(get surface_dim)" = "#282a36" ]; then echo "FAIL surface_dim not darkened"; fail=1; fi

# jumlah token tepat 16
n="$(printf '%s\n' "$out" | grep -c '=')"
[ "$n" = 16 ] || { echo "FAIL token count: $n"; fail=1; }

[ "$fail" = 0 ] && echo "PASS test_map" || exit 1
```

- [ ] **Step 3: Jalankan, pastikan gagal**

Run: `bash tests/test_map.sh`
Expected: FAIL — `bin/map.sh` belum ada.

- [ ] **Step 4: Tulis `bin/map.sh`**

```bash
#!/usr/bin/env bash
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"

# Baca satu key dari palette ghostty. Untuk 'palette N', key = "pN".
_pget() {
  local file="$1" key="$2"
  if [[ "$key" == p* ]]; then
    grep -E "^palette *= *${key#p}=" "$file" | head -1 | sed 's/.*=//'
  else
    grep -E "^${key} *=" "$file" | head -1 | sed 's/^[^=]*= *//'
  fi
}

# palette_to_tokens <file> -> 16 baris "token=#hex"
palette_to_tokens() {
  local f="$1" bg fg p0 p1 p2 p3 p4 p5 p6 p7 p8 p9
  p0="$(_pget "$f" p0)"
  [ -n "$p0" ] || die "bukan palette ghostty valid: $f"
  bg="$(_pget "$f" background)"; fg="$(_pget "$f" foreground)"
  p1="$(_pget "$f" p1)"; p2="$(_pget "$f" p2)"; p3="$(_pget "$f" p3)"
  p4="$(_pget "$f" p4)"; p5="$(_pget "$f" p5)"; p6="$(_pget "$f" p6)"
  p7="$(_pget "$f" p7)"; p8="$(_pget "$f" p8)"; p9="$(_pget "$f" p9)"
  : "${p9:=$p3}"  # fallback peach
  printf 'panel_bg=%s\n' "$bg"
  printf 'surface0=%s\n' "$p0"
  printf 'surface1=%s\n' "$p8"
  printf 'surface_dim=%s\n' "$(darken_hex "$bg" 8)"
  printf 'overlay0=%s\n' "$p8"
  printf 'overlay1=%s\n' "$p7"
  printf 'text=%s\n' "$fg"
  printf 'subtext0=%s\n' "$p7"
  printf 'accent=%s\n' "$p4"
  printf 'mauve=%s\n' "$p5"
  printf 'green=%s\n' "$p2"
  printf 'yellow=%s\n' "$p3"
  printf 'red=%s\n' "$p1"
  printf 'blue=%s\n' "$p4"
  printf 'teal=%s\n' "$p6"
  printf 'peach=%s\n' "$p9"
}
```

- [ ] **Step 5: Jalankan, pastikan lulus**

Run: `bash tests/test_map.sh`
Expected: `PASS test_map`

- [ ] **Step 6: Commit**

```bash
git add bin/map.sh tests/test_map.sh tests/fixtures/dracula-default
git commit -m "feat: map ghostty palette to 16 herdr theme tokens"
```

---
### Task 3: Fetch palette (bundle-or-live) dengan cache

**Files:**
- Create: `bin/fetch.sh`
- Create: `themes/dracula-default` (salin dari fixture — bundle nyata pertama)
- Test: `tests/test_fetch.sh`

**Interfaces:**
- Consumes: `bin/lib.sh` (`is_valid_slug`, `die`, `CACHE_DIR`).
- Produces: `bin/fetch.sh` mengekspor `resolve_palette <slug>` → mencetak path file palette yang bisa dibaca (`themes/<slug>` bila ada; else cache; else fetch ke cache). Exit 1 bila slug invalid / offline+tak ada / fetch gagal. `PLUGIN_ROOT` var = akar plugin.

- [ ] **Step 1: Buat bundle `themes/dracula-default`** (isi sama dgn `tests/fixtures/dracula-default` dari Task 2 — salin file).

- [ ] **Step 2: Tulis test yang gagal**

`tests/test_fetch.sh`:
```bash
#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source bin/fetch.sh

fail=0
# bundle resolve tanpa network
p="$(resolve_palette dracula-default)"
[ -f "$p" ] && grep -q '^palette = 0=' "$p" || { echo "FAIL bundle resolve"; fail=1; }

# slug invalid ditolak
if resolve_palette "../etc" 2>/dev/null; then echo "FAIL slug guard"; fail=1; fi

[ "$fail" = 0 ] && echo "PASS test_fetch" || exit 1
```

- [ ] **Step 3: Jalankan, pastikan gagal**

Run: `bash tests/test_fetch.sh`
Expected: FAIL — `bin/fetch.sh` belum ada.

- [ ] **Step 4: Tulis `bin/fetch.sh`**

```bash
#!/usr/bin/env bash
source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
PLUGIN_ROOT="${HERDR_PLUGIN_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
BASE_URL="https://terminalcolors.com/downloads/ghostty"

# resolve_palette <slug> -> path file palette (stdout)
resolve_palette() {
  local slug="$1"
  is_valid_slug "$slug" || die "slug tidak valid: $slug"

  local bundle="$PLUGIN_ROOT/themes/$slug"
  [ -f "$bundle" ] && { printf '%s\n' "$bundle"; return 0; }

  local cache="$CACHE_DIR/$slug"
  [ -f "$cache" ] && { printf '%s\n' "$cache"; return 0; }

  mkdir -p "$CACHE_DIR"
  local tmp; tmp="$(mktemp)"
  if ! curl -fsSL -m 15 "$BASE_URL/$slug" -o "$tmp" 2>/dev/null; then
    rm -f "$tmp"; die "gagal fetch '$slug' (offline atau tidak ada). Butuh internet."
  fi
  grep -q '^palette = 0=' "$tmp" || { rm -f "$tmp"; die "respons bukan palette valid untuk '$slug'"; }
  mv "$tmp" "$cache"
  printf '%s\n' "$cache"
}
```

- [ ] **Step 5: Jalankan, pastikan lulus**

Run: `bash tests/test_fetch.sh`
Expected: `PASS test_fetch`

- [ ] **Step 6: Commit**

```bash
git add bin/fetch.sh themes/dracula-default tests/test_fetch.sh
git commit -m "feat: resolve palette from bundle or live fetch with cache"
```

---

### Task 4: Terapkan theme ke config.toml + reload

**Files:**
- Create: `bin/apply.sh`
- Test: `tests/test_apply.sh`

**Interfaces:**
- Consumes: `bin/map.sh` (`palette_to_tokens`), `bin/fetch.sh` (`resolve_palette`), `bin/lib.sh` (`CONFIG_PATH`, `die`).
- Produces: `bin/apply.sh` — dijalankan `apply.sh <slug>`. Menulis/mengganti blok `[theme.custom]` di `CONFIG_PATH`, backup sekali/hari, lalu `herdr server reload-config`. Fungsi internal `write_custom_block <config> <<<tokens>` dapat diuji terpisah.

- [ ] **Step 1: Tulis test yang gagal**

`tests/test_apply.sh`:
```bash
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

[ "$fail" = 0 ] && echo "PASS test_apply" || exit 1
```

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `bash tests/test_apply.sh`
Expected: FAIL — `bin/apply.sh` belum ada.

- [ ] **Step 3: Tulis `bin/apply.sh`**

```bash
#!/usr/bin/env bash
_here="$(dirname "${BASH_SOURCE[0]}")"
source "$_here/map.sh"
source "$_here/fetch.sh"

# Hapus blok [theme.custom] lama (sampai header berikutnya / EOF), lalu append baru.
write_custom_block() {
  local cfg="$1" tokens="$2"
  awk '
    /^\[theme\.custom\]/ { skip=1; next }
    skip && /^\[/ { skip=0 }
    !skip { print }
  ' "$cfg" > "$cfg.tmp"
  {
    printf '\n[theme.custom]\n'
    while IFS='=' read -r k v; do
      [ -n "$k" ] && printf '%s = "%s"\n' "$k" "$v"
    done <<< "$tokens"
  } >> "$cfg.tmp"
  mv "$cfg.tmp" "$cfg"
}

main() {
  local slug="${1:-}"
  [ -n "$slug" ] || die "pemakaian: apply.sh <slug>"
  [ -f "$CONFIG_PATH" ] || die "config tak ditemukan: $CONFIG_PATH"

  local palette tokens
  palette="$(resolve_palette "$slug")"
  tokens="$(palette_to_tokens "$palette")"

  local stamp; stamp="$(date +%Y%m%d)"
  local bak="$CONFIG_PATH.bak-$stamp"
  [ -f "$bak" ] || cp "$CONFIG_PATH" "$bak"

  write_custom_block "$CONFIG_PATH" "$tokens"

  if herdr server reload-config >/dev/null 2>&1; then
    printf 'Theme "%s" diterapkan.\n' "$slug"
  else
    printf 'Theme "%s" ditulis, tapi reload gagal. Coba: herdr server reload-config\n' "$slug" >&2
  fi
}

# Jalankan main hanya bila dieksekusi langsung (bukan di-source oleh test).
if [ "${BASH_SOURCE[0]}" = "$0" ]; then main "$@"; fi
```

- [ ] **Step 4: Jalankan, pastikan lulus**

Run: `bash tests/test_apply.sh`
Expected: `PASS test_apply`

- [ ] **Step 5: Commit**

```bash
git add bin/apply.sh tests/test_apply.sh
git commit -m "feat: write [theme.custom] block idempotently and reload"
```

---
### Task 5: Bundle theme populer + katalog live

**Files:**
- Create: `themes/gruvbox-dark`, `themes/nord-default`, `themes/tokyo-night-default`, `themes/catppuccin-mocha`, `themes/solarized-dark`, `themes/one-dark`, `themes/monokai-default`, `themes/everforest-dark`, `themes/rose-pine`, `themes/kanagawa-default`, `themes/ayu-dark`, `themes/night-owl`, `themes/material-default`, `themes/tokyo-night-storm` (14 tambahan; total 15 dgn dracula)
- Create: `themes/index.txt` (katalog slug live kurasi)
- Test: `tests/test_bundle.sh`

**Interfaces:**
- Consumes: `bin/map.sh` (validasi tiap bundle bisa dipetakan).
- Produces: direktori `themes/` berisi ≥15 palette ghostty valid + `index.txt` (satu slug per baris).

- [ ] **Step 1: Unduh 14 bundle** (jalankan sekali; commit hasilnya sbg file statis):

```bash
cd "$(git rev-parse --show-toplevel)"
slugs="gruvbox-dark nord-default tokyo-night-default catppuccin-mocha solarized-dark one-dark monokai-default everforest-dark rose-pine kanagawa-default ayu-dark night-owl material-default tokyo-night-storm"
for s in $slugs; do
  curl -fsSL -m 20 "https://terminalcolors.com/downloads/ghostty/$s" -o "themes/$s" \
    && grep -q '^palette = 0=' "themes/$s" \
    && echo "ok $s" || echo "SKIP $s (cek slug di terminalcolors.com)"
done
```

Jika sebuah slug gagal (situs pakai nama/variant berbeda), buka `https://terminalcolors.com/` cari slug benar (pola `<name>-<variant>`), ganti, ulang. Pastikan akhirnya ≥15 file valid di `themes/`.

- [ ] **Step 2: Tulis `themes/index.txt`** — katalog live (bundle + slug populer lain). Minimal isi ke-15 bundle plus tambahan; satu slug per baris, contoh:

```
dracula-default
gruvbox-dark
gruvbox-light
nord-default
tokyo-night-default
tokyo-night-storm
tokyo-night-day
catppuccin-mocha
catppuccin-latte
catppuccin-frappe
catppuccin-macchiato
solarized-dark
solarized-light
one-dark
one-light
monokai-default
everforest-dark
everforest-light
rose-pine
rose-pine-moon
rose-pine-dawn
kanagawa-default
ayu-dark
ayu-mirage
ayu-light
night-owl
material-default
github-dark
github-light
```
(Slug yang belum terkonfirmasi tetap boleh — `resolve_palette` akan fetch saat dipilih; bila 404 tampil error, tak merusak config.)

- [ ] **Step 3: Tulis test yang gagal**

`tests/test_bundle.sh`:
```bash
#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source bin/map.sh

fail=0
n=0
for f in themes/*; do
  [ "$(basename "$f")" = "index.txt" ] && continue
  if palette_to_tokens "$f" >/dev/null 2>&1; then n=$((n+1)); else echo "FAIL invalid bundle: $f"; fail=1; fi
done
[ "$n" -ge 15 ] || { echo "FAIL bundle count: $n (<15)"; fail=1; }
[ -f themes/index.txt ] && grep -q '^dracula-default$' themes/index.txt || { echo "FAIL index.txt"; fail=1; }

[ "$fail" = 0 ] && echo "PASS test_bundle ($n themes)" || exit 1
```

- [ ] **Step 4: Jalankan, pastikan lulus** (bundle sudah ada dari Step 1-2)

Run: `bash tests/test_bundle.sh`
Expected: `PASS test_bundle (15+ themes)`. Jika <15 valid, ulang Step 1 untuk slug yang gagal.

- [ ] **Step 5: Commit**

```bash
git add themes/
git commit -m "feat: bundle 15 popular themes + live catalog index"
```

---

### Task 6: Picker popup (fzf + preview swatch)

**Files:**
- Create: `bin/picker.sh`
- Test: `tests/test_preview.sh`

**Interfaces:**
- Consumes: `bin/fetch.sh` (`resolve_palette`), `bin/map.sh` (`palette_to_tokens`), `bin/apply.sh` (dipanggil sbg subprocess).
- Produces: `bin/picker.sh` (entrypoint pane). Fungsi internal `swatch <slug>` mencetak preview 16 warna via ANSI truecolor untuk fzf `--preview`.

- [ ] **Step 1: Tulis test yang gagal**

`tests/test_preview.sh`:
```bash
#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
export HERDR_PLUGIN_ROOT="$PWD"
source bin/picker.sh

out="$(swatch dracula-default)"
fail=0
# memuat kode escape truecolor (48;2)
printf '%s' "$out" | grep -q $'\033\[48;2;' || { echo "FAIL no truecolor bg in swatch"; fail=1; }
# menyebut beberapa hex
printf '%s' "$out" | grep -q '#282a36' || { echo "FAIL bg hex missing"; fail=1; }

[ "$fail" = 0 ] && echo "PASS test_preview" || exit 1
```

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `bash tests/test_preview.sh`
Expected: FAIL — `bin/picker.sh` belum ada.

- [ ] **Step 3: Tulis `bin/picker.sh`**

```bash
#!/usr/bin/env bash
set -euo pipefail
_here="$(dirname "${BASH_SOURCE[0]}")"
source "$_here/fetch.sh"
source "$_here/map.sh"
PLUGIN_ROOT="${HERDR_PLUGIN_ROOT:-$(cd "$_here/.." && pwd)}"

# Blok warna truecolor: bg=hex, teks hex.
_block() { local hex="${1#\#}"; printf '\033[48;2;%d;%d;%dm  %s  \033[0m' \
  "$((16#${hex:0:2}))" "$((16#${hex:2:2}))" "$((16#${hex:4:2}))" "$1"; }

# swatch <slug>: preview token utama (dipakai fzf --preview). Fetch bila perlu.
swatch() {
  local slug="$1" p tok
  p="$(resolve_palette "$slug" 2>/dev/null)" || { echo "(gagal memuat $slug — butuh internet)"; return 0; }
  tok="$(palette_to_tokens "$p")"
  printf '%s\n\n' "$slug"
  local key
  for key in panel_bg text accent red green yellow blue teal mauve peach; do
    printf '%-10s ' "$key"
    _block "$(printf '%s\n' "$tok" | grep "^$key=" | cut -d= -f2-)"
    printf '\n'
  done
}

# Bila dipanggil fzf untuk preview satu baris.
if [ "${1:-}" = "--preview" ]; then swatch "$2"; exit 0; fi

main() {
  command -v fzf >/dev/null || { echo "fzf tidak terpasang"; read -r; exit 1; }
  local list; list="$(cat "$PLUGIN_ROOT/themes/index.txt")"
  local sel
  sel="$(printf '%s\n' "$list" | fzf \
    --prompt="theme> " \
    --preview="bash '$PLUGIN_ROOT/bin/picker.sh' --preview {}" \
    --preview-window=right,50%)" || exit 0
  [ -n "$sel" ] || exit 0
  bash "$PLUGIN_ROOT/bin/apply.sh" "$sel"
  echo; echo "Tekan Enter untuk tutup."; read -r
}

if [ "${BASH_SOURCE[0]}" = "$0" ] && [ "${1:-}" != "--preview" ]; then main "$@"; fi
```

- [ ] **Step 4: Jalankan, pastikan lulus**

Run: `bash tests/test_preview.sh`
Expected: `PASS test_preview`

- [ ] **Step 5: Commit**

```bash
git add bin/picker.sh tests/test_preview.sh
git commit -m "feat: fzf theme picker with truecolor swatch preview"
```

---
### Task 7: Installer — link plugin + auto-patch keybind

**Files:**
- Create: `bin/install.sh`
- Create: `bin/uninstall.sh`
- Test: `tests/test_install.sh`

**Interfaces:**
- Consumes: `bin/lib.sh` (`CONFIG_PATH`, `die`).
- Produces: `bin/install.sh` — `herdr plugin link <root>` + tambah `[[keys.command]] prefix+t` (idempoten, backup). `bin/uninstall.sh` — `herdr plugin unlink` + hapus blok keybind. Fungsi internal `add_keybind <config>` / `remove_keybind <config>` teruji.

- [ ] **Step 1: Tulis test yang gagal**

`tests/test_install.sh`:
```bash
#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
source bin/install.sh

tmp="$(mktemp -d)"; cfg="$tmp/config.toml"
printf '[keys]\nprefix = "cmd+b"\n' > "$cfg"

fail=0
add_keybind "$cfg"
grep -q 'command = "herdr-theme-picker.open"' "$cfg" || { echo "FAIL keybind added"; fail=1; }
grep -q 'key = "prefix+t"' "$cfg" || { echo "FAIL key line"; fail=1; }

# idempoten
add_keybind "$cfg"
n="$(grep -c 'command = "herdr-theme-picker.open"' "$cfg")"
[ "$n" = 1 ] || { echo "FAIL dup keybind: $n"; fail=1; }

# remove
remove_keybind "$cfg"
grep -q 'herdr-theme-picker.open' "$cfg" && { echo "FAIL keybind not removed"; fail=1; } || true

[ "$fail" = 0 ] && echo "PASS test_install" || exit 1
```

- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `bash tests/test_install.sh`
Expected: FAIL — `bin/install.sh` belum ada.

- [ ] **Step 3: Tulis `bin/install.sh`**

```bash
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
```

- [ ] **Step 4: Tulis `bin/uninstall.sh`**

```bash
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
```

Catatan: `install.sh` di-source oleh `uninstall.sh`; karena guard `BASH_SOURCE == $0`, `main` install TIDAK jalan saat di-source. Aman.

- [ ] **Step 5: Jalankan, pastikan lulus**

Run: `bash tests/test_install.sh`
Expected: `PASS test_install`

- [ ] **Step 6: Commit**

```bash
git add bin/install.sh bin/uninstall.sh tests/test_install.sh
git commit -m "feat: installer links plugin and auto-patches prefix+t keybind"
```

---

### Task 8: Runner tes + README + verifikasi live end-to-end

**Files:**
- Create: `tests/run.sh`
- Create: `README.md`

**Interfaces:**
- Consumes: semua `tests/test_*.sh`.
- Produces: `tests/run.sh` menjalankan seluruh test, exit ≠0 bila ada yang gagal.

- [ ] **Step 1: Tulis `tests/run.sh`**

```bash
#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "$0")/.."
rc=0
for t in tests/test_*.sh; do
  echo "== $t =="
  bash "$t" || rc=1
done
[ "$rc" = 0 ] && echo "ALL PASS" || echo "SOME FAILED"
exit "$rc"
```

- [ ] **Step 2: Jalankan seluruh suite**

Run: `bash tests/run.sh`
Expected: `ALL PASS` (test_lib, test_map, test_fetch, test_apply, test_bundle, test_preview, test_install).

- [ ] **Step 3: Verifikasi apply nyata di config sandbox** (tanpa sentuh config asli)

```bash
export HERDR_CONFIG_PATH="$(mktemp -d)/config.toml"
printf '[keys]\nprefix="cmd+b"\n\n[theme]\nname="solarized"\n' > "$HERDR_CONFIG_PATH"
bash bin/apply.sh gruvbox-dark
grep -A20 '^\[theme.custom\]' "$HERDR_CONFIG_PATH"
unset HERDR_CONFIG_PATH
```
Expected: blok `[theme.custom]` dgn 16 token hex. (reload akan gagal/di-skip di sandbox — normal.)

- [ ] **Step 4: Tulis `README.md`**

```markdown
# herdr-theme-picker

Pilih theme warna apa pun dari [terminalcolors.com](https://terminalcolors.com)
dan terapkan ke UI Herdr lewat `[theme.custom]`.

## Pasang

    bash bin/install.sh

Menautkan plugin ke Herdr dan menambah keybind `prefix+t` (prefix default `cmd+b`).

## Pakai

Tekan **prefix+t** → popup fzf. Ketik untuk cari, panah untuk lihat preview
warna, Enter untuk terapkan. Herdr reload otomatis.

- 15 theme populer tersedia offline; sisanya diambil dari terminalcolors.com
  saat dipilih (butuh internet).
- Config yang disentuh: `~/.config/herdr/config.toml` (blok `[theme.custom]`).
  Backup dibuat otomatis (`config.toml.bak-YYYYMMDD`).

## Copot

    bash bin/uninstall.sh

## Cakupan

Mengubah warna UI Herdr (panel, sidebar, aksen). TIDAK mengubah warna sel
terminal (ANSI) — itu diatur emulator terminalmu (mis. Ghostty).

## Uji

    bash tests/run.sh
```

- [ ] **Step 5: Commit**

```bash
git add tests/run.sh README.md
git commit -m "test: suite runner + README + e2e apply verification"
```

- [ ] **Step 6: Verifikasi instalasi nyata** (opsional, minta user konfirmasi dulu karena menyentuh config Herdr asli & keybind)

```bash
bash bin/install.sh
# lalu di Herdr: tekan prefix+t, pilih 'nord-default', pastikan UI berubah.
```





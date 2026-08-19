# herdr-theme-picker — Design

Tanggal: 2026-08-19
Status: disetujui untuk implementasi

## Masalah

Herdr hanya menyediakan sedikit theme bawaan (`catppuccin`, `tokyo-night`,
`gruvbox`, `solarized`, `terminal`, dll) via `[theme] name = "..."`. User ingin
memakai theme lain — ratusan palette dari terminalcolors.com — tanpa menyusun
`[theme.custom]` manual.

## Tujuan

Plugin herdr lokal `herdr-theme-picker`:
- Keybind `prefix+t` membuka popup picker theme.
- Katalog: ~15 theme populer di-bundle offline + fetch on-demand dari
  terminalcolors.com saat online.
- Pilih theme → tulis `[theme.custom]` ke `config.toml` herdr → `herdr server
  reload-config` → UI herdr langsung berubah.
- Scope: UI chrome herdr saja (bukan ANSI sel terminal). Reversible.

## Non-tujuan (YAGNI)

- Recolor config Ghostty / app lain.
- Editor warna manual per-token.
- Sinkronisasi auto light/dark.

## Temuan riset (terkonfirmasi)

- Config: `~/.config/herdr/config.toml`. `[theme] name=` + opsional
  `[theme.custom]` dengan token: `accent, panel_bg, surface0, surface1,
  surface_dim, overlay0, overlay1, text, subtext0, mauve, green, yellow, red,
  blue, teal, peach`. Nilai: hex / rgb() / named / reset.
- `herdr server reload-config` memuat ulang config di server berjalan.
- Keybind: `[[keys.command]]` dengan `key`, `type="plugin_action"`,
  `command="<plugin_id>.<action_id>"`. Prefix user = `cmd+b`; `prefix+t` valid.
- Plugin: `herdr-plugin.toml` manifest + `actions` + `panes` (placement popup).
  Contoh `ray.plugin-manager` & `termscope` sudah pakai pola popup + fzf.
- terminalcolors.com download stabil:
  `https://terminalcolors.com/downloads/ghostty/<name>-<variant>`
  → plaintext `background/foreground/selection/cursor-color/palette N=#hex`.
  Format ghostty paling ringkas untuk di-parse.

## Arsitektur

Semua bash + fzf + curl (dependency minim, konsisten dgn plugin terpasang).

```
herdr-theme-picker/
  herdr-plugin.toml         # manifest: action open, pane picker popup, keybind hint
  bin/
    picker.sh               # popup: fzf daftar theme (bundle + live), preview warna
    apply.sh                # fetch/baca palette → map → tulis [theme.custom] → reload
    map.sh                  # fungsi: ghostty-palette → token herdr (dipakai apply)
    install.sh              # link plugin + auto-patch [[keys.command]] prefix+t (backup)
    uninstall.sh            # unlink + hapus keybind (restore backup bila ada)
  themes/                   # ~15 palette bundle offline (format ghostty mentah)
    dracula-default
    gruvbox-dark
    nord-default
    tokyo-night-default
    catppuccin-mocha
    ...
  README.md
```

### Alur

1. `prefix+t` → `theme-picker.open` action → buka pane popup `picker`.
2. `picker.sh`:
   - Daftar = union nama file di `themes/` + (jika online) daftar theme live.
     Live-list: karena tak ada API JSON publik, sertakan file `themes/index.txt`
     berisi `name/variant` kurasi (~200 slug) untuk fetch on-demand; entri bundle
     ditandai `●`, sisanya `○`.
   - fzf preview: render swatch 16 warna via ANSI truecolor dari palette
     (bundle dibaca lokal; live entry di-fetch+cache ke `$cache/`).
   - Pilih → panggil `apply.sh <name-variant>`.
3. `apply.sh`:
   - Sumber palette: `themes/<slug>` bila ada, else fetch
     `downloads/ghostty/<slug>` → cache `~/.cache/herdr-theme-picker/`.
   - `map.sh` memetakan palette→16 token herdr (tabel di bawah).
   - Tulis blok `[theme.custom]` ke `config.toml` (ganti blok lama bila ada;
     backup `config.toml.bak-<ts>` sekali per hari). Set `[theme].name` tetap,
     custom menimpa.
   - `herdr server reload-config`. Tampilkan notifikasi.

### Mapping palette ghostty → token herdr

Token herdr bergaya catppuccin (semantik UI), palette terminal = 16 ANSI +
bg/fg. Pemetaan:

| token herdr | sumber palette |
|-------------|----------------|
| panel_bg    | `background` |
| surface0    | `palette 0` (black) |
| surface1    | `palette 8` (bright black) |
| surface_dim | `background` digelapkan ~8% (fallback: palette 0) |
| overlay0    | `palette 8` |
| overlay1    | `palette 7` (white) |
| text        | `foreground` |
| subtext0    | `palette 7` |
| accent      | `palette 4` (blue) — token warna aksen utama |
| mauve       | `palette 5` (magenta) |
| green       | `palette 2` |
| yellow      | `palette 3` |
| red         | `palette 1` |
| blue        | `palette 4` |
| teal        | `palette 6` (cyan) |
| peach       | `palette 9` (bright red) fallback `palette 3` |

`surface_dim` digelapkan via aritmetika hex sederhana di bash (kurangi tiap
kanal RGB ~8%, clamp ≥0). Sisanya salin langsung.

### Keybind (auto-patch)

`install.sh` menambah ke `config.toml` user (idempoten, backup dulu):

```toml
[[keys.command]]
key = "prefix+t"
type = "plugin_action"
command = "herdr-theme-picker.open"
description = "pick a color theme"
```

Cek dulu apakah blok dgn `command="herdr-theme-picker.open"` sudah ada → skip.

## Penanganan error

- Offline & theme bukan bundle → pesan "butuh internet untuk theme ini",
  tetap tampilkan bundle.
- Fetch gagal / HTTP≠200 / body tak berisi `palette 0=` → batal, config tak
  disentuh.
- `config.toml` tak ada → error jelas, tak membuat file baru.
- reload-config gagal → laporkan tapi custom sudah tertulis (idempoten, aman
  ulang).
- Validasi slug: hanya `[a-z0-9-]+` sebelum masuk URL (cegah injeksi path).

## Testing

- `bin/map.sh` self-check: assert-based bash test — beri palette dracula
  dikenal, verifikasi 16 token termapping benar (mis. `text=#f8f8f2`,
  `red=#ff5555`, `panel_bg=#282a36`, `surface_dim` lebih gelap dari bg).
- `darken_hex` unit: `darken #282a36 8` → nilai lebih kecil, tetap 6-digit.
- Validasi slug: input jahat `../../etc` ditolak.
- Manual: `apply.sh dracula-default` di sandbox `HERDR_CONFIG_PATH` sementara,
  cek blok `[theme.custom]` tertulis & reload sukses.

## Simplifikasi sengaja (ponytail)

- Live-list dari `index.txt` kurasi statis, bukan scraping/API — terminalcolors
  tak sediakan API. Upgrade bila situs merilis endpoint JSON.
- Parser palette hanya baca format ghostty (satu format, paling ringkas).

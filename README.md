# herdr-theme-picker

Pilih theme warna apa pun dari [terminalcolors.com](https://terminalcolors.com)
dan terapkan ke UI [Herdr](https://herdr.dev) — lewat satu keybind `prefix+t`.

Herdr hanya menyediakan segelintir theme bawaan (`catppuccin`, `tokyo-night`,
`gruvbox`, `solarized`, `terminal`, dll). Plugin ini membuka ratusan palette
dari terminalcolors.com: pilih dari popup, warnanya langsung dipetakan ke
`[theme.custom]` di `config.toml` dan Herdr reload otomatis.

> **Cakupan:** plugin mewarnai **UI chrome Herdr** (panel, sidebar, aksen).
> Plugin **tidak** mengubah warna sel terminal (ANSI 16-color) — itu diatur
> oleh emulator terminalmu (Ghostty, iTerm, dll), bukan Herdr.

---

## Fitur

- **`prefix+t`** → popup fzf dengan preview swatch warna truecolor.
- **19 theme populer** ter-bundle offline (Dracula, Gruvbox, Nord, Tokyo
  Night, Catppuccin, Solarized, One Dark, Everforest, Rosé Pine, Kanagawa,
  Ayu, GitHub, …).
- **Live-fetch** theme lain dari terminalcolors.com saat dipilih (butuh
  internet), di-cache lokal.
- **Backup otomatis** `config.toml` sebelum menulis.
- **Idempoten & reversible** — copot bersih lewat `uninstall.sh`.

---

## Kompatibilitas

| | |
|---|---|
| Herdr | ≥ 0.8.0 |
| OS | macOS, Linux |
| Dependency | `bash`, `curl`, [`fzf`](https://github.com/junegunn/fzf) |

`fzf` biasanya sudah ada bila kamu pakai plugin Herdr lain (file-picker,
termscope). Jika belum: `brew install fzf` (macOS) atau paket distro-mu.

---

## Install

### A. Via installer (disarankan)

```bash
git clone https://github.com/qintmb/herder-theme-picker.git
cd herder-theme-picker
bash bin/install.sh
```

Installer akan:
1. `herdr plugin link` folder ini ke Herdr.
2. Menambah keybind `prefix+t` ke `config.toml` (backup dulu, idempoten).
3. `herdr server reload-config`.

### B. Manual

```bash
git clone https://github.com/qintmb/herder-theme-picker.git ~/.config/herdr/plugins/local/herdr-theme-picker
herdr plugin link ~/.config/herdr/plugins/local/herdr-theme-picker
```

Lalu tambah keybind ke `~/.config/herdr/config.toml`:

```toml
[[keys.command]]
key = "prefix+t"
type = "plugin_action"
command = "herdr-theme-picker.open"
description = "pick a color theme"
```

Reload: `herdr server reload-config`.

---

## Pakai

Tekan **`prefix+t`** (prefix default Herdr = `cmd+b`, jadi `cmd+b` lalu `t`).

- Popup fzf terbuka. Ketik untuk memfilter, panah untuk navigasi.
- Panel kanan menampilkan preview warna theme.
- **Enter** → theme diterapkan, Herdr reload otomatis.
- **Esc** → batal.

Theme bertanda ada di `themes/index.txt`. Yang ter-bundle offline berlaku
tanpa internet; sisanya di-fetch saat dipilih.

Menjalankan tanpa keybind:

```bash
herdr plugin pane open --plugin herdr-theme-picker --entrypoint picker --placement popup --focus
```

---

## Cara kerja

```
prefix+t → picker.sh (fzf)
             │  pilih slug
             ▼
          apply.sh <slug>
             │  1. resolve_palette  → themes/<slug>  atau  fetch terminalcolors.com (cache)
             │  2. palette_to_tokens → 16 token [theme.custom]
             │  3. tulis config.toml (backup config.toml.bak-YYYYMMDD)
             │  4. herdr server reload-config
             ▼
          UI Herdr berubah
```

Palette dibaca dalam **format ghostty** (`background`, `foreground`,
`palette N=#hex`), lalu dipetakan ke token UI Herdr:

| token Herdr | sumber palette | | token Herdr | sumber palette |
|---|---|---|---|---|
| `panel_bg` | `background` | | `text` | `foreground` |
| `surface0` | `palette 0` | | `subtext0` | `palette 7` |
| `surface1` | `palette 8` | | `accent` / `blue` | `palette 4` |
| `surface_dim` | `background` −8% | | `red` | `palette 1` |
| `overlay0` | `palette 8` | | `green` | `palette 2` |
| `overlay1` | `palette 7` | | `yellow` | `palette 3` |
| `mauve` | `palette 5` | | `teal` | `palette 6` |
| `peach` | `palette 9` | | | |

---

## Modifikasi

### Menambah theme offline

Simpan file format-ghostty di `themes/<slug>` lalu daftarkan slug-nya ke
`themes/index.txt`:

```
background = #1e1e2e
foreground = #cdd6f4
selection-background = #585b70
cursor-color = #f5e0dc
palette = 0=#45475a
palette = 1=#f38ba8
...
palette = 15=#a6adc8
```

Cara cepat mengunduh dari terminalcolors.com:

```bash
curl -fsSL "https://terminalcolors.com/downloads/ghostty/<slug>" -o "themes/<slug>"
echo "<slug>" >> themes/index.txt
```

Slug hanya boleh `[a-z0-9-]` (divalidasi; mencegah path/URL injection).

### Mengubah pemetaan warna

Edit `bin/map.sh` → fungsi `palette_to_tokens`. Setiap baris `printf` memetakan
satu token Herdr ke sumber palette. `bin/lib.sh` menyediakan `darken_hex <#hex>
<persen>` untuk turunan warna.

### Mengubah keybind

Ganti `key = "prefix+t"` di `config.toml` ke kombinasi lain (mis. `prefix+shift+t`).

---

## Uji

```bash
bash tests/run.sh
```

Menjalankan self-check berbasis assert (tanpa framework): pemetaan palette,
`darken_hex`, validasi slug, penulisan `[theme.custom]` idempoten, patch
keybind, dan render swatch.

---

## Copot

```bash
bash bin/uninstall.sh
```

Menghapus keybind dari `config.toml`, `herdr plugin unlink`, dan reload.

---

## Lisensi

MIT.

## Tags

`herdr` · `herdr-plugin`

# herdr-theme-picker

Pilih theme warna apa pun dari [terminalcolors.com](https://terminalcolors.com)
dan terapkan ke UI Herdr lewat `[theme.custom]`.

## Pasang

    bash bin/install.sh

Menautkan plugin ke Herdr dan menambah keybind `prefix+t` (prefix default `cmd+b`).

## Pakai

Tekan **prefix+t** → popup fzf. Ketik untuk cari, panah untuk lihat preview
warna, Enter untuk terapkan. Herdr reload otomatis.

- 19 theme populer tersedia offline; sisanya diambil dari terminalcolors.com
  saat dipilih (butuh internet).
- Config yang disentuh: `~/.config/herdr/config.toml` (blok `[theme.custom]`).
  Backup dibuat otomatis (`config.toml.bak-YYYYMMDD`).

## Format theme (untuk tambah bundle offline)

File di `themes/<slug>` memakai format ghostty:

    background = #282a36
    foreground = #f8f8f2
    palette = 0=#21222c
    ...
    palette = 15=#ffffff

Tambah slug ke `themes/index.txt` agar muncul di picker.

## Copot

    bash bin/uninstall.sh

## Cakupan

Mengubah warna UI Herdr (panel, sidebar, aksen). TIDAK mengubah warna sel
terminal (ANSI) — itu diatur emulator terminalmu (mis. Ghostty).

## Uji

    bash tests/run.sh

# Windows test plan

Native checks for the `go-side-by-side` branch. The code cross-compiles and the
Windows-only tests compile on macOS, but nothing below has run on Windows yet.

Reference setup: a Windows laptop running `herdr --remote <mac>` in Windows
Terminal, with the Mac as the Herdr server.

## 0. Setup

```powershell
git clone https://github.com/wapenshaw/herdr-theme-picker.git
cd herdr-theme-picker
git checkout go-side-by-side
go version          # 1.24 or newer
fzf --version       # must be on PATH
go build -C go -trimpath -ldflags="-s -w" -o ../herdr-theme-picker.exe ./cmd/herdr-theme-picker
```

Run every command below from the checkout so the binary finds `themes\`.
Note your config path before starting: `HERDR_CONFIG_PATH` if set, else
`%XDG_CONFIG_HOME%\herdr\config.toml` if `XDG_CONFIG_HOME` is set, else
`%APPDATA%\herdr\config.toml`.

## 1. Automated tests

```powershell
go vet ./...
go test -C go -count=1 ./...
```

Expected: all `ok`. These run only on Windows:

| Test | Checks |
| --- | --- |
| `TestHelperWritesToAnIsolatedClientConsole` | The helper attaches to a hidden fake `herdr.exe` console and writes to `CONOUT$`. |
| `TestExactHelperDoesNotFallBackToAnotherHerdr` | A non-Herdr PID reports unavailable instead of attaching to a real client. |
| `TestCLICommandsNeverReceiveHostQueries` | `herdr server stop`, `status`, `--help`, and similar are never targeted. |
| `TestClientDiscoveryPreservesSessionArguments` | `--session "work projects"` survives Windows command-line parsing. |
| `TestInvalidExplicitPIDIsRejectedBeforeHelper` | An invalid explicit PID errors without spawning the helper. |
| `TestStandaloneStateWindowsCompatibility` | `%LOCALAPPDATA%` state paths and the `applied` marker location. |
| `TestWindowsEditorShim` | Editor launch through the Windows shim. |

The tests use a temporary home/config/state and a fake `herdr`. They must not
change your real config or recolor your terminal; check both afterwards.

## 2. Picker on the laptop (`--remote` client)

In a Windows Terminal tab **outside** Herdr, while another tab runs
`herdr --remote <mac>`:

```powershell
.\herdr-theme-picker.exe picker
```

| Check | Expected |
| --- | --- |
| fzf opens with the theme list | Search is at the top; `✓` marks the current theme. |
| Preview pane | Swatches render via pwsh or powershell with no quoting errors; it updates as you move. |
| Press Enter on a theme with clearly different colors | The popup closes. A warning says reload failed (no local server) and suggests `prefix+shift+r`. Exit code is 0. |
| Config | Your config now ends with `[theme.custom]`; the rest is untouched. On the first change of the day, `config.toml.bak-YYYYMMDD` exists. If the config was missing, it was created. |
| Herdr window before reload | Pane text, prompt colors, background, and foreground already show the new palette (OSC 4/10/11 through the console helper). |
| Press `prefix+shift+r` in the Herdr window | The Herdr UI (sidebar, tabs, borders) switches to the new theme. |
| Picker tab afterwards | No flashing window. Keys and the cursor still work; the console mode was restored. |
| Other Windows Terminal tabs without Herdr | Unchanged. |
| Mac | Its theme and terminal are unchanged. |

If pane colors do not change, run step 3 to separate discovery from the write.
Then report whether Windows Terminal ignored the sequences or the helper failed
(any stderr text).

## 3. `sync` and client discovery

```powershell
Get-Process herdr | Select-Object Id, ProcessName, StartTime
.\herdr-theme-picker.exe sync
.\herdr-theme-picker.exe sync --client <herdr --remote PID>
.\herdr-theme-picker.exe sync --client $PID      # this PowerShell
```

| Check | Expected |
| --- | --- |
| `sync` | Exit 0; the `herdr --remote` window gets the applied colors again. |
| `sync --client <client PID>` | Exit 0; only that window recolors. |
| `sync --client` with the PowerShell PID | `PID … is not a Herdr client`; exit 1. |
| `sync` with no Herdr window open | `No Herdr client terminal found on this machine.`; exit 1. |
| A new Windows Terminal window running `herdr --remote <mac>` | Starts with Windows Terminal's own colors. Running `sync` applies the theme. |
| Two `herdr --remote` windows, then apply a theme | Both recolor. |

## 4. Theme management on Windows

| Action | Expected |
| --- | --- |
| Copy a Ghostty palette, choose **+ Add from clipboard…**, enter a name | Read via `Get-Clipboard`; saved, applied, and marked `★`. |
| **Tab** (add with editor), with `$env:EDITOR` unset | Notepad opens a template; saving and closing adds the theme. |
| `$env:EDITOR = 'code --wait'`, then **Tab** | VS Code opens and the picker waits. |
| **Ctrl-E** on a `★` theme, change one hex, choose **Save & apply** | The override is saved, then applied. |
| **Ctrl-D** on a `★` theme, answer `y` | The theme is removed. The bundled themes are still there. |
| Apply an unknown slug: `.\herdr-theme-picker.exe apply nonexistent` | Fails with message containing terminalcolors.com link; exit 1. |

## 5. Config edge cases

Back up your config first, or point `$env:HERDR_CONFIG_PATH` at a copy.

| Setup | Expected after applying twice |
| --- | --- |
| Config with `auto_switch = true` and `[theme.custom.light]` / `[theme.custom.dark]` | Both layers survive unchanged. |
| CRLF line endings, as written by Notepad | Still CRLF; no mixed endings. |
| Applying the same theme twice | The second apply leaves the file byte-identical (`Get-FileHash`). |
| `$env:XDG_CONFIG_HOME = "$env:TEMP\xdg"` | The config is written under `$env:TEMP\xdg\herdr\config.toml`, matching Herdr. |

## 6. Optional: Herdr server on Windows

Only if you also run a local Herdr server on Windows (`herdr` without
`--remote`):

```powershell
herdr plugin link "$PWD"
herdr server reload-config
```

| Check | Expected |
| --- | --- |
| `prefix+t` in the local Windows Herdr | The fzf popup opens. Enter applies and reloads with no warning; the window recolors. |
| Theme picker: sync terminal colors action | Recolors the local Herdr window. |
| `Get-Process herdr-theme-picker` while idle | Nothing running; no background daemon. |
| Mac attached to this Windows server via `--remote`, no local Windows client | `prefix+t` from the Mac shows "Themes belong to your client machine". |

## What to report back

- The `go test` output if anything fails.
- For each failed row: the command, the full stderr, and a screenshot if colors
  look wrong.
- Windows Terminal version (`wt --version`, or Settings → About) and the
  PowerShell version.

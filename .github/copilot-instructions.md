# TimeSyncer — Copilot instructions

## Project

NTP time synchronizer with system tray for Windows / macOS / Linux, written in Go.

**Module:** `timesyncer`  
**Go:** 1.26.2, `go 1.25.0` in `go.mod`  
**Path:** `/Users/Aliaksandr_Shynkevich/DEV/Projects/timesyncer`

## Dependencies

| Module | Purpose |
|---|---|
| `github.com/getlantern/systray v1.2.2` | Tray icon; ICO on Windows, PNG on macOS/Linux |
| `github.com/beevik/ntp v1.5.0` | NTP client |
| `golang.org/x/sys v0.43.0` | Low-level OS APIs |
| `goversioninfo` (via `go run`) | VERSIONINFO + manifest + icon → `.syso` (Windows) |

## Build

```bash
go build ./...            # quick compilation check (cross-compile, no CGo)
./build.sh                # all platforms
./build.sh windows/amd64
./build.sh darwin/arm64
./build.sh linux          # requires Docker (CGo + GTK)
```

- Windows ldflags: `-H windowsgui` (no console window), `-s -w`
- Output: `dist/<os>-<arch>/`
- macOS: `dist/darwin-arm64/TimeSyncer.app` (`.app` bundle, raw binary removed)

## go:generate (before Windows build)

```bash
python3 tools/genicons/genicons.py
goversioninfo -o timesyncer_windows.syso res/versioninfo.json
```

`res/versioninfo.json` — VERSIONINFO metadata: publisher, copyright, version, icon path, manifest path.

## Icons

| File | Purpose |
|---|---|
| `assets/icon.png` | Tray (macOS/Linux), `go:embed` in `assets/icon_other.go` |
| `assets/icon.ico` | Tray (Windows), `go:embed` in `assets/icon_windows.go` |
| `res/timesyncer.ico` | Binary/installer icon (Windows) |
| `res/timesyncer.icns` | `.app` bundle icon (macOS) |
| `tools/genicons/genicons.py` | Generator (Python 3 + Pillow) |

## Privileges

| Platform | Method | Check file |
|---|---|---|
| Windows | UAC elevation (requireAdministrator in manifest) | `internal/ntp/check_windows.go` — `token.IsElevated()` |
| macOS | `sudoers NOPASSWD /bin/date` | `internal/ntp/check_darwin.go` — `sudo -n /bin/date --help` |
| Linux | root or CAP_SYS_TIME | `internal/ntp/check_linux.go` |

## Setting time (macOS)

```
sudo -n /bin/date -u <fmt>
```
BSD date format: `"010215042006.05"` → MMDDhhmmCCYY.SS  
Install script writes `/etc/sudoers.d/timesyncer`: `ALL ALL=(root) NOPASSWD: /bin/date`

## Autostart (Windows)

PowerShell `Register-ScheduledTask` with `RunLevel Highest`.

**Do NOT use:**
- `HKCU\Run` — cannot launch elevated applications
- `schtasks /Create /XML` — requires UTF-16 LE; we write UTF-8 → parse error

Code: `internal/autostart/autostart_windows.go`  
Constants: `internal/autostart/const_windows.go` (only `taskName`)

## Logging

`main.go:initLog()` — writes to `os.UserConfigDir()/timesyncer/timesyncer.log`

| Platform | Path |
|---|---|
| Windows | `%APPDATA%\timesyncer\timesyncer.log` |
| Linux | `~/.config/timesyncer/timesyncer.log` |
| macOS | `~/Library/Application Support/timesyncer/timesyncer.log` |

All `log.Printf(...)` in the app goes there automatically.

## Error display

`internal/errdialog/` — native dialog without a console window.  
Used only for privilege check failure at startup (`main.go`).

## Sentinel errors

`internal/ntp/const.go`: `ErrUserCancelled`  
`internal/tray/tray.go`: `doSync()` silently restores the previous label on `ErrUserCancelled`

## Rules

- **After every code change** run `go build ./...` and verify the build is clean.

## Important gotchas

- **`check_darwin.go`** — may disappear when created via editor tool. If missing — create via `cat > file << 'EOF'` in terminal.
- **`schtasks /Create /XML`** — does not work with UTF-8 XML, use PowerShell instead.
- **`New-ScheduledTaskSettingsSet`** — older PowerShell versions lack `-StopIfGoingOnBatteries` / `-DisallowStartIfOnBatteries`; use only `-ExecutionTimeLimit 0`.
- **systray on Windows** — accepts ICO, not PNG.
- **Linux builds from macOS** — always via Docker (CGo + GTK).
- **Linux launch** — never via `sudo` (GTK cannot open display). Use `setcap cap_sys_time+ep <binary>`, then run as a normal user.

## File structure

```
main.go                                   # entry point, initLog(), go:generate
assets/icon_windows.go                    # go:embed icon.ico
assets/icon_other.go                      # go:embed icon.png (!windows)
internal/ntp/sync.go                      # NTP query + apply
internal/ntp/time_darwin.go               # sudo date -u
internal/ntp/time_windows.go              # SetSystemTime via Win32
internal/ntp/time_linux.go                # clock_settime via syscall
internal/tray/tray.go                     # systray event loop, doSync()
internal/tray/labels.go                   # all UI strings
internal/autostart/autostart_windows.go   # PowerShell Register-ScheduledTask
internal/autostart/autostart_darwin.go    # LaunchAgent plist
internal/autostart/autostart_linux.go     # XDG .desktop
internal/config/config.go                 # JSON config in UserConfigDir
internal/errdialog/dialog_windows.go      # MessageBoxW
res/versioninfo.json                      # Windows binary metadata
res/timesyncer.manifest                   # UAC requireAdministrator + DPI
scripts/install_darwin.sh                 # copies binary + writes sudoers
scripts/run_linux.sh                      # setcap + launch without install
build.sh                                  # multi-platform build script
Dockerfile.linux                          # CGo+GTK for Linux cross-build
tools/genicons/genicons.py                # icon generator (Python 3 + Pillow)
```


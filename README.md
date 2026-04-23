# TimeSyncer

A lightweight system-tray NTP time synchroniser for Windows, macOS, and Linux — written in Go.

Sits in your tray, lets you pick an NTP server, and syncs on demand or automatically every hour.

## Download

**[→ Latest release](https://github.com/Alex-and-Soft/timesyncer/releases/latest)**

| Platform | File |
|---|---|
| Windows (x64) | `timesyncer-*-windows-amd64.zip` |
| Windows (ARM64) | `timesyncer-*-windows-arm64.zip` |
| macOS (Apple Silicon) | `timesyncer-*-darwin-arm64.zip` |
| macOS (Intel) | `timesyncer-*-darwin-amd64.zip` |
| Linux (x64) | `timesyncer-*-linux-amd64.tar.gz` |

## Features

- **System tray UI** — unobtrusive icon with a context menu
- **Server list** — choose from 9 public NTP pools (Google, Cloudflare, Windows, Apple, pool.ntp.org…)
- **Sync Now** — one-click synchronisation with live offset feedback
- **Auto-sync** — optional background sync every hour (toggle on/off)
- **Start with OS** — toggle autostart from the tray menu:
  - Windows → Task Scheduler logon task with highest privileges
  - macOS → `~/Library/LaunchAgents/com.timesyncer.plist`
  - Linux → XDG `~/.config/autostart/timesyncer.desktop`
- **Update notifications** — checks GitHub Releases on startup and shows a tray item if a new version is available

## Requirements & privileges

| Platform | Requirement |
|---|---|
| Windows | Windows 8 or later; run as Administrator |
| macOS | On first launch the app requests your Mac password once via a standard system dialog, then never again |
| Linux | `root` or `CAP_SYS_TIME` capability |

## Installation

### Windows

1. Download and unzip `timesyncer-*-windows-amd64.zip`
2. Right-click `timesyncer.exe` → **Run as administrator**
3. The tray icon appears — use **Start with OS** to enable autostart

### macOS

1. Download and unzip `timesyncer-*-darwin-arm64.zip` (or `amd64` for Intel)
2. Open `TimeSyncer.app`
3. On first launch a standard macOS password dialog appears — enter your Mac password once
4. Done — use **Start with OS** in the tray menu to enable autostart

### Linux

```bash
sudo cp timesyncer /usr/local/bin/
sudo setcap cap_sys_time+ep /usr/local/bin/timesyncer
timesyncer &   # run as normal user — no sudo needed after setcap
```

> **Do not run with `sudo`** — it strips `DISPLAY`/`XAUTHORITY` and GTK fails to open the display.

## Building

### Build script

[build.sh](build.sh) builds for all supported targets and places binaries in `dist/`.

```bash
./build.sh              # all platforms
./build.sh windows
./build.sh darwin
./build.sh linux        # requires Docker (CGo + GTK)
./build.sh windows/amd64
```

**Linux builds from macOS/Windows** require Docker (CGo + GTK). Make sure [Docker Desktop](https://www.docker.com/products/docker-desktop/) or [Rancher Desktop](https://rancherdesktop.io/) is running.

**Linux native build** — install GTK dev packages first:
```bash
sudo apt install gcc libgtk-3-dev libappindicator3-dev   # Debian/Ubuntu
sudo dnf install gcc gtk3-devel libappindicator-gtk3-devel  # Fedora
sudo pacman -S gcc gtk3 libappindicator-gtk3              # Arch
```

### go:generate (Windows resources, run once before Windows build)

```bash
python3 tools/genicons/genicons.py
go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest -64 -o timesyncer_windows.syso res/versioninfo.json
```

## Logs

| Platform | Path |
|---|---|
| Windows | `%APPDATA%\timesyncer\timesyncer.log` |
| macOS | `~/Library/Application Support/timesyncer/timesyncer.log` |
| Linux | `~/.config/timesyncer/timesyncer.log` |

## Configuration

Settings are stored as JSON alongside the log file (`config.json`). Created automatically on first run.

```json
{
  "servers": ["pool.ntp.org", "time.google.com", "time.cloudflare.com"],
  "selected_server": "pool.ntp.org",
  "auto_sync_enabled": true
}
```

## Dependencies

| Module | Purpose |
|---|---|
| [`github.com/getlantern/systray`](https://github.com/getlantern/systray) | Cross-platform system tray |
| [`github.com/beevik/ntp`](https://github.com/beevik/ntp) | NTP client + clock offset |
| [`golang.org/x/sys`](https://pkg.go.dev/golang.org/x/sys) | Low-level OS APIs |

## License

Copyright 2026 Aliaksandr Shynkevich  
Licensed under the [GNU General Public License v3.0](LICENSE).


# TimeSyncer

A lightweight system-tray NTP time synchroniser for Windows, Linux, and macOS — written in Go.

Why is your Windows clock always wrong? It drifts by minutes, and the built-in `w32tm` fails silently.  
TimeSyncer sits in your tray, lets you pick a server from a list, and syncs on demand or automatically every hour.

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
- **Server list** — choose from 9 public NTP pools (Google, Cloudflare, Windows, Apple, pool.ntp.org…); the active server is checkmarked
- **Sync Now** — one-click synchronisation with live offset feedback
- **Auto-sync** — optional background sync every hour (toggle on/off)
- **Start with OS** — toggle autostart without any installer:
  - Windows → Task Scheduler logon task with highest privileges
  - Linux → XDG `~/.config/autostart/timesyncer.desktop`
  - macOS → `~/Library/LaunchAgents/com.timesyncer.plist`
- **Single binary, zero dependencies** — just copy and run
- **Windows binary metadata** — publisher name, product version, and icon embedded in the `.exe`

## Requirements

| Platform | Privilege needed |
|---|---|
| Windows | **Windows 8 or later**, run as Administrator (required by `SetSystemTime`) |
| Linux | **root** or `CAP_SYS_TIME` |
| macOS | **root** (`sudo`) |

## Building

### Quick build (current platform)

```bash
go build -o timesyncer .
```

### Build script

[build.sh](build.sh) builds for all supported targets and places binaries in `dist/`.

```bash
# All platforms
./build.sh

# One platform family
./build.sh windows
./build.sh linux
./build.sh darwin

# Single target
./build.sh windows/amd64
./build.sh linux/arm64
./build.sh darwin/arm64
```

Output files follow the pattern `dist/timesyncer-<version>-<os>-<arch>[.exe]`.  
Version is taken from the latest git tag (`git describe`), or `dev` if no tag exists.

**Supported targets:**

| OS      | Architectures            | Cross-compile from macOS/Windows |
|---------|--------------------------|----------------------------------|
| Windows | `amd64`, `arm64`         | ✅ (no CGo needed)               |
| Linux   | `amd64`, `arm64`, `arm`  | ✅ via Docker (see below)        |
| macOS   | `amd64`, `arm64`         | ✅ (Xcode clang, same host OS)   |

> **Linux builds from macOS/Windows:** `systray` on Linux uses CGo + GTK libraries.  
> The build script automatically uses **Docker** for Linux targets when not running on Linux.  
> Make sure [Rancher Desktop](https://rancherdesktop.io/) or [Docker Desktop](https://www.docker.com/products/docker-desktop/) is running, then `./build.sh linux` works as usual.

> **Building natively on Linux:** no Docker needed. Just install the GTK dev packages first:
> ```bash
> # Ubuntu / Debian
> sudo apt install gcc libgtk-3-dev libappindicator3-dev
>
> # Fedora / RHEL
> sudo dnf install gcc gtk3-devel libappindicator-gtk3-devel
>
> # Arch Linux
> sudo pacman -S gcc gtk3 libappindicator-gtk3
> ```
> Then run `./build.sh linux` as usual — Docker is not involved.

### Embed the Windows manifest (once, on a Windows box)

The manifest enforces Administrator elevation and DPI awareness.  
Run `go generate` before building for Windows:

```bash
go generate          # installs akavel/rsrc and produces timesyncer_windows.syso
GOOS=windows GOARCH=amd64 go build -o timesyncer.exe .
```

The `.syso` file is picked up automatically by the Go toolchain.

## Windows SmartScreen warning

Windows SmartScreen shows a warning for any unsigned `.exe` from an unknown publisher. TimeSyncer embeds full version metadata (company name, copyright, product version) into the binary, so Windows Properties and the UAC prompt show "Aliaksandr Shynkevich" instead of "Unknown Publisher".

To **fully eliminate** the SmartScreen warning, the binary must be Authenticode-signed with a trusted certificate:

| Option | Cost | Notes |
|---|---|---|
| [Azure Trusted Signing](https://azure.microsoft.com/en-us/products/trusted-signing) | ~$10/month | Microsoft's own service; builds SmartScreen reputation quickly |
| [SignPath Foundation](https://about.signpath.io/product/open-source) | Free | For open-source projects; requires GitHub repo |
| OV Code Signing certificate | ~$60–300/year | From Certum, DigiCert, Sectigo, etc. |

Without signing, users can bypass the warning via **More info → Run anyway**.

## Installation

### Windows

1. Build or download `timesyncer.exe`
2. Right-click → **Run as administrator**
3. The tray icon appears — use **Start with OS** to register autostart

### Linux

```bash
sudo cp timesyncer /usr/local/bin/
sudo setcap cap_sys_time+ep /usr/local/bin/timesyncer
timesyncer &   # run as normal user — no sudo needed after setcap
```

> **Do not run with `sudo`** — it strips `DISPLAY`/`XAUTHORITY` and GTK fails to open the display.  
> `setcap` grants only the clock-setting capability; everything else runs as your user.
>
> If `setcap` is not available, install it first:
> ```bash
> sudo apt install libcap2-bin   # Debian / Ubuntu
> sudo dnf install libcap        # Fedora / RHEL
> sudo pacman -S libcap          # Arch
> ```

### macOS

```bash
sudo cp timesyncer /usr/local/bin/
sudo timesyncer &     # enable autostart via tray menu
```

## Logs

Application events and errors are written to a log file in the platform config directory:

| Platform | Path |
|---|---|
| Windows | `%APPDATA%\Roaming\timesyncer\timesyncer.log` |
| Linux | `~/.config/timesyncer/timesyncer.log` |
| macOS | `~/Library/Application Support/timesyncer/timesyncer.log` |

## Configuration

Settings are stored as JSON in the platform config directory:

| Platform | Path |
|---|---|
| Windows | `%APPDATA%\timesyncer\config.json` |
| Linux | `~/.config/timesyncer/config.json` |
| macOS | `~/Library/Application Support/timesyncer/config.json` |

The file is created automatically on first run. You can edit it manually:

```json
{
  "servers": [
    "pool.ntp.org",
    "time.google.com",
    "time.cloudflare.com"
  ],
  "selected_server": "pool.ntp.org",
  "auto_sync_enabled": true
}
```

## Project Structure

```
timesyncer/
├── main.go                          # entry point + go:generate directive
├── go.mod
├── res/
│   └── timesyncer.manifest          # Windows UAC + DPI manifest
├── assets/
│   └── icon.go                      # programmatically generated tray icon
└── internal/
    ├── config/
    │   └── config.go                # JSON config load/save
    ├── ntp/
    │   ├── const.go                 # shared error format strings
    │   ├── const_windows.go
    │   ├── sync.go                  # NTP query + sync logic
    │   ├── time_windows.go          # SetSystemTime via Win32 API
    │   ├── time_linux.go            # clock_settime via syscall
    │   └── time_darwin.go           # settimeofday via syscall
    ├── autostart/
    │   ├── const_windows.go         # registry key constants
    │   ├── const_linux.go           # .desktop template
    │   ├── const_darwin.go          # plist template
    │   ├── autostart_windows.go
    │   ├── autostart_linux.go
    │   └── autostart_darwin.go
    └── tray/
        ├── labels.go                # all UI strings and format constants
        └── tray.go                  # systray event loop
```

## Dependencies

| Module | Purpose |
|---|---|
| [`github.com/getlantern/systray`](https://github.com/getlantern/systray) | Cross-platform system tray |
| [`github.com/beevik/ntp`](https://github.com/beevik/ntp) | NTP client + clock offset |
| [`golang.org/x/sys`](https://pkg.go.dev/golang.org/x/sys) | Low-level OS APIs (Win32, unix) |

## License

Copyright 2026 Aliaksandr Shynkevich  
Licensed under the [GNU General Public License v3.0](LICENSE).

TimeSyncer is free software: you can redistribute it and/or modify it under the terms of the GNU GPL as published by the Free Software Foundation, version 3 or later. Forks and derivative works **must** remain open-source under the same license.

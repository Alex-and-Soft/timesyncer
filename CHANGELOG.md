# Changelog

## v1.0.1 — 2026-04-18

### Fixed

- **Windows autostart** — the task now starts correctly on laptops running on battery power. Previously, Windows could skip launching the app if the device was unplugged.
- **Windows tray icon not appearing after autostart** — added a wait for the system tray (Explorer shell) to be ready before initializing the tray icon. Previously, the icon could silently disappear if the app started too early during login.

---

## v1.0.0 — 2026-04-18

First public release.

### What's included

**Sync**
- One-click clock synchronisation via the tray menu — picks the best time from a public NTP server and sets your system clock in under a second
- Automatic background sync every hour (can be turned off)
- Live offset feedback — the menu shows exactly how far off your clock was (e.g. *Synced — offset −4.2 s*)
- Sync on startup — clock is corrected as soon as the app launches

**Server list**
- 9 built-in public NTP servers: pool.ntp.org, Google, Cloudflare, Apple, Windows Time, and regional pool variants
- Click any server in the menu to switch to it instantly; the active one is checkmarked

**Start with OS**
- Toggle autostart directly from the tray menu — no installer or admin panel needed
- Windows: registered as a Task Scheduler job so it launches with full privileges at login
- macOS: installed as a LaunchAgent that starts at login
- Linux: creates an XDG autostart entry in `~/.config/autostart/`

**Platform support**

| Platform | How time is set | Privilege required |
|---|---|---|
| Windows | Win32 `SetSystemTime` | Run as Administrator (UAC prompt on first launch) |
| macOS | `/bin/date` via sudoers | One-time `install.sh` sets up passwordless access |
| Linux | `clock_settime` syscall | `CAP_SYS_TIME` granted by `install.sh` or `run.sh` |

**Logging**
- All events and errors are written to a log file — useful if something goes wrong:
  - Windows: `%APPDATA%\timesyncer\timesyncer.log`
  - macOS: `~/Library/Application Support/timesyncer/timesyncer.log`
  - Linux: `~/.config/timesyncer/timesyncer.log`

**Packages**

| Platform | Download |
|---|---|
| Windows x64 | `timesyncer-v1.0.0-windows-amd64.zip` |
| Windows ARM64 | `timesyncer-v1.0.0-windows-arm64.zip` |
| macOS Apple Silicon | `timesyncer-v1.0.0-darwin-arm64.zip` |
| macOS Intel | `timesyncer-v1.0.0-darwin-amd64.zip` |
| Linux x64 | `timesyncer-v1.0.0-linux-amd64.tar.gz` |

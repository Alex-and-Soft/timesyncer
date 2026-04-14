#Requires -RunAsAdministrator
# Copyright (C) 2026 Aliaksandr Shynkevich
# Licensed under the GNU General Public License v3.0
#
# TimeSyncer Windows installer
# ----------------------------
# Copies the binary to Program Files, creates a Start Menu shortcut,
# and registers the app with the Windows startup registry key.
# Must be run as Administrator (required for SetSystemTime).

$ErrorActionPreference = "Stop"

$AppName    = "TimeSyncer"
$BinaryName = "timesyncer.exe"
$InstallDir = "$env:ProgramFiles\$AppName"
$BinaryDst  = "$InstallDir\$BinaryName"
$RunKey     = "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Run"

function Write-OK   { param($msg) Write-Host "[OK] $msg" -ForegroundColor Green }
function Write-Warn { param($msg) Write-Host "[!]  $msg" -ForegroundColor Yellow }
function Write-Fail { param($msg) Write-Host "[X]  $msg" -ForegroundColor Red; exit 1 }

Write-Host ""
Write-Host "TimeSyncer installer for Windows" -ForegroundColor Cyan
Write-Host ""

# ── Locate binary ──────────────────────────────────────────────────────────────
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$BinarySrc = Join-Path $ScriptDir $BinaryName
if (-not (Test-Path $BinarySrc)) { Write-Fail "Binary not found: $BinarySrc" }

# ── Install binary ─────────────────────────────────────────────────────────────
Write-OK "Installing $BinaryName → $BinaryDst"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Copy-Item -Force $BinarySrc $BinaryDst

# ── Add to system PATH (optional, for calling timesyncer from terminal) ────────
$currentPath = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
if ($currentPath -notlike "*$InstallDir*") {
    [System.Environment]::SetEnvironmentVariable(
        "Path", "$currentPath;$InstallDir", "Machine")
    Write-OK "Added $InstallDir to system PATH"
}

# ── Autostart via Run registry key ────────────────────────────────────────────
$choice = Read-Host "Register TimeSyncer as autostart for all users? [Y/n]"
if ($choice -eq "" -or $choice -match "^[Yy]$") {
    Set-ItemProperty -Path $RunKey -Name $AppName -Value "`"$BinaryDst`""
    Write-OK "Registered in HKLM Run key (starts for all users on login)"
}

# ── Start Menu shortcut ────────────────────────────────────────────────────────
$ShortcutPath = "$env:ProgramData\Microsoft\Windows\Start Menu\Programs\$AppName.lnk"
$Shell = New-Object -ComObject WScript.Shell
$Shortcut = $Shell.CreateShortcut($ShortcutPath)
$Shortcut.TargetPath  = $BinaryDst
$Shortcut.Description = "TimeSyncer NTP clock synchroniser"
$Shortcut.Save()
Write-OK "Start Menu shortcut created"

Write-Host ""
Write-Host "  Binary:    $BinaryDst"
Write-Host "  Shortcut:  $ShortcutPath"
Write-Host ""
Write-OK "Installation complete."
Write-Host ""
Write-Host "NOTE: TimeSyncer requires Administrator privileges to set the system clock."
Write-Host "      It will request elevation via UAC on each launch when needed."
Write-Host ""
Write-Host "To uninstall:"
Write-Host "  Remove-ItemProperty -Path '$RunKey' -Name '$AppName'"
Write-Host "  Remove-Item -Recurse '$InstallDir'"
Write-Host "  Remove-Item '$ShortcutPath'"

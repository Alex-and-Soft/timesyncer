#!/usr/bin/env bash
# Copyright (C) 2026 Aliaksandr Shynkevich
# Licensed under the GNU General Public License v3.0
#
# TimeSyncer Linux installer
# --------------------------
# Copies the binary to /usr/local/bin and grants it the CAP_SYS_TIME
# capability so it can set the system clock without running as root.
# Optionally installs a systemd user service for autostart.

set -euo pipefail

BINARY_NAME="timesyncer"
INSTALL_PATH="/usr/local/bin/${BINARY_NAME}"
SYSTEMD_USER_DIR="${HOME}/.config/systemd/user"
SERVICE_NAME="timesyncer.service"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
info()  { echo -e "${GREEN}[✓]${NC} $*"; }
warn()  { echo -e "${YELLOW}[!]${NC} $*"; }
error() { echo -e "${RED}[✗]${NC} $*" >&2; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="${SCRIPT_DIR}/${BINARY_NAME}"
[[ -f "${BINARY}" ]] || error "Binary not found: ${BINARY}"

echo "TimeSyncer installer for Linux"
echo ""

# ── Install binary ────────────────────────────────────────────────────────────
info "Installing ${BINARY_NAME} → ${INSTALL_PATH}"
sudo cp "${BINARY}" "${INSTALL_PATH}"
sudo chmod 755 "${INSTALL_PATH}"

# ── Grant CAP_SYS_TIME capability ─────────────────────────────────────────────
# This allows the binary to change the system clock without being root.
if command -v setcap &>/dev/null; then
  info "Granting CAP_SYS_TIME capability (no root needed at runtime)"
  sudo setcap cap_sys_time+ep "${INSTALL_PATH}"
else
  warn "setcap not found (install libcap2-bin on Debian/Ubuntu or libcap on Fedora/Arch)"
  warn "Falling back to setuid root — less preferred but functional"
  sudo chown root:root "${INSTALL_PATH}"
  sudo chmod u+s "${INSTALL_PATH}"
fi

# ── Systemd user service (optional) ───────────────────────────────────────────
if command -v systemctl &>/dev/null; then
  echo ""
  read -r -p "Install systemd user service for autostart? [Y/n] " answer
  answer="${answer:-Y}"
  if [[ "${answer}" =~ ^[Yy]$ ]]; then
    mkdir -p "${SYSTEMD_USER_DIR}"
    cat > "${SYSTEMD_USER_DIR}/${SERVICE_NAME}" <<EOF
[Unit]
Description=TimeSyncer NTP clock synchroniser
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=${INSTALL_PATH}
Restart=on-failure
RestartSec=10

[Install]
WantedBy=default.target
EOF
    systemctl --user daemon-reload
    systemctl --user enable "${SERVICE_NAME}"
    systemctl --user start  "${SERVICE_NAME}"
    info "Systemd user service enabled and started"
  fi
else
  warn "systemd not found — start TimeSyncer manually or add it to your desktop autostart"
fi

echo ""
echo "  Binary:   ${INSTALL_PATH}"
echo ""
info "Installation complete."
echo ""
echo -e "${YELLOW}IMPORTANT:${NC} Always run TimeSyncer as a normal user (not root/sudo):"
echo "  timesyncer &"
echo ""
echo "To uninstall:"
echo "  systemctl --user disable --now ${SERVICE_NAME}"
echo "  rm \${HOME}/.config/systemd/user/${SERVICE_NAME} ${INSTALL_PATH}"

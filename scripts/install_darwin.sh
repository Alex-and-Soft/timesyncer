#!/usr/bin/env bash
# Copyright (C) 2026 Aliaksandr Shynkevich
# Licensed under the GNU General Public License v3.0
#
# TimeSyncer macOS installer
# --------------------------
# Copies the binary to /usr/local/bin, grants passwordless sudo for /bin/date,
# and registers a LaunchAgent so TimeSyncer starts automatically on login.
# Password is requested only once during this installation.

set -euo pipefail

BINARY_NAME="timesyncer"
INSTALL_DIR="/usr/local/bin"
LAUNCH_AGENT_DIR="$HOME/Library/LaunchAgents"
PLIST_ID="com.timesyncer"
PLIST_PATH="${LAUNCH_AGENT_DIR}/${PLIST_ID}.plist"
INSTALL_PATH="${INSTALL_DIR}/${BINARY_NAME}"
SUDOERS_FILE="/etc/sudoers.d/timesyncer"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
info()    { echo -e "${GREEN}[✓]${NC} $*"; }
warn()    { echo -e "${YELLOW}[!]${NC} $*"; }
error()   { echo -e "${RED}[✗]${NC} $*" >&2; exit 1; }

# ── Locate binary ─────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="${SCRIPT_DIR}/${BINARY_NAME}"
[[ -f "${BINARY}" ]] || error "Binary not found: ${BINARY}"

# ── Install binary ────────────────────────────────────────────────────────────
echo "TimeSyncer installer for macOS"
echo ""

if [[ ! -d "${INSTALL_DIR}" ]]; then
  warn "${INSTALL_DIR} does not exist — creating (requires sudo)"
  sudo mkdir -p "${INSTALL_DIR}"
fi

info "Installing ${BINARY_NAME} → ${INSTALL_PATH}"
sudo cp "${BINARY}" "${INSTALL_PATH}"
sudo chmod 755 "${INSTALL_PATH}"

# ── sudoers: passwordless /bin/date ──────────────────────────────────────────
info "Granting passwordless sudo for /bin/date → ${SUDOERS_FILE}"
echo "ALL ALL=(root) NOPASSWD: /bin/date" | sudo tee "${SUDOERS_FILE}" > /dev/null
sudo chmod 440 "${SUDOERS_FILE}"
# Validate the sudoers file to avoid locking out sudo
if ! sudo visudo -cf "${SUDOERS_FILE}" &>/dev/null; then
  sudo rm -f "${SUDOERS_FILE}"
  error "sudoers validation failed — file removed, installation aborted"
fi

# ── LaunchAgent plist ─────────────────────────────────────────────────────────
mkdir -p "${LAUNCH_AGENT_DIR}"

info "Creating LaunchAgent: ${PLIST_PATH}"
cat > "${PLIST_PATH}" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>${PLIST_ID}</string>
    <key>ProgramArguments</key>
    <array>
        <string>${INSTALL_PATH}</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
    <key>StandardErrorPath</key>
    <string>${HOME}/Library/Logs/timesyncer.log</string>
</dict>
</plist>
EOF

# ── Load the agent ────────────────────────────────────────────────────────────
if launchctl list | grep -q "${PLIST_ID}" 2>/dev/null; then
  info "Reloading existing LaunchAgent…"
  launchctl unload "${PLIST_PATH}" 2>/dev/null || true
fi

launchctl load "${PLIST_PATH}"
info "LaunchAgent loaded — TimeSyncer is now running"

echo ""
echo "  Binary:       ${INSTALL_PATH}"
echo "  Sudoers:      ${SUDOERS_FILE}  (NOPASSWD for /bin/date)"
echo "  LaunchAgent:  ${PLIST_PATH}"
echo "  Logs:         ~/Library/Logs/timesyncer.log"
echo ""
echo "  Time sync will run silently — no password prompts."
echo ""
info "Installation complete."

# ── Uninstall hint ────────────────────────────────────────────────────────────
echo ""
echo "To uninstall:"
echo "  launchctl unload ${PLIST_PATH}"
echo "  sudo rm ${PLIST_PATH} ${INSTALL_PATH} ${SUDOERS_FILE}"

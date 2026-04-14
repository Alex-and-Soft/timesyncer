#!/usr/bin/env bash
# Copyright (C) 2026 Aliaksandr Shynkevich
# Licensed under the GNU General Public License v3.0
#
# Quick-start script for TimeSyncer on Linux.
# Grants CAP_SYS_TIME to the binary (once, requires sudo),
# then launches it as the current user — no full installation needed.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="${SCRIPT_DIR}/timesyncer"

[[ -f "${BINARY}" ]] || { echo "Error: binary not found at ${BINARY}"; exit 1; }

# Grant capability if not already set
if ! getcap "${BINARY}" 2>/dev/null | grep -q "cap_sys_time"; then
  echo "Granting CAP_SYS_TIME capability (requires sudo, one-time)..."
  sudo setcap cap_sys_time+ep "${BINARY}"
fi

echo "Starting TimeSyncer..."
exec "${BINARY}"

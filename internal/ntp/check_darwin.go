// Copyright (C) 2026 Aliaksandr Shynkevich
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

//go:build darwin

package ntp

import (
	"errors"
	"os/exec"
)

const sudoersFile = "/etc/sudoers.d/timesyncer"
const sudoersRule = `ALL ALL=(root) NOPASSWD: /bin/date`

// setupScript is run via osascript with administrator privileges.
// It writes the sudoers rule and sets the required permissions.
// Full paths are used because do shell script has a minimal PATH.
const setupScript = `/bin/echo '` + sudoersRule + `' > ` + sudoersFile +
	` && /bin/chmod 440 ` + sudoersFile

// CheckPrivileges ensures passwordless sudo for /bin/date is configured.
// On first launch it requests administrator access via the standard macOS
// password dialog (osascript) and writes /etc/sudoers.d/timesyncer automatically.
// Subsequent launches skip the dialog entirely.
func CheckPrivileges() error {
	if canSudoDate() {
		return nil
	}

	// Ask the user for admin credentials via the native macOS dialog.
	err := exec.Command("osascript", "-e",
		`do shell script "`+setupScript+`" with administrator privileges`,
	).Run()
	if err != nil {
		// User cancelled the dialog or entered wrong password.
		return errors.New("TimeSyncer needs administrator access to set the system clock.\n\nPlease try again and enter your Mac password when prompted.")
	}

	if !canSudoDate() {
		return errors.New("Failed to configure the required permission.\nPlease contact support or run: sudo bash install_darwin.sh")
	}
	return nil
}

// canSudoDate reports whether passwordless sudo for /bin/date works.
// BSD date on macOS does not support --help, so we run it with no arguments
// (it just prints the current time and exits 0).
func canSudoDate() bool {
	return exec.Command("sudo", "-n", "/bin/date").Run() == nil
}

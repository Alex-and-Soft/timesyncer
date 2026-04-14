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
	"fmt"
	"os/exec"
)

// CheckPrivileges verifies that passwordless sudo for /bin/date is configured.
// The install script writes /etc/sudoers.d/timesyncer to grant this.
// We probe with `sudo -n /bin/date --help` (no-password, non-destructive).
func CheckPrivileges() error {
	err := exec.Command("sudo", "-n", "/bin/date", "--help").Run()
	if err == nil {
		return nil
	}
	return fmt.Errorf(
		"TimeSyncer is not properly installed.\n\n" +
			"Please run the install script to grant the required\n" +
			"permission for setting the system clock:\n\n" +
			"  sudo bash install.sh",
	)
}

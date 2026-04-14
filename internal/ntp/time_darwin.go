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
	"strings"
	"time"
)

// setSystemTime sets the system clock via osascript.
// macOS shows a native admin-password dialog on first call per session;
// subsequent calls within ~5 minutes are silent (macOS auth cache).
// setSystemTime sets the system clock via sudo date.
// The install script registers /etc/sudoers.d/timesyncer granting NOPASSWD
// for /bin/date, so no password dialog appears after installation.
func setSystemTime(t time.Time) error {
	// BSD date(1) set format: [[[[mm]dd]HH]MM[[cc]yy][.SS]]
	// Go reference time: Mon Jan 2 15:04:05 MST 2006
	//   mm=01  dd=02  HH=15  MM=04  ccyy=2006  .SS=.05
	dateStr := t.UTC().Format("010215042006.05")
	out, err := exec.Command("sudo", "-n", "/bin/date", "-u", dateStr).CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf(errDarwinSetTime, msg)
	}
	return nil
}

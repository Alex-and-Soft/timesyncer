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

//go:build linux

package ntp

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CheckPrivileges returns an error if the process cannot set the system clock.
// On Linux this requires either root (uid 0) or the CAP_SYS_TIME capability,
// which the install script grants via: sudo setcap cap_sys_time+ep <binary>
func CheckPrivileges() error {
	if os.Getuid() == 0 {
		return nil
	}
	if hasSysTimeCap() {
		return nil
	}
	return fmt.Errorf(
		"TimeSyncer needs permission to set the system clock.\n\n"+
			"Run the install script, or grant the capability manually:\n"+
			"  sudo setcap cap_sys_time+ep %s",
		os.Args[0],
	)
}

// hasSysTimeCap checks bit 25 (CAP_SYS_TIME) in the process effective
// capability set, read from /proc/self/status.
func hasSysTimeCap() bool {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "CapEff:") {
			continue
		}
		hex := strings.TrimSpace(strings.TrimPrefix(line, "CapEff:"))
		capEff, err := strconv.ParseUint(hex, 16, 64)
		if err != nil {
			return false
		}
		const capSysTime = uint64(1) << 25
		return capEff&capSysTime != 0
	}
	return false
}

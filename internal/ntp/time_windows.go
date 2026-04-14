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

//go:build windows

package ntp

import (
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

func setSystemTime(t time.Time) error {
	u := t.UTC()
	st := windows.Systemtime{
		Year:         uint16(u.Year()),
		Month:        uint16(u.Month()),
		DayOfWeek:    uint16(u.Weekday()), // ignored by SetSystemTime, but set for correctness
		Day:          uint16(u.Day()),
		Hour:         uint16(u.Hour()),
		Minute:       uint16(u.Minute()),
		Second:       uint16(u.Second()),
		Milliseconds: uint16(u.Nanosecond() / 1_000_000),
	}
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	proc := kernel32.NewProc("SetSystemTime")
	r, _, err := proc.Call(uintptr(unsafe.Pointer(&st)))
	if r == 0 {
		return fmt.Errorf(errWinSetTime, err)
	}
	return nil
}

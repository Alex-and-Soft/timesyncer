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

package main

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	user32      = syscall.NewLazyDLL("user32.dll")
	findWindowW = user32.NewProc("FindWindowW")
)

// waitForTray waits for Explorer's system tray to become available.
func waitForTray() {
	cls, _ := syscall.UTF16PtrFromString("Shell_TrayWnd")
	for i := 0; i < 60; i++ {
		hwnd, _, _ := findWindowW.Call(uintptr(unsafe.Pointer(cls)), 0)
		if hwnd != 0 {
			// Give Explorer a moment to finish notification area setup.
			time.Sleep(2 * time.Second)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

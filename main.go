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

package main

//go:generate python3 tools/genicons/genicons.py
//go:generate go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest -64 -o timesyncer_windows.syso res/versioninfo.json

import (
	"log"
	"os"
	"path/filepath"

	"timesyncer/internal/errdialog"
	"timesyncer/internal/ntp"
	"timesyncer/internal/tray"

	"github.com/getlantern/systray"
)

func initLog() {
	dir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	logDir := filepath.Join(dir, "timesyncer")
	_ = os.MkdirAll(logDir, 0700)
	f, err := os.OpenFile(filepath.Join(logDir, "timesyncer.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	initLog()
	log.Println("starting")
	waitForTray()
	if err := ntp.CheckPrivileges(); err != nil {
		errdialog.Show("TimeSyncer — insufficient privileges", err.Error())
		os.Exit(1)
	}
	systray.Run(tray.OnReady, tray.OnExit)
}

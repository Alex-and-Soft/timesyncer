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

//go:build !windows

// Package errdialog shows a blocking native error dialog suitable for GUI
// applications that may not have a console window attached.
package errdialog

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Show prints the error to stderr. On macOS it also shows a native dialog via
// osascript; on Linux it tries zenity then xmessage as fallback.
func Show(title, msg string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", title, msg)

	switch runtime.GOOS {
	case "darwin":
		// AppleScript strings don't support \n — split into lines and join
		// with the AppleScript line-break literal: " & return & "
		lines := strings.Split(msg, "\n")
		var parts []string
		for _, l := range lines {
			parts = append(parts, `"`+strings.ReplaceAll(l, `"`, `\"`)+`"`)
		}
		apMsg := strings.Join(parts, " & return & ")
		script := fmt.Sprintf(
			`display dialog %s with title %q buttons {"OK"} default button "OK" with icon stop`,
			apMsg, title,
		)
		_ = exec.Command("osascript", "-e", script).Run()

	case "linux":
		// try GUI dialogs in order; silently skip if not installed
		for _, args := range [][]string{
			{"zenity", "--error", "--title=" + title, "--text=" + msg, "--no-wrap"},
			{"xmessage", "-center", title + "\n\n" + msg},
		} {
			if _, err := exec.LookPath(args[0]); err == nil {
				_ = exec.Command(args[0], args[1:]...).Run()
				return
			}
		}
	}
}

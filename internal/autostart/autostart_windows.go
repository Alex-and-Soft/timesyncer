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

package autostart

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

// Enable registers a Task Scheduler logon task via PowerShell.
func Enable() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	u, err := user.Current()
	if err != nil {
		return err
	}
	// Escape single quotes for PowerShell string literals.
	exeEsc := strings.ReplaceAll(exe, `'`, `''`)
	userEsc := strings.ReplaceAll(u.Username, `'`, `''`)

	script := fmt.Sprintf(
		`$a=New-ScheduledTaskAction -Execute '%s'; `+
			`$t=New-ScheduledTaskTrigger -AtLogOn -User '%s'; `+
			`$p=New-ScheduledTaskPrincipal -UserId '%s' -LogonType Interactive -RunLevel Highest; `+
			`$s=New-ScheduledTaskSettingsSet -ExecutionTimeLimit 0; `+
			`$s.DisallowStartIfOnBatteries=$false; `+
			`$s.StopIfGoingOnBatteries=$false; `+
			`Register-ScheduledTask -TaskName '%s' -Action $a -Trigger $t -Principal $p -Settings $s -Force | Out-Null`,
		exeEsc, userEsc, userEsc, taskName,
	)

	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = hiddenWindow()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powershell: %w\n%s", err, out)
	}
	return nil
}

// Disable removes the Task Scheduler logon task.
func Disable() error {
	cmd := exec.Command("schtasks", "/Delete", "/F", "/TN", taskName)
	cmd.SysProcAttr = hiddenWindow()
	err := cmd.Run()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 1 {
			return nil // task not found — already disabled
		}
	}
	return err
}

// IsEnabled reports whether the logon task exists.
func IsEnabled() bool {
	cmd := exec.Command("schtasks", "/Query", "/TN", taskName)
	cmd.SysProcAttr = hiddenWindow()
	return cmd.Run() == nil
}

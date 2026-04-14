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

package tray

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/getlantern/systray"
	"timesyncer/assets"
	"timesyncer/internal/autostart"
	"timesyncer/internal/config"
	ntpsync "timesyncer/internal/ntp"
)

var (
	cfg           *config.Config
	serverItems   []*systray.MenuItem
	syncNowItem   *systray.MenuItem
	lastSyncItem  *systray.MenuItem
	autoSyncItem  *systray.MenuItem
	autostartItem *systray.MenuItem
	lastSyncLabel string // current text of lastSyncItem, restored on user cancel
)

// OnReady is called by systray once the tray infrastructure is ready.
func OnReady() {
	cfg = config.Load()

	systray.SetIcon(assets.Icon())
	systray.SetTooltip(lblTooltip)

	// ── Server list ──────────────────────────────────────────────────────────
	serverItems = make([]*systray.MenuItem, len(cfg.Servers))
	for i, srv := range cfg.Servers {
		item := systray.AddMenuItem(srv, fmt.Sprintf(lblServerTipFmt, srv))
		if srv == cfg.SelectedServer {
			item.Check()
		}
		serverItems[i] = item
		go func(idx int, server string, mi *systray.MenuItem) {
			for range mi.ClickedCh {
				selectServer(idx, server)
			}
		}(i, srv, item)
	}

	systray.AddSeparator()

	// ── Actions ───────────────────────────────────────────────────────────────
	syncNowItem = systray.AddMenuItem(lblSyncNow, lblSyncNowTip)
	lastSyncLabel = lblLastSyncNever
	lastSyncItem = systray.AddMenuItem(lastSyncLabel, "")
	lastSyncItem.Disable()

	systray.AddSeparator()

	// ── Settings ──────────────────────────────────────────────────────────────
	autoSyncItem = systray.AddMenuItem(autoSyncLabel(), lblAutoSync)
	systray.AddSeparator()
	autostartItem = systray.AddMenuItem(autostartLabel(), lblAutostart)

	systray.AddSeparator()
	quitItem := systray.AddMenuItem(lblQuit, lblQuitTip)

	// ── Event loops ───────────────────────────────────────────────────────────
	go func() {
		for range syncNowItem.ClickedCh {
			go doSync()
		}
	}()
	go func() {
		for range autoSyncItem.ClickedCh {
			toggleAutoSync()
		}
	}()
	go func() {
		for range autostartItem.ClickedCh {
			toggleAutostart()
		}
	}()
	go func() {
		<-quitItem.ClickedCh
		systray.Quit()
	}()

	// Background hourly auto-sync
	go autoSyncLoop()

	// Sync immediately on startup
	go doSync()
}

// OnExit is called by systray just before the process exits.
func OnExit() {
	if err := cfg.Save(); err != nil {
		log.Printf("timesyncer: save config: %v", err)
	}
}

// selectServer marks the chosen server as active.
func selectServer(idx int, server string) {
	cfg.SelectedServer = server
	for i, item := range serverItems {
		if i == idx {
			item.Check()
		} else {
			item.Uncheck()
		}
	}
	_ = cfg.Save()
}

// doSync queries the selected NTP server and applies the time.
func doSync() {
	syncNowItem.Disable()
	prev := lastSyncLabel
	lastSyncItem.SetTitle(lblSyncing)

	result, err := ntpsync.Sync(cfg.SelectedServer)
	if err != nil {
		if errors.Is(err, ntpsync.ErrUserCancelled) {
			// User dismissed the privilege prompt — restore previous label silently.
			lastSyncItem.SetTitle(prev)
			syncNowItem.Enable()
			return
		}
		log.Printf("timesyncer: sync: %v", err)
		lastSyncLabel = fmt.Sprintf(lblSyncFailed, err)
		lastSyncItem.SetTitle(lastSyncLabel)
		syncNowItem.Enable()
		return
	}

	offset := result.Offset.Round(time.Millisecond)
	lastSyncLabel = fmt.Sprintf(lblLastSyncFmt,
		result.Time.Local().Format(lblTimeFmt),
		offset,
	)
	lastSyncItem.SetTitle(lastSyncLabel)
	syncNowItem.Enable()
}

// autoSyncLoop fires doSync every hour while auto-sync is enabled.
func autoSyncLoop() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		if cfg.AutoSyncEnabled {
			go doSync()
		}
	}
}

func toggleAutoSync() {
	cfg.AutoSyncEnabled = !cfg.AutoSyncEnabled
	autoSyncItem.SetTitle(autoSyncLabel())
	_ = cfg.Save()
}

func toggleAutostart() {
	if autostart.IsEnabled() {
		if err := autostart.Disable(); err != nil {
			log.Printf("timesyncer: disable autostart: %v", err)
			return
		}
	} else {
		if err := autostart.Enable(); err != nil {
			log.Printf("timesyncer: enable autostart: %v", err)
			return
		}
	}
	autostartItem.SetTitle(autostartLabel())
}

func autoSyncLabel() string {
	if cfg.AutoSyncEnabled {
		return lblAutoSyncOn
	}
	return lblAutoSyncOff
}

func autostartLabel() string {
	if autostart.IsEnabled() {
		return lblAutostartOn
	}
	return lblAutostartOff
}

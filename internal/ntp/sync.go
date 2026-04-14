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

package ntp

import (
	"fmt"
	"time"

	beevikntp "github.com/beevik/ntp"
)

// SyncResult holds the outcome of an NTP operation.
type SyncResult struct {
	Server string
	Offset time.Duration
	Time   time.Time
}

// Query fetches the current time offset from the given NTP server.
func Query(server string) (*SyncResult, error) {
	resp, err := beevikntp.Query(server)
	if err != nil {
		return nil, fmt.Errorf(errQueryFmt, server, err)
	}
	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf(errInvalidFmt, server, err)
	}
	return &SyncResult{
		Server: server,
		Offset: resp.ClockOffset,
		Time:   time.Now().Add(resp.ClockOffset),
	}, nil
}

// Sync fetches NTP time and applies it to the system clock.
// Requires elevated privileges (Administrator on Windows, root on Unix).
func Sync(server string) (*SyncResult, error) {
	result, err := Query(server)
	if err != nil {
		return nil, err
	}
	if err := setSystemTime(result.Time); err != nil {
		return nil, fmt.Errorf(errSetTimeFmt, err)
	}
	return result, nil
}

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

import "errors"

// ErrUserCancelled is returned when the user dismisses the privilege prompt.
// Callers should handle it silently (no error indicator in the UI).
var ErrUserCancelled = errors.New("user cancelled")

const (
	errQueryFmt   = "NTP query to %q: %w"
	errInvalidFmt = "NTP response from %q invalid: %w"
	errSetTimeFmt = "set system time: %w"
)

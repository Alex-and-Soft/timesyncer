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

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds all persistent application settings.
type Config struct {
	Servers         []string `json:"servers"`
	SelectedServer  string   `json:"selected_server"`
	AutoSyncEnabled bool     `json:"auto_sync_enabled"`
}

// DefaultServers is the built-in list of public NTP servers.
var DefaultServers = []string{
	"pool.ntp.org",
	"time.google.com",
	"time.cloudflare.com",
	"time.windows.com",
	"time.apple.com",
	"0.pool.ntp.org",
	"1.pool.ntp.org",
	"2.pool.ntp.org",
	"3.pool.ntp.org",
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "timesyncer", "config.json"), nil
}

// Load reads the configuration from disk.
// Returns safe defaults if the file is missing or unreadable.
func Load() *Config {
	path, err := configPath()
	if err != nil {
		return defaults()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return defaults()
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaults()
	}
	if len(cfg.Servers) == 0 {
		cfg.Servers = DefaultServers
	}
	if cfg.SelectedServer == "" {
		cfg.SelectedServer = cfg.Servers[0]
	}
	return &cfg
}

// Save writes the configuration to disk.
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func defaults() *Config {
	return &Config{
		Servers:         DefaultServers,
		SelectedServer:  DefaultServers[0],
		AutoSyncEnabled: true,
	}
}

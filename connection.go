// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func accountID(c Config) string { return strings.TrimRight(c.Server, "/") + "\n" + c.Username }
func (s *Store) bindAccountLocked(c Config) error {
	id := accountID(c)
	if s.state.Account == id {
		return nil
	}
	next := s.clone()
	if next.Account != "" {
		if s.busy {
			return errors.New("wait for sync to finish")
		}
		if len(next.Pending) > 0 {
			return errors.New("sync or discard pending changes before changing accounts")
		}
		next.CalendarColors = nil
		next.Collections = []Collection{}
		next.Resources = map[string]Resource{}
		next.LastSync = ""
	}
	next.Account = id
	return s.commit(next)
}
func (s *Store) ensureAccount(path string) error {
	d, err := loadDAV(path)
	if err != nil {
		return nil
	} // Cached data remains usable offline/unconfigured.
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.bindAccountLocked(d.config)
}
func connectionInfo(path string) map[string]any {
	d, err := loadDAV(path)
	if err != nil {
		return map[string]any{"server": "https://cloud.example.com", "username": "", "has_password": false}
	}
	return map[string]any{"server": d.config.Server, "username": d.config.Username, "has_password": true}
}
func writeConfig(path string, c Config) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return errors.New("cannot create private connection directory")
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".connection-*")
	if err != nil {
		return errors.New("cannot save connection settings")
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(b)
	}
	if err == nil {
		err = f.Sync()
	}
	ce := f.Close()
	if err == nil {
		err = ce
	}
	if err == nil {
		err = os.Rename(f.Name(), path)
	}
	if err != nil {
		return errors.New("cannot save connection settings")
	}
	return nil
}
func (s *Store) saveConnection(path string, c Config, timezone string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.busy {
		return errors.New("wait for sync to finish")
	}
	c.Server = strings.TrimRight(strings.TrimSpace(c.Server), "/")
	c.Username = strings.TrimSpace(c.Username)
	old, err := loadDAV(path)
	if err == nil {
		if err = s.bindAccountLocked(old.config); err != nil {
			return err
		}
		c.Timezone = old.config.Timezone
		if c.Password == "" && accountID(c) == accountID(old.config) {
			c.Password = old.config.Password
		}
	} else {
		c.Timezone = timezone
	}
	if _, err = newDAV(c); err != nil {
		return err
	}
	if s.state.Account != accountID(c) && len(s.state.Pending) > 0 {
		return errors.New("sync or discard pending changes before changing accounts")
	}
	// Mark an unidentified old cache so it cannot be shown as another account's data.
	if s.state.Account == "" && len(s.state.Resources) > 0 {
		next := s.clone()
		next.Account = "unidentified"
		if err = s.commit(next); err != nil {
			return err
		}
	}
	if err = writeConfig(path, c); err != nil {
		return err
	}
	return s.bindAccountLocked(c)
}
func (s *Store) preferences(format, sort string) error {
	if format != "" && format != "12h" && format != "24h" {
		return errors.New("choose 12h or 24h")
	}
	if sort != "" && sort != "scheduled" && sort != "due" && sort != "priority" && sort != "title" {
		return errors.New("unknown task sort")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.clone()
	if format != "" {
		next.TimeFormat = format
	}
	if sort != "" {
		next.TaskSort = sort
	}
	return s.commit(next)
}

// SPDX-License-Identifier: GPL-3.0-only
package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type LauncherSettings struct {
	Enabled bool `json:"enabled"`
}

func (s *Store) launcherPath() string { return filepath.Join(filepath.Dir(s.path), "launcher.json") }

// A separate, non-secret file lets the home-screen hook read this preference
// without starting the backend or reading the account cache.
func (s *Store) launcherSettings() (LauncherSettings, error) {
	value := LauncherSettings{Enabled: true}
	b, err := os.ReadFile(s.launcherPath())
	if os.IsNotExist(err) {
		return value, nil
	}
	if err != nil {
		return value, err
	}
	err = json.Unmarshal(b, &value)
	return value, err
}

func (s *Store) saveLauncher(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := s.launcherPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	b, _ := json.Marshal(LauncherSettings{Enabled: enabled})
	f, err := os.CreateTemp(filepath.Dir(path), ".launcher-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(b)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(f.Name(), path)
	}
	if err != nil {
		return errors.New("cannot save calendar icon setting")
	}
	return nil
}

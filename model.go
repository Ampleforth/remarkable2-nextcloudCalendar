// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Collection struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Events   bool   `json:"events"`
	Tasks    bool   `json:"tasks"`
	Writable bool   `json:"writable"`
	Enabled  bool   `json:"enabled"`
}
type Resource struct {
	URL      string `json:"url"`
	Calendar string `json:"calendar"`
	ETag     string `json:"etag"`
	Data     string `json:"data"`
}
type Change struct {
	Resource
	Delete bool   `json:"delete"`
	Error  string `json:"error,omitempty"`
}
type State struct {
	Account     string              `json:"account,omitempty"`
	TimeFormat  string              `json:"time_format,omitempty"`
	TaskSort    string              `json:"task_sort,omitempty"`
	Collections []Collection        `json:"collections"`
	Resources   map[string]Resource `json:"resources"`
	Pending     map[string]Change   `json:"pending"`
	LastSync    string              `json:"last_sync"`
}
type Store struct {
	mu      sync.Mutex
	path    string
	state   State
	busy    bool
	message string
}

func openStore(path string) (*Store, error) {
	s := &Store{path: path, state: State{Resources: map[string]Resource{}, Pending: map[string]Change{}, Collections: []Collection{}}}
	b, err := os.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(b, &s.state)
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if s.state.Resources == nil || s.state.Pending == nil {
		return nil, errors.New("invalid cache; preserve it and restore a backup")
	}
	return s, nil
}

// Commit only after fsync and atomic replacement. Failed persistence never acknowledges an edit.
func (s *Store) commit(next State) error {
	b, err := json.Marshal(next)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(s.path), ".calendar-*")
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
		err = os.Rename(f.Name(), s.path)
	}
	if err != nil {
		return err
	}
	s.state = next
	return nil
}
func (s *Store) clone() State {
	b, _ := json.Marshal(s.state)
	var next State
	_ = json.Unmarshal(b, &next)
	return next
}
func (s *Store) snapshot() State { s.mu.Lock(); defer s.mu.Unlock(); return s.clone() }
func (s *Store) status() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	pending := []map[string]any{}
	for url, c := range s.state.Pending {
		title := "Untitled item"
		if cal, err := decode(c.Data); err == nil {
			for _, component := range cal.Children {
				if component.Name == "VEVENT" || component.Name == "VTODO" {
					title = text(component, "SUMMARY")
					break
				}
			}
		}
		pending = append(pending, map[string]any{"id": url, "title": title, "delete": c.Delete, "error": c.Error})
	}
	format := s.state.TimeFormat
	if format == "" {
		format = "24h"
	}
	sort := s.state.TaskSort
	if sort == "" {
		sort = "scheduled"
	}
	return map[string]any{"busy": s.busy, "message": s.message, "last_sync": s.state.LastSync, "pending": pending, "collections": s.state.Collections, "time_format": format, "task_sort": sort}
}
func nowStamp() string { return time.Now().UTC().Format(time.RFC3339) }

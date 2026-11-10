// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *Store) startSync(configPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.busy {
		return errors.New("sync already running")
	}
	d, err := loadDAV(configPath)
	if err != nil {
		return err
	}
	s.busy = true
	s.message = "Connecting to Nextcloud…"
	go func() {
		err := s.syncDAV(d)
		s.mu.Lock()
		defer s.mu.Unlock()
		s.busy = false
		if err != nil {
			s.message = err.Error()
		} else {
			s.message = "Sync complete"
		}
	}()
	return nil
}
func (s *Store) mutate(fn func(*State) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := s.clone()
	if err := fn(&next); err != nil {
		return err
	}
	return s.commit(next)
}
func (s *Store) syncDAV(d *DAV) error {
	collections, err := d.discover()
	if err != nil {
		return err
	}
	old := s.snapshot()
	for n := range collections {
		for _, c := range old.Collections {
			if collections[n].ID == c.ID {
				collections[n].Enabled = c.Enabled
			}
		}
	}
	if err = s.mutate(func(next *State) error { next.Collections = collections; return nil }); err != nil {
		return err
	}
	collectionMap := map[string]Collection{}
	for _, c := range collections {
		collectionMap[c.ID] = c
	}
	failures := 0
	for id, p := range old.Pending {
		coll, ok := collectionMap[p.Calendar]
		var writeErr error
		if !ok || !coll.Writable {
			writeErr = errors.New("calendar unavailable or read-only")
		} else if p.Error != "" {
			// A conflict is sticky: another Sync never retries it as a blind overwrite.
			writeErr = errors.New(p.Error)
		} else {
			writeErr = s.flushOne(d, id, p)
		}
		if writeErr != nil {
			failures++
			if err = s.mutate(func(next *State) error {
				c := next.Pending[id]
				c.Error = writeErr.Error()
				next.Pending[id] = c
				return nil
			}); err != nil {
				return err
			}
		}
	}
	loc, _ := time.LoadLocation(d.config.Timezone)
	now := time.Now().In(loc)
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, -1, 0)
	end := start.AddDate(0, 8, 0)
	for _, c := range collections {
		if !c.Enabled {
			continue
		}
		resources, err := d.fetch(c, start, end)
		if err != nil {
			return fmt.Errorf("%s: %w; previous cache kept", c.Name, err)
		}
		// Validate every raw resource before replacing a working cache.
		for _, r := range resources {
			if _, err := decode(r.Data); err != nil {
				return fmt.Errorf("%s: invalid calendar data; previous cache kept", c.Name)
			}
		}
		if err = s.mutate(func(next *State) error {
			for id, r := range next.Resources {
				if r.Calendar == c.ID {
					delete(next.Resources, id)
				}
			}
			for id, r := range resources {
				next.Resources[id] = r
			}
			return nil
		}); err != nil {
			return err
		}
	}
	if err = s.mutate(func(next *State) error { next.LastSync = nowStamp(); return nil }); err != nil {
		return err
	}
	if failures > 0 {
		return fmt.Errorf("downloads complete; %d pending change(s) need review in Settings", failures)
	}
	return nil
}
func (s *Store) flushOne(d *DAV, id string, p Change) error {
	// Reconcile first: a previous PUT/DELETE response may have been lost.
	data, headers, status, err := d.request("GET", id, "", "")
	exists := err == nil
	if err != nil && status != 404 {
		return errors.New("could not check pending resource; retry in Settings")
	}
	if p.Delete && !exists {
		return s.ack(id, nil)
	}
	if !p.Delete && exists && sameCalendar(string(data), p.Data) {
		r := p.Resource
		r.Data = string(data)
		r.ETag = headers.Get("ETag")
		return s.ack(id, &r)
	}
	if p.ETag == "" && exists {
		return errors.New("conflict: creation URL already exists; review before retrying")
	}
	if p.ETag != "" && (!exists || headers.Get("ETag") == "" || headers.Get("ETag") != p.ETag) {
		return errors.New("conflict: server item changed or was deleted; discard the pending edit to keep the server version")
	}
	method := "PUT"
	etag := p.ETag
	if etag == "" {
		etag = "*"
	}
	if p.Delete {
		method = "DELETE"
	}
	_, _, status, err = d.request(method, id, p.Data, etag)
	if err != nil {
		if status == 412 || status == 409 {
			return errors.New("conflict: server changed; discard pending edit to keep the server version")
		}
		return errors.New("write not confirmed; retry in Settings to reconcile safely")
	}
	if p.Delete {
		return s.ack(id, nil)
	}
	// Persist the server representation and new ETag, never an invented ETag.
	data, headers, _, err = d.request("GET", id, "", "")
	if err != nil || headers.Get("ETag") == "" {
		return errors.New("write sent; confirmation pending, retry in Settings")
	}
	if _, err = decode(string(data)); err != nil {
		return errors.New("server returned invalid data after write")
	}
	r := p.Resource
	r.Data = string(data)
	r.ETag = headers.Get("ETag")
	return s.ack(id, &r)
}
func (s *Store) ack(id string, r *Resource) error {
	return s.mutate(func(next *State) error {
		delete(next.Pending, id)
		if r == nil {
			delete(next.Resources, id)
		} else {
			next.Resources[id] = *r
		}
		return nil
	})
}
func (s *Store) resolvePending(id, action string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.busy {
		return errors.New("wait for sync to finish")
	}
	next := s.clone()
	p, ok := next.Pending[id]
	if !ok {
		return errors.New("pending change not found")
	}
	switch action {
	case "discard":
		delete(next.Pending, id)
	case "retry":
		if strings.HasPrefix(p.Error, "conflict:") {
			return errors.New("conflicts require review; discard to keep the server version")
		}
		p.Error = ""
		next.Pending[id] = p
	default:
		return errors.New("unknown action")
	}
	return s.commit(next)
}

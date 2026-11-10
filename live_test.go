// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

// Opt-in only: creates unique temporary resources and removes them on completion.
// Never include credentials in command arguments or test output.
func TestLiveRoundtrip(t *testing.T) {
	config := os.Getenv("CALENDAR_LIVE_CONFIG")
	target := os.Getenv("CALENDAR_LIVE_COLLECTION")
	if config == "" || target == "" {
		t.Skip("live account test is opt-in")
	}
	d, err := loadDAV(config)
	if err != nil {
		t.Fatal(err)
	}
	collections, err := d.discover()
	if err != nil {
		t.Fatal(err)
	}
	var collection Collection
	for _, c := range collections {
		if c.ID == target {
			collection = c
		}
	}
	if !collection.Writable || !collection.Events || !collection.Tasks {
		t.Fatal("test collection must support events, tasks, and writes")
	}
	s, _ := testStore(t)
	s.state.Collections = []Collection{collection}
	created := map[string]bool{}
	defer func() {
		for id := range created {
			_, h, status, err := d.request("GET", id, "", "")
			if status == 404 {
				continue
			}
			if err != nil || h.Get("ETag") == "" {
				t.Errorf("temporary resource cleanup needs review: %s", id)
				continue
			}
			if _, _, _, err = d.request("DELETE", id, "", h.Get("ETag")); err != nil {
				t.Errorf("temporary resource cleanup failed: %s", id)
			}
		}
	}()
	flush := func() {
		t.Helper()
		for id, p := range s.snapshot().Pending {
			created[id] = true
			if err := s.flushOne(d, id, p); err != nil {
				t.Fatal(err)
			}
		}
	}
	loc, _ := time.LoadLocation(d.config.Timezone)
	now := time.Now().In(loc)
	start := now.AddDate(0, 0, 1).Format("2006-01-02") + "T09:00"
	end := start[:10] + "T09:15"
	for _, kind := range []string{"VEVENT", "VTODO"} {
		edit := Edit{Item: Item{Calendar: target, Kind: kind, Title: "reMarkable sync verification (temporary)", Start: start, End: end}}
		if err = s.queue(edit, loc); err != nil {
			t.Fatal(err)
		}
		flush()
		a := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		items, err := s.items(a, a.AddDate(0, 0, 3), loc)
		if err != nil {
			t.Fatal(err)
		}
		var current Item
		for _, i := range items {
			if i.Kind == kind {
				current = i
			}
		}
		if current.ID == "" {
			t.Fatal("created item not found")
		}
		edit.Item = current
		edit.Title = "reMarkable sync verification updated (temporary)"
		if kind == "VTODO" {
			edit.Completed = true
		}
		if err = s.queue(edit, loc); err != nil {
			t.Fatal(err)
		}
		flush()
		items, err = s.items(a, a.AddDate(0, 0, 3), loc)
		if err != nil {
			t.Fatal(err)
		}
		for _, i := range items {
			if i.ID == current.ID {
				current = i
			}
		}
		if !strings.Contains(current.Title, "updated") || kind == "VTODO" && !current.Completed {
			t.Fatal("update did not roundtrip")
		}
		if kind == "VTODO" {
			edit.Item = current
			edit.Completed = false
			if err = s.queue(edit, loc); err != nil {
				t.Fatal(err)
			}
			flush()
			items, err = s.items(a, a.AddDate(0, 0, 3), loc)
			if err != nil {
				t.Fatal(err)
			}
			for _, i := range items {
				if i.ID == current.ID {
					current = i
				}
			}
			if current.Completed {
				t.Fatal("task did not reopen")
			}
		}
		edit.Item = current
		edit.Delete = true
		if err = s.queue(edit, loc); err != nil {
			t.Fatal(err)
		}
		flush()
		_, _, status, _ := d.request("GET", current.ID, "", "")
		if status != 404 {
			t.Fatal("temporary item was not deleted")
		}
		delete(created, current.ID)
	}
}

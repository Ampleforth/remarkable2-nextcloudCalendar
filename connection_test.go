// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
package main

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConnectionSettingsPrivacyAndAccountIsolation(t *testing.T) {
	s, c := testStore(t)
	path := filepath.Join(t.TempDir(), "private", "config.json")
	handler := api(s, path, time.UTC)
	call := func(method, route, body string) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, route, strings.NewReader(body))
		r.Header.Set("X-RemarkableCalendar-Request", "sync")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	first := `{"server":"https://nextcloud.example","username":"one","app_password":"secret-one"}`
	if w := call("POST", "/connection", first); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("connection file must be private")
	}
	if w := call("GET", "/connection", ""); strings.Contains(w.Body.String(), "secret-one") || strings.Contains(w.Body.String(), "app_password") {
		t.Fatal("saved password leaked")
	}
	if w := call("POST", "/connection", `{"server":"https://nextcloud.example","username":"one","app_password":""}`); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	d, err := loadDAV(path)
	if err != nil || d.config.Password != "secret-one" {
		t.Fatal("blank unchanged password not retained")
	}
	s.state.Collections = []Collection{c}
	s.state.Resources["cached"] = Resource{URL: "cached", Calendar: c.ID, Data: "cached"}
	s.state.Pending["pending"] = Change{Resource: Resource{URL: "pending", Calendar: c.ID}}
	if w := call("POST", "/connection", `{"server":"https://nextcloud.example","username":"two","app_password":"secret-two"}`); w.Code == 200 {
		t.Fatal("changed account with unsynced edits")
	}
	d, _ = loadDAV(path)
	if d.config.Username != "one" {
		t.Fatal("rejected account replaced credentials")
	}
	if w := call("POST", "/connection", `{"server":"https://nextcloud.example","username":"one","app_password":"rotated"}`); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if len(s.state.Pending) != 1 {
		t.Fatal("password rotation lost pending edits")
	}
	delete(s.state.Pending, "pending")
	if w := call("POST", "/connection", `{"server":"https://nextcloud.example","username":"two","app_password":"secret-two"}`); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if len(s.state.Resources) != 0 || len(s.state.Collections) != 0 {
		t.Fatal("previous account data mixed into new account")
	}
	if w := call("POST", "/connection", `{"server":"http://nextcloud.example","username":"two","app_password":"secret-two"}`); w.Code == 200 {
		t.Fatal("insecure connection accepted")
	}
}
func TestPreferencesPersistAndValidate(t *testing.T) {
	s, _ := testStore(t)
	if err := s.preferences("12h", "priority"); err != nil {
		t.Fatal(err)
	}
	restored, err := openStore(s.path)
	if err != nil {
		t.Fatal(err)
	}
	if restored.status()["time_format"] != "12h" || restored.status()["task_sort"] != "priority" {
		t.Fatal("preferences lost after restart")
	}
	if err = s.preferences("13h", "scheduled"); err == nil {
		t.Fatal("invalid time format accepted")
	}
	if err = s.preferences("24h", "unknown"); err == nil {
		t.Fatal("invalid task sort accepted")
	}
}
func TestTaskScheduledDatePriorityAndCompletionPreservation(t *testing.T) {
	s, c := testStore(t)
	edit := Edit{Item: Item{Calendar: c.ID, Kind: "VTODO", Title: "Scheduled task", Scheduled: "2026-11-10T09:15", Due: "2026-11-12T17:00", Priority: 1}}
	if err := s.queue(edit, time.UTC); err != nil {
		t.Fatal(err)
	}
	item := mustItems(t, s)[0]
	if item.Scheduled != "2026-11-10T09:15" || item.Priority != 1 {
		t.Fatalf("scheduled date/priority not parsed: %+v", item)
	}
	edit.Item = item
	edit.Completed = true
	if err := s.queue(edit, time.UTC); err != nil {
		t.Fatal(err)
	}
	item = mustItems(t, s)[0]
	if !item.Completed || item.Scheduled != "2026-11-10T09:15" || item.Due != "2026-11-12T17:00" || item.Priority != 1 {
		t.Fatal("completion changed schedule or priority")
	}
	b, _ := json.Marshal(s.status())
	if strings.Contains(string(b), "app_password") {
		t.Fatal("credentials in status")
	}
}

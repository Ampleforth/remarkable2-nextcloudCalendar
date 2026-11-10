// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testStore(t *testing.T) (*Store, Collection) {
	t.Helper()
	s, e := openStore(filepath.Join(t.TempDir(), "state.json"))
	if e != nil {
		t.Fatal(e)
	}
	c := Collection{ID: "https://nextcloud.example/cal/", Name: "Personal", Events: true, Tasks: true, Writable: true, Enabled: true}
	s.state.Collections = []Collection{c}
	return s, c
}
func testRange() (time.Time, time.Time) {
	a := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	return a, a.AddDate(0, 1, 0)
}
func mustItems(t *testing.T, s *Store) []Item {
	t.Helper()
	a, b := testRange()
	items, e := s.items(a, b, time.UTC)
	if e != nil {
		t.Fatal(e)
	}
	return items
}
func TestOfflineTaskLifecycle(t *testing.T) {
	s, c := testStore(t)
	e := Edit{Item: Item{Calendar: c.ID, Kind: "VTODO", Title: "Buy milk"}}
	if err := s.queue(e, time.UTC); err != nil {
		t.Fatal(err)
	}
	items := mustItems(t, s)
	if len(items) != 1 || items[0].Due != "" || !items[0].Pending {
		t.Fatalf("undated offline task missing: %+v", items)
	}
	reopened, err := openStore(s.path)
	if err != nil {
		t.Fatal(err)
	}
	items = mustItems(t, reopened)
	e.Item = items[0]
	e.Completed = true
	if err = reopened.queue(e, time.UTC); err != nil {
		t.Fatal(err)
	}
	i := mustItems(t, reopened)[0]
	if !i.Completed {
		t.Fatal("not completed")
	}
	cal, _ := decode(reopened.state.Pending[i.ID].Data)
	var data string
	data, _ = encode(cal)
	if !strings.Contains(data, "COMPLETED:") || !strings.Contains(data, "PERCENT-COMPLETE:100") {
		t.Fatal(data)
	}
	e.Item = i
	e.Completed = false
	if err = reopened.queue(e, time.UTC); err != nil {
		t.Fatal(err)
	}
	i = mustItems(t, reopened)[0]
	data = reopened.state.Pending[i.ID].Data
	if i.Completed || strings.Contains(data, "\r\nCOMPLETED:") || !strings.Contains(data, "PERCENT-COMPLETE:0") {
		t.Fatal(data)
	}
	e.Item = i
	e.Delete = true
	if err = reopened.queue(e, time.UTC); err != nil {
		t.Fatal(err)
	}
	if len(reopened.state.Pending) != 0 || len(mustItems(t, reopened)) != 0 {
		t.Fatal("deleting an unsynced creation must cancel the outbox entry")
	}
}
func TestEditPreservesPropertiesAndRejectsStaleRevision(t *testing.T) {
	s, c := testStore(t)
	r := Resource{URL: c.ID + "one.ics", Calendar: c.ID, ETag: `"v1"`, Data: "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//Test//EN\r\nBEGIN:VEVENT\r\nUID:one\r\nSUMMARY:Old\r\nDTSTART:20261110T090000Z\r\nDTEND:20261110T100000Z\r\nX-CUSTOM:keep me\r\nBEGIN:VALARM\r\nACTION:DISPLAY\r\nTRIGGER:-PT10M\r\nDESCRIPTION:Reminder\r\nEND:VALARM\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"}
	s.state.Resources[r.URL] = r
	e := Edit{Item: mustItems(t, s)[0]}
	e.Title = "New"
	if err := s.queue(e, time.UTC); err != nil {
		t.Fatal(err)
	}
	data := s.state.Pending[r.URL].Data
	if !strings.Contains(data, "X-CUSTOM:keep me") || !strings.Contains(data, "BEGIN:VALARM") {
		t.Fatal(data)
	}
	if err := s.queue(e, time.UTC); err == nil {
		t.Fatal("stale local revision accepted")
	}
}
func TestRecurringExceptionsAndAllDay(t *testing.T) {
	s, c := testStore(t)
	r := Resource{URL: c.ID + "series.ics", Calendar: c.ID, ETag: `"v1"`, Data: "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//Test//EN\r\nBEGIN:VEVENT\r\nUID:series\r\nSUMMARY:Daily\r\nDTSTART:20261110T090000Z\r\nDTEND:20261110T100000Z\r\nRRULE:FREQ=DAILY;COUNT=4\r\nEXDATE:20261111T090000Z,20261113T090000Z\r\nEND:VEVENT\r\nBEGIN:VEVENT\r\nUID:series\r\nRECURRENCE-ID:20261112T090000Z\r\nSUMMARY:Moved\r\nDTSTART:20261112T120000Z\r\nDTEND:20261112T130000Z\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"}
	s.state.Resources[r.URL] = r
	items := mustItems(t, s)
	if len(items) != 2 || items[0].Writable || items[1].Start != "2026-11-12T12:00" {
		t.Fatalf("%+v", items)
	}
	if err := s.queue(Edit{Item: items[0]}, time.UTC); err == nil {
		t.Fatal("recurring write accepted")
	}
	e := Edit{Item: Item{Calendar: c.ID, Kind: "VEVENT", Title: "Holiday", AllDay: true, Start: "2026-11-20", End: "2026-11-22"}}
	if err := s.queue(e, time.UTC); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, i := range mustItems(t, s) {
		if i.Title == "Holiday" {
			found = true
			if !i.AllDay || i.Start != "2026-11-20T00:00" || i.End != "2026-11-22T00:00" {
				t.Fatalf("%+v", i)
			}
		}
	}
	if !found {
		t.Fatal("all-day event missing")
	}
}
func TestValidationAndReadonly(t *testing.T) {
	s, c := testStore(t)
	e := Edit{Item: Item{Calendar: c.ID, Kind: "VEVENT", Title: "Invalid", Start: "2026-11-10T10:00", End: "2026-11-10T09:00"}}
	if err := s.queue(e, time.UTC); err == nil {
		t.Fatal("negative duration accepted")
	}
	if len(s.state.Pending) != 0 {
		t.Fatal("invalid edit persisted")
	}
	e.Kind = "VTODO"
	e.Due = "2026-02-30"
	if err := s.queue(e, time.UTC); err == nil {
		t.Fatal("invalid date accepted")
	}
	s.state.Collections[0].Writable = false
	e.Due = ""
	if err := s.queue(e, time.UTC); err == nil {
		t.Fatal("read-only edit accepted")
	}
	s.state.Collections[0].Writable = true
	s.busy = true
	if err := s.queue(e, time.UTC); err == nil {
		t.Fatal("edit during sync accepted")
	}
}
func TestConditionalWriteAndConflict(t *testing.T) {
	s, _ := testStore(t)
	var stored string
	etag := `"one"`
	puts := 0
	lastMatch := ""
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != "test" || password != "secret" {
			t.Error("missing auth")
			w.WriteHeader(401)
			return
		}
		switch r.Method {
		case "GET":
			if stored == "" {
				w.WriteHeader(404)
				return
			}
			w.Header().Set("ETag", etag)
			fmt.Fprint(w, stored)
		case "PUT":
			puts++
			lastMatch = r.Header.Get("If-None-Match")
			if lastMatch == "" {
				lastMatch = r.Header.Get("If-Match")
			}
			b, _ := io.ReadAll(r.Body)
			stored = string(b)
			etag = `"two"`
			w.WriteHeader(201)
		case "DELETE":
			if r.Header.Get("If-Match") != etag {
				t.Error("missing delete precondition")
			}
			stored = ""
			w.WriteHeader(204)
		}
	}))
	defer server.Close()
	d, e := newDAV(Config{Server: server.URL, Username: "test", Password: "secret", Timezone: "UTC"})
	if e != nil {
		t.Fatal(e)
	}
	d.client = server.Client()
	c := s.state.Collections[0]
	c.ID = server.URL + "/cal/"
	s.state.Collections[0] = c
	if e = s.queue(Edit{Item: Item{Calendar: c.ID, Kind: "VTODO", Title: "Task"}}, time.UTC); e != nil {
		t.Fatal(e)
	}
	var id string
	var p Change
	for id, p = range s.state.Pending {
		break
	}
	if e = s.flushOne(d, id, p); e != nil {
		t.Fatal(e)
	}
	if lastMatch != "*" || len(s.state.Pending) != 0 || s.state.Resources[id].ETag != `"two"` {
		t.Fatalf("creation precondition or ack: %s %+v", lastMatch, s.state)
	}
	// An acknowledged PUT with a lost response is reconciled by GET, without a second PUT.
	s.state.Pending[id] = p
	if e = s.flushOne(d, id, p); e != nil {
		t.Fatal(e)
	}
	if puts != 1 {
		t.Fatal("duplicate creation")
	}
	item := mustItems(t, s)[0]
	edit := Edit{Item: item}
	edit.Title = "Offline edit"
	if e = s.queue(edit, time.UTC); e != nil {
		t.Fatal(e)
	}
	etag = `"changed-on-server"`
	e = s.flushOne(d, id, s.state.Pending[id])
	if e == nil || !strings.HasPrefix(e.Error(), "conflict:") {
		t.Fatal("server edit not protected", e)
	}
	if puts != 1 {
		t.Fatal("conflict overwrote server")
	}
	s.state.Pending[id] = Change{Resource: s.state.Resources[id], Delete: true}
	etag = s.state.Resources[id].ETag
	if e = s.flushOne(d, id, s.state.Pending[id]); e != nil {
		t.Fatal(e)
	}
	if len(s.state.Resources) != 0 {
		t.Fatal("delete not acknowledged")
	}
	// A second deletion after a lost response succeeds on 404.
	if e = s.flushOne(d, id, Change{Resource: p.Resource, Delete: true}); e != nil {
		t.Fatal(e)
	}
}
func TestDiscoveryAndUnboundedTaskQuery(t *testing.T) {
	var server *httptest.Server
	taskQuery := false
	server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		body := string(b)
		w.WriteHeader(207)
		props := ""
		href := r.URL.Path
		switch {
		case strings.Contains(body, "current-user-principal"):
			props = `<d:current-user-principal><d:href>/principal/</d:href></d:current-user-principal>`
		case strings.Contains(body, "calendar-home-set"):
			props = `<c:calendar-home-set><d:href>/calendars/test/</d:href></c:calendar-home-set>`
		case r.Method == "PROPFIND":
			href = "/calendars/test/personal/"
			props = `<d:displayname>Personal</d:displayname><d:resourcetype><c:calendar/></d:resourcetype><d:current-user-privilege-set><d:privilege><d:write/></d:privilege></d:current-user-privilege-set><c:supported-calendar-component-set><c:comp name="VEVENT"/><c:comp name="VTODO"/></c:supported-calendar-component-set>`
		case strings.Contains(body, `name="VTODO"`):
			taskQuery = true
			if strings.Contains(body, "time-range") {
				t.Error("task query must include undated tasks")
			}
			fmt.Fprint(w, `<d:multistatus xmlns:d="DAV:"/>`)
			return
		case r.Method == "REPORT":
			fmt.Fprint(w, `<d:multistatus xmlns:d="DAV:"/>`)
			return
		}
		fmt.Fprintf(w, `<d:multistatus xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav"><d:response><d:href>%s</d:href><d:propstat><d:prop>%s</d:prop><d:status>HTTP/1.1 200 OK</d:status></d:propstat></d:response></d:multistatus>`, href, props)
	}))
	defer server.Close()
	d, e := newDAV(Config{Server: server.URL, Username: "test", Password: "secret", Timezone: "UTC"})
	if e != nil {
		t.Fatal(e)
	}
	d.client = server.Client()
	collections, e := d.discover()
	if e != nil {
		t.Fatal(e)
	}
	if len(collections) != 1 || !collections[0].Writable || !collections[0].Tasks || !collections[0].Events {
		t.Fatalf("%+v", collections)
	}
	a, b := testRange()
	if _, e = d.fetch(collections[0], a, b); e != nil {
		t.Fatal(e)
	}
	if !taskQuery {
		t.Fatal("tasks not queried")
	}
	if _, e = d.resolve(server.URL, "https://evil.example/cal"); e == nil {
		t.Fatal("cross-origin credentials allowed")
	}
	if _, e = d.resolve(server.URL, "http://"+strings.TrimPrefix(server.URL, "https://")); e == nil {
		t.Fatal("HTTP downgrade allowed")
	}
}
func TestHTTPMutationGuard(t *testing.T) {
	s, _ := testStore(t)
	handler := api(s, "missing.json", time.UTC)
	for _, origin := range []string{"", "https://evil.example"} {
		r := httptest.NewRequest("POST", "/sync", strings.NewReader("{}"))
		if origin != "" {
			r.Header.Set("Origin", origin)
			r.Header.Set("X-RemarkableCalendar-Request", "sync")
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != 403 {
			t.Fatalf("unsafe mutation accepted: %d", w.Code)
		}
	}
	r := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
}

func TestInvalidDAVResponseCannotReplaceCache(t *testing.T) {
	for _, data := range []string{`<html>Login page</html>`, `<multistatus/>`, `<d:multistatus xmlns:d="DAV:"><broken>`} {
		if _, err := parseDAV([]byte(data)); err == nil {
			t.Fatalf("accepted non-CalDAV response: %s", data)
		}
	}
}
func TestAllDayAcrossDST(t *testing.T) {
	s, c := testStore(t)
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	err = s.queue(Edit{Item: Item{Calendar: c.ID, Kind: "VEVENT", Title: "DST weekend", AllDay: true, Start: "2026-10-31", End: "2026-11-02"}}, loc)
	if err != nil {
		t.Fatal(err)
	}
	a := time.Date(2026, 10, 30, 0, 0, 0, 0, loc)
	items, err := s.items(a, a.AddDate(0, 0, 5), loc)
	if err != nil || len(items) != 1 || items[0].End != "2026-11-02T00:00" {
		t.Fatalf("all-day DST span: %+v %v", items, err)
	}
}

func TestLostLocalSaveResponseDoesNotDuplicateCreation(t *testing.T) {
	s, c := testStore(t)
	edit := Edit{Item: Item{Calendar: c.ID, Kind: "VTODO", Title: "Only once"}, CreateID: "0123456789abcdef0123456789abcdef"}
	if err := s.queue(edit, time.UTC); err != nil {
		t.Fatal(err)
	}
	if err := s.queue(edit, time.UTC); err == nil {
		t.Fatal("duplicate save should direct user to the saved item")
	}
	if len(s.state.Pending) != 1 {
		t.Fatal("duplicate offline creation")
	}
}

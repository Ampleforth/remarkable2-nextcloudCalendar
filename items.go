// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	ical "github.com/emersion/go-ical"
)

type Item struct {
	ID           string `json:"id"`
	Calendar     string `json:"calendar"`
	CalendarName string `json:"calendar_name"`
	Kind         string `json:"kind"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Location     string `json:"location"`
	Start        string `json:"start"`
	End          string `json:"end"`
	Due          string `json:"due"`
	Scheduled    string `json:"scheduled"`
	Priority     int    `json:"priority"`
	AllDay       bool   `json:"all_day"`
	Completed    bool   `json:"completed"`
	Recurring    bool   `json:"recurring"`
	Writable     bool   `json:"writable"`
	Pending      bool   `json:"pending"`
	Error        string `json:"error"`
	Revision     string `json:"revision"`
}
type Edit struct {
	CaptureID string `json:"capture_id,omitempty"`
	Item
	Delete   bool   `json:"delete"`
	CreateID string `json:"create_id,omitempty"`
}

func decode(data string) (*ical.Calendar, error) {
	return ical.NewDecoder(strings.NewReader(data)).Decode()
}
func encode(cal *ical.Calendar) (string, error) {
	var b bytes.Buffer
	err := ical.NewEncoder(&b).Encode(cal)
	return b.String(), err
}
func text(c *ical.Component, name string) string { v, _ := c.Props.Text(name); return v }
func setInteger(c *ical.Component, name string, value int) {
	p := ical.NewProp(name)
	p.Value = fmt.Sprint(value)
	c.Props.Set(p)
}
func stamp(c *ical.Component, name string, loc *time.Location) (time.Time, error) {
	return c.Props.DateTime(name, loc)
}
func revision(r Resource) string {
	h := sha256.Sum256([]byte(r.ETag + "\x00" + r.Data))
	return hex.EncodeToString(h[:])
}
func recurring(cal *ical.Calendar) bool {
	count := 0
	for _, c := range cal.Children {
		if c.Name == "VEVENT" || c.Name == "VTODO" {
			count++
			if c.Props.Get("RRULE") != nil || c.Props.Get("RECURRENCE-ID") != nil || c.Props.Get("RDATE") != nil || c.Props.Get("EXDATE") != nil {
				return true
			}
		}
	}
	return count > 1
}
func editable(cal *ical.Calendar) bool {
	if recurring(cal) {
		return false
	}
	for _, c := range cal.Children {
		if c.Name == "VEVENT" || c.Name == "VTODO" {
			if c.Props.Get("ATTENDEE") != nil || c.Props.Get("ORGANIZER") != nil {
				return false
			}
		}
	}
	return true
}
func itemFor(r Resource, c *ical.Component, coll Collection, loc *time.Location, canEdit, recurs bool) Item {
	i := Item{ID: r.URL, Calendar: r.Calendar, CalendarName: coll.Name, Kind: c.Name, Title: text(c, "SUMMARY"), Description: text(c, "DESCRIPTION"), Location: text(c, "LOCATION"), Writable: coll.Writable && canEdit, Recurring: recurs, Revision: revision(r)}
	if t, e := stamp(c, "DTSTART", loc); e == nil && !t.IsZero() {
		i.Start = t.In(loc).Format("2006-01-02T15:04")
		i.AllDay = c.Props.Get("DTSTART").ValueType() == ical.ValueDate
	}
	if c.Name == "VEVENT" {
		ev := ical.Event{Component: c}
		if t, e := ev.DateTimeEnd(loc); e == nil && !t.IsZero() {
			i.End = t.In(loc).Format("2006-01-02T15:04")
		}
	}
	if p := c.Props.Get("DUE"); p != nil {
		if t, e := p.DateTime(loc); e == nil {
			i.Due = t.In(loc).Format("2006-01-02")
			if p.ValueType() != ical.ValueDate {
				i.Due = t.In(loc).Format("2006-01-02T15:04")
			}
		}
	}
	percent := 0
	if p := c.Props.Get("PERCENT-COMPLETE"); p != nil {
		percent, _ = p.Int()
	}
	i.Completed = text(c, "STATUS") == "COMPLETED" || percent == 100
	if c.Name == "VTODO" {
		i.Scheduled = i.Start
		if p := c.Props.Get("DTSTART"); p != nil && p.ValueType() == ical.ValueDate && len(i.Scheduled) >= 10 {
			i.Scheduled = i.Scheduled[:10]
		}
		if p := c.Props.Get("PRIORITY"); p != nil {
			i.Priority, _ = p.Int()
		}
	}
	return i
}
func (s *Store) items(start, end time.Time, loc *time.Location) ([]Item, error) {
	state := s.snapshot()
	out := []Item{}
	collections := map[string]Collection{}
	for _, c := range state.Collections {
		collections[c.ID] = c
	}
	resources := state.Resources
	for id, p := range state.Pending {
		if p.Delete {
			delete(resources, id)
		} else {
			resources[id] = p.Resource
		}
	}
	for id, r := range resources {
		coll, ok := collections[r.Calendar]
		if !ok || !coll.Enabled {
			continue
		}
		cal, err := decode(r.Data)
		if err != nil {
			return nil, fmt.Errorf("cannot parse cached resource in %s", coll.Name)
		}
		recurs := recurring(cal)
		canEdit := editable(cal)
		// Detached instances replace the master instance identified by RECURRENCE-ID.
		exceptions := map[string]bool{}
		for _, c := range cal.Children {
			if p := c.Props.Get("RECURRENCE-ID"); p != nil {
				t, e := p.DateTime(loc)
				if e != nil {
					return nil, e
				}
				if p.Params.Get("RANGE") != "" {
					return nil, errors.New("THISANDFUTURE recurrence is unsupported")
				}
				exceptions[text(c, "UID")+t.UTC().Format(time.RFC3339)] = true
			}
		}
		for _, c := range cal.Children {
			if c.Name != "VEVENT" && c.Name != "VTODO" {
				continue
			}
			if text(c, "STATUS") == "CANCELLED" {
				continue
			}
			i := itemFor(r, c, coll, loc, canEdit, recurs)
			if p, ok := state.Pending[id]; ok {
				i.Pending = true
				i.Error = p.Error
			}
			if c.Name == "VTODO" {
				out = append(out, i)
				continue
			}
			a, err := stamp(c, "DTSTART", loc)
			if err != nil {
				return nil, err
			}
			ev := ical.Event{Component: c}
			b, err := ev.DateTimeEnd(loc)
			if err != nil {
				return nil, err
			}
			occurrences := []time.Time{a}
			if c.Props.Get("RRULE") != nil {
				// go-ical expects individual EXDATE/RDATE values, whereas RFC 5545 also permits lists.
				for _, name := range []string{"EXDATE", "RDATE"} {
					props := c.Props.Values(name)
					c.Props.Del(name)
					for _, p := range props {
						for _, v := range strings.Split(p.Value, ",") {
							copy := p
							copy.Value = v
							c.Props.Add(&copy)
						}
					}
				}
				set, err := c.RecurrenceSet(loc)
				if err != nil {
					return nil, err
				}
				occurrences = []time.Time{}
				it := set.Iterator()
				for n := 0; ; n++ {
					t, ok := it()
					if !ok || !t.Before(end) {
						break
					}
					if n >= 200000 {
						return nil, errors.New("recurrence expansion limit exceeded")
					}
					if !t.Before(start.Add(-b.Sub(a))) {
						occurrences = append(occurrences, t)
					}
					if len(occurrences) > 5000 {
						return nil, errors.New("too many recurring events in date range")
					}
				}
			} else if c.Props.Get("RDATE") != nil {
				return nil, errors.New("RDATE-only recurrence is unsupported")
			}
			for _, t := range occurrences {
				if c.Props.Get("RECURRENCE-ID") == nil && exceptions[text(c, "UID")+t.UTC().Format(time.RFC3339)] {
					continue
				}
				finish := t.Add(b.Sub(a))
				if i.AllDay {
					ay, am, ad := a.Date()
					by, bm, bd := b.Date()
					days := int(time.Date(by, bm, bd, 0, 0, 0, 0, time.UTC).Sub(time.Date(ay, am, ad, 0, 0, 0, 0, time.UTC)).Hours() / 24)
					if days < 1 {
						days = 1
					}
					finish = t.AddDate(0, 0, days)
				}
				if !t.Before(end) || finish.Before(start) || finish.Equal(start) && !finish.Equal(t) {
					continue
				}
				copy := i
				copy.Start = t.In(loc).Format("2006-01-02T15:04")
				copy.End = finish.In(loc).Format("2006-01-02T15:04")
				out = append(out, copy)
			}
		}
	}
	sort.Slice(out, func(a, b int) bool {
		if out[a].Start == out[b].Start {
			return out[a].Title < out[b].Title
		}
		return out[a].Start < out[b].Start
	})
	return out, nil
}
func parseInput(v string, dateOnly bool, loc *time.Location) (time.Time, error) {
	layout := "2006-01-02T15:04"
	if dateOnly {
		layout = "2006-01-02"
	}
	t, e := time.ParseInLocation(layout, v, loc)
	if e != nil || t.Format(layout) != v {
		return time.Time{}, errors.New("use YYYY-MM-DD dates and HH:MM times")
	}
	return t, nil
}
func (s *Store) queue(e Edit, loc *time.Location) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.busy {
		return errors.New("wait for sync to finish")
	}
	next := s.clone()
	if e.CaptureID != "" {
		if e.ID != "" || e.Kind != "VTODO" || e.Delete {
			return errors.New("capture must create a task")
		}
		if next.Captures[e.CaptureID] {
			return nil
		}
	}
	var coll Collection
	found := false
	for _, c := range next.Collections {
		if c.ID == e.Calendar {
			coll = c
			found = true
		}
	}
	if !found || !coll.Writable {
		return errors.New("choose a writable calendar")
	}
	if e.Kind != "VEVENT" && e.Kind != "VTODO" {
		return errors.New("unsupported item type")
	}
	if e.Kind == "VEVENT" && !coll.Events || e.Kind == "VTODO" && !coll.Tasks {
		return errors.New("calendar does not support this item type")
	}
	var r Resource
	var cal *ical.Calendar
	var c *ical.Component
	if e.ID != "" {
		var ok bool
		r, ok = next.Resources[e.ID]
		if p, pending := next.Pending[e.ID]; pending {
			r = p.Resource
			ok = true
		}
		if !ok || r.Calendar != e.Calendar {
			return errors.New("item no longer available")
		}
		if revision(r) != e.Revision {
			return errors.New("item changed; reopen it before editing")
		}
		var err error
		cal, err = decode(r.Data)
		if err != nil {
			return err
		}
		if !editable(cal) {
			return errors.New("recurring events and invitations are read-only")
		}
		for _, child := range cal.Children {
			if child.Name == e.Kind {
				c = child
				break
			}
		}
		if c == nil {
			return errors.New("item type mismatch")
		}
	} else {
		if e.Delete {
			return errors.New("missing item")
		}
		id := make([]byte, 16)
		if _, err := rand.Read(id); err != nil {
			return err
		}
		uid := hex.EncodeToString(id)
		if e.CreateID != "" {
			decoded, err := hex.DecodeString(e.CreateID)
			if err != nil || len(decoded) != 16 {
				return errors.New("invalid creation identifier")
			}
			uid = e.CreateID
		}
		target := coll.ID + uid + ".ics"
		if _, ok := next.Pending[target]; ok {
			return errors.New("this item was already saved; cancel the editor and reopen it")
		}
		if _, ok := next.Resources[target]; ok {
			return errors.New("this item was already saved; cancel the editor and reopen it")
		}
		r = Resource{URL: target, Calendar: coll.ID}
		cal = ical.NewCalendar()
		cal.Props.SetText("VERSION", "2.0")
		cal.Props.SetText("PRODID", "-//reMarkable Nextcloud//EN")
		c = ical.NewComponent(e.Kind)
		cal.Children = append(cal.Children, c)
		c.Props.SetText("UID", uid+"@remarkable")
	}
	if e.Delete {
		if r.ETag == "" {
			delete(next.Pending, r.URL)
		} else {
			next.Pending[r.URL] = Change{Resource: r, Delete: true}
		}
		return s.commit(next)
	}
	if strings.TrimSpace(e.Title) == "" || len(e.Title) > 4096 {
		return errors.New("a title of at most 4096 bytes is required")
	}
	c.Props.SetText("SUMMARY", e.Title)
	c.Props.SetText("DESCRIPTION", e.Description)
	c.Props.SetText("LOCATION", e.Location)
	c.Props.SetDateTime("DTSTAMP", time.Now().UTC())
	c.Props.SetDateTime("LAST-MODIFIED", time.Now().UTC())
	sequence := 0
	if p := c.Props.Get("SEQUENCE"); p != nil {
		sequence, _ = p.Int()
	}
	setInteger(c, "SEQUENCE", sequence+1)
	if e.Kind == "VEVENT" {
		a, err := parseInput(e.Start, e.AllDay, loc)
		if err != nil {
			return err
		}
		b, err := parseInput(e.End, e.AllDay, loc)
		if err != nil {
			return err
		}
		if !b.After(a) {
			return errors.New("end must be later than start; all-day end dates are exclusive")
		}
		c.Props.Del("DURATION")
		if e.AllDay {
			c.Props.SetDate("DTSTART", a)
			c.Props.SetDate("DTEND", b)
		} else {
			c.Props.SetDateTime("DTSTART", a.UTC())
			c.Props.SetDateTime("DTEND", b.UTC())
		}
	} else {
		if e.Priority < 0 || e.Priority > 9 {
			return errors.New("priority must be 0–9")
		}
		original := itemFor(r, c, coll, loc, true, false)
		if e.Priority != original.Priority {
			setInteger(c, "PRIORITY", e.Priority)
		}
		if e.Scheduled != original.Scheduled {
			if e.Scheduled == "" {
				c.Props.Del("DTSTART")
				c.Props.Del("DURATION")
			} else {
				t, err := parseInput(e.Scheduled, len(e.Scheduled) == 10, loc)
				if err != nil {
					return err
				}
				if len(e.Scheduled) == 10 {
					c.Props.SetDate("DTSTART", t)
				} else {
					c.Props.SetDateTime("DTSTART", t.UTC())
				}
			}
		}
		if e.Due == "" {
			c.Props.Del("DUE")
		} else {
			dateOnly := len(e.Due) == 10
			t, err := parseInput(e.Due, dateOnly, loc)
			if err != nil {
				return err
			}
			if p := c.Props.Get("DTSTART"); p != nil {
				start, err := p.DateTime(loc)
				if err != nil {
					return err
				}
				if !t.After(start) || dateOnly != (p.ValueType() == ical.ValueDate) {
					return errors.New("task due date must follow its existing start with the same date type")
				}
			}
			c.Props.Del("DURATION")
			if dateOnly {
				c.Props.SetDate("DUE", t)
			} else {
				c.Props.SetDateTime("DUE", t.UTC())
			}
		}
		if e.Completed {
			c.Props.SetText("STATUS", "COMPLETED")
			setInteger(c, "PERCENT-COMPLETE", 100)
			c.Props.SetDateTime("COMPLETED", time.Now().UTC())
		} else {
			c.Props.SetText("STATUS", "NEEDS-ACTION")
			setInteger(c, "PERCENT-COMPLETE", 0)
			c.Props.Del("COMPLETED")
		}
	}
	var err error
	r.Data, err = encode(cal)
	if err != nil {
		return err
	}
	next.Pending[r.URL] = Change{Resource: r}
	if e.CaptureID != "" {
		if next.Captures == nil {
			next.Captures = map[string]bool{}
		}
		next.Captures[e.CaptureID] = true
	}
	return s.commit(next)
}
func sameCalendar(a, b string) bool {
	ca, ea := decode(a)
	cb, eb := decode(b)
	if ea != nil || eb != nil {
		return false
	}
	ja, _ := json.Marshal(ca)
	jb, _ := json.Marshal(cb)
	return bytes.Equal(ja, jb)
}

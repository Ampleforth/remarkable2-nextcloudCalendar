// SPDX-License-Identifier: GPL-3.0-only
package main

import (
	"os"
	"testing"
)

func TestColorDisplayDetection(t *testing.T) {
	for _, tc := range []struct {
		model string
		color bool
	}{
		{"reMarkable Ferrari\n", true}, {"reMarkable Chiappa\x00", true},
		{"reMarkable 2.0", false}, {"reMarkable 1.0", false}, {"reMarkable Paper Pure", false},
		{"desktop", false}, {"", false}, {"new color device", false},
	} {
		if got := colorModel(tc.model); got != tc.color {
			t.Errorf("%q: got %v", tc.model, got)
		}
	}
	read := func(path string) ([]byte, error) {
		if path == "/proc/device-tree/model" {
			return []byte("reMarkable Chiappa\x00"), nil
		}
		return nil, os.ErrNotExist
	}
	if !detectColorDisplay(read) {
		t.Fatal("model fallback failed")
	}
	if detectColorDisplay(func(string) ([]byte, error) { return nil, os.ErrNotExist }) {
		t.Fatal("missing device identity enabled color")
	}
}

func TestCalendarColorsPersistAndRejectInvalidValues(t *testing.T) {
	s, c := testStore(t)
	for _, bad := range []string{"blue", "#123", "#12345678", "#badhex", "url(file:///private)"} {
		if err := s.setCalendarColor(c.ID, bad); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
	if err := s.setCalendarColor("missing", "#0082c9"); err == nil {
		t.Fatal("unknown calendar accepted")
	}
	if err := s.setCalendarColor(c.ID, "#7952B3"); err != nil {
		t.Fatal(err)
	}
	reopened, err := openStore(s.path)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.state.CalendarColors[c.ID] != "#7952b3" {
		t.Fatal("color did not persist")
	}
	s.busy = true
	if err = s.setCalendarColor(c.ID, "#0082c9"); err == nil {
		t.Fatal("color update during sync accepted")
	}
	s.busy = false
	if err = os.Remove(s.path); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(s.path, 0700); err != nil {
		t.Fatal(err)
	}
	if err = s.setCalendarColor(c.ID, "#0082c9"); err == nil {
		t.Fatal("failed persistence acknowledged")
	}
	if s.state.CalendarColors[c.ID] != "#7952b3" {
		t.Fatal("failed save changed color")
	}
}

func TestCalendarColorsClearOnAccountSwitch(t *testing.T) {
	s, c := testStore(t)
	s.state.Account = "previous"
	if err := s.setCalendarColor(c.ID, "#0082c9"); err != nil {
		t.Fatal(err)
	}
	if err := s.bindAccountLocked(Config{Server: "https://other.example", Username: "test"}); err != nil {
		t.Fatal(err)
	}
	if len(s.state.CalendarColors) != 0 {
		t.Fatal("previous account colors retained")
	}
}

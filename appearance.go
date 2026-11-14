// SPDX-License-Identifier: GPL-3.0-only
package main

import (
	"errors"
	"os"
	"regexp"
	"strings"
)

func colorModel(model string) bool {
	switch strings.ToLower(strings.TrimSpace(strings.TrimRight(model, "\x00"))) {
	case "remarkable ferrari", "remarkable chiappa":
		return true
	default:
		return false
	}
}

func detectColorDisplay(read func(string) ([]byte, error)) bool {
	for _, path := range []string{"/sys/devices/soc0/machine", "/proc/device-tree/model", "/sys/firmware/devicetree/base/model"} {
		b, err := read(path)
		if err == nil && strings.TrimSpace(strings.TrimRight(string(b), "\x00")) != "" {
			return colorModel(string(b))
		}
	}
	return false
}

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func (s *Store) setCalendarColor(id, color string) error {
	if !hexColor.MatchString(color) {
		return errors.New("choose a color as #RRGGBB")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.busy {
		return errors.New("wait for sync to finish")
	}
	found := false
	for _, c := range s.state.Collections {
		if c.ID == id {
			found = true
			break
		}
	}
	if !found {
		return errors.New("calendar not found")
	}
	next := s.clone()
	if next.CalendarColors == nil {
		next.CalendarColors = map[string]string{}
	}
	next.CalendarColors[id] = strings.ToLower(color)
	return s.commit(next)
}

func nativeColorDisplay() bool { return detectColorDisplay(os.ReadFile) }

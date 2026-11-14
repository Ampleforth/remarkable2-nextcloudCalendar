// SPDX-License-Identifier: GPL-3.0-only
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Capture struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

// The spool is independent of the account: nothing is uploaded until the user
// explicitly chooses a destination in the ordinary task editor and saves it.
func (s *Store) captures() ([]Capture, error) {
	dir := filepath.Join(filepath.Dir(s.path), "intake")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	saved := s.snapshot().Captures
	result := []Capture{}
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		if info.Size() > 128*1024 {
			return nil, fmt.Errorf("capture %s is too large", entry.Name())
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(append([]byte(entry.Name()+"\x00"), b...))
		id := hex.EncodeToString(sum[:])
		if saved[id] {
			continue
		}
		var v struct {
			Summary  string `json:"summary"`
			Document string `json:"source_doc_uuid"`
			Page     string `json:"source_page_key"`
			Label    string `json:"source_label"`
			Captured string `json:"captured_at"`
		}
		if err := json.Unmarshal(b, &v); err != nil {
			return nil, fmt.Errorf("capture %s is incomplete or invalid; retry after capture finishes", entry.Name())
		}
		if strings.TrimSpace(v.Summary) == "" {
			return nil, fmt.Errorf("capture %s has no task text", entry.Name())
		}
		result = append(result, Capture{ID: id, Summary: v.Summary, Description: fmt.Sprintf("Captured from %s\nNotebook: %s\nPage: %s\nCaptured: %s", v.Label, v.Document, v.Page, v.Captured)})
	}
	return result, nil
}

// Consume a capture without creating or deleting any calendar resource. The same
// durable ledger used by Save keeps discarded captures hidden after restart.
func (s *Store) discardCapture(id string) error {
	b, err := hex.DecodeString(id)
	if err != nil || len(b) != sha256.Size {
		return errors.New("invalid capture identifier")
	}
	captures, err := s.captures()
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.busy {
		return errors.New("wait for sync to finish")
	}
	if s.state.Captures[id] {
		return nil
	}
	found := false
	for _, capture := range captures {
		if capture.ID == id {
			found = true
			break
		}
	}
	if !found {
		return errors.New("capture no longer available")
	}
	next := s.clone()
	if next.Captures == nil {
		next.Captures = map[string]bool{}
	}
	next.Captures[id] = true
	return s.commit(next)
}

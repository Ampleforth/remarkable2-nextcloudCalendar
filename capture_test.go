// SPDX-License-Identifier: GPL-3.0-only
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCaptureReviewAndDurableImport(t *testing.T) {
	s, c := testStore(t)
	if _, err := s.captures(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(filepath.Dir(s.path), "intake", "cap_123.json")
	if err := os.WriteFile(path, []byte(`{"summary":"Buy milk","source_doc_uuid":"notebook-id","source_page_key":"idx:2","source_label":"Shopping · p.3"}`), 0600); err != nil {
		t.Fatal(err)
	}
	captures, err := s.captures()
	if err != nil || len(captures) != 1 {
		t.Fatalf("captures: %v %v", captures, err)
	}
	if len(mustItems(t, s)) != 0 {
		t.Fatal("capture must wait for review")
	}
	e := Edit{CaptureID: captures[0].ID, Item: Item{Calendar: c.ID, Kind: "VTODO", Title: captures[0].Summary, Description: captures[0].Description}}
	bad := e
	bad.Calendar = "missing"
	if err = s.queue(bad, time.UTC); err == nil {
		t.Fatal("invalid destination accepted")
	}
	captures, _ = s.captures()
	if len(captures) != 1 {
		t.Fatal("failed save lost capture")
	}
	if err = s.queue(e, time.UTC); err != nil {
		t.Fatal(err)
	}
	reopened, err := openStore(s.path)
	if err != nil {
		t.Fatal(err)
	}
	captures, err = reopened.captures()
	if err != nil || len(captures) != 0 {
		t.Fatalf("saved capture offered again: %v %v", captures, err)
	}
	if err = reopened.queue(e, time.UTC); err != nil {
		t.Fatal(err)
	}
	items := mustItems(t, reopened)
	if len(items) != 1 || items[0].Description != e.Description {
		t.Fatalf("duplicate or lost provenance: %+v", items)
	}
}

func TestCaptureMalformedFilePreserved(t *testing.T) {
	s, _ := testStore(t)
	s.captures()
	path := filepath.Join(filepath.Dir(s.path), "intake", "cap_bad.json")
	os.WriteFile(path, []byte(`{"summary":`), 0600)
	if _, err := s.captures(); err == nil {
		t.Fatal("partial file accepted")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("partial file removed")
	}
}

func TestDiscardCaptureDurableAndNoCalendarChanges(t *testing.T) {
	s, _ := testStore(t)
	if _, err := s.captures(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(filepath.Dir(s.path), "intake", "cap_discard.json")
	if err := os.WriteFile(path, []byte(`{"summary":"Discard me"}`), 0600); err != nil {
		t.Fatal(err)
	}
	captures, err := s.captures()
	if err != nil {
		t.Fatal(err)
	}
	if err = s.discardCapture(captures[0].ID); err != nil {
		t.Fatal(err)
	}
	reopened, err := openStore(s.path)
	if err != nil {
		t.Fatal(err)
	}
	remaining, err := reopened.captures()
	if err != nil || len(remaining) != 0 {
		t.Fatalf("discard not durable: %v %v", remaining, err)
	}
	if len(reopened.state.Pending) != 0 || len(reopened.state.Resources) != 0 {
		t.Fatal("discard modified calendar items")
	}
	if err = reopened.discardCapture(captures[0].ID); err != nil {
		t.Fatal(err)
	}
	if err = reopened.discardCapture("../state.json"); err == nil {
		t.Fatal("invalid ID accepted")
	}
	if err = reopened.discardCapture(strings.Repeat("0", 64)); err == nil {
		t.Fatal("unknown ID accepted")
	}
}

func TestDiscardCapturePersistenceFailurePreservesCapture(t *testing.T) {
	s, _ := testStore(t)
	s.captures()
	path := filepath.Join(filepath.Dir(s.path), "intake", "cap_retry.json")
	if err := os.WriteFile(path, []byte(`{"summary":"Keep me"}`), 0600); err != nil {
		t.Fatal(err)
	}
	captures, err := s.captures()
	if err != nil {
		t.Fatal(err)
	}
	// A directory at the destination prevents atomic replacement of the state.
	if err = os.Mkdir(s.path, 0700); err != nil {
		t.Fatal(err)
	}
	if err = s.discardCapture(captures[0].ID); err == nil {
		t.Fatal("failed persistence acknowledged")
	}
	remaining, err := s.captures()
	if err != nil || len(remaining) != 1 {
		t.Fatalf("capture lost: %v %v", remaining, err)
	}
}

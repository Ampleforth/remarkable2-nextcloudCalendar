// SPDX-License-Identifier: GPL-3.0-only
package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLauncherPreferencePersistsIndependently(t *testing.T) {
	s, _ := testStore(t)
	value, err := s.launcherSettings()
	if err != nil || !value.Enabled {
		t.Fatalf("default: %+v %v", value, err)
	}
	if err = s.saveLauncher(false); err != nil {
		t.Fatal(err)
	}
	reopened, err := openStore(s.path)
	if err != nil {
		t.Fatal(err)
	}
	value, err = reopened.launcherSettings()
	if err != nil || value.Enabled {
		t.Fatalf("disabled preference lost: %+v %v", value, err)
	}
	info, err := os.Stat(s.launcherPath())
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatal("preference file should be private")
	}
	if err = reopened.saveLauncher(true); err != nil {
		t.Fatal(err)
	}
	value, err = s.launcherSettings()
	if err != nil || !value.Enabled {
		t.Fatalf("enable: %+v %v", value, err)
	}
}

func TestLauncherAPIValidationAndFailure(t *testing.T) {
	s, _ := testStore(t)
	handler := api(s, filepath.Join(t.TempDir(), "missing-config.json"), time.UTC)
	for _, tc := range []struct {
		body string
		code int
	}{
		{`{}`, 400}, {`{"enabled":null}`, 400}, {`{"enabled":"no"}`, 400}, {`{"enabled":false}`, 200},
	} {
		req := httptest.NewRequest("POST", "/launcher", strings.NewReader(tc.body))
		req.Header.Set("X-RemarkableCalendar-Request", "sync")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != tc.code {
			t.Fatalf("%s: status %d", tc.body, response.Code)
		}
	}
	value, err := s.launcherSettings()
	if err != nil || value.Enabled {
		t.Fatalf("false not saved: %+v %v", value, err)
	}
	if err = os.Remove(s.launcherPath()); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(s.launcherPath(), 0700); err != nil {
		t.Fatal(err)
	}
	if err = s.saveLauncher(true); err == nil {
		t.Fatal("failed write acknowledged")
	}
}

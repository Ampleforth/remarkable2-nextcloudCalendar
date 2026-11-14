// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
	_ "time/tzdata"
)

func api(s *Store, configPath string, loc *time.Location) http.Handler {
	mux := http.NewServeMux()
	respond := func(w http.ResponseWriter, v any, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if err != nil {
			w.WriteHeader(400)
			v = map[string]string{"error": err.Error()}
		}
		_ = json.NewEncoder(w).Encode(v)
	}
	decodeBody := func(w http.ResponseWriter, r *http.Request, v any) error {
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128*1024))
		d.DisallowUnknownFields()
		return d.Decode(v)
	}
	mux.HandleFunc("POST /colors", func(w http.ResponseWriter, r *http.Request) {
		var v struct {
			ID    string `json:"id"`
			Color string `json:"color"`
		}
		err := decodeBody(w, r, &v)
		if err == nil {
			err = s.setCalendarColor(v.ID, v.Color)
		}
		respond(w, s.status(), err)
	})
	mux.HandleFunc("GET /launcher", func(w http.ResponseWriter, r *http.Request) {
		v, err := s.launcherSettings()
		respond(w, v, err)
	})
	mux.HandleFunc("POST /launcher", func(w http.ResponseWriter, r *http.Request) {
		var v struct {
			Enabled *bool `json:"enabled"`
		}
		err := decodeBody(w, r, &v)
		if err == nil && v.Enabled == nil {
			err = errors.New("enabled is required")
		}
		if err == nil {
			err = s.saveLauncher(*v.Enabled)
		}
		respond(w, map[string]bool{"ok": true}, err)
	})
	mux.HandleFunc("GET /captures", func(w http.ResponseWriter, r *http.Request) {
		v, err := s.captures()
		respond(w, map[string]any{"captures": v}, err)
	})
	mux.HandleFunc("POST /captures/discard", func(w http.ResponseWriter, r *http.Request) {
		var v struct {
			ID string `json:"id"`
		}
		err := decodeBody(w, r, &v)
		if err == nil {
			err = s.discardCapture(v.ID)
		}
		respond(w, map[string]bool{"ok": true}, err)
	})
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) { respond(w, s.status(), nil) })
	mux.HandleFunc("GET /connection", func(w http.ResponseWriter, r *http.Request) { respond(w, connectionInfo(configPath), nil) })
	mux.HandleFunc("POST /connection", func(w http.ResponseWriter, r *http.Request) {
		var c Config
		err := decodeBody(w, r, &c)
		if err == nil {
			err = s.saveConnection(configPath, c, loc.String())
		}
		respond(w, connectionInfo(configPath), err)
	})
	mux.HandleFunc("POST /preferences", func(w http.ResponseWriter, r *http.Request) {
		var p struct {
			TimeFormat string `json:"time_format"`
			TaskSort   string `json:"task_sort"`
		}
		err := decodeBody(w, r, &p)
		if err == nil {
			err = s.preferences(p.TimeFormat, p.TaskSort)
		}
		respond(w, s.status(), err)
	})
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		a, e := time.ParseInLocation("2006-01-02", r.URL.Query().Get("start"), loc)
		if e != nil {
			respond(w, nil, e)
			return
		}
		b, e := time.ParseInLocation("2006-01-02", r.URL.Query().Get("end"), loc)
		if e != nil || !b.After(a) || b.Sub(a) > 370*24*time.Hour {
			respond(w, nil, errors.New("choose a date range of at most one year"))
			return
		}
		items, e := s.items(a, b, loc)
		respond(w, map[string]any{"items": items}, e)
	})
	mux.HandleFunc("POST /sync", func(w http.ResponseWriter, r *http.Request) {
		respond(w, map[string]bool{"ok": true}, s.startSync(configPath))
	})
	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		var e Edit
		err := decodeBody(w, r, &e)
		if err == nil {
			err = s.queue(e, loc)
		}
		respond(w, map[string]bool{"ok": true}, err)
	})
	mux.HandleFunc("POST /pending", func(w http.ResponseWriter, r *http.Request) {
		var v struct {
			ID     string `json:"id"`
			Action string `json:"action"`
		}
		err := decodeBody(w, r, &v)
		if err == nil {
			err = s.resolvePending(v.ID, v.Action)
		}
		respond(w, map[string]bool{"ok": true}, err)
	})
	mux.HandleFunc("POST /collections", func(w http.ResponseWriter, r *http.Request) {
		var v struct {
			ID      string `json:"id"`
			Enabled bool   `json:"enabled"`
		}
		err := decodeBody(w, r, &v)
		if err == nil {
			s.mu.Lock()
			defer s.mu.Unlock()
			if s.busy {
				err = errors.New("wait for sync to finish")
			} else {
				next := s.clone()
				found := false
				for n := range next.Collections {
					if next.Collections[n].ID == v.ID {
						next.Collections[n].Enabled = v.Enabled
						found = true
					}
				}
				if !found {
					err = errors.New("calendar not found")
				} else {
					err = s.commit(next)
				}
			}
		}
		respond(w, map[string]bool{"ok": true}, err)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Native QML XHR can send this header. Browser pages cannot pass a CORS preflight.
		if r.Method != "GET" && r.Header.Get("X-RemarkableCalendar-Request") != "sync" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.Header.Get("Origin") != "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if err := s.ensureAccount(configPath); err != nil {
			respond(w, nil, err)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
func lifecycle(path string) (net.Conn, error) {
	var conn net.Conn
	var err error
	for n := 0; n < 50; n++ {
		conn, err = net.Dial("unixpacket", path)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		return nil, err
	}
	go func() {
		defer os.Exit(0)
		buf := make([]byte, 8)
		for {
			n, e := conn.Read(buf)
			if e != nil || n != 8 {
				return
			}
			kind := binary.NativeEndian.Uint32(buf[:4])
			length := binary.NativeEndian.Uint32(buf[4:])
			if length > 10*1024*1024 {
				return
			}
			if length > 0 {
				b := make([]byte, length)
				n, e = conn.Read(b)
				if e != nil || uint32(n) != length {
					return
				}
			}
			if kind == ^uint32(0) {
				return
			}
		}
	}()
	return conn, nil
}
func main() {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".local/share/nextcloud-calendar")
	config := flag.String("config", filepath.Join(home, ".config/nextcloud-calendar/config.json"), "private Nextcloud config")
	cache := flag.String("cache", filepath.Join(base, "state.json"), "offline state")
	address := flag.String("listen", "127.0.0.1:38766", "loopback HTTP address")
	timezone := flag.String("timezone", "UTC", "offline display timezone")
	flag.Parse()
	if d, e := loadDAV(*config); e == nil {
		*timezone = d.config.Timezone
	}
	loc, e := time.LoadLocation(*timezone)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	host, _, e := net.SplitHostPort(*address)
	if e != nil || (host != "127.0.0.1" && host != "::1") {
		fmt.Fprintln(os.Stderr, "listen must be a loopback address")
		os.Exit(1)
	}
	s, e := openStore(*cache)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	server := &http.Server{Addr: *address, Handler: api(s, *config, loc), ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 10 * time.Second, MaxHeaderBytes: 8192}
	listener, e := net.Listen("tcp", *address)
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	defer listener.Close()
	// Bind HTTP before AppLoad starts the frontend, so its first request can connect.
	if flag.NArg() > 0 {
		conn, e := lifecycle(flag.Arg(0))
		if e != nil {
			fmt.Fprintln(os.Stderr, "AppLoad lifecycle connection failed")
			os.Exit(1)
		}
		defer conn.Close()
	}
	fmt.Fprintln(os.Stderr, "Nextcloud Calendar backend:", *address)
	if e = server.Serve(listener); e != nil && e != http.ErrServerClosed && e != io.EOF {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}

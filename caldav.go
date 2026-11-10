// SPDX-License-Identifier: GPL-3.0-only
// Copyright (C) 2026 remarkable2-nextcloudCalendar contributors
package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"app_password"`
	Timezone string `json:"timezone"`
}
type DAV struct {
	config Config
	base   *url.URL
	client *http.Client
}

func loadDAV(path string) (*DAV, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("configure Nextcloud first (%s)", path)
	}
	if info.Mode().Perm()&0077 != 0 {
		return nil, errors.New("Nextcloud config must have permissions 600")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err = json.Unmarshal(b, &c); err != nil {
		return nil, errors.New("invalid Nextcloud configuration")
	}
	return newDAV(c)
}
func newDAV(c Config) (*DAV, error) {
	u, err := url.Parse(c.Server)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("server must be an HTTPS origin or installation path")
	}
	if c.Username == "" || c.Password == "" {
		return nil, errors.New("Nextcloud username and app password required")
	}
	if _, err = time.LoadLocation(c.Timezone); err != nil {
		return nil, errors.New("choose an IANA timezone")
	}
	d := &DAV{config: c, base: u}
	d.client = &http.Client{Timeout: 45 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	return d, nil
}

// DAV hrefs are untrusted server data. Never forward credentials to another origin.
func (d *DAV) resolve(base, href string) (string, error) {
	b, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	u, err := b.Parse(href)
	if err != nil || u.Scheme != "https" || !strings.EqualFold(u.Host, d.base.Host) || u.User != nil || u.Fragment != "" || u.RawQuery != "" {
		return "", errors.New("unsafe CalDAV resource URL")
	}
	return u.String(), nil
}
func (d *DAV) request(method, target, body, etag string) ([]byte, http.Header, int, error) {
	target, err := d.resolve(d.base.String(), target)
	if err != nil {
		return nil, nil, 0, err
	}
	req, err := http.NewRequest(method, target, strings.NewReader(body))
	if err != nil {
		return nil, nil, 0, err
	}
	req.SetBasicAuth(d.config.Username, d.config.Password)
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	if method == "PROPFIND" || method == "REPORT" {
		req.Header.Set("Depth", "1")
	}
	if method == "PUT" {
		req.Header.Set("Content-Type", "text/calendar; charset=utf-8")
	}
	if etag == "*" {
		req.Header.Set("If-None-Match", "*")
	} else if etag != "" {
		req.Header.Set("If-Match", etag)
	}
	res, err := d.client.Do(req)
	if err != nil {
		return nil, nil, 0, errors.New("Nextcloud connection failed; check Wi-Fi and server address")
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, 32*1024*1024+1))
	if err != nil {
		return nil, nil, res.StatusCode, err
	}
	if len(b) > 32*1024*1024 {
		return nil, nil, res.StatusCode, errors.New("CalDAV response exceeds 32 MiB")
	}
	if res.StatusCode >= 300 {
		return b, res.Header, res.StatusCode, fmt.Errorf("Nextcloud returned HTTP %d", res.StatusCode)
	}
	return b, res.Header, res.StatusCode, nil
}

type davProp struct {
	Principal struct {
		Href string `xml:"href"`
	} `xml:"current-user-principal"`
	Home struct {
		Href string `xml:"href"`
	} `xml:"calendar-home-set"`
	Name  string `xml:"displayname"`
	Types struct {
		Calendar *struct{} `xml:"calendar"`
	} `xml:"resourcetype"`
	Components struct {
		Values []struct {
			Name string `xml:"name,attr"`
		} `xml:"comp"`
	} `xml:"supported-calendar-component-set"`
	Privileges struct {
		Values []struct {
			Write   *struct{} `xml:"write"`
			All     *struct{} `xml:"all"`
			Content *struct{} `xml:"write-content"`
			Bind    *struct{} `xml:"bind"`
			Unbind  *struct{} `xml:"unbind"`
		} `xml:"privilege"`
	} `xml:"current-user-privilege-set"`
	ETag string `xml:"getetag"`
	Data string `xml:"calendar-data"`
}
type davResponse struct {
	Href   string `xml:"href"`
	Status string `xml:"status"`
	Props  []struct {
		Status string  `xml:"status"`
		Prop   davProp `xml:"prop"`
	} `xml:"propstat"`
}

func parseDAV(b []byte) ([]davResponse, error) {
	var v struct {
		XMLName   xml.Name      `xml:"DAV: multistatus"`
		Responses []davResponse `xml:"response"`
	}
	err := xml.Unmarshal(b, &v)
	return v.Responses, err
}
func propfindBody(props string) string {
	return `<?xml version="1.0"?><d:propfind xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav"><d:prop>` + props + `</d:prop></d:propfind>`
}
func (d *DAV) discover() ([]Collection, error) {
	endpoint := strings.TrimRight(d.base.String(), "/") + "/remote.php/dav/"
	findHref := func(target, props string, home bool) (string, error) {
		b, _, _, err := d.request("PROPFIND", target, propfindBody(props), "")
		if err != nil {
			return "", err
		}
		rows, err := parseDAV(b)
		if err != nil {
			return "", err
		}
		for _, r := range rows {
			for _, ps := range r.Props {
				if !strings.Contains(ps.Status, " 200 ") {
					continue
				}
				h := ps.Prop.Principal.Href
				if home {
					h = ps.Prop.Home.Href
				}
				if h != "" {
					return d.resolve(target, h)
				}
			}
		}
		return "", errors.New("CalDAV discovery property missing")
	}
	principal, err := findHref(endpoint, "<d:current-user-principal/>", false)
	if err != nil {
		return nil, err
	}
	home, err := findHref(principal, "<c:calendar-home-set/>", true)
	if err != nil {
		return nil, err
	}
	b, _, _, err := d.request("PROPFIND", home, propfindBody("<d:displayname/><d:resourcetype/><d:current-user-privilege-set/><c:supported-calendar-component-set/>"), "")
	if err != nil {
		return nil, err
	}
	rows, err := parseDAV(b)
	if err != nil {
		return nil, err
	}
	out := []Collection{}
	for _, r := range rows {
		c := Collection{Enabled: true}
		calendar := false
		for _, ps := range r.Props {
			if !strings.Contains(ps.Status, " 200 ") {
				continue
			}
			p := ps.Prop
			if p.Types.Calendar != nil {
				calendar = true
			}
			if p.Name != "" {
				c.Name = p.Name
			}
			for _, component := range p.Components.Values {
				if component.Name == "VEVENT" {
					c.Events = true
				}
				if component.Name == "VTODO" {
					c.Tasks = true
				}
			}
			write, bind, unbind := false, false, false
			for _, priv := range p.Privileges.Values {
				if priv.All != nil || priv.Write != nil {
					write = true
					bind = true
					unbind = true
				}
				write = write || priv.Content != nil
				bind = bind || priv.Bind != nil
				unbind = unbind || priv.Unbind != nil
			}
			c.Writable = c.Writable || (write && bind && unbind)
		}
		if !calendar {
			continue
		}
		c.ID, err = d.resolve(home, r.Href)
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(c.ID, "/") {
			c.ID += "/"
		}
		if c.Name == "" {
			c.Name = c.ID
		}
		out = append(out, c)
	}
	return out, nil
}
func (d *DAV) fetch(c Collection, start, end time.Time) (map[string]Resource, error) {
	out := map[string]Resource{}
	for _, kind := range []string{"VEVENT", "VTODO"} {
		if kind == "VEVENT" && !c.Events || kind == "VTODO" && !c.Tasks {
			continue
		}
		filter := ""
		if kind == "VEVENT" {
			filter = fmt.Sprintf(`<c:time-range start="%s" end="%s"/>`, start.UTC().Format("20060102T150405Z"), end.UTC().Format("20060102T150405Z"))
		}
		body := `<c:calendar-query xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav"><d:prop><d:getetag/><c:calendar-data/></d:prop><c:filter><c:comp-filter name="VCALENDAR"><c:comp-filter name="` + kind + `">` + filter + `</c:comp-filter></c:comp-filter></c:filter></c:calendar-query>`
		b, _, _, err := d.request("REPORT", c.ID, body, "")
		if err != nil {
			return nil, err
		}
		rows, err := parseDAV(b)
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			var etag, data string
			for _, ps := range r.Props {
				if !strings.Contains(ps.Status, " 200 ") {
					return nil, errors.New("incomplete CalDAV query; cache retained")
				}
				if ps.Prop.ETag != "" {
					etag = ps.Prop.ETag
				}
				if ps.Prop.Data != "" {
					data = ps.Prop.Data
				}
			}
			if data == "" || etag == "" {
				return nil, errors.New("CalDAV resource missing data or ETag; cache retained")
			}
			target, err := d.resolve(c.ID, r.Href)
			if err != nil {
				return nil, err
			}
			if !strings.HasPrefix(target, c.ID) {
				return nil, errors.New("resource outside calendar collection")
			}
			out[target] = Resource{target, c.ID, etag, data}
		}
	}
	return out, nil
}

package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const tinySVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"><rect width="1" height="1"/></svg>`

func TestFaviconNeedsASignedInUser(t *testing.T) {
	e := newEnv(t)
	mustStatus(t, e.browser().do("GET", "/api/favicon?url=a.lan", nil), 401)
}

func TestFaviconIsFoundFromThePageOrTheDefaultPath(t *testing.T) {
	site := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Write([]byte(`<html><head><link rel="shortcut icon" href="/static/logo.svg"></head></html>`))
		case "/static/logo.svg":
			w.Header().Set("Content-Type", "image/svg+xml")
			w.Write([]byte(tinySVG))
		default:
			http.NotFound(w, r)
		}
	}))
	defer site.Close()
	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/favicon.ico" {
			w.Write([]byte{0, 0, 1, 0, 1, 0, 1, 1, 0, 0, 1, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0})
			return
		}
		w.Write([]byte("<html>no icon links</html>"))
	}))
	defer plain.Close()
	none := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html>nothing here</html>")) // even /favicon.ico answers with HTML
	}))
	defer none.Close()

	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")
	get := func(u string) resp { return b.do("GET", "/api/favicon?url="+url.QueryEscape(u), nil) }

	r := get(site.URL)
	mustStatus(t, r, 200)
	if r.Header().Get("Content-Type") != "image/svg+xml" || !strings.HasPrefix(r.Body.String(), "<svg") {
		t.Errorf("linked icon: %v %q", r.Header(), r.Body.String())
	}
	if !strings.Contains(r.Header().Get("Content-Security-Policy"), "sandbox") {
		t.Error("icon is not sandboxed")
	}

	r = get(plain.URL)
	mustStatus(t, r, 200)
	if r.Header().Get("Content-Type") != "image/x-icon" {
		t.Errorf("default path icon: %v", r.Header())
	}

	mustStatus(t, get(none.URL), 404) // HTML is never passed off as an icon
	mustStatus(t, b.do("GET", "/api/favicon?url=javascript:alert(1)", nil), 400)
}

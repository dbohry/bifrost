package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dbohry/bifrost/internal/status"
	"github.com/dbohry/bifrost/internal/store"
)

func (r resp) statuses() map[string]any {
	m, _ := r.json()["statuses"].(map[string]any)
	return m
}

func TestStatusNeedsASignedInUser(t *testing.T) {
	e := newEnv(t)
	anon := e.browser()
	mustStatus(t, anon.do("GET", "/api/status", nil), 401)
	mustStatus(t, anon.do("GET", "/api/status/settings", nil), 401)
	mustStatus(t, anon.do("PUT", "/api/status/settings", m{"interval": 60, "timeout": 5}), 401)
	mustStatus(t, anon.do("POST", "/api/status/test", m{"url": "a.lan"}), 401)
}

func TestAppsRememberTheirStatusCheck(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")

	// off by default, with the app's own address
	r := b.do("POST", "/api/apps", m{"name": "Plain", "url": "plain.lan"})
	mustStatus(t, r, 201)
	if a := r.app(); a["checkEnabled"] != false || a["checkUrl"] != "" || a["checkStrict"] != false || a["checkInsecure"] != false {
		t.Errorf("defaults = %v", a)
	}

	r = b.do("POST", "/api/apps", m{"name": "Nas", "url": "nas.example.com", "checkEnabled": true, "checkUrl": "192.168.1.20:8000", "checkStrict": true, "checkInsecure": true})
	mustStatus(t, r, 201)
	a := r.app()
	if a["checkEnabled"] != true || a["checkUrl"] != "http://192.168.1.20:8000" || a["checkStrict"] != true || a["checkInsecure"] != true {
		t.Errorf("saved = %v", a)
	}

	// turning it off keeps the rest and it survives a reload
	id := int64(a["id"].(float64))
	r = b.do("PUT", "/api/apps/"+strconv.FormatInt(id, 10), m{"name": "Nas", "url": "nas.example.com", "checkEnabled": false, "checkUrl": "192.168.1.20:8000"})
	mustStatus(t, r, 200)
	got := b.do("GET", "/api/apps", nil).apps()[1].(map[string]any)
	if got["checkEnabled"] != false || got["checkUrl"] != "http://192.168.1.20:8000" {
		t.Errorf("after update = %v", got)
	}
}

func TestTheAddressToCheckIsValidatedLikeAnyAddress(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")
	for _, bad := range []string{"javascript:alert(1)", "ftp://x.lan", "http://user:pw@x.lan", "http://has space"} {
		r := b.do("POST", "/api/apps", m{"name": "A", "url": "a.lan", "checkEnabled": true, "checkUrl": bad})
		mustStatus(t, r, 400)
		if !strings.Contains(r.json()["error"].(string), "address to check") {
			t.Errorf("%q: %v", bad, r.json()["error"])
		}
	}
}

func TestStatusSettings(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")

	r := b.do("GET", "/api/status/settings", nil)
	mustStatus(t, r, 200)
	if s := r.json()["settings"].(map[string]any); s["interval"] != float64(60) || s["timeout"] != float64(5) {
		t.Errorf("defaults = %v", s)
	}

	mustStatus(t, b.do("PUT", "/api/status/settings", m{"interval": 300, "timeout": 10}), 200)
	if s := b.do("GET", "/api/status/settings", nil).json()["settings"].(map[string]any); s["interval"] != float64(300) || s["timeout"] != float64(10) {
		t.Errorf("saved = %v", s)
	}

	for name, body := range map[string]m{
		"interval too short": {"interval": 9, "timeout": 5},
		"interval too long":  {"interval": 86401, "timeout": 5},
		"timeout zero":       {"interval": 60, "timeout": 0},
		"timeout too long":   {"interval": 600, "timeout": 61},
		"timeout > interval": {"interval": 10, "timeout": 30},
	} {
		if r := b.do("PUT", "/api/status/settings", body); r.Code != 400 {
			t.Errorf("%s: status %d", name, r.Code)
		}
	}
	// custom values at the edges are fine
	mustStatus(t, b.do("PUT", "/api/status/settings", m{"interval": 10, "timeout": 10}), 200)
	mustStatus(t, b.do("PUT", "/api/status/settings", m{"interval": 86400, "timeout": 60}), 200)
}

func TestStatusResultsOnlyShowTheUsersOwnApps(t *testing.T) {
	e := newEnv(t)
	alex := e.browser()
	alex.setup("alex", "correct-password")

	var answer status.Outcome
	e.s.monitor = status.NewMonitor(e.s.store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	e.s.monitor.SetCheck(func(context.Context, store.CheckTarget) status.Outcome { return answer })

	answer = status.Outcome{Up: true, Latency: 12 * time.Millisecond, Code: 200}
	r := alex.do("POST", "/api/apps", m{"name": "Jelly", "url": "jelly.lan", "checkEnabled": true})
	mustStatus(t, r, 201)
	id := strconv.FormatInt(int64(r.app()["id"].(float64)), 10)
	mustStatus(t, alex.do("POST", "/api/apps", m{"name": "Unchecked", "url": "u.lan"}), 201)

	// nothing yet: the first look has not happened
	if got := alex.do("GET", "/api/status", nil).statuses(); len(got) != 0 {
		t.Errorf("before any look: %v", got)
	}
	e.s.monitor.Step(context.Background())
	e.s.monitor.Wait()

	got := alex.do("GET", "/api/status", nil).statuses()
	res, ok := got[id].(map[string]any)
	if len(got) != 1 || !ok || res["state"] != "up" || res["latencyMs"] != float64(12) || res["address"] != "http://jelly.lan" {
		t.Fatalf("statuses = %v", got)
	}

	// another user sees nothing of it
	made := alex.do("POST", "/api/users", m{"username": "sam", "isAdmin": false})
	mustStatus(t, made, 201)
	temp := made.json()["temporaryPassword"].(string)
	sam := e.browser()
	sam.login("sam", temp, false)
	mustStatus(t, sam.do("POST", "/api/me/password", m{"current": temp, "new": "sams-password-1"}), 204)
	sam = e.browser()
	sam.login("sam", "sams-password-1", false)
	if got := sam.do("GET", "/api/status", nil).statuses(); len(got) != 0 {
		t.Errorf("Sam sees %v", got)
	}

	// turning the check off removes the result
	mustStatus(t, alex.do("PUT", "/api/apps/"+id, m{"name": "Jelly", "url": "jelly.lan", "checkEnabled": false}), 200)
	e.s.monitor.Step(context.Background())
	if got := alex.do("GET", "/api/status", nil).statuses(); len(got) != 0 {
		t.Errorf("after turning it off: %v", got)
	}
}

func TestTestNowLooksAtAnAddress(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
	t.Cleanup(up.Close)

	r := b.do("POST", "/api/status/test", m{"url": up.URL})
	mustStatus(t, r, 200)
	if j := r.json(); j["up"] != true || j["code"] != float64(503) {
		t.Errorf("any answer = %v", j)
	}
	r = b.do("POST", "/api/status/test", m{"url": up.URL, "strict": true})
	if j := r.json(); j["up"] != false || j["reason"] != "Answered with HTTP 503" {
		t.Errorf("strict = %v", j)
	}
	mustStatus(t, b.do("POST", "/api/status/test", m{"url": "javascript:alert(1)"}), 400)
	mustStatus(t, b.do("POST", "/api/status/test", m{"url": ""}), 400)
}

func TestTestNowIsRateLimited(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	limited := false
	for i := 0; i < 40 && !limited; i++ {
		limited = b.do("POST", "/api/status/test", m{"url": srv.URL}).Code == 429
	}
	if !limited {
		t.Error("the test button was never limited")
	}
}

func TestStatusChecksTravelInAnExport(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")
	mustStatus(t, b.do("POST", "/api/apps", m{"name": "Nas", "url": "nas.lan", "checkEnabled": true, "checkUrl": "10.0.0.5:8000", "checkStrict": true}), 201)

	exp := b.do("GET", "/api/export", nil)
	var f m
	if err := json.Unmarshal(exp.Body.Bytes(), &f); err != nil {
		t.Fatal(err)
	}
	f["mode"] = "replace"
	mustStatus(t, b.do("POST", "/api/import", f), 200)
	a := b.do("GET", "/api/apps", nil).apps()[0].(map[string]any)
	if a["checkEnabled"] != true || a["checkUrl"] != "http://10.0.0.5:8000" || a["checkStrict"] != true || a["checkInsecure"] != false {
		t.Errorf("after a round trip = %v", a)
	}
}

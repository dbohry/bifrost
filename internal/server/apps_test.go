package server

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/dbohry/bifrost/internal/auth"
)

func TestCleanURL(t *testing.T) {
	ok := map[string]string{
		"http://jellyfin.lan:8096":         "http://jellyfin.lan:8096",
		"https://example.com/path?q=1#top": "https://example.com/path?q=1#top",
		"  https://example.com  ":          "https://example.com",
		"jellyfin.lan:8096":                "http://jellyfin.lan:8096",
		"192.168.1.10":                     "http://192.168.1.10",
		"192.168.1.10:9000/admin":          "http://192.168.1.10:9000/admin",
		"[::1]:8080":                       "http://[::1]:8080",
		"HTTP://Example.COM":               "http://Example.COM",
	}
	for in, want := range ok {
		if got, msg := cleanURL(in); msg != "" || got != want {
			t.Errorf("cleanURL(%q) = %q, %q; want %q", in, got, msg, want)
		}
	}

	bad := []string{
		"", "   ",
		"javascript:alert(1)",
		"JaVaScRiPt://x",
		"data:text/html,<script>alert(1)</script>",
		"ftp://files.lan",
		"file:///etc/passwd",
		"http://",
		"https:///nohost",
		"http://user:pass@host.lan",
		"http://has space.lan",
		"http://host.lan/\x00",
		"http://exa\nmple.com",
		"http://[bad",
		"//host.lan",
		strings.Repeat("a", maxURLBytes+1),
	}
	for _, in := range bad {
		if got, msg := cleanURL(in); msg == "" {
			t.Errorf("cleanURL(%q) = %q, want it rejected", in, got)
		}
	}
}

func TestCleanApp(t *testing.T) {
	yes, no := true, false

	in, msg := cleanApp(appBody{Name: "  Jellyfin ", URL: "jf.lan", Category: "  Home   media ", Favourite: true})
	if msg != "" {
		t.Fatal(msg)
	}
	if in.Name != "Jellyfin" || in.URL != "http://jf.lan" || in.Category != "Home media" || !in.Favourite || !in.NewTab {
		t.Errorf("in = %+v", in)
	}

	if in, _ = cleanApp(appBody{Name: "a", URL: "a.lan", NewTab: &no}); in.NewTab || in.Category != defaultCategory {
		t.Errorf("explicit newTab=false and empty category: %+v", in)
	}
	if in, _ = cleanApp(appBody{Name: "a", URL: "a.lan", NewTab: &yes}); !in.NewTab {
		t.Errorf("newTab=true: %+v", in)
	}

	bad := map[string]appBody{
		"empty name":       {Name: "  ", URL: "a.lan"},
		"long name":        {Name: strings.Repeat("x", 41), URL: "a.lan"},
		"control in name":  {Name: "a\x07b", URL: "a.lan"},
		"empty url":        {Name: "a"},
		"long category":    {Name: "a", URL: "a.lan", Category: strings.Repeat("x", 41)},
		"control category": {Name: "a", URL: "a.lan", Category: "a\x00"},
	}
	for name, body := range bad {
		if _, msg := cleanApp(body); msg == "" {
			t.Errorf("%s: expected an error", name)
		}
	}

	// 40 characters is fine, counted as characters rather than bytes.
	if _, msg := cleanApp(appBody{Name: strings.Repeat("é", 40), URL: "a.lan"}); msg != "" {
		t.Errorf("40 accented characters rejected: %s", msg)
	}
}

func (r resp) app() map[string]any {
	a, _ := r.json()["app"].(map[string]any)
	return a
}

func (r resp) apps() []any {
	a, _ := r.json()["apps"].([]any)
	return a
}

func TestAppsNeedASignedInUser(t *testing.T) {
	e := newEnv(t)
	anon := e.browser()
	mustStatus(t, anon.do("GET", "/api/apps", nil), 401)
	mustStatus(t, anon.do("POST", "/api/apps", m{"name": "a", "url": "a.lan"}), 401)
	mustStatus(t, anon.do("PUT", "/api/apps/1", m{"name": "a", "url": "a.lan"}), 401)
	mustStatus(t, anon.do("DELETE", "/api/apps/1", nil), 401)
}

func TestAppsAreBlockedUntilTheTemporaryPasswordIsChanged(t *testing.T) {
	e := newEnv(t)
	e.browser().setup("alex", "correct-password")
	temp, _ := auth.ResetPassword(context.Background(), e.s.store, "alex")

	b := e.browser()
	b.login("alex", temp, false)
	r := b.do("GET", "/api/apps", nil)
	mustStatus(t, r, 403)
	if r.json()["code"] != "password_change_required" {
		t.Errorf("body = %s", r.Body)
	}
	mustStatus(t, b.do("POST", "/api/apps", m{"name": "a", "url": "a.lan"}), 403)
}

func TestAppLifecycle(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")

	r := b.do("GET", "/api/apps", nil)
	mustStatus(t, r, 200)
	if len(r.apps()) != 0 || r.Body.String() == "" || !strings.Contains(r.Body.String(), `"apps":[]`) {
		t.Errorf("an empty list must be [] not null: %s", r.Body)
	}

	r = b.do("POST", "/api/apps", m{"name": "Jellyfin", "url": "jellyfin.lan:8096", "category": "Media", "favourite": true})
	mustStatus(t, r, 201)
	app := r.app()
	if app["name"] != "Jellyfin" || app["url"] != "http://jellyfin.lan:8096" || app["category"] != "Media" ||
		app["favourite"] != true || app["newTab"] != true || app["position"] != float64(1) {
		t.Errorf("app = %v", app)
	}
	id := fmt.Sprint(int64(app["id"].(float64)))

	mustStatus(t, b.do("POST", "/api/apps", m{"name": "Sonarr", "url": "sonarr.lan", "category": "media", "newTab": false}), 201)
	r = b.do("GET", "/api/apps", nil)
	if list := r.apps(); len(list) != 2 {
		t.Fatalf("apps = %v", list)
	} else if second := list[1].(map[string]any); second["category"] != "Media" || second["newTab"] != false {
		t.Errorf("second app = %v (category case should match the first)", second)
	}

	r = b.do("PUT", "/api/apps/"+id, m{"name": "Jellyfin 2", "url": "https://jf.lan", "category": "Video", "favourite": false, "newTab": true})
	mustStatus(t, r, 200)
	if app = r.app(); app["name"] != "Jellyfin 2" || app["category"] != "Video" || app["favourite"] != false || app["url"] != "https://jf.lan" {
		t.Errorf("updated = %v", app)
	}

	mustStatus(t, b.do("DELETE", "/api/apps/"+id, nil), 204)
	mustStatus(t, b.do("DELETE", "/api/apps/"+id, nil), 404)
	mustStatus(t, b.do("PUT", "/api/apps/"+id, m{"name": "x", "url": "x.lan"}), 404)
	if list := b.do("GET", "/api/apps", nil).apps(); len(list) != 1 {
		t.Errorf("apps left = %v", list)
	}
}

func TestAppValidationOverHTTP(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")

	for name, body := range map[string]m{
		"no name":         {"name": "", "url": "a.lan"},
		"no url":          {"name": "a", "url": ""},
		"javascript link": {"name": "a", "url": "javascript:alert(1)"},
		"credentials":     {"name": "a", "url": "http://u:p@a.lan"},
	} {
		r := b.do("POST", "/api/apps", body)
		if r.Code != 400 || r.errMsg() == "" {
			t.Errorf("%s: status %d, body %s", name, r.Code, r.Body)
		}
	}

	req := m{"name": "a", "url": "a.lan"}
	for _, bad := range []string{"abc", "0", "-1", "1.5", "99999999999999999999"} {
		mustStatus(t, b.do("PUT", "/api/apps/"+bad, req), 404)
		mustStatus(t, b.do("DELETE", "/api/apps/"+bad, nil), 404)
	}
	if list := b.do("GET", "/api/apps", nil).apps(); len(list) != 0 {
		t.Errorf("rejected requests created apps: %v", list)
	}
}

func TestOneUsersAppsAreInvisibleToAnother(t *testing.T) {
	e := newEnv(t)
	alex := e.browser()
	alex.setup("alex", "correct-password")

	// There is no "add user" screen yet, so create the second user directly.
	hash, _ := auth.HashPassword("sam-password-1")
	if _, err := e.s.db.Exec(`INSERT INTO users (username, password_hash, created_at) VALUES ('sam', ?, 1)`, hash); err != nil {
		t.Fatal(err)
	}
	sam := e.browser()
	mustStatus(t, sam.login("sam", "sam-password-1", false), 200)

	r := alex.do("POST", "/api/apps", m{"name": "Private", "url": "private.lan"})
	mustStatus(t, r, 201)
	id := fmt.Sprint(int64(r.app()["id"].(float64)))

	if list := sam.do("GET", "/api/apps", nil).apps(); len(list) != 0 {
		t.Errorf("sam sees alex's apps: %v", list)
	}
	mustStatus(t, sam.do("PUT", "/api/apps/"+id, m{"name": "stolen", "url": "x.lan"}), 404)
	mustStatus(t, sam.do("DELETE", "/api/apps/"+id, nil), 404)

	r = alex.do("GET", "/api/apps", nil)
	if list := r.apps(); len(list) != 1 || list[0].(map[string]any)["name"] != "Private" {
		t.Errorf("alex's app was touched: %s", r.Body)
	}
}

func idsOf(r resp) []int64 {
	var out []int64
	for _, a := range r.apps() {
		out = append(out, int64(a.(map[string]any)["id"].(float64)))
	}
	return out
}

func TestReorderingOverHTTP(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")

	var made []int64
	for _, name := range []string{"one", "two", "three"} {
		r := b.do("POST", "/api/apps", m{"name": name, "url": name + ".lan", "category": "x"})
		mustStatus(t, r, 201)
		made = append(made, int64(r.app()["id"].(float64)))
	}

	mustStatus(t, b.do("PUT", "/api/apps/order", m{"ids": []int64{made[2], made[0], made[1]}}), 204)
	got := idsOf(b.do("GET", "/api/apps", nil))
	if len(got) != 3 || got[0] != made[2] || got[1] != made[0] || got[2] != made[1] {
		t.Errorf("order = %v", got)
	}

	// An out-of-date list is refused with a message, and changes nothing.
	r := b.do("PUT", "/api/apps/order", m{"ids": []int64{made[0], made[1]}})
	if r.Code != 409 || !strings.Contains(r.errMsg(), "Reload") {
		t.Errorf("stale order: %d %s", r.Code, r.Body)
	}
	mustStatus(t, b.do("PUT", "/api/apps/order", m{"ids": nil}), 409)
	if again := idsOf(b.do("GET", "/api/apps", nil)); again[0] != made[2] {
		t.Errorf("a refused order changed things: %v", again)
	}
}

func TestReorderingNeedsASignedInUserAndOwnsItsApps(t *testing.T) {
	e := newEnv(t)
	alex := e.browser()
	alex.setup("alex", "correct-password")
	r := alex.do("POST", "/api/apps", m{"name": "mine", "url": "mine.lan"})
	mine := int64(r.app()["id"].(float64))

	mustStatus(t, e.browser().do("PUT", "/api/apps/order", m{"ids": []int64{mine}}), 401)

	r = alex.do("POST", "/api/users", m{"username": "samuel"})
	temp := r.json()["temporaryPassword"].(string)
	sam := e.browser()
	sam.login("samuel", temp, false)
	sam.do("POST", "/api/me/password", m{"current": temp, "new": "samuel-pass-1"})

	// Another user cannot reorder with someone else's id, and the id "order" is not an app id.
	mustStatus(t, sam.do("PUT", "/api/apps/order", m{"ids": []int64{mine}}), 409)
	mustStatus(t, sam.do("PUT", "/api/apps/order", m{"ids": []int64{}}), 204)
	if got := idsOf(alex.do("GET", "/api/apps", nil)); len(got) != 1 || got[0] != mine {
		t.Errorf("alex's apps changed: %v", got)
	}
}

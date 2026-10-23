package server

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTheCatalogueIsListedForAnyone(t *testing.T) {
	e := newEnv(t)
	r := e.browser().do("GET", "/api/catalog", nil)
	mustStatus(t, r, 200)

	cats, _ := r.json()["categories"].([]any)
	apps, _ := r.json()["apps"].([]any)
	if len(cats) != 5 || cats[0] != "Media" || len(apps) != 30 {
		t.Fatalf("categories %v, %d apps", cats, len(apps))
	}
	first := apps[0].(map[string]any)
	if first["id"] != "jellyfin" || first["name"] != "Jellyfin" || first["category"] != "Media" || first["port"] != float64(8096) {
		t.Errorf("first app = %v", first)
	}
}

func TestEveryCatalogueIconIsServedSafely(t *testing.T) {
	e := newEnv(t)
	list := e.browser().do("GET", "/api/catalog", nil).json()["apps"].([]any)

	for _, a := range list {
		id := a.(map[string]any)["id"].(string)
		rec := e.browser().do("GET", "/api/catalog/icons/"+id+".svg", nil)
		if rec.Code != 200 {
			t.Errorf("%s: status %d", id, rec.Code)
			continue
		}
		h := rec.Header()
		if h.Get("Content-Type") != "image/svg+xml" || !strings.Contains(h.Get("Cache-Control"), "max-age") {
			t.Errorf("%s: headers %v", id, h)
		}
		// Opened directly, an icon must not be able to run anything.
		if csp := h.Get("Content-Security-Policy"); !strings.Contains(csp, "sandbox") || !strings.Contains(csp, "default-src 'none'") {
			t.Errorf("%s: Content-Security-Policy = %q", id, csp)
		}
		if h.Get("X-Content-Type-Options") != "nosniff" || !strings.HasPrefix(rec.Body.String(), "<svg") {
			t.Errorf("%s: not served as a plain SVG", id)
		}
	}
}

func TestOnlyCatalogueIconsCanBeFetched(t *testing.T) {
	e := newEnv(t)
	for _, path := range []string{
		"/api/catalog/icons/nope.svg",
		"/api/catalog/icons/jellyfin",     // no extension
		"/api/catalog/icons/jellyfin.png", // wrong extension
		"/api/catalog/icons/.svg",
		"/api/catalog/icons/..%2f..%2fcatalog.json",
		"/api/catalog/icons/%2e%2e%2fgo.mod.svg",
		"/api/catalog/icons/catalog.json",
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", path, nil)
		e.h.ServeHTTP(rec, req)
		if rec.Code != 404 {
			t.Errorf("%s: status %d, want 404", path, rec.Code)
		}
	}
}

func TestAppsCanUseACatalogueIcon(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")

	r := b.do("POST", "/api/apps", m{"name": "Jellyfin", "url": "jf.lan", "category": "Media", "icon": "jellyfin"})
	mustStatus(t, r, 201)
	if r.app()["icon"] != "jellyfin" {
		t.Errorf("app = %v", r.app())
	}
	id := idOf(r.app()["id"])

	// No icon is fine, and is what an app without one has.
	r = b.do("POST", "/api/apps", m{"name": "Custom", "url": "c.lan"})
	mustStatus(t, r, 201)
	if r.app()["icon"] != "" {
		t.Errorf("icon = %v", r.app()["icon"])
	}

	// Changing and removing the icon.
	r = b.do("PUT", "/api/apps/"+id, m{"name": "Jellyfin", "url": "jf.lan", "category": "Media", "icon": "plex"})
	mustStatus(t, r, 200)
	if r.app()["icon"] != "plex" {
		t.Errorf("after change: %v", r.app())
	}
	r = b.do("PUT", "/api/apps/"+id, m{"name": "Jellyfin", "url": "jf.lan", "category": "Media", "icon": ""})
	if r.app()["icon"] != "" {
		t.Errorf("after removal: %v", r.app())
	}

	// And it is in the list.
	b.do("PUT", "/api/apps/"+id, m{"name": "Jellyfin", "url": "jf.lan", "category": "Media", "icon": "jellyfin"})
	list := b.do("GET", "/api/apps", nil).apps()
	if list[0].(map[string]any)["icon"] != "jellyfin" {
		t.Errorf("list = %v", list)
	}
}

func TestAnIconOutsideTheCatalogueIsRefused(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")

	for _, icon := range []string{"nope", "../../etc/passwd", "https://evil.example/x.svg", "javascript:alert(1)", `"><script>`, "JELLYFIN"} {
		r := b.do("POST", "/api/apps", m{"name": "x", "url": "x.lan", "icon": icon})
		if r.Code != 400 || !strings.Contains(r.errMsg(), "icon") {
			t.Errorf("icon %q: %d %s", icon, r.Code, r.Body)
		}
	}
	if list := b.do("GET", "/api/apps", nil).apps(); len(list) != 0 {
		t.Errorf("refused requests created apps: %v", list)
	}
}

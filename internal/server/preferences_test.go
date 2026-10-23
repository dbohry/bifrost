package server

import (
	"context"
	"testing"

	"github.com/dbohry/bifrost/internal/auth"
)

func (r resp) prefs() map[string]any {
	p, _ := r.json()["preferences"].(map[string]any)
	return p
}

func TestThemesAreListedForAnyone(t *testing.T) {
	e := newEnv(t)
	r := e.browser().do("GET", "/api/themes", nil)
	mustStatus(t, r, 200)

	list, _ := r.json()["themes"].([]any)
	if len(list) != 7 {
		t.Fatalf("%d themes, want 7", len(list))
	}
	first := list[0].(map[string]any)
	colors := first["colors"].(map[string]any)
	if first["id"] != "bifrost-light" || first["mode"] != "light" || colors["onAccent"] == nil || len(colors) != 10 {
		t.Errorf("first theme = %v", first)
	}
}

func TestPreferencesDefaultsAndRoundTrip(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	mustStatus(t, b.do("GET", "/api/preferences", nil), 401)
	b.setup("alex", "correct-password")

	r := b.do("GET", "/api/preferences", nil)
	mustStatus(t, r, 200)
	want := map[string]any{"mode": "system", "lightTheme": "bifrost-light", "darkTheme": "bifrost-dark", "logo": "app"}
	for k, v := range want {
		if r.prefs()[k] != v {
			t.Errorf("default %s = %v, want %v", k, r.prefs()[k], v)
		}
	}

	save := m{"mode": "dark", "lightTheme": "solarized-light", "darkTheme": "dracula", "logo": "spectrum"}
	r = b.do("PUT", "/api/preferences", save)
	mustStatus(t, r, 200)
	for k, v := range save {
		if r.prefs()[k] != v {
			t.Errorf("saved %s = %v, want %v", k, r.prefs()[k], v)
		}
	}

	// They follow the user to another browser.
	other := e.browser()
	mustStatus(t, other.login("alex", "correct-password", false), 200)
	r = other.do("GET", "/api/preferences", nil)
	if r.prefs()["darkTheme"] != "dracula" || r.prefs()["logo"] != "spectrum" {
		t.Errorf("another browser sees %v", r.prefs())
	}
}

func TestPreferencesAreValidated(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")

	ok := func() m {
		return m{"mode": "light", "lightTheme": "bifrost-light", "darkTheme": "nord", "logo": "arches"}
	}
	with := func(k string, v any) m { p := ok(); p[k] = v; return p }

	for name, body := range map[string]m{
		"bad mode":                      with("mode", "sepia"),
		"unknown light theme":           with("lightTheme", "nope"),
		"a dark theme as the light one": with("lightTheme", "dracula"),
		"unknown dark theme":            with("darkTheme", "nope"),
		"a light theme as the dark one": with("darkTheme", "bifrost-light"),
		"bad logo":                      with("logo", "star"),
		"empty":                         m{},
		"css in a theme id":             with("darkTheme", "nord;}body{display:none"),
	} {
		if r := b.do("PUT", "/api/preferences", body); r.Code != 400 || r.errMsg() == "" {
			t.Errorf("%s: %d %s", name, r.Code, r.Body)
		}
	}
	mustStatus(t, b.do("PUT", "/api/preferences", ok()), 200)

	// A refused change leaves the saved preferences alone.
	b.do("PUT", "/api/preferences", with("mode", "sepia"))
	if got := b.do("GET", "/api/preferences", nil).prefs(); got["mode"] != "light" || got["logo"] != "arches" {
		t.Errorf("preferences changed: %v", got)
	}
}

func TestEachUserHasTheirOwnPreferences(t *testing.T) {
	e := newEnv(t)
	alex, sam, _ := adminAndUser(t, e)

	mustStatus(t, alex.do("PUT", "/api/preferences", m{"mode": "dark", "lightTheme": "bifrost-light", "darkTheme": "gruvbox", "logo": "spectrum"}), 200)
	r := sam.do("GET", "/api/preferences", nil)
	if r.prefs()["mode"] != "system" || r.prefs()["logo"] != "app" {
		t.Errorf("sam inherited alex's preferences: %v", r.prefs())
	}
}

func TestPreferencesWorkWhileThePasswordIsTemporary(t *testing.T) {
	e := newEnv(t)
	e.browser().setup("alex", "correct-password")
	temp, _ := auth.ResetPassword(context.Background(), e.s.store, "alex")

	b := e.browser()
	b.login("alex", temp, false)
	// The "choose a new password" screen is themed too.
	mustStatus(t, b.do("GET", "/api/preferences", nil), 200)
}

func TestAThemeThatHasGoneFallsBackToTheDefault(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	r := b.setup("alex", "correct-password")
	id := int64(r.user()["id"].(float64))

	// Saved earlier, when a theme that no longer exists was still there.
	_, err := e.s.db.Exec(`INSERT INTO preferences (user_id, mode, light_theme, dark_theme, logo) VALUES (?, 'dark', 'old-light', 'old-dark', 'arches')`, id)
	if err != nil {
		t.Fatal(err)
	}
	got := b.do("GET", "/api/preferences", nil).prefs()
	if got["lightTheme"] != "bifrost-light" || got["darkTheme"] != "bifrost-dark" || got["logo"] != "arches" || got["mode"] != "dark" {
		t.Errorf("prefs = %v", got)
	}
}

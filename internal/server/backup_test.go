package server

import (
	"encoding/json"
	"strings"
	"testing"
)

func appNames(r resp) []string {
	var out []string
	for _, a := range r.apps() {
		out = append(out, a.(map[string]any)["name"].(string))
	}
	return out
}

func TestExportNeedsASignedInUserAndHoldsOnlyTheirData(t *testing.T) {
	e := newEnv(t)
	mustStatus(t, e.browser().do("GET", "/api/export", nil), 401)

	alex := e.browser()
	alex.setup("alex", "correct-password")
	mustStatus(t, alex.do("POST", "/api/apps", m{"name": "Jellyfin", "url": "jf.lan", "category": "Media", "favourite": true, "icon": "jellyfin"}), 201)
	mustStatus(t, alex.do("POST", "/api/apps", m{"name": "Wiki", "url": "wiki.lan"}), 201)
	mustStatus(t, alex.do("PUT", "/api/preferences", m{"mode": "dark", "lightTheme": "bifrost-light", "darkTheme": "nord", "logo": "arches"}), 200)

	r := alex.do("GET", "/api/export", nil)
	mustStatus(t, r, 200)
	if !strings.Contains(r.Header().Get("Content-Disposition"), "attachment") {
		t.Errorf("Content-Disposition = %q", r.Header().Get("Content-Disposition"))
	}
	var f map[string]any
	if err := json.Unmarshal(r.Body.Bytes(), &f); err != nil {
		t.Fatal(err)
	}
	if f["format"] != "bifrost-export" || f["version"] != float64(1) {
		t.Errorf("header = %v %v", f["format"], f["version"])
	}
	apps := f["apps"].([]any)
	if len(apps) != 2 || apps[0].(map[string]any)["name"] != "Jellyfin" || apps[0].(map[string]any)["icon"] != "jellyfin" {
		t.Errorf("apps = %v", apps)
	}
	if _, has := apps[0].(map[string]any)["id"]; has {
		t.Error("ids must not be exported")
	}
	if f["preferences"].(map[string]any)["darkTheme"] != "nord" {
		t.Errorf("preferences = %v", f["preferences"])
	}
	for _, secret := range []string{"password", "hash", "token", "session"} {
		if strings.Contains(strings.ToLower(r.Body.String()), secret) {
			t.Errorf("the export mentions %q", secret)
		}
	}
}

func TestImportMergeAndReplace(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")
	mustStatus(t, b.do("POST", "/api/apps", m{"name": "Jellyfin", "url": "jf.lan", "category": "Media"}), 201)

	file := func(mode string, extra m) m {
		out := m{"format": "bifrost-export", "version": 1, "mode": mode, "apps": []m{
			{"name": "jellyfin", "url": "http://jf.lan", "category": "media", "newTab": true},
			{"name": "Sonarr", "url": "sonarr.lan", "category": "Media", "favourite": true, "newTab": false, "icon": "sonarr"},
			{"name": "Mystery", "url": "m.lan", "icon": "no-such-icon", "newTab": true},
		}}
		for k, v := range extra {
			out[k] = v
		}
		return out
	}

	// Merge: the duplicate (same name in any case, same address) is skipped.
	r := b.do("POST", "/api/import", file("merge", nil))
	mustStatus(t, r, 200)
	if j := r.json(); j["added"] != float64(2) || j["skipped"] != float64(1) || j["removed"] != float64(0) {
		t.Errorf("merge result = %v", j)
	}
	got := b.do("GET", "/api/apps", nil)
	if names := appNames(got); strings.Join(names, ",") != "Jellyfin,Sonarr,Mystery" {
		t.Errorf("apps after merge = %v", names)
	}
	sonarr := got.apps()[1].(map[string]any)
	if sonarr["newTab"] != false || sonarr["favourite"] != true || sonarr["icon"] != "sonarr" {
		t.Errorf("sonarr = %v", sonarr)
	}
	if got.apps()[2].(map[string]any)["icon"] != "" {
		t.Error("an unknown icon should be dropped")
	}

	// Importing the same file again adds nothing.
	r = b.do("POST", "/api/import", file("merge", nil))
	if j := r.json(); j["added"] != float64(0) || j["skipped"] != float64(3) {
		t.Errorf("second merge = %v", j)
	}

	// Replace swaps everything and applies appearance only when asked.
	prefs := m{"mode": "dark", "lightTheme": "bifrost-light", "darkTheme": "dracula", "logo": "spectrum"}
	r = b.do("POST", "/api/import", file("replace", m{"preferences": prefs}))
	mustStatus(t, r, 200)
	if j := r.json(); j["added"] != float64(3) || j["removed"] != float64(3) || j["appearanceApplied"] != false {
		t.Errorf("replace result = %v", j)
	}
	if p := b.do("GET", "/api/preferences", nil).prefs(); p["darkTheme"] == "dracula" {
		t.Error("appearance was applied without being asked")
	}
	r = b.do("POST", "/api/import", file("replace", m{"preferences": prefs, "appearance": true}))
	if r.json()["appearanceApplied"] != true {
		t.Errorf("result = %v", r.json())
	}
	if p := b.do("GET", "/api/preferences", nil).prefs(); p["darkTheme"] != "dracula" || p["logo"] != "spectrum" {
		t.Errorf("prefs = %v", p)
	}

	// An export round-trips.
	exported := b.do("GET", "/api/export", nil)
	var f m
	json.Unmarshal(exported.Body.Bytes(), &f)
	f["mode"] = "replace"
	mustStatus(t, b.do("POST", "/api/import", f), 200)
	if names := appNames(b.do("GET", "/api/apps", nil)); strings.Join(names, ",") != "jellyfin,Sonarr,Mystery" {
		t.Errorf("apps after a round trip = %v", names)
	}
}

func TestImportIsAllOrNothing(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")
	mustStatus(t, b.do("POST", "/api/apps", m{"name": "Keep", "url": "keep.lan"}), 201)

	bad := m{"format": "bifrost-export", "version": 1, "mode": "replace", "apps": []m{
		{"name": "Fine", "url": "fine.lan"},
		{"name": "Evil", "url": "javascript:alert(1)"},
	}}
	r := b.do("POST", "/api/import", bad)
	mustStatus(t, r, 400)
	if !strings.Contains(r.json()["error"].(string), "App 2") {
		t.Errorf("the message should name the app: %v", r.json()["error"])
	}
	if names := appNames(b.do("GET", "/api/apps", nil)); strings.Join(names, ",") != "Keep" {
		t.Errorf("a refused import changed the dashboard: %v", names)
	}
}

func TestImportRefusesWrongFilesAndModes(t *testing.T) {
	e := newEnv(t)
	b := e.browser()
	b.setup("alex", "correct-password")
	ok := []m{{"name": "A", "url": "a.lan"}}
	for name, body := range map[string]m{
		"other format":  {"format": "heimdall", "version": 1, "mode": "merge", "apps": ok},
		"other version": {"format": "bifrost-export", "version": 2, "mode": "merge", "apps": ok},
		"no mode":       {"format": "bifrost-export", "version": 1, "apps": ok},
		"odd mode":      {"format": "bifrost-export", "version": 1, "mode": "wipe", "apps": ok},
	} {
		mustStatus(t, b.do("POST", "/api/import", body), 400)
		_ = name
	}
	many := make([]m, maxImportApps+1)
	for i := range many {
		many[i] = m{"name": "A", "url": "a.lan"}
	}
	mustStatus(t, b.do("POST", "/api/import", m{"format": "bifrost-export", "version": 1, "mode": "merge", "apps": many}), 400)
	mustStatus(t, e.browser().do("POST", "/api/import", m{"format": "bifrost-export", "version": 1, "mode": "merge", "apps": ok}), 401)
}

func TestImportOnlyTouchesTheSignedInUser(t *testing.T) {
	e := newEnv(t)
	alex := e.browser()
	alex.setup("alex", "correct-password")
	mustStatus(t, alex.do("POST", "/api/apps", m{"name": "Alex's", "url": "a.lan"}), 201)
	r := alex.do("POST", "/api/users", m{"username": "sam", "isAdmin": false})
	mustStatus(t, r, 201)
	temp := r.json()["temporaryPassword"].(string)

	sam := e.browser()
	sam.login("sam", temp, false)
	mustStatus(t, sam.do("POST", "/api/me/password", m{"current": temp, "new": "sams-password-1"}), 204)
	sam = e.browser()
	sam.login("sam", "sams-password-1", false)
	mustStatus(t, sam.do("POST", "/api/import", m{"format": "bifrost-export", "version": 1, "mode": "replace", "apps": []m{{"name": "Sam's", "url": "s.lan"}}}), 200)

	if names := appNames(alex.do("GET", "/api/apps", nil)); strings.Join(names, ",") != "Alex's" {
		t.Errorf("Alex's apps = %v", names)
	}
	if names := appNames(sam.do("GET", "/api/apps", nil)); strings.Join(names, ",") != "Sam's" {
		t.Errorf("Sam's apps = %v", names)
	}
}

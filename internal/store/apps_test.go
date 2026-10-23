package store

import (
	"testing"
	"time"
)

func newUsers(t *testing.T) (*Store, User, User) {
	t.Helper()
	st := newStore(t)
	alex, err := st.CreateFirstAdmin(ctx, "alex", "hash", now)
	if err != nil {
		t.Fatal(err)
	}
	sam, err := st.CreateUser(ctx, "sam", "hash", false, false, now)
	if err != nil {
		t.Fatal(err)
	}
	return st, alex, sam
}

func input(name, category string) AppInput {
	return AppInput{Name: name, URL: "http://" + name + ".lan", Category: category, NewTab: true}
}

func TestCreateAndListApps(t *testing.T) {
	st, alex, _ := newUsers(t)

	a, err := st.CreateApp(ctx, alex.ID, AppInput{Name: "Jellyfin", URL: "http://jf.lan", Category: "Media", Favourite: true, NewTab: false}, now)
	if err != nil {
		t.Fatal(err)
	}
	if a.ID == 0 || a.UserID != alex.ID || a.Name != "Jellyfin" || a.URL != "http://jf.lan" ||
		a.Category != "Media" || !a.Favourite || a.NewTab || a.Position != 1 || !a.CreatedAt.Equal(now) {
		t.Errorf("unexpected app: %+v", a)
	}

	b, _ := st.CreateApp(ctx, alex.ID, input("sonarr", "Media"), now)
	if b.Position != 2 {
		t.Errorf("second app position = %d, want 2", b.Position)
	}

	apps, err := st.ListApps(ctx, alex.ID)
	if err != nil || len(apps) != 2 || apps[0].ID != a.ID || apps[1].ID != b.ID {
		t.Errorf("list = %+v, err %v", apps, err)
	}
}

func TestListAppsIsNeverNil(t *testing.T) {
	st, alex, _ := newUsers(t)
	apps, err := st.ListApps(ctx, alex.ID)
	if err != nil || apps == nil || len(apps) != 0 {
		t.Errorf("apps = %#v, err %v", apps, err)
	}
}

func TestAppsArePrivateToTheirOwner(t *testing.T) {
	st, alex, sam := newUsers(t)
	mine, _ := st.CreateApp(ctx, alex.ID, input("jellyfin", "Media"), now)
	theirs, _ := st.CreateApp(ctx, sam.ID, input("gitea", "Dev"), now)

	// Positions are per user.
	if mine.Position != 1 || theirs.Position != 1 {
		t.Errorf("positions: %d, %d, want 1 and 1", mine.Position, theirs.Position)
	}
	if list, _ := st.ListApps(ctx, sam.ID); len(list) != 1 || list[0].ID != theirs.ID {
		t.Errorf("sam sees %+v", list)
	}

	if _, err := st.UpdateApp(ctx, sam.ID, mine.ID, input("hijack", "x"), now); err != ErrNotFound {
		t.Errorf("update of another user's app: err = %v, want ErrNotFound", err)
	}
	if err := st.DeleteApp(ctx, sam.ID, mine.ID); err != ErrNotFound {
		t.Errorf("delete of another user's app: err = %v, want ErrNotFound", err)
	}
	if list, _ := st.ListApps(ctx, alex.ID); len(list) != 1 || list[0].Name != "jellyfin" {
		t.Errorf("alex's app was changed: %+v", list)
	}
}

func TestCategoryCaseIsReused(t *testing.T) {
	st, alex, sam := newUsers(t)
	st.CreateApp(ctx, alex.ID, input("a", "Media"), now)

	b, _ := st.CreateApp(ctx, alex.ID, input("b", "media"), now)
	if b.Category != "Media" {
		t.Errorf("category = %q, want the existing spelling Media", b.Category)
	}
	// Another user's spelling does not leak across.
	c, _ := st.CreateApp(ctx, sam.ID, input("c", "media"), now)
	if c.Category != "media" {
		t.Errorf("category = %q, want media", c.Category)
	}

	// Editing reuses it too.
	d, _ := st.CreateApp(ctx, alex.ID, input("d", "Tools"), now)
	d, err := st.UpdateApp(ctx, alex.ID, d.ID, input("d", "MEDIA"), now)
	if err != nil || d.Category != "Media" {
		t.Errorf("after update: %+v, err %v", d, err)
	}
}

func TestChangingTheCaseOfALonelyCategory(t *testing.T) {
	st, alex, _ := newUsers(t)
	a, _ := st.CreateApp(ctx, alex.ID, input("a", "media"), now)

	a, err := st.UpdateApp(ctx, alex.ID, a.ID, input("a", "Media"), now)
	if err != nil || a.Category != "Media" {
		t.Errorf("category = %q, err %v: an app must be able to fix its own category's case", a.Category, err)
	}
}

func TestUpdateApp(t *testing.T) {
	st, alex, _ := newUsers(t)
	a, _ := st.CreateApp(ctx, alex.ID, input("old", "Media"), now)
	later := now.Add(time.Hour)

	got, err := st.UpdateApp(ctx, alex.ID, a.ID, AppInput{Name: "new", URL: "https://new.lan", Category: "Tools", Favourite: true, NewTab: false}, later)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "new" || got.URL != "https://new.lan" || got.Category != "Tools" || !got.Favourite || got.NewTab ||
		got.Position != a.Position || !got.CreatedAt.Equal(now) || !got.UpdatedAt.Equal(later) {
		t.Errorf("unexpected app: %+v", got)
	}
	if _, err := st.UpdateApp(ctx, alex.ID, 999, input("x", "x"), now); err != ErrNotFound {
		t.Errorf("unknown app: err = %v", err)
	}
}

func TestDeleteApp(t *testing.T) {
	st, alex, _ := newUsers(t)
	a, _ := st.CreateApp(ctx, alex.ID, input("a", "x"), now)
	b, _ := st.CreateApp(ctx, alex.ID, input("b", "x"), now)

	if err := st.DeleteApp(ctx, alex.ID, a.ID); err != nil {
		t.Fatal(err)
	}
	if err := st.DeleteApp(ctx, alex.ID, a.ID); err != ErrNotFound {
		t.Errorf("second delete: err = %v, want ErrNotFound", err)
	}
	// A new app still goes to the end, never reusing a position.
	c, _ := st.CreateApp(ctx, alex.ID, input("c", "x"), now)
	if c.Position <= b.Position {
		t.Errorf("new position %d should come after %d", c.Position, b.Position)
	}
}

func TestAppsGoWhenTheirUserGoes(t *testing.T) {
	st, alex, sam := newUsers(t)
	st.CreateApp(ctx, sam.ID, input("a", "x"), now)
	st.CreateApp(ctx, alex.ID, input("b", "x"), now)

	if _, err := st.db.Exec(`DELETE FROM users WHERE id = ?`, sam.ID); err != nil {
		t.Fatal(err)
	}
	var n int
	st.db.QueryRow(`SELECT COUNT(*) FROM apps`).Scan(&n)
	if n != 1 {
		t.Errorf("%d apps left, want 1 (only alex's)", n)
	}
}

func ids(apps []App) []int64 {
	out := make([]int64, len(apps))
	for i, a := range apps {
		out[i] = a.ID
	}
	return out
}

func TestReorderApps(t *testing.T) {
	st, alex, _ := newUsers(t)
	a, _ := st.CreateApp(ctx, alex.ID, input("a", "x"), now)
	b, _ := st.CreateApp(ctx, alex.ID, input("b", "x"), now)
	c, _ := st.CreateApp(ctx, alex.ID, input("c", "y"), now)

	if err := st.ReorderApps(ctx, alex.ID, []int64{c.ID, a.ID, b.ID}); err != nil {
		t.Fatal(err)
	}
	list, _ := st.ListApps(ctx, alex.ID)
	if got := ids(list); got[0] != c.ID || got[1] != a.ID || got[2] != b.ID {
		t.Errorf("order = %v, want [%d %d %d]", got, c.ID, a.ID, b.ID)
	}
	for i, app := range list {
		if app.Position != int64(i+1) {
			t.Errorf("position %d = %d, want %d (renumbered from 1)", i, app.Position, i+1)
		}
	}

	// A new app still goes last.
	d, _ := st.CreateApp(ctx, alex.ID, input("d", "x"), now)
	if d.Position != 4 {
		t.Errorf("new app position = %d, want 4", d.Position)
	}
}

func TestReorderAppsRefusesAnythingButTheUsersOwnApps(t *testing.T) {
	st, alex, sam := newUsers(t)
	a, _ := st.CreateApp(ctx, alex.ID, input("a", "x"), now)
	b, _ := st.CreateApp(ctx, alex.ID, input("b", "x"), now)
	theirs, _ := st.CreateApp(ctx, sam.ID, input("s", "x"), now)

	for name, order := range map[string][]int64{
		"missing one":    {a.ID},
		"extra id":       {a.ID, b.ID, 999},
		"duplicate":      {a.ID, a.ID},
		"someone else's": {a.ID, theirs.ID},
		"empty":          {},
		"nil":            nil,
	} {
		if err := st.ReorderApps(ctx, alex.ID, order); err != ErrOrderMismatch {
			t.Errorf("%s: err = %v, want ErrOrderMismatch", name, err)
		}
	}

	// Nothing changed, for either user.
	if list, _ := st.ListApps(ctx, alex.ID); list[0].ID != a.ID || list[1].ID != b.ID || list[0].Position != 1 {
		t.Errorf("alex's order changed: %+v", list)
	}
	if list, _ := st.ListApps(ctx, sam.ID); list[0].Position != 1 {
		t.Errorf("sam's app changed: %+v", list)
	}
}

func TestReorderAppsForAUserWithNoApps(t *testing.T) {
	st, alex, _ := newUsers(t)
	if err := st.ReorderApps(ctx, alex.ID, []int64{}); err != nil {
		t.Errorf("empty list for an empty dashboard: %v", err)
	}
}

func TestAppIcons(t *testing.T) {
	st, alex, _ := newUsers(t)

	in := input("jellyfin", "Media")
	in.Icon = "jellyfin"
	a, err := st.CreateApp(ctx, alex.ID, in, now)
	if err != nil || a.Icon != "jellyfin" {
		t.Fatalf("created %+v, %v", a, err)
	}
	if plain, _ := st.CreateApp(ctx, alex.ID, input("custom", "Media"), now); plain.Icon != "" {
		t.Errorf("an app with no icon has %q", plain.Icon)
	}

	// The icon can be changed and removed, and survives a listing.
	in.Icon = "plex"
	a, _ = st.UpdateApp(ctx, alex.ID, a.ID, in, now)
	if a.Icon != "plex" {
		t.Errorf("after update: %q", a.Icon)
	}
	in.Icon = ""
	a, _ = st.UpdateApp(ctx, alex.ID, a.ID, in, now)
	list, _ := st.ListApps(ctx, alex.ID)
	if a.Icon != "" || list[0].Icon != "" {
		t.Errorf("icon not removed: %q / %q", a.Icon, list[0].Icon)
	}
}

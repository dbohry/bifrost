package store

import (
	"context"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dbohry/bifrost/internal/db"
)

var ctx = context.Background()

func newStore(t *testing.T) *Store {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := db.Migrate(conn); err != nil {
		t.Fatal(err)
	}
	return New(conn)
}

var now = time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)

func TestCreateFirstAdminOnlyOnce(t *testing.T) {
	st := newStore(t)

	if has, _ := st.HasUsers(ctx); has {
		t.Fatal("a new database must have no users")
	}
	u, err := st.CreateFirstAdmin(ctx, "alex", "hash", now)
	if err != nil {
		t.Fatal(err)
	}
	if !u.IsAdmin || u.MustChangePassword || u.Username != "alex" || !u.CreatedAt.Equal(now) {
		t.Errorf("unexpected user: %+v", u)
	}
	if _, err := st.CreateFirstAdmin(ctx, "sam", "hash", now); err != ErrSetupClosed {
		t.Errorf("second admin: err = %v, want ErrSetupClosed", err)
	}
	if has, _ := st.HasUsers(ctx); !has {
		t.Error("HasUsers should be true now")
	}
}

func TestCreateFirstAdminRace(t *testing.T) {
	st := newStore(t)

	var wins, closed atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := st.CreateFirstAdmin(ctx, "user"+string(rune('a'+i)), "hash", now)
			switch err {
			case nil:
				wins.Add(1)
			case ErrSetupClosed:
				closed.Add(1)
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if wins.Load() != 1 || closed.Load() != 19 {
		t.Errorf("wins=%d closed=%d, want exactly one winner", wins.Load(), closed.Load())
	}
}

func TestUsernamesAreCaseInsensitive(t *testing.T) {
	st := newStore(t)
	st.CreateFirstAdmin(ctx, "Alex", "hash", now)

	u, err := st.UserByUsername(ctx, "aLEX")
	if err != nil || u.Username != "Alex" {
		t.Errorf("lookup: %+v, %v", u, err)
	}
	if _, err := st.UserByUsername(ctx, "nobody"); err != ErrNotFound {
		t.Errorf("unknown user: err = %v", err)
	}
	if _, err := st.UserByID(ctx, 999); err != ErrNotFound {
		t.Errorf("unknown id: err = %v", err)
	}
}

func TestSetPassword(t *testing.T) {
	st := newStore(t)
	u, _ := st.CreateFirstAdmin(ctx, "alex", "old", now)

	if err := st.SetPassword(ctx, u.ID, "new", true); err != nil {
		t.Fatal(err)
	}
	got, _ := st.UserByID(ctx, u.ID)
	if got.PasswordHash != "new" || !got.MustChangePassword {
		t.Errorf("unexpected user: %+v", got)
	}
	if err := st.SetPassword(ctx, 999, "x", false); err != ErrNotFound {
		t.Errorf("unknown user: err = %v", err)
	}
}

func TestSessions(t *testing.T) {
	st := newStore(t)
	u, _ := st.CreateFirstAdmin(ctx, "alex", "hash", now)

	token, hash, err := NewToken()
	if err != nil {
		t.Fatal(err)
	}
	if hash == token || HashToken(token) != hash {
		t.Error("the stored value must be the hash of the token, not the token")
	}
	if err := st.CreateSession(ctx, hash, u.ID, true, now, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	sess, user, err := st.SessionUser(ctx, hash, now.Add(30*time.Minute))
	if err != nil || user.ID != u.ID || !sess.Persistent {
		t.Fatalf("lookup: %+v %+v %v", sess, user, err)
	}
	if _, _, err := st.SessionUser(ctx, hash, now.Add(time.Hour)); err != ErrNotFound {
		t.Errorf("expired session: err = %v, want ErrNotFound", err)
	}

	if err := st.ExtendSession(ctx, hash, now.Add(3*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := st.SessionUser(ctx, hash, now.Add(2*time.Hour)); err != nil {
		t.Errorf("extended session should still be valid: %v", err)
	}

	if err := st.DeleteSession(ctx, hash); err != nil {
		t.Fatal(err)
	}
	if _, _, err := st.SessionUser(ctx, hash, now); err != ErrNotFound {
		t.Errorf("deleted session: err = %v", err)
	}
}

func TestDeleteUserSessionsKeepsOne(t *testing.T) {
	st := newStore(t)
	u, _ := st.CreateFirstAdmin(ctx, "alex", "hash", now)
	for _, h := range []string{"a", "b", "c"} {
		st.CreateSession(ctx, h, u.ID, false, now, now.Add(time.Hour))
	}

	if err := st.DeleteUserSessions(ctx, u.ID, "b"); err != nil {
		t.Fatal(err)
	}
	for h, want := range map[string]bool{"a": false, "b": true, "c": false} {
		_, _, err := st.SessionUser(ctx, h, now)
		if (err == nil) != want {
			t.Errorf("session %s alive = %v, want %v", h, err == nil, want)
		}
	}

	st.DeleteUserSessions(ctx, u.ID, "")
	if _, _, err := st.SessionUser(ctx, "b", now); err != ErrNotFound {
		t.Error("empty keep should delete everything")
	}
}

func TestPurgeExpiredSessions(t *testing.T) {
	st := newStore(t)
	u, _ := st.CreateFirstAdmin(ctx, "alex", "hash", now)
	st.CreateSession(ctx, "old", u.ID, false, now, now.Add(time.Minute))
	st.CreateSession(ctx, "new", u.ID, false, now, now.Add(time.Hour))

	n, err := st.PurgeExpiredSessions(ctx, now.Add(2*time.Minute))
	if err != nil || n != 1 {
		t.Fatalf("purged %d, err %v, want 1", n, err)
	}
	if _, _, err := st.SessionUser(ctx, "new", now); err != nil {
		t.Errorf("valid session was purged: %v", err)
	}
}

func TestCreateAndListUsers(t *testing.T) {
	st := newStore(t)
	alex, _ := st.CreateFirstAdmin(ctx, "alex", "h1", now)

	sam, err := st.CreateUser(ctx, "Sam", "h2", false, true, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if sam.IsAdmin || !sam.MustChangePassword || sam.Username != "Sam" || !sam.LastSeenAt.IsZero() || !sam.CreatedAt.Equal(now.Add(time.Hour)) {
		t.Errorf("unexpected user: %+v", sam)
	}
	jo, err := st.CreateUser(ctx, "jo", "h3", true, false, now)
	if err != nil || !jo.IsAdmin || jo.MustChangePassword {
		t.Errorf("admin user: %+v, %v", jo, err)
	}

	users, err := st.ListUsers(ctx)
	if err != nil || len(users) != 3 || users[0].ID != alex.ID || users[1].ID != sam.ID || users[2].ID != jo.ID {
		t.Errorf("users = %+v, err %v", users, err)
	}
}

func TestCreateUserRejectsATakenUsernameInAnyCase(t *testing.T) {
	st := newStore(t)
	st.CreateFirstAdmin(ctx, "Alex", "h", now)

	for _, name := range []string{"Alex", "alex", "ALEX"} {
		if _, err := st.CreateUser(ctx, name, "h", false, false, now); err != ErrUsernameTaken {
			t.Errorf("%q: err = %v, want ErrUsernameTaken", name, err)
		}
	}
	if users, _ := st.ListUsers(ctx); len(users) != 1 {
		t.Errorf("%d users, want 1", len(users))
	}
}

func TestTouchUserSetsLastSeen(t *testing.T) {
	st := newStore(t)
	u, _ := st.CreateFirstAdmin(ctx, "alex", "h", now)

	later := now.Add(3 * time.Hour)
	if err := st.TouchUser(ctx, u.ID, later); err != nil {
		t.Fatal(err)
	}
	got, _ := st.UserByID(ctx, u.ID)
	if !got.LastSeenAt.Equal(later) {
		t.Errorf("LastSeenAt = %v, want %v", got.LastSeenAt, later)
	}
	// It also comes back from a session lookup and from the list.
	st.CreateSession(ctx, "tok", u.ID, false, now, now.Add(time.Hour))
	if _, viaSession, _ := st.SessionUser(ctx, "tok", now); !viaSession.LastSeenAt.Equal(later) {
		t.Errorf("session user LastSeenAt = %v", viaSession.LastSeenAt)
	}
	if users, _ := st.ListUsers(ctx); !users[0].LastSeenAt.Equal(later) {
		t.Errorf("listed LastSeenAt = %v", users[0].LastSeenAt)
	}
}

func TestDeleteUserRemovesTheirAppsAndSessions(t *testing.T) {
	st := newStore(t)
	st.CreateFirstAdmin(ctx, "alex", "h", now)
	sam, _ := st.CreateUser(ctx, "sam", "h", false, false, now)
	st.CreateSession(ctx, "sams", sam.ID, false, now, now.Add(time.Hour))
	st.CreateApp(ctx, sam.ID, AppInput{Name: "a", URL: "http://a.lan", Category: "x", NewTab: true}, now)

	if err := st.DeleteUser(ctx, sam.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := st.UserByID(ctx, sam.ID); err != ErrNotFound {
		t.Errorf("user still there: %v", err)
	}
	if _, _, err := st.SessionUser(ctx, "sams", now); err != ErrNotFound {
		t.Errorf("session survived: %v", err)
	}
	var n int
	st.db.QueryRow(`SELECT COUNT(*) FROM apps`).Scan(&n)
	if n != 0 {
		t.Errorf("%d apps left", n)
	}
	if err := st.DeleteUser(ctx, sam.ID); err != ErrNotFound {
		t.Errorf("second delete: err = %v", err)
	}
}

func TestPreferences(t *testing.T) {
	st := newStore(t)
	alex, _ := st.CreateFirstAdmin(ctx, "alex", "h", now)
	sam, _ := st.CreateUser(ctx, "sam", "h", false, false, now)

	if _, err := st.GetPreferences(ctx, alex.ID); err != ErrNotFound {
		t.Fatalf("new user: err = %v, want ErrNotFound", err)
	}

	p := Preferences{Mode: "dark", LightTheme: "solarized-light", DarkTheme: "dracula", Logo: "spectrum"}
	if err := st.SetPreferences(ctx, alex.ID, p); err != nil {
		t.Fatal(err)
	}
	if got, err := st.GetPreferences(ctx, alex.ID); err != nil || got != p {
		t.Errorf("got %+v, %v; want %+v", got, err, p)
	}

	// Saving again replaces, and other users are untouched.
	p2 := Preferences{Mode: "system", LightTheme: "bifrost-light", DarkTheme: "nord", Logo: "arches"}
	if err := st.SetPreferences(ctx, alex.ID, p2); err != nil {
		t.Fatal(err)
	}
	if got, _ := st.GetPreferences(ctx, alex.ID); got != p2 {
		t.Errorf("after replace: %+v", got)
	}
	if _, err := st.GetPreferences(ctx, sam.ID); err != ErrNotFound {
		t.Errorf("sam's preferences appeared: %v", err)
	}

	// They go when the user goes.
	st.SetPreferences(ctx, sam.ID, p)
	st.DeleteUser(ctx, sam.ID)
	var n int
	st.db.QueryRow(`SELECT COUNT(*) FROM preferences`).Scan(&n)
	if n != 1 {
		t.Errorf("%d preference rows, want 1", n)
	}
}

func TestPreferencesRefuseValuesOutsideTheirSets(t *testing.T) {
	st := newStore(t)
	u, _ := st.CreateFirstAdmin(ctx, "alex", "h", now)
	for name, p := range map[string]Preferences{
		"mode": {Mode: "sepia", LightTheme: "a", DarkTheme: "b", Logo: "app"},
		"logo": {Mode: "dark", LightTheme: "a", DarkTheme: "b", Logo: "star"},
	} {
		if err := st.SetPreferences(ctx, u.ID, p); err == nil {
			t.Errorf("%s: expected the database to refuse it", name)
		}
	}
}

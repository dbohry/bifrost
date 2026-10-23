package store

import (
	"testing"
	"time"
)

func TestStatusSettingsDefaultAndSave(t *testing.T) {
	st, alex, sam := newUsers(t)
	v, err := st.GetStatusSettings(ctx, alex.ID)
	if err != nil || v.Interval != DefaultCheckInterval || v.Timeout != DefaultCheckTimeout {
		t.Fatalf("defaults = %+v, %v", v, err)
	}
	if err := st.SetStatusSettings(ctx, alex.ID, StatusSettings{Interval: 5 * time.Minute, Timeout: 10 * time.Second}); err != nil {
		t.Fatal(err)
	}
	if v, _ := st.GetStatusSettings(ctx, alex.ID); v.Interval != 5*time.Minute || v.Timeout != 10*time.Second {
		t.Errorf("saved = %+v", v)
	}
	if v, _ := st.GetStatusSettings(ctx, sam.ID); v.Interval != DefaultCheckInterval {
		t.Errorf("another user's settings changed: %+v", v)
	}
}

func TestCheckTargetsUseTheOwnersSettingsAndTheCheckAddress(t *testing.T) {
	st, alex, sam := newUsers(t)
	a, _ := st.CreateApp(ctx, alex.ID, AppInput{Name: "On", URL: "http://on.lan", Category: "X", NewTab: true, CheckEnabled: true, CheckStrict: true}, now)
	b, _ := st.CreateApp(ctx, alex.ID, AppInput{Name: "Elsewhere", URL: "http://pub.example.com", Category: "X", NewTab: true, CheckEnabled: true, CheckURL: "http://10.0.0.5:8000", CheckInsecure: true}, now)
	st.CreateApp(ctx, alex.ID, input("off", "X"), now)
	c, _ := st.CreateApp(ctx, sam.ID, AppInput{Name: "Sams", URL: "http://sam.lan", Category: "X", NewTab: true, CheckEnabled: true}, now)
	st.SetStatusSettings(ctx, alex.ID, StatusSettings{Interval: 2 * time.Minute, Timeout: 10 * time.Second})

	targets, err := st.CheckTargets(ctx)
	if err != nil || len(targets) != 3 {
		t.Fatalf("targets = %+v, %v", targets, err)
	}
	by := map[int64]CheckTarget{}
	for _, tg := range targets {
		by[tg.AppID] = tg
	}
	if x := by[a.ID]; x.URL != "http://on.lan" || !x.Strict || x.Insecure || x.Interval != 2*time.Minute || x.Timeout != 10*time.Second {
		t.Errorf("On = %+v", x)
	}
	if x := by[b.ID]; x.URL != "http://10.0.0.5:8000" || !x.Insecure || x.Strict {
		t.Errorf("Elsewhere should use its check address: %+v", x)
	}
	if x := by[c.ID]; x.UserID != sam.ID || x.Interval != DefaultCheckInterval || x.Timeout != DefaultCheckTimeout {
		t.Errorf("Sams = %+v", x)
	}
}

func TestStatusHoursAddUpAndStayWithTheirOwner(t *testing.T) {
	st, alex, sam := newUsers(t)
	a, _ := st.CreateApp(ctx, alex.ID, input("a", "X"), now)

	if err := st.AddStatusCounts(ctx, []HourCount{{AppID: a.ID, Hour: 100, Checks: 60, Failures: 2}, {AppID: a.ID, Hour: 101, Checks: 5, Failures: 5}}); err != nil {
		t.Fatal(err)
	}
	// the same hour again adds to it
	if err := st.AddStatusCounts(ctx, []HourCount{{AppID: a.ID, Hour: 100, Checks: 10, Failures: 1}}); err != nil {
		t.Fatal(err)
	}
	// an app that does not exist is skipped without failing the batch
	if err := st.AddStatusCounts(ctx, []HourCount{{AppID: 9999, Hour: 100, Checks: 1}, {AppID: a.ID, Hour: 102, Checks: 1}}); err != nil {
		t.Fatalf("unknown app: %v", err)
	}

	got, err := st.StatusHours(ctx, alex.ID, a.ID, 100)
	if err != nil || len(got) != 3 || got[0].Checks != 70 || got[0].Failures != 3 || got[1].Failures != 5 {
		t.Fatalf("hours = %+v, %v", got, err)
	}
	if got, _ := st.StatusHours(ctx, alex.ID, a.ID, 101); len(got) != 2 {
		t.Errorf("since 101 = %+v", got)
	}
	if got, _ := st.StatusHours(ctx, sam.ID, a.ID, 0); len(got) != 0 {
		t.Errorf("Sam can read Alex's history: %+v", got)
	}

	if err := st.PruneStatusHours(ctx, 101); err != nil {
		t.Fatal(err)
	}
	if got, _ := st.StatusHours(ctx, alex.ID, a.ID, 0); len(got) != 2 || got[0].Hour != 101 {
		t.Errorf("after pruning = %+v", got)
	}

	// deleting the app deletes its history
	if err := st.DeleteApp(ctx, alex.ID, a.ID); err != nil {
		t.Fatal(err)
	}
	var n int
	st.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM status_hours`).Scan(&n)
	if n != 0 {
		t.Errorf("%d history rows left after deleting the app", n)
	}
}

package status

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/dbohry/bifrost/internal/store"
)

type fakeSource struct {
	mu      sync.Mutex
	targets []store.CheckTarget
}

func (f *fakeSource) CheckTargets(context.Context) ([]store.CheckTarget, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]store.CheckTarget(nil), f.targets...), nil
}

func (f *fakeSource) set(t ...store.CheckTarget) {
	f.mu.Lock()
	f.targets = t
	f.mu.Unlock()
}

// rig is a monitor with a clock and an answer that the test controls.
type rig struct {
	m   *Monitor
	src *fakeSource
	t0  time.Time
	now time.Time
	up  bool
	n   int
	mu  sync.Mutex
}

func newRig() *rig {
	r := &rig{src: &fakeSource{}, t0: time.Unix(1_000_000, 0), up: true}
	r.now = r.t0
	r.m = NewMonitor(r.src, slog.New(slog.NewTextHandler(io.Discard, nil)))
	r.m.now = func() time.Time { r.mu.Lock(); defer r.mu.Unlock(); return r.now }
	r.m.check = func(context.Context, store.CheckTarget) Outcome {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.n++
		if r.up {
			return Outcome{Up: true, Latency: 20 * time.Millisecond, Code: 200}
		}
		return Outcome{Reason: "Connection refused"}
	}
	return r
}

// step moves the clock on, runs the monitor once, and waits for the looks to finish.
func (r *rig) step(after time.Duration) {
	r.mu.Lock()
	r.now = r.now.Add(after)
	r.mu.Unlock()
	r.m.Step(context.Background())
	for i := 0; i < 200; i++ {
		r.m.mu.Lock()
		busy := false
		for _, e := range r.m.entries {
			busy = busy || e.running
		}
		r.m.mu.Unlock()
		if !busy {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func (r *rig) result(app int64) (Result, bool) {
	res, ok := r.m.For(1)[app]
	return res, ok
}

func target(app int64) store.CheckTarget {
	return store.CheckTarget{AppID: app, UserID: 1, URL: "http://a.lan", Interval: time.Minute, Timeout: 5 * time.Second}
}

func TestAnAppGoesUpAndIsLookedAtOnSchedule(t *testing.T) {
	r := newRig()
	r.src.set(target(7))

	r.step(0)
	res, ok := r.result(7)
	if !ok || res.State != Up || res.Latency != 20*time.Millisecond || res.Address != "http://a.lan" {
		t.Fatalf("after the first look: %+v", res)
	}
	r.step(30 * time.Second)
	if r.n != 1 {
		t.Errorf("looked again after 30 s: %d looks", r.n)
	}
	r.step(31 * time.Second)
	if r.n != 2 {
		t.Errorf("did not look after a minute: %d looks", r.n)
	}
}

func TestItTakesTwoMissesToGoDownAndOneHitToComeBack(t *testing.T) {
	r := newRig()
	r.src.set(target(7))
	r.step(0)
	up, _ := r.result(7)

	r.up = false
	r.step(61 * time.Second)
	res, _ := r.result(7)
	if res.State != Up || res.Reason != "" {
		t.Fatalf("one miss changed things: %+v", res)
	}
	// the second look comes soon, not after a whole interval
	r.step(5 * time.Second)
	res, _ = r.result(7)
	if res.State != Down || res.Reason != "Connection refused" || !res.Since.After(up.Since) {
		t.Fatalf("two misses: %+v", res)
	}
	downSince := res.Since

	r.step(61 * time.Second)
	res, _ = r.result(7)
	if res.State != Down || !res.Since.Equal(downSince) {
		t.Errorf("staying down moved Since: %+v", res)
	}

	r.up = true
	r.step(61 * time.Second)
	res, _ = r.result(7)
	if res.State != Up || res.Reason != "" || !res.Since.After(downSince) {
		t.Errorf("coming back: %+v", res)
	}
}

func TestAFirstMissLeavesItChecking(t *testing.T) {
	r := newRig()
	r.up = false
	r.src.set(target(7))
	r.step(0)
	if res, _ := r.result(7); res.State != Checking {
		t.Errorf("after one miss on a new app: %+v", res)
	}
	r.step(5 * time.Second)
	if res, _ := r.result(7); res.State != Down {
		t.Errorf("after two: %+v", res)
	}
}

func TestChangedAndRemovedAppsStartOver(t *testing.T) {
	r := newRig()
	r.src.set(target(7), target(8))
	r.step(0)
	if len(r.m.For(1)) != 2 {
		t.Fatal("both apps should have results")
	}

	changed := target(7)
	changed.URL = "http://b.lan"
	r.up = false
	r.src.set(changed) // 8 is gone
	r.step(1 * time.Second)
	if _, ok := r.result(8); ok {
		t.Error("a removed app kept its result")
	}
	if res, _ := r.result(7); res.State != Checking || res.Address != "http://b.lan" {
		t.Errorf("a changed address should start again: %+v", res)
	}
}

func TestResultsBelongToTheirUser(t *testing.T) {
	r := newRig()
	other := target(9)
	other.UserID = 2
	r.src.set(target(7), other)
	r.step(0)
	if got := r.m.For(1); len(got) != 1 || got[7].State != Up {
		t.Errorf("user 1 sees %v", got)
	}
	if got := r.m.For(2); len(got) != 1 {
		t.Errorf("user 2 sees %v", got)
	}
	if got := r.m.For(3); len(got) != 0 {
		t.Errorf("user 3 sees %v", got)
	}
}

func TestAShorterIntervalTakesEffectAtOnce(t *testing.T) {
	r := newRig()
	r.src.set(target(7))
	r.step(0)
	faster := target(7)
	faster.Interval = 10 * time.Second
	r.src.set(faster)
	r.step(11 * time.Second)
	if r.n != 2 {
		t.Errorf("looks = %d, want a second look after 11 s", r.n)
	}
}

type fakeHistory struct {
	mu     sync.Mutex
	saved  []store.HourCount
	pruned []int64
	fail   bool
}

func (f *fakeHistory) AddStatusCounts(_ context.Context, c []store.HourCount) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail {
		return context.DeadlineExceeded
	}
	f.saved = append(f.saved, c...)
	return nil
}

func (f *fakeHistory) PruneStatusHours(_ context.Context, before int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pruned = append(f.pruned, before)
	return nil
}

func TestEveryLookIsCountedForTheHistoryBar(t *testing.T) {
	r := newRig()
	h := &fakeHistory{}
	r.m.SetHistory(h)
	r.src.set(target(7))

	r.step(0) // hit
	r.up = false
	r.step(61 * time.Second) // miss
	r.step(5 * time.Second)  // miss, now down
	r.up = true
	r.step(61 * time.Second) // hit

	var pend, pendFails int64
	for _, c := range r.m.Pending(7) {
		pend += c.Checks
		pendFails += c.Failures
	}
	if pend != 4 || pendFails != 2 {
		t.Fatalf("pending = %+v", r.m.Pending(7))
	}

	r.m.Flush(context.Background())
	if len(r.m.Pending(7)) != 0 {
		t.Error("counts were not cleared after saving")
	}
	var checks, fails int64
	for _, c := range h.saved {
		checks += c.Checks
		fails += c.Failures
	}
	if checks != 4 || fails != 2 || len(h.pruned) != 1 {
		t.Errorf("saved %+v, pruned %v", h.saved, h.pruned)
	}
	if want := r.now.Unix()/3600 - keepHours; h.pruned[0] != want {
		t.Errorf("pruned before %d, want %d", h.pruned[0], want)
	}
}

func TestCountsAreKeptWhenSavingFails(t *testing.T) {
	r := newRig()
	h := &fakeHistory{fail: true}
	r.m.SetHistory(h)
	r.src.set(target(7))
	r.step(0)
	r.m.Flush(context.Background())
	if got := r.m.Pending(7); len(got) != 1 || got[0].Checks != 1 {
		t.Fatalf("after a failed save: %+v", got)
	}
	r.step(61 * time.Second)
	h.mu.Lock()
	h.fail = false
	h.mu.Unlock()
	r.m.Flush(context.Background())
	var checks int64
	for _, c := range h.saved {
		checks += c.Checks
	}
	if checks != 2 || len(r.m.Pending(7)) != 0 {
		t.Errorf("saved %+v pending %+v", h.saved, r.m.Pending(7))
	}
}

func TestOwnsOnlyForTheRightUser(t *testing.T) {
	r := newRig()
	r.src.set(target(7))
	r.step(0)
	if !r.m.Owns(1, 7) || r.m.Owns(2, 7) || r.m.Owns(1, 8) {
		t.Error("ownership is wrong")
	}
}

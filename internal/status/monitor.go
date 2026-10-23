package status

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/dbohry/bifrost/internal/store"
)

// State is what Bifrost currently believes about an app.
type State string

const (
	// Checking means no verdict yet: a first look is under way, or the first one failed and
	// a second is coming.
	Checking State = "checking"
	Up       State = "up"
	Down     State = "down"
)

const (
	// failuresToGoDown is how many looks in a row must fail before an app is shown as offline,
	// so one dropped packet does not light up the dashboard.
	failuresToGoDown = 2
	// retryAfter is how soon a failed look is repeated, if the normal wait is longer.
	retryAfter = 5 * time.Second
	// tick is how often the monitor sees what is due.
	tick = time.Second
	// flushEvery is how often the counts for the history bar are saved, in one batch.
	flushEvery = 5 * time.Minute
	// keepHours is how long history is kept. The card shows 24 hours; the rest is headroom.
	keepHours = 7 * 24
	// parallel is how many apps are looked at at once, so a slow one never holds up the rest.
	parallel = 16
)

// Result is the latest news about one app.
type Result struct {
	State       State
	Since       time.Time // when the state last changed
	LastChecked time.Time
	Latency     time.Duration
	Reason      string // why it is down, in words
	Address     string // what was looked at
}

// Source says which apps want checking. The store does.
type Source interface {
	CheckTargets(ctx context.Context) ([]store.CheckTarget, error)
}

// History saves the counts behind the 24-hour bar. The store does.
type History interface {
	AddStatusCounts(ctx context.Context, counts []store.HourCount) error
	PruneStatusHours(ctx context.Context, before int64) error
}

type entry struct {
	userID  int64
	sig     string // address and rules, so a change starts again from "checking"
	result  Result
	fails   int
	next    time.Time
	running bool
}

// Monitor looks at every app that has its check on, as often as its owner asked,
// and keeps the latest result of each in memory. A restart starts over, and the
// first looks happen straight away.
type Monitor struct {
	src Source
	log *slog.Logger
	now func() time.Time
	// check is Check, replaceable in tests.
	check func(ctx context.Context, t store.CheckTarget) Outcome

	mu      sync.Mutex
	entries map[int64]*entry
	sem     chan struct{}

	history History
	// pending holds counts not yet saved: app, then hour.
	pending map[int64]map[int64]*store.HourCount
}

// NewMonitor returns a monitor that has not started. Call Run.
func NewMonitor(src Source, log *slog.Logger) *Monitor {
	return &Monitor{
		src:     src,
		log:     log,
		now:     time.Now,
		entries: map[int64]*entry{},
		pending: map[int64]map[int64]*store.HourCount{},
		sem:     make(chan struct{}, parallel),
		check: func(ctx context.Context, t store.CheckTarget) Outcome {
			return Check(ctx, t.URL, t.Timeout, t.Strict, t.Insecure)
		},
	}
}

// SetHistory makes the monitor save the counts behind the history bar. Without it they are
// kept in memory only.
func (m *Monitor) SetHistory(h History) {
	m.mu.Lock()
	m.history = h
	m.mu.Unlock()
}

// SetCheck replaces how an app is looked at. Tests use it to give fixed answers.
func (m *Monitor) SetCheck(f func(ctx context.Context, t store.CheckTarget) Outcome) {
	m.mu.Lock()
	m.check = f
	m.mu.Unlock()
}

// Wait returns once no look is under way. Tests use it after Step.
func (m *Monitor) Wait() {
	for {
		m.mu.Lock()
		busy := false
		for _, e := range m.entries {
			busy = busy || e.running
		}
		m.mu.Unlock()
		if !busy {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// Run checks until ctx ends.
func (m *Monitor) Run(ctx context.Context) {
	t := time.NewTicker(tick)
	defer t.Stop()
	f := time.NewTicker(flushEvery)
	defer f.Stop()
	for {
		m.Step(ctx)
		select {
		case <-ctx.Done():
			// Save what has not been saved yet. ctx is over, so use a fresh one.
			end, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			m.Flush(end)
			cancel()
			return
		case <-f.C:
			m.Flush(ctx)
		case <-t.C:
		}
	}
}

// Step starts a look at every app that is due. It does not wait for them.
func (m *Monitor) Step(ctx context.Context) {
	targets, err := m.src.CheckTargets(ctx)
	if err != nil {
		if ctx.Err() == nil {
			m.log.Error("list status checks", "err", err)
		}
		return
	}
	now := m.now()

	m.mu.Lock()
	live := make(map[int64]bool, len(targets))
	var due []store.CheckTarget
	for _, t := range targets {
		live[t.AppID] = true
		sig := signature(t)
		e := m.entries[t.AppID]
		if e == nil || e.sig != sig {
			// New, or the address or rules changed: forget what was known and look again now.
			e = &entry{userID: t.UserID, sig: sig, next: now, result: Result{State: Checking, Since: now, Address: t.URL}}
			m.entries[t.AppID] = e
		}
		// A shorter interval takes effect without waiting out the old one.
		if !e.running && !e.result.LastChecked.IsZero() {
			if latest := e.result.LastChecked.Add(t.Interval); e.next.After(latest) && e.fails == 0 {
				e.next = latest
			}
		}
		if !e.running && !now.Before(e.next) {
			e.running = true
			due = append(due, t)
		}
	}
	for id := range m.entries {
		if !live[id] {
			delete(m.entries, id)
		}
	}
	m.mu.Unlock()

	for _, t := range due {
		m.sem <- struct{}{}
		go func() {
			defer func() { <-m.sem }()
			m.mu.Lock()
			check := m.check
			m.mu.Unlock()
			m.finish(t, check(ctx, t))
		}()
	}
}

func signature(t store.CheckTarget) string {
	s := t.URL
	if t.Strict {
		s += "|strict"
	}
	if t.Insecure {
		s += "|insecure"
	}
	return s
}

// finish records what a look found and decides when to look again.
func (m *Monitor) finish(t store.CheckTarget, o Outcome) {
	now := m.now()
	m.mu.Lock()
	defer m.mu.Unlock()

	e := m.entries[t.AppID]
	if e == nil || e.sig != signature(t) {
		return // the app was removed or changed while we were looking
	}
	e.running = false
	e.result.LastChecked = now
	e.result.Address = t.URL
	m.count(t.AppID, now, !o.Up)

	if o.Up {
		e.fails = 0
		e.result.Latency = o.Latency
		e.result.Reason = ""
		if e.result.State != Up {
			e.result.State, e.result.Since = Up, now
		}
		e.next = now.Add(t.Interval)
		return
	}

	e.fails++
	if e.fails >= failuresToGoDown {
		e.result.Latency = 0
		e.result.Reason = o.Reason
		if e.result.State != Down {
			e.result.State, e.result.Since = Down, now
		}
		e.next = now.Add(t.Interval)
		return
	}
	// First miss: say nothing new yet, and try again soon.
	e.next = now.Add(min(t.Interval, retryAfter))
}

// For returns the latest results for one user's apps, by app id. Apps with no
// result yet are left out.
func (m *Monitor) For(userID int64) map[int64]Result {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[int64]Result{}
	for id, e := range m.entries {
		if e.userID == userID {
			out[id] = e.result
		}
	}
	return out
}

// count adds one look to the hour it was made in. The caller holds the lock.
func (m *Monitor) count(app int64, at time.Time, failed bool) {
	hour := at.Unix() / 3600
	byHour := m.pending[app]
	if byHour == nil {
		byHour = map[int64]*store.HourCount{}
		m.pending[app] = byHour
	}
	c := byHour[hour]
	if c == nil {
		c = &store.HourCount{AppID: app, Hour: hour}
		byHour[hour] = c
	}
	c.Checks++
	if failed {
		c.Failures++
	}
}

// Owns reports whether the monitor has a result for this app and it belongs to this user.
func (m *Monitor) Owns(userID, app int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	e := m.entries[app]
	return e != nil && e.userID == userID
}

// Pending returns the counts for one app that are not saved yet.
func (m *Monitor) Pending(app int64) []store.HourCount {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []store.HourCount
	for _, c := range m.pending[app] {
		out = append(out, *c)
	}
	return out
}

// Flush saves the counts gathered since the last time, and deletes history that has aged out.
// If saving fails, the counts are kept and tried again with the next batch.
func (m *Monitor) Flush(ctx context.Context) {
	m.mu.Lock()
	h := m.history
	taken := m.pending
	m.pending = map[int64]map[int64]*store.HourCount{}
	m.mu.Unlock()
	if h == nil {
		m.restore(taken)
		return
	}

	var batch []store.HourCount
	for _, byHour := range taken {
		for _, c := range byHour {
			batch = append(batch, *c)
		}
	}
	if err := h.AddStatusCounts(ctx, batch); err != nil {
		m.log.Error("save status history", "err", err)
		m.restore(taken)
		return
	}
	if err := h.PruneStatusHours(ctx, m.now().Unix()/3600-keepHours); err != nil {
		m.log.Error("prune status history", "err", err)
	}
}

// restore puts counts back that could not be saved.
func (m *Monitor) restore(taken map[int64]map[int64]*store.HourCount) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for app, byHour := range taken {
		for hour, c := range byHour {
			cur := m.pending[app]
			if cur == nil {
				cur = map[int64]*store.HourCount{}
				m.pending[app] = cur
			}
			if existing := cur[hour]; existing != nil {
				existing.Checks += c.Checks
				existing.Failures += c.Failures
			} else {
				cur[hour] = c
			}
		}
	}
}

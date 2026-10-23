package server

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/dbohry/bifrost/internal/status"
	"github.com/dbohry/bifrost/internal/store"
)

// What a user may choose for how often to check and how long to wait.
const (
	minInterval = 10 * time.Second
	maxInterval = 24 * time.Hour
	minTimeout  = 1 * time.Second
	maxTimeout  = 60 * time.Second
)

type resultJSON struct {
	State       string `json:"state"`
	Since       int64  `json:"since"`
	LastChecked int64  `json:"lastChecked"`
	LatencyMS   int64  `json:"latencyMs"`
	Reason      string `json:"reason"`
	Address     string `json:"address"`
}

// getStatus returns the latest result of every app of the signed-in user that
// has its check on and has been looked at. The keys are app ids.
func (s *server) getStatus(w http.ResponseWriter, r *http.Request, user store.User, _ store.Session) {
	out := map[string]resultJSON{}
	for id, res := range s.monitor.For(user.ID) {
		out[strconv.FormatInt(id, 10)] = resultJSON{
			State: string(res.State), Since: res.Since.Unix(), LastChecked: res.LastChecked.Unix(),
			LatencyMS: res.Latency.Milliseconds(), Reason: res.Reason, Address: res.Address,
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"statuses": out})
}

type statusSettingsJSON struct {
	// Interval and Timeout are in seconds.
	Interval int64 `json:"interval"`
	Timeout  int64 `json:"timeout"`
}

func (s *server) getStatusSettings(w http.ResponseWriter, r *http.Request, user store.User, _ store.Session) {
	v, err := s.store.GetStatusSettings(r.Context(), user.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]statusSettingsJSON{"settings": {Interval: int64(v.Interval / time.Second), Timeout: int64(v.Timeout / time.Second)}})
}

func (s *server) setStatusSettings(w http.ResponseWriter, r *http.Request, user store.User, _ store.Session) {
	var in statusSettingsJSON
	if !decode(w, r, &in) {
		return
	}
	v := store.StatusSettings{Interval: time.Duration(in.Interval) * time.Second, Timeout: time.Duration(in.Timeout) * time.Second}
	switch {
	case in.Interval < int64(minInterval/time.Second) || in.Interval > int64(maxInterval/time.Second):
		writeError(w, http.StatusBadRequest, "Check every must be between 10 seconds and 24 hours.")
		return
	case in.Timeout < int64(minTimeout/time.Second) || in.Timeout > int64(maxTimeout/time.Second):
		writeError(w, http.StatusBadRequest, "Give up after must be between 1 and 60 seconds.")
		return
	case v.Timeout > v.Interval:
		writeError(w, http.StatusBadRequest, "Give up after cannot be longer than the time between checks.")
		return
	}
	if err := s.store.SetStatusSettings(r.Context(), user.ID, v); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]statusSettingsJSON{"settings": in})
}

type testBody struct {
	URL      string `json:"url"`
	Strict   bool   `json:"strict"`
	Insecure bool   `json:"insecure"`
}

// testStatus looks at an address once, right now, so the form can say whether
// the settings work before they are saved.
func (s *server) testStatus(w http.ResponseWriter, r *http.Request, user store.User, _ store.Session) {
	var in testBody
	if !decode(w, r, &in) {
		return
	}
	address, msg := cleanURL(in.URL)
	if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	key := strconv.FormatInt(user.ID, 10)
	if ok, _ := s.testByUser.Allow(key); !ok {
		writeError(w, http.StatusTooManyRequests, "Too many tests. Wait a minute and try again.")
		return
	}
	s.testByUser.Fail(key) // each test counts, so the limit is on how many per minute

	settings, err := s.store.GetStatusSettings(r.Context(), user.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	o := status.Check(r.Context(), address, settings.Timeout, in.Strict, in.Insecure)
	writeJSON(w, http.StatusOK, map[string]any{
		"up": o.Up, "latencyMs": o.Latency.Milliseconds(), "code": o.Code, "reason": o.Reason,
		"address": address,
	})
}

// historyBars is how many hours the card shows, oldest first, ending with the current hour.
const historyBars = 24

type hourJSON struct {
	Start    int64 `json:"start"` // Unix time the hour began
	Checks   int64 `json:"checks"`
	Failures int64 `json:"failures"`
}

// statusHistory returns the last 24 hours of checks for one app, as 24 hourly counts
// and the share of checks that passed. An app that is not the user's has no history.
func (s *server) statusHistory(w http.ResponseWriter, r *http.Request, user store.User, _ store.Session) {
	id, ok := appID(w, r)
	if !ok {
		return
	}
	current := s.now().Unix() / 3600
	first := current - historyBars + 1

	saved, err := s.store.StatusHours(r.Context(), user.ID, id, first)
	if err != nil {
		serverError(w, err)
		return
	}
	byHour := map[int64]*hourJSON{}
	for h := first; h <= current; h++ {
		byHour[h] = &hourJSON{Start: h * 3600}
	}
	add := func(c store.HourCount) {
		if b := byHour[c.Hour]; b != nil {
			b.Checks += c.Checks
			b.Failures += c.Failures
		}
	}
	for _, c := range saved {
		add(c)
	}
	// Counts waiting to be saved belong to the latest hours. Only apps the user owns get this far:
	// the monitor's results are per user.
	if len(saved) > 0 || s.monitor.Owns(user.ID, id) {
		for _, c := range s.monitor.Pending(id) {
			add(c)
		}
	}

	out := make([]hourJSON, 0, historyBars)
	var checks, failures int64
	for h := first; h <= current; h++ {
		out = append(out, *byHour[h])
		checks += byHour[h].Checks
		failures += byHour[h].Failures
	}
	var uptime *float64
	if checks > 0 {
		v := math.Round(float64(checks-failures)/float64(checks)*1000) / 10
		uptime = &v
	}
	writeJSON(w, http.StatusOK, map[string]any{"hours": out, "uptime": uptime})
}

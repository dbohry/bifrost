package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Defaults for a user who has never changed the status settings.
const (
	DefaultCheckInterval = 60 * time.Second
	DefaultCheckTimeout  = 5 * time.Second
)

// StatusSettings is how often a user's apps are checked, and how long to wait
// for an answer.
type StatusSettings struct {
	Interval time.Duration
	Timeout  time.Duration
}

// GetStatusSettings returns a user's settings, or the defaults if they have none.
func (s *Store) GetStatusSettings(ctx context.Context, userID int64) (StatusSettings, error) {
	var interval, timeout int64
	err := s.db.QueryRowContext(ctx,
		`SELECT interval_seconds, timeout_seconds FROM status_settings WHERE user_id = ?`, userID).
		Scan(&interval, &timeout)
	if errors.Is(err, sql.ErrNoRows) {
		return StatusSettings{Interval: DefaultCheckInterval, Timeout: DefaultCheckTimeout}, nil
	}
	if err != nil {
		return StatusSettings{}, err
	}
	return StatusSettings{Interval: time.Duration(interval) * time.Second, Timeout: time.Duration(timeout) * time.Second}, nil
}

// SetStatusSettings saves a user's settings. The caller has validated them.
func (s *Store) SetStatusSettings(ctx context.Context, userID int64, v StatusSettings) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO status_settings (user_id, interval_seconds, timeout_seconds) VALUES (?, ?, ?)
		 ON CONFLICT (user_id) DO UPDATE SET
		   interval_seconds = excluded.interval_seconds, timeout_seconds = excluded.timeout_seconds`,
		userID, int64(v.Interval/time.Second), int64(v.Timeout/time.Second))
	return err
}

// CheckTarget is one app that wants to be checked, with its owner's settings.
type CheckTarget struct {
	AppID    int64
	UserID   int64
	URL      string // the address to look at: the check address, or else the app's own
	Strict   bool
	Insecure bool
	Interval time.Duration
	Timeout  time.Duration
}

// CheckTargets lists every app, of every user, that has its status check on.
func (s *Store) CheckTargets(ctx context.Context) ([]CheckTarget, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT a.id, a.user_id, CASE WHEN a.check_url <> '' THEN a.check_url ELSE a.url END,
		        a.check_strict, a.check_insecure,
		        COALESCE(t.interval_seconds, ?), COALESCE(t.timeout_seconds, ?)
		   FROM apps a LEFT JOIN status_settings t ON t.user_id = a.user_id
		  WHERE a.check_enabled = 1`,
		int64(DefaultCheckInterval/time.Second), int64(DefaultCheckTimeout/time.Second))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CheckTarget
	for rows.Next() {
		var t CheckTarget
		var interval, timeout int64
		if err := rows.Scan(&t.AppID, &t.UserID, &t.URL, &t.Strict, &t.Insecure, &interval, &timeout); err != nil {
			return nil, err
		}
		t.Interval = time.Duration(interval) * time.Second
		t.Timeout = time.Duration(timeout) * time.Second
		out = append(out, t)
	}
	return out, rows.Err()
}

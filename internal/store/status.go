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

// HourCount is how many checks were made for one app in one hour, and how many failed.
type HourCount struct {
	AppID    int64
	Hour     int64 // Unix time divided by 3600
	Checks   int64
	Failures int64
}

// AddStatusCounts adds counts to the hours they belong to, in one transaction. An app that
// has been deleted since is skipped.
func (s *Store) AddStatusCounts(ctx context.Context, counts []HourCount) error {
	if len(counts) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, c := range counts {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO status_hours (app_id, hour, checks, failures)
			 SELECT id, ?, ?, ? FROM apps WHERE id = ?
			 ON CONFLICT (app_id, hour) DO UPDATE SET
			   checks = checks + excluded.checks, failures = failures + excluded.failures`,
			c.Hour, c.Checks, c.Failures, c.AppID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// StatusHours returns the saved counts for one of a user's apps from hour `since` on. An app
// that is not the user's gives nothing.
func (s *Store) StatusHours(ctx context.Context, userID, appID, since int64) ([]HourCount, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT h.app_id, h.hour, h.checks, h.failures
		   FROM status_hours h JOIN apps a ON a.id = h.app_id
		  WHERE a.id = ? AND a.user_id = ? AND h.hour >= ? ORDER BY h.hour`,
		appID, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []HourCount
	for rows.Next() {
		var c HourCount
		if err := rows.Scan(&c.AppID, &c.Hour, &c.Checks, &c.Failures); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// PruneStatusHours deletes counts for hours before `before`.
func (s *Store) PruneStatusHours(ctx context.Context, before int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM status_hours WHERE hour < ?`, before)
	return err
}

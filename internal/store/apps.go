package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// App is one tile on a user's dashboard.
type App struct {
	ID        int64
	UserID    int64
	Name      string
	URL       string
	Category  string
	Favourite bool
	NewTab    bool
	// Icon is the catalogue id of the icon to show, or "" for the first letter of the name.
	Icon string
	// Status check: whether to look at the app now and then, at which address (empty for the
	// app's own), whether only a 2xx answer counts, and whether a self-signed certificate is fine.
	CheckEnabled  bool
	CheckURL      string
	CheckStrict   bool
	CheckInsecure bool
	Position      int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// AppInput is what a user supplies when adding or editing an app. The caller
// has already validated and cleaned it.
type AppInput struct {
	Name      string
	URL       string
	Category  string
	Favourite bool
	NewTab    bool
	Icon      string

	CheckEnabled  bool
	CheckURL      string
	CheckStrict   bool
	CheckInsecure bool
}

const appColumns = `id, user_id, name, url, category, favourite, new_tab, icon, check_enabled, check_url, check_strict, check_insecure, position, created_at, updated_at`

func scanApp(row interface{ Scan(...any) error }) (App, error) {
	var a App
	var created, updated int64
	err := row.Scan(&a.ID, &a.UserID, &a.Name, &a.URL, &a.Category, &a.Favourite, &a.NewTab, &a.Icon, &a.CheckEnabled, &a.CheckURL, &a.CheckStrict, &a.CheckInsecure, &a.Position, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return App{}, ErrNotFound
	}
	if err != nil {
		return App{}, err
	}
	a.CreatedAt = time.Unix(created, 0).UTC()
	a.UpdatedAt = time.Unix(updated, 0).UTC()
	return a, nil
}

// ListApps returns a user's apps in dashboard order.
func (s *Store) ListApps(ctx context.Context, userID int64) ([]App, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+appColumns+` FROM apps WHERE user_id = ? ORDER BY position, id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps := []App{}
	for rows.Next() {
		a, err := scanApp(rows)
		if err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, rows.Err()
}

// A category typed in a different case from one the user already has reuses
// the existing spelling, so "media" and "Media" do not become two groups.
// The ? placeholders are: user, category, excluded app id (0 for none), category.
const reuseCategory = `COALESCE(
	(SELECT category FROM apps WHERE user_id = ? AND category = ? COLLATE NOCASE AND id <> ? LIMIT 1), ?)`

// CreateApp adds an app at the end of the user's dashboard.
func (s *Store) CreateApp(ctx context.Context, userID int64, in AppInput, now time.Time) (App, error) {
	return scanApp(s.db.QueryRowContext(ctx,
		`INSERT INTO apps (user_id, name, url, category, favourite, new_tab, icon,
		                   check_enabled, check_url, check_strict, check_insecure, position, created_at, updated_at)
		 VALUES (?, ?, ?, `+reuseCategory+`, ?, ?, ?, ?, ?, ?, ?,
		         (SELECT COALESCE(MAX(position), 0) + 1 FROM apps WHERE user_id = ?), ?, ?)
		 RETURNING `+appColumns,
		userID, in.Name, in.URL,
		userID, in.Category, 0, in.Category,
		in.Favourite, in.NewTab, in.Icon,
		in.CheckEnabled, in.CheckURL, in.CheckStrict, in.CheckInsecure,
		userID, now.Unix(), now.Unix()))
}

// UpdateApp replaces an app's details. It returns ErrNotFound when the app
// does not exist or belongs to someone else.
func (s *Store) UpdateApp(ctx context.Context, userID, id int64, in AppInput, now time.Time) (App, error) {
	return scanApp(s.db.QueryRowContext(ctx,
		`UPDATE apps SET name = ?, url = ?, category = `+reuseCategory+`, favourite = ?, new_tab = ?, icon = ?,
		        check_enabled = ?, check_url = ?, check_strict = ?, check_insecure = ?, updated_at = ?
		  WHERE id = ? AND user_id = ?
		 RETURNING `+appColumns,
		in.Name, in.URL,
		userID, in.Category, id, in.Category,
		in.Favourite, in.NewTab, in.Icon,
		in.CheckEnabled, in.CheckURL, in.CheckStrict, in.CheckInsecure,
		now.Unix(), id, userID))
}

// ErrOrderMismatch means a new order did not list exactly the user's apps,
// for example because the page was out of date.
var ErrOrderMismatch = errors.New("order does not match the user's apps")

// ReorderApps sets the dashboard order. ids must list every one of the user's
// apps exactly once, first to last. The whole change happens or none of it.
func (s *Store) ReorderApps(ctx context.Context, userID int64, ids []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `SELECT id FROM apps WHERE user_id = ?`, userID)
	if err != nil {
		return err
	}
	owned := map[int64]bool{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		owned[id] = true
	}
	if err := rows.Close(); err != nil {
		return err
	}

	if len(ids) != len(owned) {
		return ErrOrderMismatch
	}
	seen := make(map[int64]bool, len(ids))
	for _, id := range ids {
		if !owned[id] || seen[id] {
			return ErrOrderMismatch
		}
		seen[id] = true
	}

	for i, id := range ids {
		if _, err := tx.ExecContext(ctx, `UPDATE apps SET position = ? WHERE id = ? AND user_id = ?`, i+1, id, userID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// DeleteApp removes an app. It returns ErrNotFound when the app does not exist
// or belongs to someone else.
func (s *Store) DeleteApp(ctx context.Context, userID, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM apps WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// ImportResult says what ImportApps did.
type ImportResult struct {
	Added   int
	Skipped int
	Removed int
}

// ImportApps adds apps to the end of a user's dashboard, in the order given.
// With replace, the user's existing apps are deleted first. Otherwise an app
// whose name (in any case) and address are already on the dashboard, or earlier
// in the same list, is skipped. The whole import happens or none of it.
func (s *Store) ImportApps(ctx context.Context, userID int64, apps []AppInput, replace bool, now time.Time) (ImportResult, error) {
	var res ImportResult
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}
	defer tx.Rollback()

	if replace {
		r, err := tx.ExecContext(ctx, `DELETE FROM apps WHERE user_id = ?`, userID)
		if err != nil {
			return res, err
		}
		n, _ := r.RowsAffected()
		res.Removed = int(n)
	}

	for _, in := range apps {
		var exists bool
		err := tx.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM apps WHERE user_id = ? AND name = ? COLLATE NOCASE AND url = ?)`,
			userID, in.Name, in.URL).Scan(&exists)
		if err != nil {
			return res, err
		}
		if exists {
			res.Skipped++
			continue
		}
		_, err = tx.ExecContext(ctx,
			`INSERT INTO apps (user_id, name, url, category, favourite, new_tab, icon,
			                   check_enabled, check_url, check_strict, check_insecure, position, created_at, updated_at)
			 VALUES (?, ?, ?, `+reuseCategory+`, ?, ?, ?, ?, ?, ?, ?,
			         (SELECT COALESCE(MAX(position), 0) + 1 FROM apps WHERE user_id = ?), ?, ?)`,
			userID, in.Name, in.URL,
			userID, in.Category, 0, in.Category,
			in.Favourite, in.NewTab, in.Icon,
			in.CheckEnabled, in.CheckURL, in.CheckStrict, in.CheckInsecure,
			userID, now.Unix(), now.Unix())
		if err != nil {
			return res, err
		}
		res.Added++
	}
	return res, tx.Commit()
}

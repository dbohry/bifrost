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
	Position  int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AppInput is what a user supplies when adding or editing an app. The caller
// has already validated and cleaned it.
type AppInput struct {
	Name      string
	URL       string
	Category  string
	Favourite bool
	NewTab    bool
}

const appColumns = `id, user_id, name, url, category, favourite, new_tab, position, created_at, updated_at`

func scanApp(row interface{ Scan(...any) error }) (App, error) {
	var a App
	var created, updated int64
	err := row.Scan(&a.ID, &a.UserID, &a.Name, &a.URL, &a.Category, &a.Favourite, &a.NewTab, &a.Position, &created, &updated)
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
		`INSERT INTO apps (user_id, name, url, category, favourite, new_tab, position, created_at, updated_at)
		 VALUES (?, ?, ?, `+reuseCategory+`, ?, ?,
		         (SELECT COALESCE(MAX(position), 0) + 1 FROM apps WHERE user_id = ?), ?, ?)
		 RETURNING `+appColumns,
		userID, in.Name, in.URL,
		userID, in.Category, 0, in.Category,
		in.Favourite, in.NewTab, userID, now.Unix(), now.Unix()))
}

// UpdateApp replaces an app's details. It returns ErrNotFound when the app
// does not exist or belongs to someone else.
func (s *Store) UpdateApp(ctx context.Context, userID, id int64, in AppInput, now time.Time) (App, error) {
	return scanApp(s.db.QueryRowContext(ctx,
		`UPDATE apps SET name = ?, url = ?, category = `+reuseCategory+`, favourite = ?, new_tab = ?, updated_at = ?
		  WHERE id = ? AND user_id = ?
		 RETURNING `+appColumns,
		in.Name, in.URL,
		userID, in.Category, id, in.Category,
		in.Favourite, in.NewTab, now.Unix(), id, userID))
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

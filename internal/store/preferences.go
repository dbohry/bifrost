package store

import (
	"context"
	"database/sql"
	"errors"
)

// Preferences is how a user has set up the look of their dashboard.
type Preferences struct {
	// Mode is "light", "dark" or "system" (follow the device).
	Mode       string
	LightTheme string
	DarkTheme  string
	// Logo is "app", "arches" or "spectrum".
	Logo string
}

// GetPreferences returns a user's saved preferences, or ErrNotFound if they
// have never saved any.
func (s *Store) GetPreferences(ctx context.Context, userID int64) (Preferences, error) {
	var p Preferences
	err := s.db.QueryRowContext(ctx,
		`SELECT mode, light_theme, dark_theme, logo FROM preferences WHERE user_id = ?`, userID).
		Scan(&p.Mode, &p.LightTheme, &p.DarkTheme, &p.Logo)
	if errors.Is(err, sql.ErrNoRows) {
		return Preferences{}, ErrNotFound
	}
	return p, err
}

// SetPreferences saves a user's preferences, replacing any earlier ones.
func (s *Store) SetPreferences(ctx context.Context, userID int64, p Preferences) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO preferences (user_id, mode, light_theme, dark_theme, logo) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (user_id) DO UPDATE SET
		   mode = excluded.mode, light_theme = excluded.light_theme,
		   dark_theme = excluded.dark_theme, logo = excluded.logo`,
		userID, p.Mode, p.LightTheme, p.DarkTheme, p.Logo)
	return err
}

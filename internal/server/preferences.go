package server

import (
	"errors"
	"net/http"

	"github.com/dbohry/bifrost/internal/store"
	"github.com/dbohry/bifrost/internal/themes"
)

var (
	modes = map[string]bool{"light": true, "dark": true, "system": true}
	logos = map[string]bool{"app": true, "arches": true, "spectrum": true}
)

// defaultPreferences is what a user has until they change something: follow
// the device, with the Bifrost themes and the app icon.
var defaultPreferences = store.Preferences{
	Mode:       "system",
	LightTheme: "bifrost-light",
	DarkTheme:  "bifrost-dark",
	Logo:       "app",
}

type preferencesJSON struct {
	Mode       string `json:"mode"`
	LightTheme string `json:"lightTheme"`
	DarkTheme  string `json:"darkTheme"`
	Logo       string `json:"logo"`
}

func toPreferencesJSON(p store.Preferences) preferencesJSON {
	return preferencesJSON{Mode: p.Mode, LightTheme: p.LightTheme, DarkTheme: p.DarkTheme, Logo: p.Logo}
}

// listThemes returns every theme. Themes are not secret, and the sign-in screen
// has no user yet, so this needs no session.
func (s *server) listThemes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string][]themes.Theme{"themes": s.themes})
}

func (s *server) theme(id string) (themes.Theme, bool) {
	for _, t := range s.themes {
		if t.ID == id {
			return t, true
		}
	}
	return themes.Theme{}, false
}

func (s *server) getPreferences(w http.ResponseWriter, r *http.Request, user store.User, _ store.Session) {
	p, err := s.store.GetPreferences(r.Context(), user.ID)
	if errors.Is(err, store.ErrNotFound) {
		p = defaultPreferences
	} else if err != nil {
		serverError(w, err)
		return
	}

	// A theme that has since gone away falls back to the default for its mode.
	if t, ok := s.theme(p.LightTheme); !ok || t.Mode != themes.Light {
		p.LightTheme = defaultPreferences.LightTheme
	}
	if t, ok := s.theme(p.DarkTheme); !ok || t.Mode != themes.Dark {
		p.DarkTheme = defaultPreferences.DarkTheme
	}
	writeJSON(w, http.StatusOK, map[string]preferencesJSON{"preferences": toPreferencesJSON(p)})
}

func (s *server) setPreferences(w http.ResponseWriter, r *http.Request, user store.User, _ store.Session) {
	var in preferencesJSON
	if !decode(w, r, &in) {
		return
	}
	if !modes[in.Mode] {
		writeError(w, http.StatusBadRequest, `Mode must be "light", "dark" or "system".`)
		return
	}
	if t, ok := s.theme(in.LightTheme); !ok || t.Mode != themes.Light {
		writeError(w, http.StatusBadRequest, "Choose one of the light themes for light mode.")
		return
	}
	if t, ok := s.theme(in.DarkTheme); !ok || t.Mode != themes.Dark {
		writeError(w, http.StatusBadRequest, "Choose one of the dark themes for dark mode.")
		return
	}
	if !logos[in.Logo] {
		writeError(w, http.StatusBadRequest, `Logo must be "app", "arches" or "spectrum".`)
		return
	}

	p := store.Preferences{Mode: in.Mode, LightTheme: in.LightTheme, DarkTheme: in.DarkTheme, Logo: in.Logo}
	if err := s.store.SetPreferences(r.Context(), user.ID, p); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]preferencesJSON{"preferences": toPreferencesJSON(p)})
}

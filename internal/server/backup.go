package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dbohry/bifrost/internal/store"
	"github.com/dbohry/bifrost/internal/themes"
)

const (
	exportFormat  = "bifrost-export"
	exportVersion = 1
	// maxImportApps and maxImportBody keep a mistaken or hostile file from
	// filling the database or the memory of a small machine.
	maxImportApps = 2000
	maxImportBody = 2 << 20
)

// exportApp is one app in an export file. It leaves out ids, positions and
// dates: the order of the list is the order of the dashboard.
type exportApp struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Category  string `json:"category"`
	Favourite bool   `json:"favourite"`
	NewTab    bool   `json:"newTab"`
	Icon      string `json:"icon"`

	CheckEnabled  bool   `json:"checkEnabled,omitempty"`
	CheckURL      string `json:"checkUrl,omitempty"`
	CheckStrict   bool   `json:"checkStrict,omitempty"`
	CheckInsecure bool   `json:"checkInsecure,omitempty"`
}

// exportFile is the whole file. It holds one person's dashboard and
// appearance. It never holds passwords or anything about other users.
type exportFile struct {
	Format      string           `json:"format"`
	Version     int              `json:"version"`
	ExportedAt  string           `json:"exportedAt,omitempty"`
	Apps        []exportApp      `json:"apps"`
	Preferences *preferencesJSON `json:"preferences,omitempty"`
}

func (s *server) exportData(w http.ResponseWriter, r *http.Request, user store.User, _ store.Session) {
	apps, err := s.store.ListApps(r.Context(), user.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	file := exportFile{Format: exportFormat, Version: exportVersion, ExportedAt: s.now().UTC().Format("2006-01-02T15:04:05Z"), Apps: make([]exportApp, len(apps))}
	for i, a := range apps {
		file.Apps[i] = exportApp{
			Name: a.Name, URL: a.URL, Category: a.Category, Favourite: a.Favourite, NewTab: a.NewTab, Icon: a.Icon,
			CheckEnabled: a.CheckEnabled, CheckURL: a.CheckURL, CheckStrict: a.CheckStrict, CheckInsecure: a.CheckInsecure,
		}
	}
	if p, err := s.store.GetPreferences(r.Context(), user.ID); err == nil {
		pj := toPreferencesJSON(p)
		file.Preferences = &pj
	} else if !errors.Is(err, store.ErrNotFound) {
		serverError(w, err)
		return
	}

	w.Header().Set("Content-Disposition", `attachment; filename="bifrost-export.json"`)
	writeJSON(w, http.StatusOK, file)
}

type importBody struct {
	exportFile
	// Mode is "merge" (add to the dashboard) or "replace" (swap it for the file's).
	Mode string `json:"mode"`
	// Appearance says whether to also restore the file's appearance.
	Appearance bool `json:"appearance"`
}

type importResult struct {
	Added             int  `json:"added"`
	Skipped           int  `json:"skipped"`
	Removed           int  `json:"removed"`
	AppearanceApplied bool `json:"appearanceApplied"`
}

func (s *server) importData(w http.ResponseWriter, r *http.Request, user store.User, _ store.Session) {
	var in importBody
	if !decodeLimit(w, r, &in, maxImportBody) {
		return
	}
	if in.Format != exportFormat {
		writeError(w, http.StatusBadRequest, "That file is not a Bifrost export.")
		return
	}
	if in.Version != exportVersion {
		writeError(w, http.StatusBadRequest, "That export comes from a different version of Bifrost and cannot be read.")
		return
	}
	if in.Mode != "merge" && in.Mode != "replace" {
		writeError(w, http.StatusBadRequest, `Mode must be "merge" or "replace".`)
		return
	}
	if len(in.Apps) > maxImportApps {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("A file can hold at most %d apps.", maxImportApps))
		return
	}

	apps := make([]store.AppInput, len(in.Apps))
	for i, a := range in.Apps {
		newTab := a.NewTab
		clean, msg := cleanApp(appBody{
			Name: a.Name, URL: a.URL, Category: a.Category, Favourite: a.Favourite, NewTab: &newTab, Icon: a.Icon,
			CheckEnabled: a.CheckEnabled, CheckURL: a.CheckURL, CheckStrict: a.CheckStrict, CheckInsecure: a.CheckInsecure,
		})
		if msg != "" {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("App %d (%q) cannot be imported: %s", i+1, a.Name, msg))
			return
		}
		// An icon this version does not know falls back to the first letter.
		if clean.Icon != "" && !s.catalog.Has(clean.Icon) {
			clean.Icon = ""
		}
		apps[i] = clean
	}

	res, err := s.store.ImportApps(r.Context(), user.ID, apps, in.Mode == "replace", s.now())
	if err != nil {
		serverError(w, err)
		return
	}
	out := importResult{Added: res.Added, Skipped: res.Skipped, Removed: res.Removed}

	if in.Appearance && in.Preferences != nil {
		if p, ok := s.validPreferences(*in.Preferences); ok {
			if err := s.store.SetPreferences(r.Context(), user.ID, p); err != nil {
				serverError(w, err)
				return
			}
			out.AppearanceApplied = true
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// validPreferences reports whether preferences name a known mode, logo and
// themes of the right kind.
func (s *server) validPreferences(in preferencesJSON) (store.Preferences, bool) {
	// A file exported before backgrounds existed has none.
	if in.Background == "" {
		in.Background = defaultPreferences.Background
	}
	l, lok := s.theme(in.LightTheme)
	d, dok := s.theme(in.DarkTheme)
	if !modes[in.Mode] || !logos[in.Logo] || !backgrounds[in.Background] || !lok || !dok || l.Mode != themes.Light || d.Mode != themes.Dark {
		return store.Preferences{}, false
	}
	return store.Preferences{Mode: in.Mode, LightTheme: in.LightTheme, DarkTheme: in.DarkTheme, Logo: in.Logo, Background: in.Background}, true
}

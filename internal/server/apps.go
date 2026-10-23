package server

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dbohry/bifrost/internal/store"
)

const (
	maxNameChars     = 40
	maxCategoryChars = 40
	maxURLBytes      = 2048
	defaultCategory  = "Other"
)

type appJSON struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	Category  string `json:"category"`
	Favourite bool   `json:"favourite"`
	NewTab    bool   `json:"newTab"`
	Position  int64  `json:"position"`
}

func toAppJSON(a store.App) appJSON {
	return appJSON{ID: a.ID, Name: a.Name, URL: a.URL, Category: a.Category, Favourite: a.Favourite, NewTab: a.NewTab, Position: a.Position}
}

// appBody is what the browser sends to add or change an app. newTab defaults
// to true when left out.
type appBody struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Category  string `json:"category"`
	Favourite bool   `json:"favourite"`
	NewTab    *bool  `json:"newTab"`
}

func (s *server) listApps(w http.ResponseWriter, r *http.Request, user store.User, _ store.Session) {
	apps, err := s.store.ListApps(r.Context(), user.ID)
	if err != nil {
		serverError(w, err)
		return
	}
	out := make([]appJSON, len(apps))
	for i, a := range apps {
		out[i] = toAppJSON(a)
	}
	writeJSON(w, http.StatusOK, map[string][]appJSON{"apps": out})
}

func (s *server) createApp(w http.ResponseWriter, r *http.Request, user store.User, _ store.Session) {
	in, ok := readAppInput(w, r)
	if !ok {
		return
	}
	app, err := s.store.CreateApp(r.Context(), user.ID, in, s.now())
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]appJSON{"app": toAppJSON(app)})
}

func (s *server) updateApp(w http.ResponseWriter, r *http.Request, user store.User, _ store.Session) {
	id, ok := appID(w, r)
	if !ok {
		return
	}
	in, ok := readAppInput(w, r)
	if !ok {
		return
	}
	app, err := s.store.UpdateApp(r.Context(), user.ID, id, in, s.now())
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "App not found.")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]appJSON{"app": toAppJSON(app)})
}

func (s *server) deleteApp(w http.ResponseWriter, r *http.Request, user store.User, _ store.Session) {
	id, ok := appID(w, r)
	if !ok {
		return
	}
	err := s.store.DeleteApp(r.Context(), user.ID, id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "App not found.")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// appID reads the {id} in the path. A malformed id is the same as an unknown
// one.
func appID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusNotFound, "App not found.")
		return 0, false
	}
	return id, true
}

func readAppInput(w http.ResponseWriter, r *http.Request) (store.AppInput, bool) {
	var body appBody
	if !decode(w, r, &body) {
		return store.AppInput{}, false
	}
	in, msg := cleanApp(body)
	if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return store.AppInput{}, false
	}
	return in, true
}

// cleanApp validates and tidies what the user typed. It returns a message the
// user can read when something is wrong.
func cleanApp(b appBody) (store.AppInput, string) {
	name := strings.TrimSpace(b.Name)
	if name == "" {
		return store.AppInput{}, "Enter a name for the app."
	}
	if utf8.RuneCountInString(name) > maxNameChars || hasControl(name) {
		return store.AppInput{}, "The name must be 40 characters or fewer."
	}

	address, msg := cleanURL(b.URL)
	if msg != "" {
		return store.AppInput{}, msg
	}

	category := strings.Join(strings.Fields(b.Category), " ")
	if category == "" {
		category = defaultCategory
	}
	if utf8.RuneCountInString(category) > maxCategoryChars || hasControl(category) {
		return store.AppInput{}, "The category must be 40 characters or fewer."
	}

	newTab := true
	if b.NewTab != nil {
		newTab = *b.NewTab
	}
	return store.AppInput{Name: name, URL: address, Category: category, Favourite: b.Favourite, NewTab: newTab}, ""
}

// cleanURL accepts http and https addresses only, so a tile can never be a
// javascript: or data: link. An address typed without a scheme gets http://.
func cleanURL(raw string) (string, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "Enter the address of the app."
	}
	if len(raw) > maxURLBytes {
		return "", "That address is too long."
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || hasControl(raw) || strings.ContainsAny(raw, " ") {
		return "", "That does not look like a valid address."
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", "The address must start with http:// or https://."
	}
	if u.Hostname() == "" {
		return "", "That does not look like a valid address."
	}
	if u.User != nil {
		return "", "Do not put a username or password in the address."
	}
	return u.String(), ""
}

func hasControl(s string) bool {
	return strings.IndexFunc(s, unicode.IsControl) >= 0
}

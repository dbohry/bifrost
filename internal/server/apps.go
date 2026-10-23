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
	// Icon is the catalogue id of the app's icon, or "" for none.
	Icon     string `json:"icon"`
	Position int64  `json:"position"`
	// Status check settings. CheckURL is empty when the app's own address is used.
	CheckEnabled  bool   `json:"checkEnabled"`
	CheckURL      string `json:"checkUrl"`
	CheckStrict   bool   `json:"checkStrict"`
	CheckInsecure bool   `json:"checkInsecure"`
}

func toAppJSON(a store.App) appJSON {
	return appJSON{
		ID: a.ID, Name: a.Name, URL: a.URL, Category: a.Category, Favourite: a.Favourite, NewTab: a.NewTab, Icon: a.Icon, Position: a.Position,
		CheckEnabled: a.CheckEnabled, CheckURL: a.CheckURL, CheckStrict: a.CheckStrict, CheckInsecure: a.CheckInsecure,
	}
}

// appBody is what the browser sends to add or change an app. newTab defaults
// to true when left out.
type appBody struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Category  string `json:"category"`
	Favourite bool   `json:"favourite"`
	NewTab    *bool  `json:"newTab"`
	Icon      string `json:"icon"`

	CheckEnabled  bool   `json:"checkEnabled"`
	CheckURL      string `json:"checkUrl"`
	CheckStrict   bool   `json:"checkStrict"`
	CheckInsecure bool   `json:"checkInsecure"`
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
	in, ok := s.readAppInput(w, r)
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
	in, ok := s.readAppInput(w, r)
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

// reorderApps saves a new dashboard order. The body lists every app id, first
// to last. If the list does not match the user's apps (another tab added or
// removed one) it is refused and the page should reload.
func (s *server) reorderApps(w http.ResponseWriter, r *http.Request, user store.User, _ store.Session) {
	var in struct {
		IDs []int64 `json:"ids"`
	}
	if !decode(w, r, &in) {
		return
	}
	err := s.store.ReorderApps(r.Context(), user.ID, in.IDs)
	if errors.Is(err, store.ErrOrderMismatch) {
		writeError(w, http.StatusConflict, "Your apps changed in another window. Reload the page and try again.")
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

func (s *server) readAppInput(w http.ResponseWriter, r *http.Request) (store.AppInput, bool) {
	var body appBody
	if !decode(w, r, &body) {
		return store.AppInput{}, false
	}
	in, msg := cleanApp(body)
	if msg == "" && in.Icon != "" && !s.catalog.Has(in.Icon) {
		msg = "That icon is not in the catalogue."
	}
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

	// The address to check is optional. Empty means "use the app's own address".
	checkURL := strings.TrimSpace(b.CheckURL)
	if checkURL != "" {
		var msg string
		if checkURL, msg = cleanURL(checkURL); msg != "" {
			return store.AppInput{}, "The address to check: " + lowerFirst(msg)
		}
	}
	return store.AppInput{
		Name: name, URL: address, Category: category, Favourite: b.Favourite, NewTab: newTab, Icon: strings.TrimSpace(b.Icon),
		CheckEnabled: b.CheckEnabled, CheckURL: checkURL, CheckStrict: b.CheckStrict, CheckInsecure: b.CheckInsecure,
	}, ""
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

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

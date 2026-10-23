package server

import (
	"net/http"
	"strings"

	"github.com/dbohry/bifrost/internal/catalog"
)

// listCatalog returns the preset apps. They are the same for everyone and are
// not secret, so no session is needed.
func (s *server) listCatalog(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"categories": catalog.Categories,
		"apps":       s.catalog.Entries,
	})
}

// catalogIcon serves one preset's icon. The id is only ever looked up in the
// catalogue, never used as a file name, so nothing outside it can be reached.
func (s *server) catalogIcon(w http.ResponseWriter, r *http.Request) {
	id, ok := strings.CutSuffix(r.PathValue("file"), ".svg")
	svg, found := s.catalog.Icon(id)
	if !ok || !found {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	h := w.Header()
	h.Set("Content-Type", "image/svg+xml")
	h.Set("Cache-Control", "public, max-age=86400")
	// The icon is meant for an <img>. If someone opens its address directly, the
	// browser must not run anything in it.
	h.Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
	w.Write(svg)
}

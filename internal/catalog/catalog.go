// Package catalog is the built-in list of popular self-hosted apps that can be
// added to a dashboard with one click, each with its icon.
//
// The list and the icons are embedded in the binary. A preset only fills in the
// add-app form: the address is always the user's own, and the app is saved like
// any other afterwards.
package catalog

import (
	"embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

//go:embed catalog.json icons/*.svg
var files embed.FS

// Categories are the groups the presets are sorted into, in display order.
var Categories = []string{"Media", "Productivity", "Network", "Monitoring", "Development"}

// Entry is one preset app.
type Entry struct {
	// ID is also the name of its icon file, and what an app stores to say which icon it uses.
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	// Port is the one the app usually listens on, shown as a hint. 80 means the plain web port.
	Port int `json:"port"`
}

// MaxIconBytes keeps the icons from bloating the binary.
const MaxIconBytes = 50 << 10

var idRE = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,39}$`)

// Catalog is the loaded list, with icons looked up by id.
type Catalog struct {
	Entries []Entry
	icons   map[string][]byte
}

// Load reads and checks the built-in catalogue. It fails on anything wrong, so a
// bad entry or icon is found by the tests rather than by a user.
func Load() (*Catalog, error) {
	raw, err := files.ReadFile("catalog.json")
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, fmt.Errorf("catalog.json: %w", err)
	}

	known := map[string]bool{}
	for _, c := range Categories {
		known[c] = true
	}

	c := &Catalog{Entries: entries, icons: map[string][]byte{}}
	names := map[string]bool{}
	for _, e := range entries {
		switch {
		case !idRE.MatchString(e.ID):
			return nil, fmt.Errorf("catalog entry %q: the id must be lowercase letters, digits and dashes", e.ID)
		case c.icons[e.ID] != nil:
			return nil, fmt.Errorf("catalog entry %q appears twice", e.ID)
		case e.Name == "" || utf8.RuneCountInString(e.Name) > 40:
			return nil, fmt.Errorf("catalog entry %q: the name must be 1 to 40 characters", e.ID)
		case names[strings.ToLower(e.Name)]:
			return nil, fmt.Errorf("catalog entry %q: the name %q is used twice", e.ID, e.Name)
		case !known[e.Category]:
			return nil, fmt.Errorf("catalog entry %q: unknown category %q", e.ID, e.Category)
		case e.Port < 1 || e.Port > 65535:
			return nil, fmt.Errorf("catalog entry %q: port %d is out of range", e.ID, e.Port)
		}
		names[strings.ToLower(e.Name)] = true

		svg, err := files.ReadFile("icons/" + e.ID + ".svg")
		if err != nil {
			return nil, fmt.Errorf("catalog entry %q has no icon: %w", e.ID, err)
		}
		if err := CheckIcon(svg); err != nil {
			return nil, fmt.Errorf("icon %q: %w", e.ID, err)
		}
		c.icons[e.ID] = svg
	}

	// An icon nobody uses is a mistake too.
	dir, _ := files.ReadDir("icons")
	for _, f := range dir {
		id := strings.TrimSuffix(f.Name(), ".svg")
		if c.icons[id] == nil {
			return nil, fmt.Errorf("icons/%s is not used by any catalogue entry", f.Name())
		}
	}
	return c, nil
}

// Has reports whether id names a preset.
func (c *Catalog) Has(id string) bool {
	_, ok := c.icons[id]
	return ok
}

// Icon returns the SVG for a preset.
func (c *Catalog) Icon(id string) ([]byte, bool) {
	svg, ok := c.icons[id]
	return svg, ok
}

// unsafe lists what must not appear in an icon. An icon is only ever shown in
// an <img>, where none of this would run, but the file is also served at its
// own address, and an SVG opened directly can run scripts. So icons are kept
// inert as well as sandboxed when served.
var unsafe = regexp.MustCompile(`(?i)<script|<foreignobject|<iframe|<object|<embed|<image|<audio|<video|<!entity|<!doctype|<\?xml-stylesheet|javascript:|data:|\son[a-z]+\s*=|@import|url\(\s*["']?\s*(https?:|//)|(xlink:)?href\s*=\s*["']\s*[^#"']`)

// CheckIcon refuses anything that is not a small, plain SVG: no scripts, event
// handlers, external references or embedded images. Links to a shape inside the
// same file (href="#a") are fine.
func CheckIcon(svg []byte) error {
	if len(svg) == 0 || len(svg) > MaxIconBytes {
		return fmt.Errorf("size %d bytes: must be between 1 byte and %d", len(svg), MaxIconBytes)
	}
	if !strings.Contains(string(svg[:min(len(svg), 400)]), "<svg") {
		return fmt.Errorf("does not start like an SVG")
	}
	if m := unsafe.Find(svg); m != nil {
		return fmt.Errorf("contains %q, which an icon must not", m)
	}
	return nil
}

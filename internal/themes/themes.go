// Package themes defines Bifrost's colour themes.
//
// A theme is a small JSON file: an id, a name, a mode (light or dark) and ten
// colours. Components only ever use those ten, so a new theme is ten values.
// The bundled themes are embedded in the binary.
package themes

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

//go:embed bundled/*.json
var bundled embed.FS

// Mode says whether a theme is meant for a light or a dark screen.
const (
	Light = "light"
	Dark  = "dark"
)

// Tokens are the ten colours every theme provides.
var Tokens = []string{"bg", "surface", "border", "text", "muted", "accent", "onAccent", "ok", "bad", "iconBg"}

// Theme is a colour theme.
type Theme struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Mode   string            `json:"mode"`
	Colors map[string]string `json:"colors"`
}

var (
	idRE  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,31}$`)
	hexRE = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
)

// Parse reads and checks a theme file. It checks the shape only: whether the
// colours are readable is a separate question, see ContrastIssues.
func Parse(data []byte) (Theme, error) {
	var t Theme
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&t); err != nil {
		return Theme{}, fmt.Errorf("not a valid theme file: %w", err)
	}

	if !idRE.MatchString(t.ID) {
		return Theme{}, fmt.Errorf("id %q must be 2 to 32 lowercase letters, digits or dashes", t.ID)
	}
	if n := utf8.RuneCountInString(t.Name); n < 1 || n > 40 {
		return Theme{}, fmt.Errorf("name must be 1 to 40 characters")
	}
	if t.Mode != Light && t.Mode != Dark {
		return Theme{}, fmt.Errorf(`mode must be "light" or "dark", not %q`, t.Mode)
	}

	known := map[string]bool{}
	for _, token := range Tokens {
		known[token] = true
		value, ok := t.Colors[token]
		if !ok {
			return Theme{}, fmt.Errorf("missing colour %q", token)
		}
		if !hexRE.MatchString(value) {
			return Theme{}, fmt.Errorf("colour %q must look like #1a2b3c, not %q", token, value)
		}
		t.Colors[token] = strings.ToLower(value)
	}
	for token := range t.Colors {
		if !known[token] {
			return Theme{}, fmt.Errorf("unknown colour %q", token)
		}
	}
	return t, nil
}

// Bundled returns the themes that ship with Bifrost: light ones first, then dark,
// in the order of their file names.
func Bundled() ([]Theme, error) {
	entries, err := bundled.ReadDir("bundled")
	if err != nil {
		return nil, err
	}
	var out []Theme
	seen := map[string]string{}
	for _, e := range entries {
		data, err := bundled.ReadFile("bundled/" + e.Name())
		if err != nil {
			return nil, err
		}
		t, err := Parse(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		if other, dup := seen[t.ID]; dup {
			return nil, fmt.Errorf("%s: id %q is also used by %s", e.Name(), t.ID, other)
		}
		seen[t.ID] = e.Name()
		out = append(out, t)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Mode == Light && out[j].Mode == Dark })
	return out, nil
}

// MinTextContrast is the contrast ratio text needs against its background
// (WCAG AA for normal text). MinUIContrast is the lower bar for icons and borders.
const (
	MinTextContrast = 4.5
	MinUIContrast   = 3.0
)

// ContrastIssues lists the readability problems of a theme, or nothing if it
// is fine. A theme with issues still works, so a caller may show these as
// warnings rather than refuse the theme.
func (t Theme) ContrastIssues() []string {
	var issues []string
	check := func(fg, bg string, min float64, what string) {
		if ratio := Contrast(t.Colors[fg], t.Colors[bg]); ratio < min {
			issues = append(issues, fmt.Sprintf("%s: %s on %s is %.2f:1, needs %.1f:1", what, fg, bg, ratio, min))
		}
	}

	for _, bg := range []string{"bg", "surface"} {
		check("text", bg, MinTextContrast, "text")
		check("muted", bg, MinTextContrast, "secondary text")
		check("accent", bg, MinUIContrast, "accent")
	}
	check("onAccent", "accent", MinTextContrast, "text on the accent colour")
	check("ok", "surface", MinUIContrast, "status colour")
	check("bad", "surface", MinUIContrast, "status colour")
	return issues
}

// Contrast is the WCAG contrast ratio between two #rrggbb colours, from 1 to 21.
func Contrast(a, b string) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func luminance(hex string) float64 {
	channel := func(s string) float64 {
		n, _ := strconv.ParseUint(s, 16, 8)
		c := float64(n) / 255
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(hex[1:3]) + 0.7152*channel(hex[3:5]) + 0.0722*channel(hex[5:7])
}

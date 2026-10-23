package themes

import (
	"math"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestBundledThemesLoad(t *testing.T) {
	list, err := Bundled()
	if err != nil {
		t.Fatal(err)
	}

	byID := map[string]Theme{}
	names := map[string]bool{}
	for _, th := range list {
		if _, dup := byID[th.ID]; dup {
			t.Errorf("duplicate id %q", th.ID)
		}
		if names[th.Name] {
			t.Errorf("duplicate name %q", th.Name)
		}
		byID[th.ID], names[th.Name] = th, true
	}

	// The themes the design promises.
	want := map[string]string{
		"bifrost-light": Light, "solarized-light": Light,
		"bifrost-dark": Dark, "dracula": Dark, "nord": Dark, "gruvbox": Dark, "solarized-dark": Dark,
	}
	for id, mode := range want {
		if th, ok := byID[id]; !ok || th.Mode != mode {
			t.Errorf("theme %q: got %+v, want mode %s", id, th, mode)
		}
	}
	if len(list) != len(want) {
		t.Errorf("%d bundled themes, want %d", len(list), len(want))
	}

	// Light themes are listed before dark ones.
	seenDark := false
	for _, th := range list {
		if th.Mode == Dark {
			seenDark = true
		} else if seenDark {
			t.Errorf("light theme %q is listed after a dark one", th.ID)
		}
	}
}

// This is the rule from the design: every bundled theme must be readable.
func TestEveryBundledThemeIsReadable(t *testing.T) {
	list, _ := Bundled()
	for _, th := range list {
		for _, issue := range th.ContrastIssues() {
			t.Errorf("%s: %s", th.ID, issue)
		}
	}
}

// The sign-in screen is drawn before any theme is fetched, using the colours in
// theme.css. They must be the Bifrost themes, or the page would change colour
// a moment after it loads.
func TestStylesheetFallbackMatchesTheDefaultThemes(t *testing.T) {
	css, err := os.ReadFile("../../frontend/src/theme.css")
	if err != nil {
		t.Skip("frontend sources not available:", err)
	}
	list, _ := Bundled()
	for _, th := range list {
		if th.ID != "bifrost-light" && th.ID != "bifrost-dark" {
			continue
		}
		block := regexp.MustCompile(`(?s)/\* theme:` + th.ID + ` \*/(.*?)\}`).FindSubmatch(css)
		if block == nil {
			t.Errorf("theme.css has no /* theme:%s */ block", th.ID)
			continue
		}
		for token, value := range th.Colors {
			name := regexp.MustCompile(`[A-Z]`).ReplaceAllStringFunc(token, func(s string) string { return "-" + strings.ToLower(s) })
			if !strings.Contains(string(block[1]), "--"+name+": "+value+";") {
				t.Errorf("theme.css %s: want --%s: %s;", th.ID, name, value)
			}
		}
	}
}

func TestContrast(t *testing.T) {
	cases := []struct {
		a, b string
		want float64
	}{
		{"#000000", "#ffffff", 21},
		{"#ffffff", "#000000", 21},
		{"#336699", "#336699", 1},
		{"#767676", "#ffffff", 4.54}, // the lightest grey that passes AA on white
		{"#777777", "#ffffff", 4.48},
	}
	for _, c := range cases {
		if got := Contrast(c.a, c.b); math.Abs(got-c.want) > 0.02 {
			t.Errorf("Contrast(%s, %s) = %.2f, want %.2f", c.a, c.b, got, c.want)
		}
	}
}

func valid() Theme {
	list, _ := Bundled()
	for _, th := range list {
		if th.ID == "bifrost-light" {
			return th
		}
	}
	panic("no bifrost-light")
}

func TestContrastIssuesAreFound(t *testing.T) {
	th := valid()
	th.Colors["muted"] = "#c0c0c0" // pale grey on near-white
	th.Colors["onAccent"] = th.Colors["accent"]
	issues := strings.Join(th.ContrastIssues(), "\n")
	for _, want := range []string{"secondary text: muted on bg", "secondary text: muted on surface", "text on the accent colour"} {
		if !strings.Contains(issues, want) {
			t.Errorf("missing %q in:\n%s", want, issues)
		}
	}
	if strings.Contains(issues, "text: text on") {
		t.Errorf("the main text colour was fine:\n%s", issues)
	}
}

const good = `{"id":"my-theme","name":"My theme","mode":"dark","colors":{
 "bg":"#101010","surface":"#202020","border":"#303030","text":"#F0F0F0","muted":"#B0B0B0",
 "accent":"#88AAFF","onAccent":"#101010","ok":"#44CC88","bad":"#FF8866","iconBg":"#282828"}}`

func TestParse(t *testing.T) {
	th, err := Parse([]byte(good))
	if err != nil {
		t.Fatal(err)
	}
	if th.Colors["text"] != "#f0f0f0" {
		t.Errorf("colours should be lower-cased, got %q", th.Colors["text"])
	}

	bad := map[string]string{
		"not json":       `nope`,
		"unknown field":  strings.Replace(good, `"mode"`, `"extra":1,"mode"`, 1),
		"bad id":         strings.Replace(good, `"my-theme"`, `"My Theme"`, 1),
		"short id":       strings.Replace(good, `"my-theme"`, `"x"`, 1),
		"empty name":     strings.Replace(good, `"My theme"`, `""`, 1),
		"bad mode":       strings.Replace(good, `"dark"`, `"sepia"`, 1),
		"missing colour": strings.Replace(good, `"iconBg":"#282828"`, `"icon":"#282828"`, 1),
		"unknown colour": strings.Replace(good, `"ok":`, `"extra":"#000000","ok":`, 1),
		"short hex":      strings.Replace(good, `"#101010","surface"`, `"#111","surface"`, 1),
		"colour name":    strings.Replace(good, `"#88AAFF"`, `"blue"`, 1),
		"css injection":  strings.Replace(good, `"#88AAFF"`, `"#88aaff;}body{display:none"`, 1),
	}
	for name, data := range bad {
		if _, err := Parse([]byte(data)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

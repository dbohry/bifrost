package catalog

import (
	"strings"
	"testing"
)

func load(t *testing.T) *Catalog {
	t.Helper()
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestTheCatalogueLoads(t *testing.T) {
	c := load(t)
	if n := len(c.Entries); n < 300 {
		t.Errorf("%d entries, want at least 300", n)
	}

	perCategory := map[string]int{}
	for _, e := range c.Entries {
		perCategory[e.Category]++
	}
	for _, name := range Categories {
		if perCategory[name] == 0 {
			t.Errorf("category %q has no apps", name)
		}
	}
}

func TestEveryEntryHasAnIconAndNothingElseDoes(t *testing.T) {
	c := load(t) // Load itself fails on a missing icon, or on one no entry uses.
	for _, e := range c.Entries {
		svg, ok := c.Icon(e.ID)
		if !ok || !c.Has(e.ID) || len(svg) == 0 {
			t.Errorf("%s: no icon", e.ID)
		}
	}
	if c.Has("nope") || c.Has("") || c.Has("../catalog") {
		t.Error("an id that is not in the catalogue was found")
	}
	if _, ok := c.Icon("nope"); ok {
		t.Error("an icon was returned for an unknown id")
	}
}

func TestIconsStayReasonablySmall(t *testing.T) {
	c := load(t)
	total := 0
	for _, e := range c.Entries {
		svg, _ := c.Icon(e.ID)
		total += len(svg)
	}
	if total > 1536<<10 {
		t.Errorf("icons add up to %d KB, which is a lot to carry in the binary", total>>10)
	}
}

func TestCheckIcon(t *testing.T) {
	good := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 8 8"><defs><circle id="a" r="2"/></defs><use xlink:href="#a" href="#a"/><path d="M0 0h8" style="fill:#fff"/></svg>`
	if err := CheckIcon([]byte(good)); err != nil {
		t.Errorf("a plain icon was refused: %v", err)
	}

	bad := map[string]string{
		"empty":             "",
		"not an svg":        "<html></html>",
		"script":            `<svg><script>alert(1)</script></svg>`,
		"script, any case":  `<svg><ScRiPt>alert(1)</ScRiPt></svg>`,
		"event handler":     `<svg onload="alert(1)"></svg>`,
		"event handler 2":   `<svg><rect onclick = "x()"/></svg>`,
		"javascript link":   `<svg><a href="javascript:alert(1)"><rect/></a></svg>`,
		"external link":     `<svg><use href="https://evil.example/x.svg#a"/></svg>`,
		"external xlink":    `<svg><use xlink:href='http://evil.example/x.svg#a'/></svg>`,
		"protocol-relative": `<svg><use href="//evil.example/x.svg#a"/></svg>`,
		"relative file":     `<svg><use href="other.svg#a"/></svg>`,
		"foreignObject":     `<svg><foreignObject><div/></foreignObject></svg>`,
		"embedded image":    `<svg><image href="data:image/png;base64,AAAA"/></svg>`,
		"data uri":          `<svg><rect style="fill:url(data:image/svg+xml;base64,AAAA)"/></svg>`,
		"css import":        `<svg><style>@import "https://evil.example/a.css";</style></svg>`,
		"css external url":  `<svg><rect style="fill:url(https://evil.example/a.svg#x)"/></svg>`,
		"entity":            `<svg><!ENTITY x SYSTEM "file:///etc/passwd"></svg>`,
		"iframe":            `<svg><iframe src="x"/></svg>`,
		"too big":           "<svg>" + strings.Repeat("a", MaxIconBytes) + "</svg>",
	}
	for name, svg := range bad {
		if err := CheckIcon([]byte(svg)); err == nil {
			t.Errorf("%s: this should have been refused", name)
		}
	}
}

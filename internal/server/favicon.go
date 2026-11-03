package server

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dbohry/bifrost/internal/store"
)

const (
	faviconTimeout   = 5 * time.Second
	maxFaviconBytes  = 256 << 10
	maxFaviconPage   = 512 << 10
	maxFaviconTries  = 4
	faviconHitTTL    = 24 * time.Hour
	faviconMissTTL   = time.Hour
	faviconCacheSize = 512
)

type faviconEntry struct {
	data    []byte // nil means "this site has no icon"
	ctype   string
	expires time.Time
}

// faviconCache remembers what was found per site, including that nothing was,
// so a dashboard full of tiles does not fetch the same pages over and over.
type faviconCache struct {
	mu sync.Mutex
	m  map[string]faviconEntry
}

func (c *faviconCache) get(key string, now time.Time) (faviconEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.m[key]
	if !ok || now.After(e.expires) {
		delete(c.m, key)
		return faviconEntry{}, false
	}
	return e, true
}

func (c *faviconCache) put(key string, e faviconEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.m == nil {
		c.m = map[string]faviconEntry{}
	}
	if len(c.m) >= faviconCacheSize {
		for k := range c.m { // drop an arbitrary entry to stay bounded
			delete(c.m, k)
			break
		}
	}
	c.m[key] = e
}

// siteFavicon serves the favicon of the site at ?url=. It answers 404 when the
// site has none, which is the page's cue to show the first letter instead.
func (s *server) siteFavicon(w http.ResponseWriter, r *http.Request, _ store.User, _ store.Session) {
	address, msg := cleanURL(r.URL.Query().Get("url"))
	if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	u, _ := url.Parse(address)
	key := u.Scheme + "://" + u.Host

	entry, ok := s.favicons.get(key, s.now())
	if !ok {
		entry = findFavicon(r.Context(), u)
		ttl := faviconHitTTL
		if entry.data == nil {
			ttl = faviconMissTTL
		}
		entry.expires = s.now().Add(ttl)
		s.favicons.put(key, entry)
	}

	h := w.Header()
	if entry.data == nil {
		h.Set("Cache-Control", "private, max-age=3600")
		writeError(w, http.StatusNotFound, "This site has no icon.")
		return
	}
	h.Set("Content-Type", entry.ctype)
	h.Set("Cache-Control", "private, max-age=86400")
	// Meant for an <img>; opened directly, nothing in it may run.
	h.Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
	w.Write(entry.data)
}

var (
	linkTag = regexp.MustCompile(`(?is)<link\b[^>]*>`)
	attrRe  = regexp.MustCompile(`(?is)([a-z-]+)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))`)
)

// pageIcons lists the icon addresses a page declares, best first: a plain
// "icon" before an "apple-touch-icon".
func pageIcons(html string, base *url.URL) []string {
	var icons, touch []string
	for _, tag := range linkTag.FindAllString(html, -1) {
		var rel, href string
		for _, m := range attrRe.FindAllStringSubmatch(tag, -1) {
			val := m[2] + m[3] + m[4]
			switch strings.ToLower(m[1]) {
			case "rel":
				rel = strings.ToLower(val)
			case "href":
				href = strings.TrimSpace(val)
			}
		}
		if href == "" || strings.HasPrefix(strings.ToLower(href), "data:") {
			continue
		}
		ref, err := url.Parse(href)
		if err != nil {
			continue
		}
		abs := base.ResolveReference(ref)
		if abs.Scheme != "http" && abs.Scheme != "https" {
			continue
		}
		switch {
		case strings.Contains(rel, "apple-touch-icon"):
			touch = append(touch, abs.String())
		case strings.Contains(rel, "icon"):
			icons = append(icons, abs.String())
		}
	}
	return append(icons, touch...)
}

// findFavicon reads the site's page for icon links, then falls back to
// /favicon.ico. An empty entry means nothing usable was found.
func findFavicon(ctx context.Context, site *url.URL) faviconEntry {
	ctx, cancel := context.WithTimeout(ctx, 3*faviconTimeout)
	defer cancel()

	client := &http.Client{
		Timeout: faviconTimeout,
		Transport: &http.Transport{
			Proxy:             nil,
			DialContext:       (&net.Dialer{Timeout: faviconTimeout}).DialContext,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // a public icon on a home server, often self-signed
			DisableKeepAlives: true,
		},
	}
	defer client.CloseIdleConnections()

	var candidates []string
	page := *site
	if page.Path == "" {
		page.Path = "/"
	}
	if body, _, final := fetch(ctx, client, page.String(), maxFaviconPage); body != nil {
		base := page
		if final != nil {
			base = *final
		}
		candidates = pageIcons(string(body), &base)
	}
	candidates = append(candidates, (&url.URL{Scheme: site.Scheme, Host: site.Host, Path: "/favicon.ico"}).String())
	if len(candidates) > maxFaviconTries {
		candidates = candidates[:maxFaviconTries]
	}

	for _, c := range candidates {
		body, ctype, _ := fetch(ctx, client, c, maxFaviconBytes)
		if kind := imageType(body, ctype); kind != "" {
			return faviconEntry{data: body, ctype: kind}
		}
	}
	return faviconEntry{}
}

// fetch GETs one address and returns the body (nil on any failure or non-200),
// its Content-Type and the address finally reached after redirects.
func fetch(ctx context.Context, client *http.Client, address string, limit int64) ([]byte, string, *url.URL) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, "", nil
	}
	req.Header.Set("User-Agent", "Bifrost-icon-lookup")
	res, err := client.Do(req)
	if err != nil {
		return nil, "", nil
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, "", nil
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, limit+1))
	if err != nil || int64(len(body)) > limit || len(body) == 0 {
		return nil, "", nil
	}
	return body, res.Header.Get("Content-Type"), res.Request.URL
}

// imageType says which image type the bytes are, or "" if they are not an
// image the browser should be given. HTML is never accepted, whatever the
// server called it.
func imageType(body []byte, declared string) string {
	if len(body) == 0 {
		return ""
	}
	head := body
	if len(head) > 512 {
		head = head[:512]
	}
	sniffed := http.DetectContentType(head)
	switch {
	case strings.HasPrefix(sniffed, "image/"):
		return sniffed
	case strings.HasPrefix(sniffed, "text/xml"), strings.HasPrefix(sniffed, "text/plain"):
		if strings.Contains(strings.ToLower(string(head)), "<svg") {
			return "image/svg+xml"
		}
	}
	// .ico files are sometimes not recognised by sniffing.
	if strings.HasPrefix(strings.ToLower(declared), "image/x-icon") || strings.HasPrefix(strings.ToLower(declared), "image/vnd.microsoft.icon") {
		if len(body) > 4 && body[0] == 0 && body[1] == 0 && body[2] == 1 && body[3] == 0 {
			return "image/x-icon"
		}
	}
	return ""
}

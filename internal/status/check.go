// Package status looks at the apps people have asked to be checked, and
// remembers whether each one answered.
package status

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"
)

// maxRedirects is how many redirects a check follows. An app that sends you to
// its login page is up; one that redirects for ever is not worth following.
const maxRedirects = 2

// Outcome is what one look at an address found.
type Outcome struct {
	// Up says whether the app counts as online.
	Up bool
	// Latency is how long the answer took. It is zero when there was none.
	Latency time.Duration
	// Code is the HTTP status of the answer, or 0 if there was none.
	Code int
	// Reason says in plain words why the app is not up. It is empty when it is.
	Reason string
}

// Check makes one plain GET request to address and decides whether the app is up.
// Any answer counts, unless strict is set, which wants a 2xx one. A self-signed
// certificate is refused unless insecure is set. No cookies or credentials are
// sent, and the body is never read.
func Check(ctx context.Context, address string, timeout time.Duration, strict, insecure bool) Outcome {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy:                 nil, // checks go straight to the app, never through a proxy
			DialContext:           (&net.Dialer{Timeout: timeout}).DialContext,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: insecure}, //nolint:gosec // the user asked for it, per app
			TLSHandshakeTimeout:   timeout,
			ResponseHeaderTimeout: timeout,
			DisableKeepAlives:     true,
		},
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) > maxRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	defer client.CloseIdleConnections()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return Outcome{Reason: "The address is not valid"}
	}
	req.Header.Set("User-Agent", "Bifrost-status-check")
	req.Header.Set("Accept", "*/*")

	start := time.Now()
	res, err := client.Do(req)
	if err != nil {
		return Outcome{Reason: reason(err, timeout)}
	}
	res.Body.Close()
	latency := time.Since(start)

	if strict && (res.StatusCode < 200 || res.StatusCode > 299) {
		return Outcome{Latency: latency, Code: res.StatusCode, Reason: fmt.Sprintf("Answered with HTTP %d", res.StatusCode)}
	}
	return Outcome{Up: true, Latency: latency, Code: res.StatusCode}
}

// reason turns a network error into a sentence a person can act on.
func reason(err error, timeout time.Duration) string {
	var dns *net.DNSError
	var cert *tls.CertificateVerificationError
	var timeoutErr interface{ Timeout() bool }
	switch {
	case errors.As(err, &cert):
		return "The certificate is not trusted"
	case errors.As(err, &dns):
		return "Could not find the host"
	case errors.Is(err, syscall.ECONNREFUSED):
		return "Connection refused"
	case errors.Is(err, syscall.EHOSTUNREACH), errors.Is(err, syscall.ENETUNREACH):
		return "Host unreachable"
	case errors.Is(err, context.DeadlineExceeded), errors.As(err, &timeoutErr) && timeoutErr.Timeout():
		return fmt.Sprintf("Timed out after %s", wholeSeconds(timeout))
	default:
		return "No answer"
	}
}

func wholeSeconds(d time.Duration) string {
	return fmt.Sprintf("%d s", int(d.Round(time.Second)/time.Second))
}

package status

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func serve(t *testing.T, h http.HandlerFunc) string {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestAnyAnswerCountsUnlessStrict(t *testing.T) {
	ok := serve(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	broken := serve(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) })
	login := serve(t, func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401) })
	ctx := context.Background()

	if o := Check(ctx, ok, time.Second, false, false); !o.Up || o.Code != 200 || o.Reason != "" {
		t.Errorf("200: %+v", o)
	}
	if o := Check(ctx, broken, time.Second, false, false); !o.Up || o.Code != 503 {
		t.Errorf("503 with any answer: %+v", o)
	}
	if o := Check(ctx, login, time.Second, false, false); !o.Up {
		t.Errorf("a login page counts as up: %+v", o)
	}
	if o := Check(ctx, broken, time.Second, true, false); o.Up || o.Reason != "Answered with HTTP 503" || o.Code != 503 {
		t.Errorf("503 with strict: %+v", o)
	}
	if o := Check(ctx, ok, time.Second, true, false); !o.Up {
		t.Errorf("200 with strict: %+v", o)
	}
}

func TestCheckSendsOnlyAPlainRequest(t *testing.T) {
	var got *http.Request
	u := serve(t, func(w http.ResponseWriter, r *http.Request) { got = r.Clone(context.Background()) })
	Check(context.Background(), u, time.Second, false, false)
	if got == nil || got.Method != "GET" || got.Header.Get("Authorization") != "" || got.Header.Get("Cookie") != "" ||
		!strings.HasPrefix(got.UserAgent(), "Bifrost") {
		t.Errorf("request = %+v", got)
	}
}

func TestRedirectsAreFollowedAFewTimesOnly(t *testing.T) {
	var u string
	u = serve(t, func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, u+"/again", http.StatusFound) })
	o := Check(context.Background(), u, time.Second, false, false)
	if !o.Up || o.Code != http.StatusFound {
		t.Errorf("an endless redirect still counts as an answer, stopped at the last one: %+v", o)
	}
	if o := Check(context.Background(), u, time.Second, true, false); o.Up {
		t.Errorf("strict: a redirect is not a success: %+v", o)
	}
}

func TestWhyAnAppIsDown(t *testing.T) {
	// nothing is listening
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := l.Addr().String()
	l.Close()
	if o := Check(context.Background(), "http://"+addr, time.Second, false, false); o.Up || o.Reason != "Connection refused" {
		t.Errorf("refused: %+v", o)
	}

	slow := serve(t, func(w http.ResponseWriter, r *http.Request) { time.Sleep(1500 * time.Millisecond) })
	start := time.Now()
	o := Check(context.Background(), slow, time.Second, false, false)
	if o.Up || o.Reason != "Timed out after 1 s" || time.Since(start) > 1400*time.Millisecond {
		t.Errorf("timeout: %+v after %s", o, time.Since(start))
	}

	if o := Check(context.Background(), "http://does-not-exist.invalid", time.Second, false, false); o.Up || o.Reason == "" {
		t.Errorf("unknown host: %+v", o)
	}
	if o := Check(context.Background(), "http://bad host", time.Second, false, false); o.Up || o.Reason == "" {
		t.Errorf("bad address: %+v", o)
	}
}

func TestSelfSignedCertificatesNeedAskingFor(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	if o := Check(context.Background(), srv.URL, time.Second, false, false); o.Up || o.Reason != "The certificate is not trusted" {
		t.Errorf("default: %+v", o)
	}
	if o := Check(context.Background(), srv.URL, time.Second, false, true); !o.Up {
		t.Errorf("insecure: %+v", o)
	}
}

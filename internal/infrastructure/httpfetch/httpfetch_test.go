package httpfetch

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const capBytes = 16

func serve(t *testing.T, h http.HandlerFunc) (*Client, string) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	u, _ := url.Parse(srv.URL)
	return New(srv.Client(), capBytes, u.Host), srv.URL
}

// FR-PRV-002: EONET labels its JSON application/rss+xml (measured 2026-09-23).
// The client hands the body on whatever the label says; Response carries no
// Content-Type, so no adapter can refuse a body for its label.
func TestFRPRV002_ABodyIsHandedOnWhateverItsContentType(t *testing.T) {
	t.Parallel()
	const body = `{"events":[]}`
	c, base := serve(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(body))
	})
	got, err := c.Get(context.Background(), base+"/api/v3/events", "")
	if err != nil || string(got.Body) != body {
		t.Errorf("Get = %q, %v; want the JSON body whatever its label", got.Body, err)
	}
}

func TestFRPRV004_SendsValidatorAndReadsNotModified(t *testing.T) {
	t.Parallel()
	c, base := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Accept = %q", r.Header.Get("Accept"))
		}
		if r.Header.Get("If-Modified-Since") == "Wed, 23 Sep 2026 11:52:04 GMT" {
			w.Header().Set("Last-Modified", "Wed, 23 Sep 2026 11:52:04 GMT")
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("Last-Modified", "Wed, 23 Sep 2026 11:52:04 GMT")
		_, _ = w.Write([]byte(`{"a":1}`))
	})
	first, err := c.Get(context.Background(), base, "")
	if err != nil || string(first.Body) != `{"a":1}` || first.NotModified || first.LastModified == "" {
		t.Fatalf("first = %+v, %v", first, err)
	}
	second, err := c.Get(context.Background(), base, first.LastModified)
	if err != nil || !second.NotModified || second.Body != nil {
		t.Errorf("second = %+v, %v", second, err)
	}
}

func TestFRPRV011_RefusesABodyOverTheCap(t *testing.T) {
	t.Parallel()
	c, base := serve(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", capBytes+1)))
	})
	if _, err := c.Get(context.Background(), base, ""); !errors.Is(err, ErrTooLarge) {
		t.Errorf("err = %v, want ErrTooLarge", err)
	}
}

func TestBodyAtTheCapIsAccepted(t *testing.T) {
	t.Parallel()
	c, base := serve(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", capBytes)))
	})
	if r, err := c.Get(context.Background(), base, ""); err != nil || len(r.Body) != capBytes {
		t.Errorf("r = %d bytes, %v", len(r.Body), err)
	}
}

func TestNFRPRIV001_RefusesAHostNotAllowed(t *testing.T) {
	t.Parallel()
	c := New(http.DefaultClient, capBytes, "earthquake.usgs.gov")
	if _, err := c.Get(context.Background(), "https://example.com/x", ""); !errors.Is(err, ErrHostNotAllowed) {
		t.Errorf("err = %v, want ErrHostNotAllowed", err)
	}
}

func TestReportsAServerError(t *testing.T) {
	t.Parallel()
	c, base := serve(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) })
	if _, err := c.Get(context.Background(), base, ""); !errors.Is(err, ErrStatus) {
		t.Errorf("err = %v, want ErrStatus", err)
	}
}

func TestReportsUnusableURLsAndDeadHosts(t *testing.T) {
	t.Parallel()
	c := New(http.DefaultClient, capBytes, "127.0.0.1:1")
	if _, err := c.Get(context.Background(), "://bad", ""); err == nil {
		t.Error("an unparseable URL was accepted")
	}
	if _, err := c.Get(context.Background(), "http://127.0.0.1:1/x", ""); err == nil {
		t.Error("a request to a closed port succeeded")
	}
}

type failingBody struct{}

func (failingBody) Read([]byte) (int, error) { return 0, errors.New("connection reset") }
func (failingBody) Close() error             { return nil }

type bodyTransport struct{}

func (bodyTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusOK, Body: failingBody{}, Header: http.Header{}}, nil
}

func TestReportsABodyThatFailsMidRead(t *testing.T) {
	t.Parallel()
	c := New(&http.Client{Transport: bodyTransport{}}, capBytes, "earthquake.usgs.gov")
	if _, err := c.Get(context.Background(), "https://earthquake.usgs.gov/x", ""); err == nil {
		t.Error("a failed read was reported as success")
	}
}

// The Smithsonian feed answers 403 to a JSON-only request, so an adapter may ask
// for what its source serves (FR-PRV-015).
func TestGetAcceptingSendsTheAskedMediaTypes(t *testing.T) {
	t.Parallel()
	c, base := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != AcceptXML {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		_, _ = w.Write([]byte("<rss/>"))
	})
	if _, err := c.GetAccepting(context.Background(), base, "", AcceptXML); err != nil {
		t.Fatalf("asking for XML: %v", err)
	}
	if _, err := c.Get(context.Background(), base, ""); err == nil {
		t.Error("a JSON-only request was not refused by the stand-in server")
	}
}

package httpfetch

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// NFR-PRIV-001: a redirect is held to the allowed hosts. An allowed host that
// answers 302 to another host is refused, not followed; a redirect within the
// allowed host is followed; an endless one stops.
func TestNFRPRIV001_ARedirectIsHeldToTheAllowedHosts(t *testing.T) {
	t.Parallel()
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("elsewhere")) }))
	t.Cleanup(other.Close)
	var allowed *httptest.Server
	allowed = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/away":
			http.Redirect(w, r, other.URL+"/x", http.StatusFound)
		case "/home":
			http.Redirect(w, r, allowed.URL+"/here", http.StatusFound)
		case "/loop":
			http.Redirect(w, r, allowed.URL+"/loop", http.StatusFound)
		default:
			_, _ = w.Write([]byte("here"))
		}
	}))
	t.Cleanup(allowed.Close)
	u, _ := url.Parse(allowed.URL)
	c := New(&http.Client{}, capBytes, u.Host)
	if got, err := c.Get(context.Background(), allowed.URL+"/away", ""); !errors.Is(err, ErrHostNotAllowed) {
		t.Errorf("a redirect to another host gave %q, %v; want ErrHostNotAllowed", got.Body, err)
	}
	if got, err := c.Get(context.Background(), allowed.URL+"/home", ""); err != nil || string(got.Body) != "here" {
		t.Errorf("a redirect within the allowed host gave %q, %v", got.Body, err)
	}
	if _, err := c.Get(context.Background(), allowed.URL+"/loop", ""); err == nil {
		t.Error("an endless redirect was followed without end")
	}
}

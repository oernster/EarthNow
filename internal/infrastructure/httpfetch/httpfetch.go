// Package httpfetch is the one way EarthNow reaches the network: a GET to an
// allowed host, capped in size, conditional when the source supports it
// (REQUIREMENTS.md NFR-PRIV-001, FR-PRV-004, FR-PRV-011).
package httpfetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Sentinel failures, each worded where it is raised.
var (
	ErrHostNotAllowed = errors.New("host not allowed")
	ErrNotHTTPS       = errors.New("not an https address")
	ErrTooLarge       = errors.New("response larger than the size cap")
	ErrStatus         = errors.New("unexpected HTTP status")
)

// Response is a body and what the source said about change.
type Response struct {
	Body         []byte
	NotModified  bool
	LastModified string
}

// Client fetches from a fixed set of hosts.
type Client struct {
	http     *http.Client
	allowed  map[string]bool
	maxBytes int64
}

// maxRedirects is the standard library's own limit, kept when the redirect
// check below replaces its default.
const maxRedirects = 10

// secureScheme is the only scheme the client sends a request over, first hop
// and every redirect alike.
const secureScheme = "https"

// New builds a client that reaches only hosts over https only; it refuses
// bodies over maxBytes. It works on its own copy of httpClient, whose redirects
// are held to the same hosts and scheme. Without that, an allowed host answering
// 302 would hand the request to any host at all (NFR-PRIV-001, reproduced
// 2026-09-24); it could also drop the request to cleartext on its own name,
// where the answer could be rewritten on the path. Each adapter is given a client of its own host, so
// one provider's redirect cannot reach another's host or a hidden layer's.
func New(httpClient *http.Client, maxBytes int64, hosts ...string) *Client {
	allowed := make(map[string]bool, len(hosts))
	for _, h := range hosts {
		allowed[h] = true
	}
	c := &Client{allowed: allowed, maxBytes: maxBytes}
	held := *httpClient
	held.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if err := c.admit(req.URL); err != nil {
			return err
		}
		if len(via) >= maxRedirects {
			return fmt.Errorf("stopped after %d redirects", maxRedirects)
		}
		return nil
	}
	c.http = &held
	return c
}

// admit is the one check every request passes before it is sent, the first
// and each redirect: https, to an allowed host.
func (c *Client) admit(u *url.URL) error {
	if u.Scheme != secureScheme {
		return fmt.Errorf("%w: %s://%s", ErrNotHTTPS, u.Scheme, u.Host)
	}
	if !c.allowed[u.Host] {
		return fmt.Errorf("%w: %s", ErrHostNotAllowed, u.Host)
	}
	return nil
}

// Accept values an adapter may ask for. JSON is what EONET and USGS serve;
// AcceptXML is what the Smithsonian's feed demands, since it answers 403 to a
// request asking only for JSON (measured 2026-09-23).
const (
	AcceptJSON = "application/json"
	AcceptXML  = "application/rss+xml, application/xml, text/xml"
)

// Get fetches rawURL asking for JSON, sending validator as If-Modified-Since when
// not empty.
func (c *Client) Get(ctx context.Context, rawURL, validator string) (Response, error) {
	return c.GetAccepting(ctx, rawURL, validator, AcceptJSON)
}

// GetAccepting is Get asking for the media types in accept.
func (c *Client) GetAccepting(ctx context.Context, rawURL, validator, accept string) (Response, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return Response{}, fmt.Errorf("parsing %q: %w", rawURL, err)
	}
	if err := c.admit(parsed); err != nil {
		return Response{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return Response{}, fmt.Errorf("building request for %s: %w", parsed.Host, err)
	}
	// EONET labels JSON as RSS whatever is asked (measured 2026-09-23); asking
	// still costs nothing and the parser never trusts the answer's label.
	req.Header.Set("Accept", accept)
	if validator != "" {
		req.Header.Set("If-Modified-Since", validator)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("requesting %s: %w", parsed.Host, err)
	}
	defer func() { _ = resp.Body.Close() }()
	switch resp.StatusCode {
	case http.StatusNotModified:
		return Response{NotModified: true, LastModified: resp.Header.Get("Last-Modified")}, nil
	case http.StatusOK:
	default:
		return Response{}, fmt.Errorf("%w: %s answered %d", ErrStatus, parsed.Host, resp.StatusCode)
	}
	// Read one byte past the cap, never more: the size a source claims is not
	// believed; nothing beyond the cap is ever held.
	body, err := io.ReadAll(io.LimitReader(resp.Body, c.maxBytes+1))
	if err != nil {
		return Response{}, fmt.Errorf("reading %s: %w", parsed.Host, err)
	}
	if int64(len(body)) > c.maxBytes {
		return Response{}, fmt.Errorf("%w: %s sent more than %d bytes", ErrTooLarge, parsed.Host, c.maxBytes)
	}
	return Response{Body: body, LastModified: resp.Header.Get("Last-Modified")}, nil
}

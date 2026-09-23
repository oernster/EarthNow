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

// New builds a client that reaches only hosts and refuses bodies over maxBytes.
func New(httpClient *http.Client, maxBytes int64, hosts ...string) *Client {
	allowed := make(map[string]bool, len(hosts))
	for _, h := range hosts {
		allowed[h] = true
	}
	return &Client{http: httpClient, allowed: allowed, maxBytes: maxBytes}
}

// Get fetches rawURL, sending validator as If-Modified-Since when not empty.
func (c *Client) Get(ctx context.Context, rawURL, validator string) (Response, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return Response{}, fmt.Errorf("parsing %q: %w", rawURL, err)
	}
	if !c.allowed[parsed.Host] {
		return Response{}, fmt.Errorf("%w: %s", ErrHostNotAllowed, parsed.Host)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return Response{}, fmt.Errorf("building request for %s: %w", parsed.Host, err)
	}
	// EONET labels JSON as RSS whatever is asked (measured 2026-09-23); asking
	// still costs nothing and the parser never trusts the answer's label.
	req.Header.Set("Accept", "application/json")
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

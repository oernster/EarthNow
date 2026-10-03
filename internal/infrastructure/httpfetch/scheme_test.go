package httpfetch

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"strings"
	"testing"
)

// hop words a URL on host answering a redirect of code to target.
func hop(host, code, target string) string {
	return "https://" + host + "/to?code=" + code + "&url=" + url.QueryEscape(target)
}

// landed is the stand-in's answer at host: the host it was reached as.
func landed(host string) string { return "https://" + host + "/ok" }

// rigCap holds any answer the stand-ins give.
const rigCap = 1 << 10

// redirectCodes are every status the client follows.
var redirectCodes = []string{"301", "302", "303", "307", "308"}

// reached reports whether any dialled address names host.
func reached(dialled []string, host string) bool {
	return slices.ContainsFunc(dialled, func(a string) bool { return strings.HasPrefix(a, host+":") })
}

// NFR-PRIV-001: every redirect code is followed within the allowed hosts and
// refused, before any dial, to a host outside them. A provider's client holds
// its own host only, so its redirect cannot reach a layer host either.
func TestNFRPRIV001_EveryRedirectCodeIsHeldToTheAllowedHosts(t *testing.T) {
	t.Parallel()
	for _, code := range redirectCodes {
		hc, rig := newDialRig(t, redirector())
		c := New(hc, rigCap, eonetHost)
		got, err := c.Get(context.Background(), hop(eonetHost, code, landed(eonetHost)), "")
		if err != nil || string(got.Body) != eonetHost {
			t.Errorf("%s within the host gave %q, %v", code, got.Body, err)
		}
		for _, away := range []string{evilHost, layerHost} {
			if _, err := c.Get(context.Background(), hop(eonetHost, code, landed(away)), ""); !errors.Is(err, ErrHostNotAllowed) {
				t.Errorf("%s to %s gave %v; want ErrHostNotAllowed", code, away, err)
			}
			if reached(rig.addresses(), away) {
				t.Errorf("%s to %s dialled it: %v", code, away, rig.addresses())
			}
		}
	}
}

// A chain through allowed hosts is followed; one that turns outward at its
// last hop is refused there, never dialling the outside host.
func TestNFRPRIV001_AChainIsHeldAtEveryHop(t *testing.T) {
	t.Parallel()
	hc, rig := newDialRig(t, redirector())
	c := New(hc, rigCap, eonetHost, usgsHost)
	through := hop(eonetHost, "302", hop(usgsHost, "307", landed(eonetHost)))
	if got, err := c.Get(context.Background(), through, ""); err != nil || string(got.Body) != eonetHost {
		t.Errorf("a chain through allowed hosts gave %q, %v", got.Body, err)
	}
	out := hop(eonetHost, "301", hop(usgsHost, "308", landed(evilHost)))
	if _, err := c.Get(context.Background(), out, ""); !errors.Is(err, ErrHostNotAllowed) || reached(rig.addresses(), evilHost) {
		t.Errorf("a chain turning outward gave %v, dialled %v", err, rig.addresses())
	}
}

// A redirect from https to plain http on the same allowed host is
// refused before the cleartext dial, so event data cannot be rewritten on the
// path (audit round 2, E-4).
func TestARedirectToPlainHTTPIsRefused(t *testing.T) {
	t.Parallel()
	hc, rig := newDialRig(t, redirector())
	c := New(hc, rigCap, eonetHost)
	for _, code := range redirectCodes {
		got, err := c.Get(context.Background(), hop(eonetHost, code, "http://"+eonetHost+"/ok"), "")
		if !errors.Is(err, ErrNotHTTPS) {
			t.Errorf("%s to http gave %q, %v; want ErrNotHTTPS", code, got.Body, err)
		}
	}
	for _, a := range rig.addresses() {
		if !strings.HasSuffix(a, ":"+httpsPort) {
			t.Errorf("a cleartext address was dialled: %v", rig.addresses())
			break
		}
	}
}

// A first request to a plain http URL is refused before any dial,
// on an allowed host as on any other.
func TestAFirstHopOverPlainHTTPIsRefused(t *testing.T) {
	t.Parallel()
	hc, rig := newDialRig(t, redirector())
	c := New(hc, rigCap, eonetHost)
	for _, raw := range []string{"http://" + eonetHost + "/ok", "ftp://" + eonetHost + "/ok", "//" + eonetHost + "/ok"} {
		if _, err := c.Get(context.Background(), raw, ""); !errors.Is(err, ErrNotHTTPS) {
			t.Errorf("%s gave %v; want ErrNotHTTPS", raw, err)
		}
	}
	if len(rig.addresses()) != 0 {
		t.Errorf("a refused request dialled %v", rig.addresses())
	}
}

// lookalikes resemble the allowed host without being it: user information in
// front, the name as a subdomain or a path, a Cyrillic letter, its punycode,
// another case, a trailing dot and another port. Case, the dot and the port fail
// closed: over-strict, never a leak.
var lookalikes = []string{
	"https://" + eonetHost + "@" + evilHost + "/ok",
	"https://" + eonetHost + "." + evilHost + "/ok",
	"https://" + evilHost + "/" + eonetHost,
	"https://еonet.gsfc.nasa.gov/ok",
	"https://xn--onet-6cd.gsfc.nasa.gov/ok",
	"https://EONET.GSFC.NASA.GOV/ok",
	"https://" + eonetHost + "./ok",
	"https://" + eonetHost + ":8443/ok",
}

// NFR-PRIV-001: a lookalike is refused as a first request and as a redirect
// target; nothing is dialled for it.
func TestNFRPRIV001_LookalikeHostsAreRefused(t *testing.T) {
	t.Parallel()
	hc, rig := newDialRig(t, redirector())
	c := New(hc, rigCap, eonetHost)
	for _, raw := range lookalikes {
		if _, err := c.Get(context.Background(), raw, ""); !errors.Is(err, ErrHostNotAllowed) {
			t.Errorf("first hop %s gave %v; want ErrHostNotAllowed", raw, err)
		}
	}
	if len(rig.addresses()) != 0 {
		t.Fatalf("a refused first hop dialled %v", rig.addresses())
	}
	for _, raw := range lookalikes {
		if _, err := c.Get(context.Background(), hop(eonetHost, "302", raw), ""); !errors.Is(err, ErrHostNotAllowed) {
			t.Errorf("redirect to %s gave %v; want ErrHostNotAllowed", raw, err)
		}
	}
	for _, a := range rig.addresses() {
		if a != eonetHost+":"+httpsPort {
			t.Errorf("a lookalike was dialled: %s", a)
		}
	}
}

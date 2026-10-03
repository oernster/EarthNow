package httpfetch

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// The real host names, so the redirect check runs against what it meets in use.
const (
	eonetHost = "eonet.gsfc.nasa.gov"
	usgsHost  = "earthquake.usgs.gov"
	layerHost = "view.eumetsat.int"
	evilHost  = "evil.example"
)

// httpsPort is where an https URL with no port is dialled; every other port
// reaches the plain stand-in, so a fall to cleartext is seen as a dial to it.
const httpsPort = "443"

// certName is the name httptest's certificate is issued for; the rig checks
// every certificate against it, since no real host name is on it.
const certName = "example.com"

// dialRig sends every connection to a loopback server, TLS for port 443 and
// plain for any other, recording each address the client asked to dial. The
// client's own redirect decisions run unchanged against the real host names;
// only DNS and the far servers are stood in for (audit round 2's method).
type dialRig struct {
	mu      sync.Mutex
	dialled []string
}

func (r *dialRig) addresses() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.dialled...)
}

// newDialRig starts both stand-ins over h and answers a client dialling them.
func newDialRig(t *testing.T, h http.Handler) (*http.Client, *dialRig) {
	t.Helper()
	secure := httptest.NewTLSServer(h)
	t.Cleanup(secure.Close)
	plain := httptest.NewServer(h)
	t.Cleanup(plain.Close)
	rig := &dialRig{}
	trusted := secure.Client().Transport.(*http.Transport).TLSClientConfig.RootCAs
	transport := &http.Transport{
		// No proxy: the test must reach its own stand-ins whatever the machine says.
		Proxy:           nil,
		TLSClientConfig: &tls.Config{RootCAs: trusted, ServerName: certName},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			rig.mu.Lock()
			rig.dialled = append(rig.dialled, addr)
			rig.mu.Unlock()
			target := plain.Listener.Addr().String()
			if _, port, _ := net.SplitHostPort(addr); port == httpsPort {
				target = secure.Listener.Addr().String()
			}
			var d net.Dialer
			return d.DialContext(ctx, network, target)
		},
	}
	t.Cleanup(transport.CloseIdleConnections)
	return &http.Client{Transport: transport}, rig
}

// redirector answers /to?code=N&url=U with a redirect of code N to U and any
// other path with the host it was reached as, so a test reads where it landed.
func redirector() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/to" {
			code := http.StatusFound
			switch r.URL.Query().Get("code") {
			case "301":
				code = http.StatusMovedPermanently
			case "303":
				code = http.StatusSeeOther
			case "307":
				code = http.StatusTemporaryRedirect
			case "308":
				code = http.StatusPermanentRedirect
			}
			http.Redirect(w, r, r.URL.Query().Get("url"), code)
			return
		}
		_, _ = w.Write([]byte(r.Host))
	})
}

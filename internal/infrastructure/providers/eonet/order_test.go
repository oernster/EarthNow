package eonet

import (
	"fmt"
	"testing"
)

// gdacsFlood is a GDACS event holding one polygon ring, written as the feed
// writes it.
func gdacsFlood(id, ring string) string {
	return fmt.Sprintf(`{"id":%q,"title":"Flood","sources":[{"id":"GDACS","url":"https://www.gdacs.org/r"}],
	 "geometry":[{"date":"2026-09-20T00:00:00Z","type":"Polygon","coordinates":[%s]}]}`, id, ring)
}

// Thailand is a ring whose 97.7 can only be a longitude, given in each order.
// Honduras is a ring that reads as valid either way round.
const (
	thailandLatFirst = `[[17.7,97.7],[17.9,97.7],[17.9,97.9],[17.7,97.7]]`
	thailandLngFirst = `[[97.7,17.7],[97.7,17.9],[97.9,17.9],[97.7,17.7]]`
	hondurasLngFirst = `[[-88.1,14.6],[-88.1,14.8],[-87.9,14.8],[-88.1,14.6]]`
	hondurasLatFirst = `[[14.6,-88.1],[14.8,-88.1],[14.8,-87.9],[14.6,-88.1]]`
)

func parseFloods(t *testing.T, floods ...string) map[string][2]float64 {
	t.Helper()
	body := `{"events":[`
	for i, f := range floods {
		if i > 0 {
			body += ","
		}
		body += f
	}
	events, dropped, err := Parse([]byte(body + `]}`))
	if err != nil || dropped != 0 {
		t.Fatalf("Parse: dropped %d, %v", dropped, err)
	}
	at := map[string][2]float64{}
	for _, e := range events {
		p := e.Observations[0].Where
		at[e.ProviderEventID] = [2]float64{p.Lat, p.Lng}
	}
	return at
}

func inHonduras(p [2]float64) bool { return p[0] > 14 && p[0] < 15 && p[1] > -89 && p[1] < -87 }
func inThailand(p [2]float64) bool { return p[0] > 17 && p[0] < 18 && p[1] > 97 && p[1] < 98 }

// If GDACS corrects its order, a ring that proves GeoJSON order carries the
// feed's undecided rings with it, rather than every flood being drawn swapped.
// A GDACS point (its wildfires) casts no vote and keeps GeoJSON order.
func TestDATA003_GDACSFollowsTheOrderItsRingsProve(t *testing.T) {
	t.Parallel()
	fire := `{"id":"F","title":"Wildfire","sources":[{"id":"GDACS","url":"https://www.gdacs.org/r"}],
	 "geometry":[{"date":"2026-09-20T00:00:00Z","type":"Point","coordinates":[97.8,17.8]}]}`
	at := parseFloods(t, gdacsFlood("T", thailandLngFirst), gdacsFlood("H", hondurasLngFirst), fire)
	if !inThailand(at["T"]) || !inHonduras(at["H"]) || !inThailand(at["F"]) {
		t.Errorf("GeoJSON-ordered feed drawn at %v", at)
	}
}

// As measured, rings that prove latitude first keep the undecided ones there.
func TestDATA003_GDACSKeepsLatitudeFirstWhileItsRingsProveIt(t *testing.T) {
	t.Parallel()
	at := parseFloods(t, gdacsFlood("T", thailandLatFirst), gdacsFlood("H", hondurasLatFirst))
	if !inThailand(at["T"]) || !inHonduras(at["H"]) {
		t.Errorf("latitude-first feed drawn at %v", at)
	}
}

// A ring that proves its order is read by it whatever the feed's vote says;
// a tie leaves the undecided rings with the measured latitude-first order.
func TestDATA003_GDACSProvenRingOverridesTheFeed(t *testing.T) {
	t.Parallel()
	at := parseFloods(t,
		gdacsFlood("A", thailandLatFirst), gdacsFlood("B", thailandLngFirst), gdacsFlood("H", hondurasLatFirst))
	if !inThailand(at["A"]) || !inThailand(at["B"]) || !inHonduras(at["H"]) {
		t.Errorf("mixed feed drawn at %v", at)
	}
}

func TestDATA003_RingOrder(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		ring         [][]float64
		latFirst, ok bool
	}{
		{"second value a longitude", [][]float64{{17, 97}}, true, true},
		{"first value a longitude", [][]float64{{97, 17}}, false, true},
		{"nothing beyond 90", [][]float64{{14, -88}}, false, false},
		{"contradicts itself", [][]float64{{97, 17}, {17, 97}}, false, false},
		{"a short vertex proves nothing", [][]float64{{97}, {14, -88}}, false, false},
	}
	for _, c := range cases {
		if latFirst, ok := ringOrder(c.ring); latFirst != c.latFirst || ok != c.ok {
			t.Errorf("%s: ringOrder = %v, %v", c.name, latFirst, ok)
		}
	}
}

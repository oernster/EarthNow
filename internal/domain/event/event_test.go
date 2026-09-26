package event

import (
	"errors"
	"math"
	"testing"
)

func TestNFRSEC003_NewPointAcceptsTheEarth(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ lat, lng float64 }{{0, 0}, {90, 180}, {-90, -180}, {61.899, -150.919}} {
		p, err := NewPoint(c.lat, c.lng)
		if err != nil || p.Lat != c.lat || p.Lng != c.lng {
			t.Errorf("NewPoint(%v, %v) = %v, %v", c.lat, c.lng, p, err)
		}
	}
}

func TestNFRSEC003_NewPointRefusesOffTheEarth(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ lat, lng float64 }{
		{90.0001, 0}, {-91, 0}, {0, 180.5}, {0, -181},
		{math.NaN(), 0}, {0, math.NaN()}, {math.Inf(1), 0}, {0, math.Inf(-1)},
	} {
		if _, err := NewPoint(c.lat, c.lng); !errors.Is(err, ErrInvalidCoordinate) {
			t.Errorf("NewPoint(%v, %v) error = %v, want ErrInvalidCoordinate", c.lat, c.lng, err)
		}
	}
}

func TestDATA002_CategoriesInKeyOrder(t *testing.T) {
	t.Parallel()
	want := []Category{Earthquake, Volcano, Wildfire, SevereStorm, Flood, Ice, Other}
	got := Categories()
	if len(got) != len(want) {
		t.Fatalf("got %d categories, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("category %d = %s, want %s", i, got[i], want[i])
		}
	}
}

func TestDATA008_IDJoinsProviderAndSourceID(t *testing.T) {
	t.Parallel()
	e := Event{Provider: USGS, ProviderEventID: "ak0261abcd"}
	if got := e.ID(); got != "USGS:ak0261abcd" {
		t.Errorf("ID() = %q", got)
	}
}

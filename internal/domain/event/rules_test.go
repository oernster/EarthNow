package event

import (
	"errors"
	"testing"
)

// FR-PRV-012: an answer whose items are all refused is a parse failure. One
// with no items is not; nor is one with any usable.
func TestFRPRV012_CheckUsable(t *testing.T) {
	t.Parallel()
	if err := CheckUsable(0, 3); !errors.Is(err, ErrNoneUsable) || err.Error() != "none of its items could be used: 3 dropped as malformed" {
		t.Errorf("none usable of three = %v", err)
	}
	for _, c := range []struct{ usable, dropped int }{{0, 0}, {1, 0}, {1, 9}} {
		if err := CheckUsable(c.usable, c.dropped); err != nil {
			t.Errorf("CheckUsable(%d, %d) = %v", c.usable, c.dropped, err)
		}
	}
}

// FR-SET-002: the minimum keeps USGS events at or above it (never one without
// a magnitude) and governs no other provider; under NoMinimum all are kept.
func TestFRSET002_MeetsMinimum(t *testing.T) {
	t.Parallel()
	quake := func(mag *float64) Event {
		o := Observation{}
		if mag != nil {
			o.Measurement = &Measurement{Value: *mag}
		}
		return Event{Provider: USGS, Observations: []Observation{o}}
	}
	three, five, below := 3.0, 4.5, -1.0
	cases := []struct {
		e       Event
		minimum float64
		want    bool
	}{
		{quake(&three), 4.5, false},
		{quake(&five), 4.5, true},
		{quake(nil), 1.0, false},
		{quake(nil), NoMinimum, true},
		{quake(&below), NoMinimum, true},
		{Event{Provider: EONET}, 4.5, true},
	}
	for i, c := range cases {
		if got := MeetsMinimum(c.e, c.minimum); got != c.want {
			t.Errorf("case %d: MeetsMinimum = %v, want %v", i, got, c.want)
		}
	}
}

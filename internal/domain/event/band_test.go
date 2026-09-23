package event

import "testing"

func TestFRMRK003_FRMRK004_Bands(t *testing.T) {
	t.Parallel()
	quake := Event{Category: Earthquake}
	cases := []struct {
		e    Event
		mag  *float64
		want int
	}{
		{quake, ptr(1.8), 1}, {quake, ptr(3.0), 2}, {quake, ptr(3.1), 2}, {quake, ptr(4.5), 3},
		{quake, ptr(5.0), 3}, {quake, ptr(6.0), 4}, {quake, ptr(6.7), 4}, {quake, nil, 0},
		{Event{Category: SevereStorm}, ptr(55), 0},
	}
	for _, c := range cases {
		o := Observation{}
		if c.mag != nil {
			o.Measurement = &Measurement{Value: *c.mag}
		}
		if got := Band(c.e, o); got != c.want {
			t.Errorf("Band(%s, %v) = %d, want %d", c.e.Category, c.mag, got, c.want)
		}
	}
}

func ptr(f float64) *float64 { return &f }

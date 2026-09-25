package window

import (
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
)

func TestFRRPL001_APositionNamesAnInstantAlongTheSpan(t *testing.T) {
	t.Parallel()
	end := utc(24, 12)
	cases := []struct {
		position float64
		want     time.Time
	}{
		{0, utc(17, 12)},
		{0.5, utc(21, 0)},
		{1, utc(24, 12)},
		{1.2, utc(24, 12)},
		{-0.3, utc(17, 12)},
	}
	for _, c := range cases {
		if got := SevenDays.At(end, c.position); !got.Equal(c.want) {
			t.Errorf("At(%v) = %v, want %v", c.position, got, c.want)
		}
	}
}

func TestFRRPL009_AReplayShowsWhatHappenedByItsInstant(t *testing.T) {
	t.Parallel()
	end := utc(24, 12)
	quake := event.Event{Category: event.Earthquake, Observations: []event.Observation{{At: time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)}}}
	before := SevenDays.Replay(end, SevenDays.position(t, end, time.Date(2026, 9, 20, 9, 59, 0, 0, time.UTC)))
	at := SevenDays.Replay(end, SevenDays.position(t, end, time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)))
	if _, ok := before.Latest(quake); ok {
		t.Error("the quake shows a minute before it happened")
	}
	if _, ok := at.Latest(quake); !ok {
		t.Error("the quake is missing at the instant it happened")
	}
	if !at.From.Equal(utc(17, 12)) {
		t.Errorf("the replay starts at %v, want the span's start", at.From)
	}
}

func TestFRRPL010_AStormsTrailGrowsWithTheReplay(t *testing.T) {
	t.Parallel()
	end := utc(24, 12)
	storm := event.Event{Category: event.SevereStorm, Observations: []event.Observation{
		{At: utc(19, 0), Where: event.Point{Lat: 10}},
		{At: utc(21, 0), Where: event.Point{Lat: 12}},
	}}
	on20 := SevenDays.Replay(end, SevenDays.position(t, end, utc(20, 0)))
	if o, ok := on20.Latest(storm); !ok || o.Where.Lat != 10 || on20.Trail(storm) != nil {
		t.Errorf("on 20 Sep: at %v, trail %v; want the 19 Sep fix and no trail", o.Where, on20.Trail(storm))
	}
	on21 := SevenDays.Replay(end, SevenDays.position(t, end, utc(21, 0)))
	if got := on21.Trail(storm); len(got) != 2 {
		t.Errorf("on 21 Sep the trail is %v, want two points", got)
	}
}

// position answers the share of the span ending at end that at lies at.
func (w Window) position(t *testing.T, end, at time.Time) float64 {
	t.Helper()
	return float64(at.Sub(end.Add(-w.Length))) / float64(w.Length)
}

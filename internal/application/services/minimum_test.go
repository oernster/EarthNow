package services

import (
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
	"github.com/oernster/EarthNow/internal/domain/window"
)

// FR-SET-002: the minimum applies at view time, live and in a replay, to USGS
// events only; the held set and the cache keep every quake, so lowering the
// minimum again offline brings them back (audit round 2, E-3).
func TestFRSET002_TheMinimumFiltersTheViewNotTheHeldSet(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	cache := &fakeCache{}
	fire := ev(event.EONET, "fire", event.Wildfire, time.Hour)
	p := &fakeProvider{name: event.USGS, fetched: ports.Fetched{Events: []event.Event{quake("small", 3, ""), quake("large", 5, "")}}}
	eonet := &fakeProvider{name: event.EONET, fetched: ports.Fetched{Events: []event.Event{fire}}}
	g := NewGlobe(NewStore(clock), clock, []ports.Provider{p, eonet})
	g.UseCache(cache)
	refreshEach(g)
	g.SetMinimum(4.5)
	if got := viewIDs(g.View("24h", Filter{}).Events); len(got) != 2 || !got["USGS:large"] || !got["EONET:fire"] {
		t.Errorf("under 4.5 the view shows %v; want the M5 and the fire", got)
	}
	r := window.Range{From: noon.Add(-time.Hour * 2), To: noon}
	if got := viewIDs(g.ReplayView(window.OneDay, r, Filter{}).Events); got["USGS:small"] {
		t.Errorf("under 4.5 a replay shows the M3: %v", got)
	}
	if snap, _ := g.store.Snapshot(event.USGS); len(snap.Events) != 2 || len(cache.held[event.USGS].Events) != 2 {
		t.Errorf("held %d, cached %d; want both quakes kept", len(snap.Events), len(cache.held[event.USGS].Events))
	}
	g.SetMinimum(event.NoMinimum)
	if got := viewIDs(g.View("24h", Filter{}).Events); len(got) != 3 {
		t.Errorf("under All the view shows %v; want every held event", got)
	}
}

func viewIDs(events []dto.Event) map[string]bool {
	out := map[string]bool{}
	for _, e := range events {
		out[e.ID] = true
	}
	return out
}

// FR-SET-002: the minimum reaches every applier at load and on a change: the
// adapter's next fetch and the view alike.
func TestFRSET002_TheMinimumReachesEveryApplier(t *testing.T) {
	t.Parallel()
	var first, second minimums
	p := NewPreferences(&fakeSettings{}, first.apply, second.apply)
	p.Load()
	chosen := p.Current()
	chosen.Magnitude = "4.5"
	p.Update(chosen)
	if len(first) != 2 || len(second) != 2 || first[1] != 4.5 || second[1] != 4.5 {
		t.Errorf("appliers were handed %v and %v; want the load's and 4.5 each", first, second)
	}
}

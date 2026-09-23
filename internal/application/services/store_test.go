package services

import (
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
	"github.com/oernster/EarthNow/internal/domain/window"
)

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

var noon = time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)

func ev(p event.Provider, id string, c event.Category, ago time.Duration) event.Event {
	return event.Event{
		Provider: p, ProviderEventID: id, Category: c, Title: id,
		Observations: []event.Observation{{At: noon.Add(-ago)}},
	}
}

func ids(shown []Shown) string {
	s := ""
	for _, x := range shown {
		s += x.Event.ID() + " "
	}
	return s
}

func TestFRTW002_VisibleHonoursTheWindowNewestFirst(t *testing.T) {
	t.Parallel()
	st := NewStore(&fakeClock{noon})
	st.Apply(event.USGS, ports.Fetched{Events: []event.Event{
		ev(event.USGS, "a", event.Earthquake, 30*time.Minute),
		ev(event.USGS, "b", event.Earthquake, 90*time.Minute),
		ev(event.USGS, "c", event.Earthquake, 10*time.Minute),
	}})
	if got := ids(st.Visible(window.OneHour, Filter{})); got != "USGS:c USGS:a " {
		t.Errorf("1 h = %q", got)
	}
	if got := ids(st.Visible(window.OneDay, Filter{})); got != "USGS:c USGS:a USGS:b " {
		t.Errorf("24 h = %q", got)
	}
}

func TestVisibleBreaksTimeTiesByID(t *testing.T) {
	t.Parallel()
	st := NewStore(&fakeClock{noon})
	st.Apply(event.USGS, ports.Fetched{Events: []event.Event{
		ev(event.USGS, "z", event.Earthquake, time.Minute),
		ev(event.USGS, "m", event.Earthquake, time.Minute),
	}})
	if got := ids(st.Visible(window.OneHour, Filter{})); got != "USGS:m USGS:z " {
		t.Errorf("tie order = %q", got)
	}
}

func TestFRFLT002_FRFLT004_FiltersHideCategoriesAndProviders(t *testing.T) {
	t.Parallel()
	st := NewStore(&fakeClock{noon})
	st.Apply(event.USGS, ports.Fetched{Events: []event.Event{ev(event.USGS, "q", event.Earthquake, time.Minute)}})
	st.Apply(event.EONET, ports.Fetched{Events: []event.Event{
		ev(event.EONET, "s", event.SevereStorm, time.Hour),
		ev(event.EONET, "f", event.Wildfire, 2*time.Hour),
	}})
	noStorms := Filter{HiddenCategories: map[event.Category]bool{event.SevereStorm: true}}
	if got := ids(st.Visible(window.OneDay, noStorms)); got != "USGS:q EONET:f " {
		t.Errorf("storms hidden = %q", got)
	}
	noUSGS := Filter{HiddenProviders: map[event.Provider]bool{event.USGS: true}}
	if got := ids(st.Visible(window.OneDay, noUSGS)); got != "EONET:s EONET:f " {
		t.Errorf("USGS hidden = %q", got)
	}
	if got := len(st.Visible(window.OneDay, Filter{})); got != 3 {
		t.Errorf("FR-FLT-003 all events = %d, want 3", got)
	}
}

func TestFRPRV004_NotModifiedKeepsEventsAndRefreshesRetrieval(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	st := NewStore(clock)
	st.Apply(event.USGS, ports.Fetched{Events: []event.Event{ev(event.USGS, "q", event.Earthquake, 0)}, Validator: "v1"})
	clock.now = noon.Add(time.Minute)
	st.Apply(event.USGS, ports.Fetched{NotModified: true})
	snap, _ := st.Snapshot(event.USGS)
	if len(snap.Events) != 1 || !snap.RetrievedAt.Equal(clock.now) || snap.Validator != "v1" {
		t.Errorf("after 304 without validator: %+v", snap)
	}
	st.Apply(event.USGS, ports.Fetched{NotModified: true, Validator: "v2"})
	if snap, _ = st.Snapshot(event.USGS); snap.Validator != "v2" {
		t.Errorf("validator = %q, want v2", snap.Validator)
	}
}

func TestFRPRV008_OneProviderNeverTouchesAnother(t *testing.T) {
	t.Parallel()
	st := NewStore(&fakeClock{noon})
	st.Apply(event.EONET, ports.Fetched{Events: []event.Event{ev(event.EONET, "s", event.SevereStorm, 0)}})
	st.Apply(event.USGS, ports.Fetched{Events: nil})
	if snap, ok := st.Snapshot(event.EONET); !ok || len(snap.Events) != 1 {
		t.Errorf("EONET set disturbed: %+v %v", snap, ok)
	}
	if _, ok := st.Snapshot(event.Provider("NONE")); ok {
		t.Error("snapshot found for a provider never applied")
	}
}

func TestDATA008_RepeatedIDKeepsTheLastCopy(t *testing.T) {
	t.Parallel()
	st := NewStore(&fakeClock{noon})
	first := ev(event.USGS, "q", event.Earthquake, time.Minute)
	second := first
	second.Title = "revised"
	st.Apply(event.USGS, ports.Fetched{Events: []event.Event{first, ev(event.USGS, "r", event.Earthquake, 0), second}})
	snap, _ := st.Snapshot(event.USGS)
	if len(snap.Events) != 2 || snap.Events[0].Title != "revised" {
		t.Errorf("events = %+v", snap.Events)
	}
}

func TestFRSTS004_RestoreServesCachedEvents(t *testing.T) {
	t.Parallel()
	st := NewStore(&fakeClock{noon})
	cached := Snapshot{Events: []event.Event{ev(event.USGS, "q", event.Earthquake, time.Hour), ev(event.USGS, "q", event.Earthquake, time.Hour)}, RetrievedAt: noon.Add(-3 * time.Hour)}
	st.Restore(event.USGS, cached)
	shown := st.Visible(window.OneDay, Filter{})
	if len(shown) != 1 || !shown[0].RetrievedAt.Equal(cached.RetrievedAt) {
		t.Errorf("restored = %+v", shown)
	}
}

func TestFRKEY004_CountsPerCategory(t *testing.T) {
	t.Parallel()
	shown := []Shown{
		{Event: ev(event.USGS, "a", event.Earthquake, 0)},
		{Event: ev(event.USGS, "b", event.Earthquake, 0)},
		{Event: ev(event.EONET, "c", event.Ice, 0)},
	}
	c := Counts(shown)
	if c[event.Earthquake] != 2 || c[event.Ice] != 1 || c[event.Volcano] != 0 {
		t.Errorf("counts = %v", c)
	}
}

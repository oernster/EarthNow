package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
)

type fakeProvider struct {
	name      event.Provider
	fetched   ports.Fetched
	err       error
	validator string
}

func (f *fakeProvider) Name() event.Provider    { return f.name }
func (f *fakeProvider) Interval() time.Duration { return time.Minute }
func (f *fakeProvider) Fetch(_ context.Context, validator string) (ports.Fetched, error) {
	f.validator = validator
	return f.fetched, f.err
}

type fakeGeocoder struct{}

func (fakeGeocoder) Describe(lat, lng float64) string { return "8 km NE of Tromsø, Troms, Norway" }

func quake(id string, mag float64, url string) event.Event {
	return event.Event{
		Provider: event.USGS, ProviderEventID: id, Category: event.Earthquake, Title: "M " + id, SourceURL: url,
		Observations: []event.Observation{{At: noon.Add(-10 * time.Minute), Where: event.Point{Lat: 61.9, Lng: -150.9}, Measurement: &event.Measurement{Value: mag, Unit: "md"}}},
	}
}

func provider(t *testing.T, g *Globe, name string) dto.Provider {
	t.Helper()
	for _, p := range g.View("24h", Filter{}).Providers {
		if p.Name == name {
			return p
		}
	}
	t.Fatalf("no status for %s", name)
	return dto.Provider{}
}

func TestFRPRV008_RefreshAllIsolatesAFailingProvider(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	usgs := &fakeProvider{name: event.USGS, fetched: ports.Fetched{Events: []event.Event{quake("a", 3.21, "https://earthquake.usgs.gov/a")}, Validator: "v1"}}
	eonet := &fakeProvider{name: event.EONET, err: errors.New("EONET: requesting eonet.gsfc.nasa.gov: timeout")}
	g := NewGlobe(NewStore(clock), clock, []ports.Provider{usgs, eonet})
	if p := provider(t, g, "USGS"); !p.Loading {
		t.Errorf("FR-STS-001 before any fetch = %+v, want loading", p)
	}
	g.RefreshAll(context.Background())
	view := g.View("24h", Filter{})
	if len(view.Events) != 1 || view.CountLine != "1 event in the last 24 h" || view.Counts["EARTHQUAKE"] != 1 {
		t.Errorf("view = %+v", view)
	}
	if p := provider(t, g, "EONET"); p.Loading || p.Problem == "" {
		t.Errorf("FR-STS-003 failed provider = %+v", p)
	}
	if p := provider(t, g, "USGS"); p.Loading || p.Stale || p.Retrieved != "Retrieved under a minute ago" {
		t.Errorf("healthy provider = %+v", p)
	}
	g.RefreshAll(context.Background())
	if usgs.validator != "v1" {
		t.Errorf("FR-PRV-004 second fetch sent %q, want v1", usgs.validator)
	}
	clock.now = noon.Add(4 * time.Minute)
	if p := provider(t, g, "USGS"); !p.Stale {
		t.Errorf("FR-STS-002 four minutes on = %+v, want stale", p)
	}
}

func TestNFROBS001_RefreshAnswersItsOutcome(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	p := &fakeProvider{name: event.USGS, fetched: ports.Fetched{Events: []event.Event{quake("a", 3, "")}, Dropped: 2}}
	g := NewGlobe(NewStore(clock), clock, []ports.Provider{p})
	out, err := g.Refresh(context.Background(), p)
	if err != nil || out.Count != 1 || out.Dropped != 2 || out.NotModified {
		t.Errorf("full = %+v, %v", out, err)
	}
	p.fetched = ports.Fetched{NotModified: true}
	if out, _ = g.Refresh(context.Background(), p); !out.NotModified || out.Count != 1 {
		t.Errorf("304 = %+v", out)
	}
	p.err = errors.New("down")
	if _, err = g.Refresh(context.Background(), p); err == nil {
		t.Error("a failure was answered as success")
	}
}

func TestViewWordsEachEvent(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	usgs := &fakeProvider{name: event.USGS, fetched: ports.Fetched{Events: []event.Event{
		quake("a", 3.21, "https://earthquake.usgs.gov/a"),
		quake("b", 5.0, "javascript:alert(1)"),
	}}}
	g := NewGlobe(NewStore(clock), clock, []ports.Provider{usgs})
	g.RefreshAll(context.Background())
	byID := map[string]dto.Event{}
	for _, e := range g.View("nonsense", Filter{}).Events {
		byID[e.ID] = e
	}
	a, b := byID["USGS:a"], byID["USGS:b"]
	if a.Measurement != "3.21 md" || a.Band != 2 || a.Reported != "Observed 10 min ago" || a.At != "2026-09-23T11:50:00Z" || a.DayOnly {
		t.Errorf("a = %+v", a)
	}
	if a.SourceURL != "https://earthquake.usgs.gov/a" || a.SourceText != "" {
		t.Errorf("FR-SEL-005 a source = %q / %q", a.SourceURL, a.SourceText)
	}
	if b.SourceURL != "" || b.SourceText != "javascript:alert(1)" || b.Band != 3 {
		t.Errorf("FR-SEL-006 b = %+v", b)
	}
	if got := g.View("nonsense", Filter{}).WindowKey; got != "24h" {
		t.Errorf("unknown window fell back to %q", got)
	}
}

func TestMeasurementWithoutUnitAndEventWithoutMeasurement(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	bare := quake("c", 4.0, "")
	bare.Observations[0].Measurement.Unit = ""
	none := quake("d", 0, "")
	none.Observations[0].Measurement = nil
	g := NewGlobe(NewStore(clock), clock, []ports.Provider{&fakeProvider{name: event.USGS, fetched: ports.Fetched{Events: []event.Event{bare, none}}}})
	g.RefreshAll(context.Background())
	for _, e := range g.View("24h", Filter{}).Events {
		if (e.ID == "USGS:c" && e.Measurement != "4") || (e.ID == "USGS:d" && e.Measurement != "") {
			t.Errorf("%s measurement = %q", e.ID, e.Measurement)
		}
	}
}

func TestFRGEO_PlaceWaitsForTheGeocoder(t *testing.T) {
	t.Parallel()
	g := NewGlobe(NewStore(&fakeClock{noon}), &fakeClock{noon}, nil)
	if got := g.Place(69.7, 19.1); got != "" {
		t.Errorf("before load = %q", got)
	}
	g.SetGeocoder(fakeGeocoder{})
	if got := g.Place(69.7, 19.1); got != "8 km NE of Tromsø, Troms, Norway" {
		t.Errorf("after load = %q", got)
	}
}

func TestFRTW001_WindowsForTheControl(t *testing.T) {
	t.Parallel()
	w := NewGlobe(nil, nil, nil).Windows()
	if len(w) != 5 || w[0].Key != "1h" || w[4].Label != "7 days" {
		t.Errorf("windows = %+v", w)
	}
}

func TestSafeURL(t *testing.T) {
	t.Parallel()
	for raw, want := range map[string]string{
		"https://x.org/p": "https://x.org/p", "http://x.org": "", "https:///nohost": "", "%zz": "", "": "",
		// FR-SEL-009: a data file is not handed to the browser as a page.
		"https://www.metoc.navy.mil/jtwc/products/ep1726.tcw": "",
	} {
		if got := SafeURL(raw); got != want {
			t.Errorf("SafeURL(%q) = %q, want %q", raw, got, want)
		}
	}
}

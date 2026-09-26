package eonet

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/oernster/EarthNow/internal/domain/event"
	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch"
)

func fixture(t *testing.T) []byte {
	t.Helper()
	body, err := os.ReadFile("testdata/open_7d.json")
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func byID(t *testing.T, events []event.Event, id string) event.Event {
	t.Helper()
	for _, e := range events {
		if e.ProviderEventID == id {
			return e
		}
	}
	t.Fatalf("%s not found", id)
	return event.Event{}
}

func TestFRPRV001_URLAsksForOpenEventsOverTheWidestWindow(t *testing.T) {
	t.Parallel()
	if got := URL(); got != "https://eonet.gsfc.nasa.gov/api/v3/events?status=all&days=7" {
		t.Errorf("URL() = %q", got)
	}
	a := New(nil)
	if a.Name() != event.EONET || a.Interval() != Interval {
		t.Error("Name or Interval wrong")
	}
}

func TestParseCapturedFixture(t *testing.T) {
	t.Parallel()
	events, dropped, err := Parse(fixture(t))
	if err != nil || dropped != 0 || len(events) != 18 {
		t.Fatalf("Parse = %d events, %d dropped, %v", len(events), dropped, err)
	}
	polo := byID(t, events, "EONET_24721")
	if polo.Category != event.SevereStorm || polo.Extras.SourceCategory != "severeStorms" || polo.Title != "Hurricane Polo" {
		t.Errorf("Polo = %+v", polo)
	}
	// FR-SEL-009: Polo's first source is a JTWC .tcw data file; the NHC page is taken.
	if polo.SourceURL != "https://www.nhc.noaa.gov/archive/2026/POLO.shtml" {
		t.Errorf("Polo source = %q, want the NHC page", polo.SourceURL)
	}
	first := polo.Observations[0]
	if first.At.Hour() != 0 || first.Precision != event.Instant {
		t.Errorf("DATA-005: Polo's 00:00 fix = %+v, want an instant", first)
	}
	if first.Measurement == nil || first.Measurement.Unit != "kts" || first.Where.Lat != 15.1 || first.Where.Lng != -104.8 {
		t.Errorf("Polo first fix = %+v", first)
	}
	for i := 1; i < len(polo.Observations); i++ {
		if polo.Observations[i].At.Before(polo.Observations[i-1].At) {
			t.Fatal("observations are not oldest first")
		}
	}
	var ice, fire int
	for _, e := range events {
		switch e.Category {
		case event.Ice:
			ice++
			for _, o := range e.Observations {
				if o.Precision != event.Day {
					t.Errorf("DATA-005: sea ice %s has an instant %v", e.ProviderEventID, o.At)
				}
			}
		case event.Wildfire:
			fire++
			if e.Observations[0].Precision != event.Instant {
				t.Errorf("wildfire %s read as day precision", e.ProviderEventID)
			}
		}
	}
	if ice != 8 || fire != 6 {
		t.Errorf("ice = %d, fire = %d; the fixture holds 8 and 6", ice, fire)
	}
}

func TestFRPRV012_UnusableBodyIsAnError(t *testing.T) {
	t.Parallel()
	for _, body := range []string{`not json`, `{"title":"EONET"}`} {
		if _, _, err := Parse([]byte(body)); err == nil {
			t.Errorf("Parse(%q) succeeded", body)
		}
	}
}

func TestFRPRV013_UnusableEventsAreDroppedAndCounted(t *testing.T) {
	t.Parallel()
	body := `{"events":[
	 {"id":"","title":"no id","geometry":[{"date":"2026-09-20T10:00:00Z","type":"Point","coordinates":[1,2]}]},
	 {"id":"E1","title":"","geometry":[{"date":"2026-09-20T10:00:00Z","type":"Point","coordinates":[1,2]}]},
	 {"id":"E2","title":"bad date","geometry":[{"date":"yesterday","type":"Point","coordinates":[1,2]}]},
	 {"id":"E3","title":"off the earth","geometry":[{"date":"2026-09-20T10:00:00Z","type":"Point","coordinates":[200,2]}]},
	 {"id":"E4","title":"short point","geometry":[{"date":"2026-09-20T10:00:00Z","type":"Point","coordinates":[1]}]},
	 {"id":"E5","title":"line","geometry":[{"date":"2026-09-20T10:00:00Z","type":"LineString","coordinates":[[1,2],[3,4]]}]},
	 {"id":"E6","title":"empty polygon","geometry":[{"date":"2026-09-20T10:00:00Z","type":"Polygon","coordinates":[]}]},
	 {"id":"E7","title":"broken ring","geometry":[{"date":"2026-09-20T10:00:00Z","type":"Polygon","coordinates":[[[1]]]}]},
	 {"id":"E8","title":"empty ring","geometry":[{"date":"2026-09-20T10:00:00Z","type":"Polygon","coordinates":[[]]}]},
	 {"id":"OK","title":"kept","geometry":[{"date":"2026-09-20T10:00:00Z","type":"Point","coordinates":[1,2]},{"date":"bad","type":"Point","coordinates":[1,2]}]}
	]}`
	events, dropped, err := Parse([]byte(body))
	if err != nil || dropped != 9 || len(events) != 1 || len(events[0].Observations) != 1 {
		t.Errorf("Parse = %+v, dropped %d, %v", events, dropped, err)
	}
}

func TestDATA006_UnknownAndUnmappedCategoriesAreOther(t *testing.T) {
	t.Parallel()
	body := `{"events":[
	 {"id":"S","title":"snow","closed":"2026-09-21T00:00:00Z","description":"source text","categories":[{"id":"snow"}],"geometry":[{"date":"2026-09-20T00:00:00Z","type":"Point","coordinates":[1,2]}]},
	 {"id":"N","title":"new","categories":[{"id":"somethingNew"}],"geometry":[{"date":"2026-09-20T01:00:00Z","type":"Point","coordinates":[1,2]}]},
	 {"id":"B","title":"bare","geometry":[{"date":"2026-09-20T01:00:00Z","type":"Point","coordinates":[1,2]}]}
	]}`
	events, _, err := Parse([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	snow := byID(t, events, "S")
	if snow.Category != event.Other || snow.Extras.SourceCategory != "snow" || snow.Status != "closed" || snow.Description != "source text" {
		t.Errorf("snow = %+v", snow)
	}
	if n := byID(t, events, "N"); n.Category != event.Other || n.Extras.SourceCategory != "somethingNew" {
		t.Errorf("new = %+v", n)
	}
	if b := byID(t, events, "B"); b.Category != event.Other || b.Status != "open" {
		t.Errorf("bare = %+v", b)
	}
}

// Amendment 45: EONET published no landslide, drought or dust haze in the year
// to 2026-09-26, so those kinds have no category of their own and are Other,
// EONET's own id kept on the event.
func TestDATA006_KindsNoSourceFeedsAreOther(t *testing.T) {
	t.Parallel()
	for _, id := range []string{"landslides", "drought", "dustHaze"} {
		body := `{"events":[{"id":"X","title":"x","categories":[{"id":"` + id + `"}],"geometry":[{"date":"2026-09-20T01:00:00Z","type":"Point","coordinates":[1,2]}]}]}`
		events, _, err := Parse([]byte(body))
		if err != nil {
			t.Fatal(err)
		}
		if e := byID(t, events, "X"); e.Category != event.Other || e.Extras.SourceCategory != id {
			t.Errorf("%s = %+v, want Other keeping its id", id, e)
		}
	}
}

func TestDATA003_PolygonIsPlacedAtItsCentroid(t *testing.T) {
	t.Parallel()
	body := `{"events":[
	 {"id":"P","title":"closed ring","geometry":[{"date":"2026-09-20T01:00:00Z","type":"Polygon","coordinates":[[[0,0],[2,0],[2,2],[0,2],[0,0]]]}]},
	 {"id":"Q","title":"open ring","geometry":[{"date":"2026-09-20T01:00:00Z","type":"Polygon","coordinates":[[[0,0],[4,0],[4,4]]]}]}
	]}`
	events, dropped, err := Parse([]byte(body))
	if err != nil || dropped != 0 {
		t.Fatalf("Parse: dropped %d, %v", dropped, err)
	}
	if p := byID(t, events, "P").Observations[0].Where; p.Lat != 1 || p.Lng != 1 {
		t.Errorf("closed ring centroid = %+v", p)
	}
	if q := byID(t, events, "Q").Observations[0].Where; q.Lng != 8.0/3 || q.Lat != 4.0/3 {
		t.Errorf("open ring centroid = %+v", q)
	}
}

// DATA-003: a polygon straddling 180 degrees is placed between its vertices,
// in the Pacific, not at the plain mean of their longitudes on the far side of
// the world. 178 and -176 lie 6 degrees apart across the line; their centre is
// -179, where a plain mean would say 1.
func TestDATA003_PolygonAcrossTheAntimeridianStaysInThePacific(t *testing.T) {
	t.Parallel()
	body := `{"events":[
	 {"id":"F","title":"across the line","geometry":[{"date":"2026-09-20T01:00:00Z","type":"Polygon","coordinates":[[[178,0],[-176,0],[-176,2],[178,2],[178,0]]]}]}
	]}`
	events, dropped, err := Parse([]byte(body))
	if err != nil || dropped != 0 {
		t.Fatalf("Parse: dropped %d, %v", dropped, err)
	}
	if f := byID(t, events, "F").Observations[0].Where; f.Lng != -179 || f.Lat != 1 {
		t.Errorf("antimeridian centroid = %+v, want lat 1, lng -179", f)
	}
}

// NFR-SEC-003: a polygon vertex outside the Earth drops the event, even where
// the vertices would average into range (-300 and 300 average to 0).
func TestDATA003_PolygonVertexOutsideTheEarthIsDropped(t *testing.T) {
	t.Parallel()
	body := `{"events":[
	 {"id":"X","title":"vertex off the Earth","geometry":[{"date":"2026-09-20T01:00:00Z","type":"Polygon","coordinates":[[[-300,0],[300,0],[0,2]]]}]}
	]}`
	events, dropped, err := Parse([]byte(body))
	if err != nil || dropped != 1 || len(events) != 0 {
		t.Errorf("Parse: %d events, dropped %d, %v; want the event dropped", len(events), dropped, err)
	}
}

// DATA-003: longitudes further than half a turn apart are compared the short
// way round in both directions; a mean past 180 is folded back.
func TestDATA003_LongitudesAreComparedTheShortWayRound(t *testing.T) {
	t.Parallel()
	cases := []struct{ lng, ref, want float64 }{
		{-176, 178, 184},
		{178, -176, -182},
		{10, 20, 10},
	}
	for _, c := range cases {
		if got := nearest(c.lng, c.ref); got != c.want {
			t.Errorf("nearest(%v, %v) = %v, want %v", c.lng, c.ref, got, c.want)
		}
	}
	for in, want := range map[float64]float64{181: -179, -181: 179, 180: 180, -12: -12} {
		if got := wrapped(in); got != want {
			t.Errorf("wrapped(%v) = %v, want %v", in, got, want)
		}
	}
}

type fakeFetcher struct {
	resp httpfetch.Response
	err  error
	url  string
}

func (f *fakeFetcher) Get(_ context.Context, rawURL, _ string) (httpfetch.Response, error) {
	f.url = rawURL
	return f.resp, f.err
}

func TestFetch(t *testing.T) {
	t.Parallel()
	ok := &fakeFetcher{resp: httpfetch.Response{Body: fixture(t)}}
	got, err := New(ok).Fetch(context.Background(), "ignored")
	if err != nil || len(got.Events) != 18 || got.Validator != "" || ok.url != URL() {
		t.Errorf("Fetch = %d events, %v, url %q", len(got.Events), err, ok.url)
	}
	down := &fakeFetcher{err: errors.New("dns failure")}
	if _, err := New(down).Fetch(context.Background(), ""); err == nil {
		t.Error("a failed fetch was reported as success")
	}
	junk := &fakeFetcher{resp: httpfetch.Response{Body: []byte("<rss/>")}}
	if _, err := New(junk).Fetch(context.Background(), ""); err == nil {
		t.Error("an unparseable body was reported as success")
	}
}

// GDACS polygons arrive [lat, lng] (measured 2026-09-23); read as GeoJSON they
// put Honduras in Antarctica. Another source's polygon keeps GeoJSON order.
func TestGDACSPolygonsAreReadLatitudeFirst(t *testing.T) {
	t.Parallel()
	body := `{"events":[
	{"id":"EONET_1","title":"Flood in Honduras","sources":[{"id":"GDACS","url":"https://www.gdacs.org/r"}],
	 "geometry":[{"date":"2026-09-20T00:00:00Z","type":"Polygon","coordinates":[[[14.6,-88.1],[14.8,-88.1],[14.8,-87.9],[14.6,-87.9],[14.6,-88.1]]]}]},
	{"id":"EONET_2","title":"Ice","sources":[{"id":"USICECENTER","url":"https://usicecenter.gov/p"}],
	 "geometry":[{"date":"2026-09-20T00:00:00Z","type":"Polygon","coordinates":[[[-60,-70],[-59,-70],[-59,-69],[-60,-69],[-60,-70]]]}]},
	{"id":"EONET_3","title":"Short vertex","sources":[{"id":"GDACS","url":"https://www.gdacs.org/r"}],
	 "geometry":[{"date":"2026-09-20T00:00:00Z","type":"Polygon","coordinates":[[[14.6],[14.8,-88.1],[14.8,-87.9]]]}]}
]}`
	events, dropped, err := Parse([]byte(body))
	if err != nil || len(events) != 2 || dropped != 1 {
		t.Fatalf("events %d, dropped %d, err %v", len(events), dropped, err)
	}
	honduras := events[0].Observations[0].Where
	if honduras.Lat < 14 || honduras.Lat > 15 || honduras.Lng > -87 || honduras.Lng < -89 {
		t.Errorf("Honduras drawn at %+v", honduras)
	}
	ice := events[1].Observations[0].Where
	if ice.Lat > -68 || ice.Lng > -58 {
		t.Errorf("the GeoJSON-ordered polygon moved to %+v", ice)
	}
}

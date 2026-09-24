package usgs

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch"
)

func fixture(t *testing.T) []byte {
	t.Helper()
	body, err := os.ReadFile("testdata/2.5_week.json")
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func TestFRPRV003_URLPicksTheHighestFeedNotAboveTheMinimum(t *testing.T) {
	t.Parallel()
	cases := map[float64]string{
		AllMagnitudes: "all",
		1.0:           "1.0",
		2.0:           "1.0",
		ownerDefault:  "2.5",
		4.5:           "4.5",
		6.0:           "4.5",
	}
	for minimum, name := range cases {
		want := "https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/" + name + "_week.geojson"
		if got := URL(minimum); got != want {
			t.Errorf("URL(%v) = %q, want %q", minimum, got, want)
		}
	}
	a := New(nil, ownerDefault)
	if a.Name() != event.USGS || a.Interval() != time.Minute {
		t.Error("Name or Interval wrong")
	}
}

// fixtureDepthKm is the first fixture quake's third coordinate, pr71534228.
const fixtureDepthKm = 42.45

// DATA-004: a USGS event's time is properties.time. DATA-010: its depth is kept
// in kilometres as the feed gives it.
func TestParseCapturedFixture(t *testing.T) {
	t.Parallel()
	all, dropped, err := Parse(fixture(t), AllMagnitudes)
	if err != nil || dropped != 0 || len(all) != 12 {
		t.Fatalf("Parse = %d, dropped %d, %v", len(all), dropped, err)
	}
	first := all[0]
	o := first.Observations[0]
	if first.ProviderEventID != "pr71534228" || first.Category != event.Earthquake || first.Description != "65 km SE of Punta Cana, Dominican Republic" {
		t.Errorf("first = %+v", first)
	}
	if !o.At.Equal(time.UnixMilli(1790167038460)) || o.Measurement.Value != 3.21 || o.Measurement.Unit != "md" || o.Precision != event.Instant {
		t.Errorf("first observation = %+v", o)
	}
	if first.Extras.DepthKm == nil || *first.Extras.DepthKm != fixtureDepthKm || first.UpdatedAt.IsZero() || first.SourceURL != "https://earthquake.usgs.gov/earthquakes/eventpage/pr71534228" {
		t.Errorf("first extras = %+v", first)
	}
	atThree, _, _ := Parse(fixture(t), ownerDefault)
	for _, e := range atThree {
		if e.Observations[0].Measurement.Value < ownerDefault {
			t.Errorf("%s at M%v kept under a 3.0 minimum", e.ProviderEventID, e.Observations[0].Measurement.Value)
		}
	}
	if len(atThree) == 0 || len(atThree) >= len(all) {
		t.Errorf("minimum 3.0 kept %d of %d; the fixture holds quakes either side of 3.0", len(atThree), len(all))
	}
}

func TestFRPRV012_UnusableBodyIsAnError(t *testing.T) {
	t.Parallel()
	for _, body := range []string{`nope`, `{"type":"FeatureCollection"}`} {
		if _, _, err := Parse([]byte(body), AllMagnitudes); err == nil {
			t.Errorf("Parse(%q) succeeded", body)
		}
	}
}

// DATA-012: a feature typed other than earthquake keeps its type and maps to
// Other; a deleted one is left out. DATA-010: the tsunami flag is kept.
func TestFRPRV013_MalformedFeaturesDroppedWithdrawnOnesLeftOut(t *testing.T) {
	t.Parallel()
	body := `{"features":[
	 {"id":"","properties":{"time":1},"geometry":{"coordinates":[1,2,3]}},
	 {"id":"notime","properties":{},"geometry":{"coordinates":[1,2,3]}},
	 {"id":"nogeom","properties":{"time":1}},
	 {"id":"short","properties":{"time":1},"geometry":{"coordinates":[1]}},
	 {"id":"offearth","properties":{"time":1},"geometry":{"coordinates":[1,95]}},
	 {"id":"gone","properties":{"time":1,"status":"deleted"},"geometry":{"coordinates":[1,2,3]}},
	 {"id":"blast","properties":{"time":1,"type":"quarry blast","place":"near a quarry","tsunami":1},"geometry":{"coordinates":[1,2]}},
	 {"id":"nomag","properties":{"time":1,"type":"earthquake"},"geometry":{"coordinates":[1,2]}}
	]}`
	events, dropped, err := Parse([]byte(body), AllMagnitudes)
	if err != nil || dropped != 5 || len(events) != 2 {
		t.Fatalf("Parse = %+v, dropped %d, %v", events, dropped, err)
	}
	blast := events[0]
	if blast.Category != event.Other || blast.Extras.SourceCategory != "quarry blast" || blast.Title != "near a quarry" || !blast.Extras.TsunamiFlag || blast.Extras.DepthKm != nil {
		t.Errorf("blast = %+v", blast)
	}
	if events[1].Observations[0].Measurement != nil {
		t.Error("a quake with no magnitude was given one")
	}
	if kept, _, _ := Parse([]byte(body), ownerDefault); len(kept) != 0 {
		t.Errorf("events without a magnitude passed a 3.0 minimum: %+v", kept)
	}
}

type fakeFetcher struct {
	resp      httpfetch.Response
	err       error
	url       string
	validator string
}

func (f *fakeFetcher) Get(_ context.Context, rawURL, validator string) (httpfetch.Response, error) {
	f.url, f.validator = rawURL, validator
	return f.resp, f.err
}

func TestFRPRV004_FetchIsConditional(t *testing.T) {
	t.Parallel()
	full := &fakeFetcher{resp: httpfetch.Response{Body: fixture(t), LastModified: "Wed, 23 Sep 2026 11:52:04 GMT"}}
	got, err := New(full, ownerDefault).Fetch(context.Background(), "")
	if err != nil || got.NotModified || got.Validator == "" || full.url != URL(ownerDefault) {
		t.Errorf("full fetch = %+v, %v", got, err)
	}
	same := &fakeFetcher{resp: httpfetch.Response{NotModified: true}}
	again, err := New(same, ownerDefault).Fetch(context.Background(), got.Validator)
	if err != nil || !again.NotModified || again.Events != nil || same.validator != "Wed, 23 Sep 2026 11:52:04 GMT" {
		t.Errorf("conditional fetch = %+v, %v, sent %q", again, err, same.validator)
	}
	// A validator from before the minimum was part of it is sent as nothing.
	if _, _ = New(same, ownerDefault).Fetch(context.Background(), "Wed, 23 Sep 2026 11:52:04 GMT"); same.validator != "" {
		t.Errorf("a bare Last-Modified was sent as %q", same.validator)
	}
}

func TestFRSET002_SetMinimumSwitchesTheNextFetch(t *testing.T) {
	t.Parallel()
	f := &fakeFetcher{resp: httpfetch.Response{Body: fixture(t)}}
	a := New(f, ownerDefault)
	a.SetMinimum(4.5)
	if _, err := a.Fetch(context.Background(), ""); err != nil || f.url != URL(4.5) {
		t.Errorf("after SetMinimum(4.5) fetched %s, %v", f.url, err)
	}
}

// ownerDefault is FR-SET-002's default minimum, whose home is the settings.
const ownerDefault = 3.0

// A validator answers "changed since?" for one feed and one filter only. 2.5
// and 3.0 share the 2.5 feed, so a 304 there would keep the 2.5 set under a
// 3.0 setting (FR-SET-002).
func TestFRSET002_ValidatorIsNotSentAcrossAMinimumChange(t *testing.T) {
	t.Parallel()
	first := &fakeFetcher{resp: httpfetch.Response{Body: fixture(t), LastModified: "Wed, 23 Sep 2026 11:52:04 GMT"}}
	got, err := New(first, 2.5).Fetch(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	next := &fakeFetcher{resp: httpfetch.Response{Body: fixture(t)}}
	if _, err := New(next, ownerDefault).Fetch(context.Background(), got.Validator); err != nil {
		t.Fatal(err)
	}
	if next.validator != "" || next.url != URL(2.5) {
		t.Errorf("after 2.5 to 3.0 the adapter sent %q to %s", next.validator, next.url)
	}
}

func TestFetchFailures(t *testing.T) {
	t.Parallel()
	if _, err := New(&fakeFetcher{err: errors.New("timeout")}, ownerDefault).Fetch(context.Background(), ""); err == nil {
		t.Error("a failed fetch was reported as success")
	}
	if _, err := New(&fakeFetcher{resp: httpfetch.Response{Body: []byte("x")}}, ownerDefault).Fetch(context.Background(), ""); err == nil {
		t.Error("an unparseable body was reported as success")
	}
}

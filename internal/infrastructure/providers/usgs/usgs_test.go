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
		AllMagnitudes:  "all",
		1.0:            "1.0",
		2.0:            "1.0",
		DefaultMinimum: "2.5",
		4.5:            "4.5",
		6.0:            "4.5",
	}
	for minimum, name := range cases {
		want := "https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/" + name + "_week.geojson"
		if got := URL(minimum); got != want {
			t.Errorf("URL(%v) = %q, want %q", minimum, got, want)
		}
	}
	a := New(nil, DefaultMinimum)
	if a.Name() != event.USGS || a.Interval() != time.Minute {
		t.Error("Name or Interval wrong")
	}
}

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
	if first.Extras.DepthKm == nil || first.UpdatedAt.IsZero() || first.SourceURL != "https://earthquake.usgs.gov/earthquakes/eventpage/pr71534228" {
		t.Errorf("first extras = %+v", first)
	}
	atThree, _, _ := Parse(fixture(t), DefaultMinimum)
	for _, e := range atThree {
		if e.Observations[0].Measurement.Value < DefaultMinimum {
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
	if kept, _, _ := Parse([]byte(body), DefaultMinimum); len(kept) != 0 {
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
	got, err := New(full, DefaultMinimum).Fetch(context.Background(), "")
	if err != nil || got.NotModified || got.Validator != "Wed, 23 Sep 2026 11:52:04 GMT" || full.url != URL(DefaultMinimum) {
		t.Errorf("full fetch = %+v, %v", got, err)
	}
	same := &fakeFetcher{resp: httpfetch.Response{NotModified: true, LastModified: "v"}}
	got, err = New(same, DefaultMinimum).Fetch(context.Background(), "Wed, 23 Sep 2026 11:52:04 GMT")
	if err != nil || !got.NotModified || got.Events != nil || same.validator != "Wed, 23 Sep 2026 11:52:04 GMT" {
		t.Errorf("conditional fetch = %+v, %v, sent %q", got, err, same.validator)
	}
	if _, err := New(&fakeFetcher{err: errors.New("timeout")}, DefaultMinimum).Fetch(context.Background(), ""); err == nil {
		t.Error("a failed fetch was reported as success")
	}
	if _, err := New(&fakeFetcher{resp: httpfetch.Response{Body: []byte("x")}}, DefaultMinimum).Fetch(context.Background(), ""); err == nil {
		t.Error("an unparseable body was reported as success")
	}
}

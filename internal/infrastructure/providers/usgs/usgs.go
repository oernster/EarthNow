// Package usgs adapts the USGS real-time earthquake GeoJSON feeds to the event
// model (REQUIREMENTS.md FR-PRV-003, FR-PRV-004, DATA-010). Nothing outside
// this package knows the feed's schema.
package usgs

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch"
)

// Host is the only host this adapter reaches.
const Host = "earthquake.usgs.gov"

// Interval matches the feed's measured Cache-Control max-age of 60 s (FR-PRV-005).
const Interval = time.Minute

// AllMagnitudes is the minimum that keeps every event, "all" in settings.
var AllMagnitudes = math.Inf(-1)

// DefaultMinimum is the owner's default of 3.0 (FR-PRV-003).
const DefaultMinimum = 3.0

// feed is one published threshold: the feeds exist only at these (measured).
type feed struct {
	minimum float64
	name    string
}

// feeds are the published week feeds, lowest threshold first.
var feeds = []feed{{AllMagnitudes, "all"}, {1.0, "1.0"}, {2.5, "2.5"}, {4.5, "4.5"}}

// earthquakeType is the USGS type that maps to Earthquake; every other type
// (quarry blast, explosion, ice quake) is Other with the type kept.
const earthquakeType = "earthquake"

// deletedStatus marks an event USGS has withdrawn; it is not shown.
const deletedStatus = "deleted"

// Fetcher is what the adapter needs from the HTTP client.
type Fetcher interface {
	Get(ctx context.Context, rawURL, validator string) (httpfetch.Response, error)
}

// Adapter is the USGS provider.
type Adapter struct {
	fetcher Fetcher
	minimum float64
}

// New builds the adapter keeping events at or above minimum magnitude.
func New(fetcher Fetcher, minimum float64) *Adapter {
	return &Adapter{fetcher: fetcher, minimum: minimum}
}

// Name implements ports.Provider.
func (a *Adapter) Name() event.Provider { return event.USGS }

// Interval implements ports.Provider.
func (a *Adapter) Interval() time.Duration { return Interval }

// URL is the week feed at the highest threshold not above minimum.
func URL(minimum float64) string {
	chosen := feeds[0]
	for _, f := range feeds {
		if f.minimum <= minimum {
			chosen = f
		}
	}
	return fmt.Sprintf("https://%s/earthquakes/feed/v1.0/summary/%s_week.geojson", Host, chosen.name)
}

// Fetch implements ports.Provider, conditional on the last Last-Modified.
func (a *Adapter) Fetch(ctx context.Context, validator string) (ports.Fetched, error) {
	resp, err := a.fetcher.Get(ctx, URL(a.minimum), validator)
	if err != nil {
		return ports.Fetched{}, fmt.Errorf("USGS: %w", err)
	}
	if resp.NotModified {
		return ports.Fetched{NotModified: true, Validator: resp.LastModified}, nil
	}
	events, dropped, err := Parse(resp.Body, a.minimum)
	if err != nil {
		return ports.Fetched{}, fmt.Errorf("USGS: %w", err)
	}
	return ports.Fetched{Events: events, Validator: resp.LastModified, Dropped: dropped}, nil
}

type wireFeature struct {
	ID         string `json:"id"`
	Properties struct {
		Mag     *float64 `json:"mag"`
		Place   string   `json:"place"`
		Time    *int64   `json:"time"`
		Updated *int64   `json:"updated"`
		URL     string   `json:"url"`
		Status  string   `json:"status"`
		Tsunami int      `json:"tsunami"`
		Type    string   `json:"type"`
		MagType string   `json:"magType"`
		Title   string   `json:"title"`
	} `json:"properties"`
	Geometry *struct {
		Coordinates []float64 `json:"coordinates"`
	} `json:"geometry"`
}

// Parse maps a feed body, keeping events at or above minimum. A malformed
// feature is dropped and counted; a withdrawn one is left out uncounted.
func Parse(body []byte, minimum float64) ([]event.Event, int, error) {
	var doc struct {
		Features *[]wireFeature `json:"features"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, 0, fmt.Errorf("parsing feed: %w", err)
	}
	if doc.Features == nil {
		return nil, 0, fmt.Errorf("parsing feed: no features list")
	}
	var out []event.Event
	dropped := 0
	for _, f := range *doc.Features {
		if f.Properties.Status == deletedStatus {
			continue
		}
		e, ok := mapFeature(f)
		if !ok {
			dropped++
			continue
		}
		if minimum != AllMagnitudes && (f.Properties.Mag == nil || *f.Properties.Mag < minimum) {
			continue
		}
		out = append(out, e)
	}
	return out, dropped, nil
}

// depthIndex is where GeoJSON puts depth, in kilometres (DATA-010).
const depthIndex = 2

func mapFeature(f wireFeature) (event.Event, bool) {
	p := f.Properties
	if f.ID == "" || p.Time == nil || f.Geometry == nil || len(f.Geometry.Coordinates) < 2 {
		return event.Event{}, false
	}
	where, err := event.NewPoint(f.Geometry.Coordinates[1], f.Geometry.Coordinates[0])
	if err != nil {
		return event.Event{}, false
	}
	o := event.Observation{At: time.UnixMilli(*p.Time).UTC(), Precision: event.Instant, Where: where}
	if p.Mag != nil {
		o.Measurement = &event.Measurement{Value: *p.Mag, Unit: p.MagType}
	}
	e := event.Event{
		Provider:        event.USGS,
		ProviderEventID: f.ID,
		Category:        event.Other,
		Title:           p.Title,
		Description:     p.Place,
		Observations:    []event.Observation{o},
		Status:          p.Status,
		SourceURL:       p.URL,
		Extras:          event.Extras{SourceCategory: p.Type, TsunamiFlag: p.Tsunami == 1, MagnitudeType: p.MagType},
	}
	if p.Type == earthquakeType {
		e.Category = event.Earthquake
	}
	if e.Title == "" {
		e.Title = p.Place
	}
	if p.Updated != nil {
		e.UpdatedAt = time.UnixMilli(*p.Updated).UTC()
	}
	if len(f.Geometry.Coordinates) > depthIndex {
		depth := f.Geometry.Coordinates[depthIndex]
		e.Extras.DepthKm = &depth
	}
	return e, true
}

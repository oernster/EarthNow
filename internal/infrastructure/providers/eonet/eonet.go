// Package eonet adapts NASA EONET v3 to the event model (REQUIREMENTS.md
// FR-PRV-001, FR-PRV-002, DATA-005, DATA-006, DATA-007). Nothing outside this
// package knows EONET's schema.
package eonet

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
	"github.com/oernster/EarthNow/internal/domain/window"
	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch"
)

// Host is the only host this adapter reaches.
const Host = "eonet.gsfc.nasa.gov"

// Interval is EONET's refresh interval (FR-PRV-005): its events change daily
// and its measured rate limit is 60 requests per unknown window (ASM-002).
const Interval = 10 * time.Minute

// hoursPerDay turns the widest window into EONET's days parameter.
const hoursPerDay = 24

// categories is DATA-006: EONET's category ids to the internal vocabulary. An
// id missing here maps to Other with the id kept in Extras.
var categories = map[string]event.Category{
	"earthquakes":  event.Earthquake,
	"volcanoes":    event.Volcano,
	"wildfires":    event.Wildfire,
	"severeStorms": event.SevereStorm,
	"floods":       event.Flood,
	"landslides":   event.Landslide,
	"drought":      event.Drought,
	"dustHaze":     event.Dust,
	"seaLakeIce":   event.Ice,
}

// Fetcher is what the adapter needs from the HTTP client.
type Fetcher interface {
	Get(ctx context.Context, rawURL, validator string) (httpfetch.Response, error)
}

// Adapter is the EONET provider.
type Adapter struct {
	fetcher Fetcher
}

// New builds the adapter over fetcher.
func New(fetcher Fetcher) *Adapter { return &Adapter{fetcher: fetcher} }

// Name implements ports.Provider.
func (a *Adapter) Name() event.Provider { return event.EONET }

// Interval implements ports.Provider.
func (a *Adapter) Interval() time.Duration { return Interval }

// URL is the request of FR-PRV-001: open events over the widest window.
func URL() string {
	days := int(window.Widest.Length / (hoursPerDay * time.Hour))
	return fmt.Sprintf("https://%s/api/v3/events?status=open&days=%d", Host, days)
}

// Fetch implements ports.Provider. EONET sends no Last-Modified (measured
// 2026-09-23), so every fetch is a full one.
func (a *Adapter) Fetch(ctx context.Context, _ string) (ports.Fetched, error) {
	resp, err := a.fetcher.Get(ctx, URL(), "")
	if err != nil {
		return ports.Fetched{}, fmt.Errorf("EONET: %w", err)
	}
	events, dropped, err := Parse(resp.Body)
	if err != nil {
		return ports.Fetched{}, fmt.Errorf("EONET: %w", err)
	}
	return ports.Fetched{Events: events, Dropped: dropped}, nil
}

type wireGeometry struct {
	Date      string          `json:"date"`
	Type      string          `json:"type"`
	Coords    json.RawMessage `json:"coordinates"`
	Magnitude *float64        `json:"magnitudeValue"`
	Unit      *string         `json:"magnitudeUnit"`
}

type wireEvent struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Closed      *string `json:"closed"`
	Categories  []struct {
		ID string `json:"id"`
	} `json:"categories"`
	Sources []struct {
		URL string `json:"url"`
	} `json:"sources"`
	Geometry []wireGeometry `json:"geometry"`
}

// Parse maps an EONET events body. A body that is not the expected document
// is an error; a single unusable event is dropped and counted (FR-PRV-012,
// FR-PRV-013). The Content-Type is never consulted (FR-PRV-002).
func Parse(body []byte) ([]event.Event, int, error) {
	var doc struct {
		Events *[]wireEvent `json:"events"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, 0, fmt.Errorf("parsing events: %w", err)
	}
	if doc.Events == nil {
		return nil, 0, fmt.Errorf("parsing events: no events list")
	}
	var out []event.Event
	dropped := 0
	for _, w := range *doc.Events {
		e, ok := mapEvent(w)
		if !ok {
			dropped++
			continue
		}
		out = append(out, e)
	}
	return out, dropped, nil
}

func mapEvent(w wireEvent) (event.Event, bool) {
	if w.ID == "" || w.Title == "" {
		return event.Event{}, false
	}
	var obs []event.Observation
	allMidnight := true
	for _, g := range w.Geometry {
		o, ok := mapGeometry(g)
		if !ok {
			continue
		}
		allMidnight = allMidnight && isMidnight(o.At)
		obs = append(obs, o)
	}
	if len(obs) == 0 {
		return event.Event{}, false
	}
	// DATA-005 as amended: day precision only when EVERY date is midnight. A
	// storm track holds genuine 00:00 fixes beside 06:00, 12:00 and 18:00.
	if allMidnight {
		for i := range obs {
			obs[i].Precision = event.Day
		}
	}
	sort.SliceStable(obs, func(i, j int) bool { return obs[i].At.Before(obs[j].At) })
	e := event.Event{
		Provider:        event.EONET,
		ProviderEventID: w.ID,
		Category:        event.Other,
		Title:           w.Title,
		Observations:    obs,
		Status:          "open",
	}
	if w.Closed != nil {
		e.Status = "closed"
	}
	if w.Description != nil {
		e.Description = *w.Description
	}
	if len(w.Categories) > 0 {
		e.Extras.SourceCategory = w.Categories[0].ID
		if c, known := categories[w.Categories[0].ID]; known {
			e.Category = c
		}
	}
	// The first source that is a page, since a storm's first source is often a
	// JTWC data file the browser would download (FR-SEL-009).
	addresses := make([]string, 0, len(w.Sources))
	for _, s := range w.Sources {
		addresses = append(addresses, s.URL)
	}
	e.SourceURL = event.FirstPage(addresses)
	return e, true
}

func isMidnight(t time.Time) bool {
	return t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0
}

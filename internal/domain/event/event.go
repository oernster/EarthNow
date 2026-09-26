// Package event is the provider-neutral model of one thing happening on Earth
// (REQUIREMENTS.md DATA-001 to DATA-011). No provider's schema reaches past the
// adapters; everything the application and the page know about an event is here.
package event

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// Category is the internal vocabulary of DATA-002. A provider's own terms never
// appear outside its adapter.
type Category string

// The categories, one constant each.
const (
	Earthquake  Category = "EARTHQUAKE"
	Volcano     Category = "VOLCANO"
	Wildfire    Category = "WILDFIRE"
	SevereStorm Category = "SEVERE_STORM"
	Flood       Category = "FLOOD"
	Ice         Category = "ICE"
	Other       Category = "OTHER"
)

// Categories answers the whole vocabulary in DATA-002 order, the order the key
// and the filter list it in (FR-KEY-001). A kind no provider has been measured
// to publish has no category of its own; it is Other (amendment 45).
func Categories() []Category {
	return []Category{Earthquake, Volcano, Wildfire, SevereStorm, Flood, Ice, Other}
}

// Provider names the public source an event came from.
type Provider string

// The providers of V1. GVP is the Smithsonian / USGS Weekly Volcanic Activity
// Report, added because EONET tracked no volcano in the month to 2026-09-23 while
// the report listed twenty (FR-PRV-015).
const (
	EONET Provider = "EONET"
	USGS  Provider = "USGS"
	GVP   Provider = "GVP"
)

// Precision says how much of an observation's time the source actually knows.
type Precision int

// Instant precision is a full timestamp; Day precision is a date only
// (DATA-005: an EONET date of exactly 00:00:00Z).
const (
	Instant Precision = iota
	Day
)

// Latitude and longitude limits of NFR-SEC-003.
const (
	maxLatitude  = 90
	maxLongitude = 180
)

// ErrInvalidCoordinate is returned for a coordinate outside the Earth or not a number.
var ErrInvalidCoordinate = errors.New("invalid coordinate")

// Point is a validated position in degrees.
type Point struct {
	Lat float64
	Lng float64
}

// NewPoint validates a coordinate before anything may be built on it.
func NewPoint(lat, lng float64) (Point, error) {
	finite := !math.IsNaN(lat) && !math.IsNaN(lng) && !math.IsInf(lat, 0) && !math.IsInf(lng, 0)
	if !finite || math.Abs(lat) > maxLatitude || math.Abs(lng) > maxLongitude {
		return Point{}, fmt.Errorf("%w: %v, %v", ErrInvalidCoordinate, lat, lng)
	}
	return Point{Lat: lat, Lng: lng}, nil
}

// Measurement is a source-reported quantity in the source's own unit (DATA-011).
type Measurement struct {
	Value float64
	Unit  string
}

// Observation is one reported sighting: where, when and how precisely the time
// is known, plus the measurement the source attached to it, if any.
type Observation struct {
	At          time.Time
	Precision   Precision
	Where       Point
	Measurement *Measurement
}

// Extras are the named provider details the product keeps (DATA-001's closed
// set). Each is empty when the source did not supply it.
type Extras struct {
	SourceCategory string
	DepthKm        *float64
	TsunamiFlag    bool
	MagnitudeType  string
}

// Event is one EarthEvent. Observations are held oldest first. UpdatedAt is
// zero when the source gives no update time.
type Event struct {
	Provider        Provider
	ProviderEventID string
	Category        Category
	Title           string
	Description     string
	Observations    []Observation
	UpdatedAt       time.Time
	Status          string
	SourceURL       string
	Extras          Extras
	// Report is the report an ongoing event comes from; zero for any other.
	Report Report
}

// StatusOpen and StatusClosed are EONET's two states for an event. A closed event
// still happened inside the window and is shown, marked as ended (FR-PRV-001).
const (
	StatusOpen   = "open"
	StatusClosed = "closed"
)

// Ended reports whether the source has closed the event.
func (e Event) Ended() bool { return e.Status == StatusClosed }

// ID is the event's identity across refreshes (DATA-008).
func (e Event) ID() string {
	return string(e.Provider) + ":" + e.ProviderEventID
}

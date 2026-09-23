// Package ports declares what the application needs from the outside world.
// Infrastructure implements these; the application never imports it
// (REQUIREMENTS.md FR-PRV-014, CON-002).
package ports

import (
	"context"
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
)

// Clock is the only way the application learns the time.
type Clock interface {
	Now() time.Time
}

// Geocoder words where a point is (FR-GEO-001 to 003).
type Geocoder interface {
	Describe(lat, lng float64) string
}

// Fetched is one successful answer from a provider.
type Fetched struct {
	// Events is the provider's whole current set; empty on NotModified.
	Events []event.Event
	// NotModified is true when the source said nothing changed since Validator.
	NotModified bool
	// Validator is what to send next time to ask "changed since?"; empty when
	// the source gives none.
	Validator string
	// Dropped counts items the adapter refused as malformed (FR-PRV-013).
	Dropped int
}

// Provider is one public source behind an adapter.
type Provider interface {
	// Name is the provider this adapter serves.
	Name() event.Provider
	// Interval is the provider's normal refresh interval (FR-PRV-005).
	Interval() time.Duration
	// Fetch retrieves the current set, sending validator when not empty.
	Fetch(ctx context.Context, validator string) (Fetched, error)
}

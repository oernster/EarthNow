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

// Snapshot is one provider's last successful set and when it arrived.
type Snapshot struct {
	Events      []event.Event
	RetrievedAt time.Time
	Validator   string
}

// SnapshotCache keeps each provider's last successful set across runs
// (FR-STS-004, DATA-009). Load answers false when nothing is held.
type SnapshotCache interface {
	Load(p event.Provider) (Snapshot, bool, error)
	Save(p event.Provider, snap Snapshot) error
}

// Settings is what the reader has chosen that outlives a run (FR-SET-001 to
// 003, FR-FLT-005). Choices are held by key; the application owns what each
// key means.
type Settings struct {
	AutoRotate       bool
	Magnitude        string
	Speed            string
	Window           string
	HiddenCategories []string
	HiddenProviders  []string
	// CloudsShown is whether the cloud layer is drawn (FR-CLD-003).
	CloudsShown bool
}

// SettingsStore keeps the settings between runs (FR-SET-004). Load starts from
// defaults, so a field the file does not hold keeps its default; it answers
// false with no error when there is no file, which is absence rather than a
// fault. Path names the file, so a notice can say which one was not read.
type SettingsStore interface {
	Load(defaults Settings) (Settings, bool, error)
	Save(Settings) error
	Path() string
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

// CloudImage is one cloud image ready to draw (FR-CLD-006, FR-CLD-007) and
// the valid time it shows.
type CloudImage struct {
	ValidTime time.Time
	PNG       []byte
}

// CloudSource is the cloud service behind its adapter (FR-CLD-004,
// FR-CLD-016). Image answers the image already drawn, so the page is handed
// pixels and fetches nothing (NFR-SEC-002); an answer that is not the image
// asked for is an error (FR-CLD-012).
type CloudSource interface {
	Latest(ctx context.Context) (time.Time, error)
	Image(ctx context.Context, validTime time.Time) ([]byte, error)
}

// CloudCache keeps the last good cloud image across runs (FR-CLD-014). Load
// answers false with no error when nothing is held.
type CloudCache interface {
	Load() (CloudImage, bool, error)
	Save(CloudImage) error
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

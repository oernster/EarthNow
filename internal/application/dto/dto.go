// Package dto is the wire between the Go side and the page. Every type here is
// stated a second time as a TypeScript interface in frontend/src/types.ts; a
// structural test compares the two (NFR-MNT-003).
package dto

// Event is one displayed event with every value the page shows already
// worded, so the page formats nothing but local time.
type Event struct {
	ID          string  `json:"id"`
	Provider    string  `json:"provider"`
	Category    string  `json:"category"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	At          string  `json:"at"`
	DayOnly     bool    `json:"dayOnly"`
	Reported    string  `json:"reported"`
	RetrievedAt string  `json:"retrievedAt"`
	Retrieved   string  `json:"retrieved"`
	Measurement string  `json:"measurement"`
	Band        int     `json:"band"`
	SourceURL   string  `json:"sourceUrl"`
	SourceText  string  `json:"sourceText"`
	// Ended is true when the source has closed the event (FR-PRV-001).
	Ended bool `json:"ended"`
}

// Provider is one source's state for the status area.
type Provider struct {
	Name    string `json:"name"`
	Loading bool   `json:"loading"`
	// Refreshing is true while a fetch runs with events already held (FR-STS-007).
	Refreshing bool   `json:"refreshing"`
	Stale      bool   `json:"stale"`
	Retrieved  string `json:"retrieved"`
	Problem    string `json:"problem"`
	// NextAttempt is filled only while Problem is: "next attempt in 4 min".
	NextAttempt string `json:"nextAttempt"`
}

// View is everything the globe shows for one window and filter.
type View struct {
	WindowKey string         `json:"windowKey"`
	CountLine string         `json:"countLine"`
	Events    []Event        `json:"events"`
	Counts    map[string]int `json:"counts"`
	Providers []Provider     `json:"providers"`
	// Notice is a standing problem the reader should know about; empty when none.
	Notice string `json:"notice"`
}

// Choice is one option of a control: a time window, a minimum magnitude.
type Choice struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// Speed is one idle rotation preset (FR-SET-003) with the period the globe
// turns at, so the page holds no rotation figure of its own.
type Speed struct {
	Key                  string `json:"key"`
	Label                string `json:"label"`
	SecondsPerRevolution int    `json:"secondsPerRevolution"`
}

// Settings is what the reader has chosen that outlives a run (FR-SET-001 to
// 003, FR-FLT-005). Every list is empty rather than absent.
type Settings struct {
	AutoRotate       bool     `json:"autoRotate"`
	Magnitude        string   `json:"magnitude"`
	Speed            string   `json:"speed"`
	WindowKey        string   `json:"windowKey"`
	HiddenCategories []string `json:"hiddenCategories"`
	HiddenProviders  []string `json:"hiddenProviders"`
	CloudsShown      bool     `json:"cloudsShown"`
	// DayNightShown is whether the day and night layer is drawn (FR-DAY-007).
	DayNightShown bool `json:"dayNightShown"`
}

// Clouds is the cloud layer's state (FR-CLD-009, FR-CLD-013). The image
// itself is asked for apart, since it is large; ValidTime tells the page when
// the one it holds has been replaced.
type Clouds struct {
	Shown bool `json:"shown"`
	// Line is the status line; empty while the layer is hidden.
	Line string `json:"line"`
	// ValidTime is the held image's valid time, RFC 3339; empty when none.
	ValidTime string `json:"validTime"`
	// Notice is a cache problem the reader should know about; empty when none.
	Notice string `json:"notice"`
	// Provider is the cloud service's entry for the provider status popover
	// (FR-CLD-011), filled only while the layer is shown. It travels here, not
	// in View.Providers, since those are the event sources the key filters.
	Provider Provider `json:"provider"`
}

// Sun is where the sun stands overhead now (FR-DAY-001), with the two
// figures the page draws the light by, so the page holds no light rule of
// its own: the twilight limit of FR-DAY-002 in degrees and FR-DAY-009's
// share of a cloud's opacity kept at night.
type Sun struct {
	Lat             float64 `json:"lat"`
	Lng             float64 `json:"lng"`
	TwilightDegrees float64 `json:"twilightDegrees"`
	CloudNightFloor float64 `json:"cloudNightFloor"`
}

// About is what the About dialog shows (FR-HLP-001).
type About struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Copyright    string   `json:"copyright"`
	Licence      string   `json:"licence"`
	Attributions []string `json:"attributions"`
}

// SettingChoices is what the settings dialog offers.
type SettingChoices struct {
	Magnitudes []Choice `json:"magnitudes"`
	Speeds     []Speed  `json:"speeds"`
}

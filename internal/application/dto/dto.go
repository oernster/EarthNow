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
}

// Provider is one source's state for the status area.
type Provider struct {
	Name      string `json:"name"`
	Loading   bool   `json:"loading"`
	Stale     bool   `json:"stale"`
	Retrieved string `json:"retrieved"`
	Problem   string `json:"problem"`
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

// Window is one choice in the time window control.
type Window struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

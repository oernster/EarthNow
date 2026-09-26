// Package services holds the application's use cases.
package services

import (
	"sort"
	"sync"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
	"github.com/oernster/EarthNow/internal/domain/window"
)

// Snapshot is the port's snapshot, named here for the store's callers.
type Snapshot = ports.Snapshot

// Shown is an event as displayed: the event plus the observation that places
// it for the selected window (DATA-003, DATA-004).
type Shown struct {
	Event       event.Event
	Observation event.Observation
	RetrievedAt time.Time
	// Trail is the storm's positions inside the window (FR-TRL-001); nil otherwise.
	Trail []event.Point
}

// Filter says what the user has switched off (FR-FLT-002, FR-FLT-004). The
// zero value hides nothing, which is "All events".
type Filter struct {
	HiddenCategories map[event.Category]bool
	HiddenProviders  map[event.Provider]bool
}

func (f Filter) hides(e event.Event) bool {
	return f.HiddenCategories[e.Category] || f.HiddenProviders[e.Provider]
}

// Store holds each provider's latest set. A failed provider keeps its last
// set, so one provider failing never touches another's (FR-PRV-008).
type Store struct {
	mu    sync.RWMutex
	clock ports.Clock
	sets  map[event.Provider]Snapshot
}

// NewStore builds an empty store reading the time from clock.
func NewStore(clock ports.Clock) *Store {
	return &Store{clock: clock, sets: map[event.Provider]Snapshot{}}
}

// Apply records a successful fetch. A not-modified answer keeps the stored
// events and refreshes only the retrieval time (FR-PRV-004). A new set
// replaces the old one whole; within it, a repeated id keeps the last copy
// (DATA-008).
func (s *Store) Apply(provider event.Provider, fetched ports.Fetched) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock.Now()
	previous := s.sets[provider]
	if fetched.NotModified {
		previous.RetrievedAt = now
		if fetched.Validator != "" {
			previous.Validator = fetched.Validator
		}
		s.sets[provider] = previous
		return
	}
	s.sets[provider] = Snapshot{Events: unique(fetched.Events), RetrievedAt: now, Validator: fetched.Validator}
}

// Restore puts a cached snapshot back, as read at start before any fetch
// (FR-STS-004).
func (s *Store) Restore(provider event.Provider, snap Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap.Events = unique(snap.Events)
	s.sets[provider] = snap
}

// Snapshot answers one provider's stored set; false when nothing is held.
func (s *Store) Snapshot(provider event.Provider) (Snapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, ok := s.sets[provider]
	return snap, ok
}

// Visible answers the events inside the window ending now and not filtered
// out, newest first, which is the order the keyboard cursor walks (NFR-KBD-004).
func (s *Store) Visible(w window.Window, f Filter) []Shown {
	return s.Within(w.Now(s.clock.Now()), f)
}

// Within answers the events inside a range and not filtered out, newest first:
// the live window's range or a replay's (FR-RPL-009). An ongoing event whose
// report is past its currency is never shown (FR-PRV-016).
func (s *Store) Within(r window.Range, f Filter) []Shown {
	now := s.clock.Now()
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Shown
	for _, snap := range s.sets {
		for _, e := range snap.Events {
			if f.hides(e) || (e.Ongoing() && !e.Report.Current(now)) {
				continue
			}
			if o, ok := r.Latest(e); ok {
				out = append(out, Shown{Event: e, Observation: o, RetrievedAt: snap.RetrievedAt, Trail: r.Trail(e)})
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].Observation.At.Equal(out[j].Observation.At) {
			return out[i].Observation.At.After(out[j].Observation.At)
		}
		return out[i].Event.ID() < out[j].Event.ID()
	})
	return out
}

// Counts answers how many shown events fall in each category (FR-KEY-004).
func Counts(shown []Shown) map[event.Category]int {
	counts := map[event.Category]int{}
	for _, s := range shown {
		counts[s.Event.Category]++
	}
	return counts
}

// unique keeps the last copy of each id, in first-seen order.
func unique(events []event.Event) []event.Event {
	index := map[string]int{}
	out := make([]event.Event, 0, len(events))
	for _, e := range events {
		if i, seen := index[e.ID()]; seen {
			out[i] = e
			continue
		}
		index[e.ID()] = len(out)
		out = append(out, e)
	}
	return out
}

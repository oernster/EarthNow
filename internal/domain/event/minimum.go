package event

import "math"

// NoMinimum is the minimum magnitude that keeps every event, those with no
// magnitude included: "All" in settings (FR-SET-002).
var NoMinimum = math.Inf(-1)

// MeetsMinimum is FR-SET-002's rule, in its one home: the USGS adapter applies
// it to what it parses and the store to what it shows, so a minimum raised
// while USGS cannot be reached hides the held quakes below it at once. The
// minimum governs USGS events only. Under NoMinimum every event is kept;
// otherwise a USGS event is kept when a magnitude it carries is at or above
// the minimum. One with no magnitude is not.
func MeetsMinimum(e Event, minimum float64) bool {
	if e.Provider != USGS || minimum == NoMinimum {
		return true
	}
	for _, o := range e.Observations {
		if o.Measurement != nil && o.Measurement.Value >= minimum {
			return true
		}
	}
	return false
}

// Package burnt holds the rules of the burnt-area layer (REQUIREMENTS.md
// 3.2.13): how the days of a window are drawn together (FR-BA-006) and how the
// layer's state is worded (FR-BA-008, FR-BA-009, FR-BA-017).
package burnt

import (
	"fmt"
	"math"
	"time"

	"github.com/oernster/EarthNow/internal/domain/freshness"
)

// The source's own red, in which every burnt pixel is drawn (FR-BA-006):
// GWIS draws (255, 0, 0), measured 2026-09-24.
const (
	Red   = math.MaxUint8
	Green = 0
	Blue  = 0
)

// The status lines with nothing to show (FR-BA-009, FR-BA-017) and while the
// first maps are on their way.
const (
	NoneYet     = "Burnt areas: none mapped yet for this window"
	Unavailable = "Burnt areas: the maps could not be retrieved"
	Retrieving  = "Burnt areas: retrieving the maps"
)

// The day layouts of FR-BA-008: "17 to 23 Sep", "30 Sep to 2 Oct", "23 Sep".
const (
	dayLayout      = "2 Jan"
	dayAloneLayout = "2"
)

// Union is FR-BA-006: each opacity in into becomes the higher of itself and
// the same pixel's in from. Planes of different sizes are joined over the
// pixels both hold; the adapter refuses an image of the wrong size before
// either reaches here.
func Union(into, from []uint8) {
	for i := range min(len(into), len(from)) {
		into[i] = max(into[i], from[i])
	}
}

// Span words the days from first to last in UTC: one day alone, a span inside
// one month with the month said once, else both months.
func Span(first, last time.Time) string {
	first, last = first.UTC(), last.UTC()
	switch {
	case first.Equal(last):
		return first.Format(dayLayout)
	case first.Month() == last.Month() && first.Year() == last.Year():
		return first.Format(dayAloneLayout) + " to " + last.Format(dayLayout)
	default:
		return first.Format(dayLayout) + " to " + last.Format(dayLayout)
	}
}

// Status is FR-BA-008: "Burnt areas: 17 to 23 Sep (UTC), retrieved 12 min ago".
func Status(first, last, retrieved, now time.Time) string {
	return fmt.Sprintf("Burnt areas: %s (UTC), retrieved %s", Span(first, last), freshness.Age(retrieved, now))
}

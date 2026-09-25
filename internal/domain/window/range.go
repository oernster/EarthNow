package window

import (
	"math"
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
)

// Range is a stretch of time an event is shown for: after From, up to and
// including To. The live window is one (FR-TW-002); a replay is another, from
// the span's start up to the replay instant (FR-RPL-009).
type Range struct {
	From, To time.Time
}

// Now is the live range of the window ending at now, stretched by ClockSkew.
func (w Window) Now(now time.Time) Range {
	return Range{From: now.Add(-w.Length), To: now.Add(ClockSkew)}
}

// At is FR-RPL-001: the replay instant a position from 0 to 1 along the
// window's span ending at end names, a position outside that held at the
// nearer end.
func (w Window) At(end time.Time, position float64) time.Time {
	share := math.Min(math.Max(position, 0), 1)
	return end.Add(-w.Length).Add(time.Duration(share * float64(w.Length)))
}

// Replay is the range a replay shows at a position: from the span's start up
// to the replay instant, so the view builds up (FR-RPL-009).
func (w Window) Replay(end time.Time, position float64) Range {
	return Range{From: end.Add(-w.Length), To: w.At(end, position)}
}

// Contains reports whether an instant falls in the range.
func (r Range) Contains(at time.Time) bool {
	return at.After(r.From) && !at.After(r.To)
}

// Latest answers the event's newest observation inside the range; false when
// none falls inside it, which means the event is not shown. That observation's
// time is the event time and its position the marker position (DATA-003).
func (r Range) Latest(e event.Event) (event.Observation, bool) {
	for i := len(e.Observations) - 1; i >= 0; i-- {
		if o := e.Observations[i]; r.Contains(o.At) {
			return o, true
		}
	}
	return event.Observation{}, false
}

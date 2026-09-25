package window

import (
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
)

// minTrailPoints is the fewest positions that make a line.
const minTrailPoints = 2

// Trail is FR-TRL-001 over the window ending at now.
func (w Window) Trail(e event.Event, now time.Time) []event.Point { return w.Now(now).Trail(e) }

// Trail is FR-TRL-001: a severe storm's positions inside the range, oldest
// first, ending at its marker position; nil for any other event or where fewer
// than two fall inside. Only the range's own fixes count, so it is recent
// movement and never an archive; a replay's range grows it (FR-RPL-010).
func (r Range) Trail(e event.Event) []event.Point {
	if e.Category != event.SevereStorm {
		return nil
	}
	var points []event.Point
	for _, o := range e.Observations {
		if r.Contains(o.At) {
			points = append(points, o.Where)
		}
	}
	if len(points) < minTrailPoints {
		return nil
	}
	return points
}

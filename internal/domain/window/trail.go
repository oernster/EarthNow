package window

import (
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
)

// minTrailPoints is the fewest positions that make a line.
const minTrailPoints = 2

// Trail is FR-TRL-001: a severe storm's positions inside the window, oldest
// first, ending at its marker position; nil for any other event or where fewer
// than two fall inside. Only the window's own fixes count, so it is recent
// movement and never an archive.
func (w Window) Trail(e event.Event, now time.Time) []event.Point {
	if e.Category != event.SevereStorm {
		return nil
	}
	var points []event.Point
	for _, o := range e.Observations {
		if w.Contains(o.At, now) {
			points = append(points, o.Where)
		}
	}
	if len(points) < minTrailPoints {
		return nil
	}
	return points
}

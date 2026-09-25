package services

import (
	"time"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/domain/freshness"
)

// Replay answers what a replay shows at a position along its span (3.2.14),
// drawing on the other use cases for each layer. The span ends where the
// replay started (FR-RPL-003); the page holds the position and the end.
type Replay struct {
	globe  *Globe
	sun    *Sun
	clouds *Clouds
	burnt  *BurntAreas
	replay *ReplayClouds
}

// NewReplay builds the use case over the layers it replays.
func NewReplay(globe *Globe, sun *Sun, clouds *Clouds, burnt *BurntAreas, replay *ReplayClouds) *Replay {
	return &Replay{globe: globe, sun: sun, clouds: clouds, burnt: burnt, replay: replay}
}

// Frame answers the replay at a position along the span of windowKey ending at
// end. While the cloud layer is shown it follows the span's cloud images; it
// answers true when a round to fetch them has become due (FR-RPL-015).
func (r *Replay) Frame(windowKey string, f Filter, end time.Time, position float64) (dto.ReplayFrame, bool) {
	w := windowOf(windowKey)
	span := w.Replay(end, position)
	at := span.To
	due := false
	if r.clouds.Status().Shown {
		due = r.replay.Follow(span.From, end)
	} else {
		r.replay.Stop()
	}
	line, provider := r.replay.Status()
	return dto.ReplayFrame{
		View:           r.globe.ReplayView(w, span, f),
		At:             at.UTC().Format(timeLayout),
		Line:           freshness.ReplayLine(at),
		Sun:            r.sun.PositionAt(at),
		CloudTime:      r.replay.Pick(at),
		CloudsLine:     line,
		CloudsProvider: provider,
		BurntKey:       r.burnt.KeyAt(w, end, at),
	}, due
}

// BurntImage answers the burnt-area image of the span's days up to the
// position (FR-RPL-013).
func (r *Replay) BurntImage(windowKey string, end time.Time, position float64) string {
	w := windowOf(windowKey)
	return r.burnt.ImageAt(w, end, w.At(end, position))
}

// End releases the replay's cloud images as the replay ends (FR-RPL-018).
func (r *Replay) End() { r.replay.Stop() }

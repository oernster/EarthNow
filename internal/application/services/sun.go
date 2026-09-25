package services

import (
	"time"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/sun"
)

// Sun is the day and night layer's use case: it answers where the sun stands
// overhead now (FR-DAY-001), with the two figures the page draws the light by
// (FR-DAY-002, FR-DAY-009), so the page holds no light rule of its own. The
// page asks again at least once a minute (FR-DAY-004).
type Sun struct {
	clock ports.Clock
}

// NewSun builds the use case over the clock.
func NewSun(clock ports.Clock) *Sun {
	return &Sun{clock: clock}
}

// Position answers the sun now as the page draws it.
func (s *Sun) Position() dto.Sun { return s.PositionAt(s.clock.Now()) }

// PositionAt answers the sun at an instant: now, else a replay's (FR-RPL-012).
func (s *Sun) PositionAt(instant time.Time) dto.Sun {
	at := sun.Subsolar(instant)
	return dto.Sun{
		Lat:             at.Lat,
		Lng:             at.Lng,
		TwilightDegrees: sun.TwilightDegrees,
		CloudNightFloor: sun.CloudNightFloor,
	}
}

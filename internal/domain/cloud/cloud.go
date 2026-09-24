// Package cloud holds the rules of the cloud layer (REQUIREMENTS.md 3.2.10):
// how an infrared brightness becomes an opacity (FR-CLD-006), what a pixel with
// no data is drawn as (FR-CLD-007) and how the held image's time is worded
// (FR-CLD-009, FR-CLD-010).
package cloud

import (
	"fmt"
	"math"
	"time"

	"github.com/oernster/EarthNow/internal/domain/freshness"
)

// The two opacity thresholds of FR-CLD-006, as brightness out of 255, measured
// by the cloud spike against EUMETSAT's cloud mask (FR-CLD-015).
const (
	ClearThreshold = 65
	CloudThreshold = 90
)

// The no-data veil of FR-CLD-007: a mid grey at 20% opacity, a target the
// spike confirms by eye in the real window.
const (
	VeilGrey    = math.MaxUint8 / 2
	VeilOpacity = 0.2
)

// Cloud is the colour every cloud pixel is drawn in; only its opacity varies.
const Cloud = math.MaxUint8

// ImageInterval is how often the service lists a new image (its time
// dimension, PT3H); FR-CLD-010's staleness counts in these.
const ImageInterval = 3 * time.Hour

// timeLayout words a valid time as "15:00".
const timeLayout = "15:04"

// Pixel is one drawn pixel, straight (not premultiplied) alpha.
type Pixel struct {
	R, G, B, A uint8
}

// Opacity is FR-CLD-006: none at or below the clear threshold, full at or
// above the cloud threshold, linear between.
func Opacity(brightness float64) float64 {
	switch {
	case brightness <= ClearThreshold:
		return 0
	case brightness >= CloudThreshold:
		return 1
	default:
		return (brightness - ClearThreshold) / (CloudThreshold - ClearThreshold)
	}
}

// Shade draws one source pixel: the veil where the source holds no data
// (FR-CLD-007), otherwise cloud white at the opacity its brightness earns.
func Shade(brightness uint8, hasData bool) Pixel {
	if !hasData {
		return Pixel{R: VeilGrey, G: VeilGrey, B: VeilGrey, A: alpha(VeilOpacity)}
	}
	return Pixel{R: Cloud, G: Cloud, B: Cloud, A: alpha(Opacity(float64(brightness)))}
}

// alpha scales an opacity to a channel value, rounded to the nearest.
func alpha(opacity float64) uint8 {
	return uint8(math.Round(opacity * math.MaxUint8))
}

// Stale is FR-CLD-010: the held image is stale once its valid time is older
// than NFR-FRESH-001's count of image intervals.
func Stale(validTime, now time.Time) bool {
	return freshness.Stale(validTime, now, ImageInterval)
}

// Status is FR-CLD-009: "Clouds: image of 15:00 UTC, 3 h ago", with
// " (stale)" appended once FR-CLD-010 says so.
func Status(validTime, now time.Time) string {
	line := fmt.Sprintf("Clouds: image of %s UTC, %s",
		validTime.UTC().Format(timeLayout), freshness.Age(validTime, now))
	if Stale(validTime, now) {
		line += " (stale)"
	}
	return line
}

// Unavailable is FR-CLD-013's status line when no image is held and the first
// fetch failed; the reason itself goes to the provider status popover.
const Unavailable = "Clouds: the cloud image could not be retrieved"

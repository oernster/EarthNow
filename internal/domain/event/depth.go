package event

import (
	"fmt"
	"math"
)

// The depth bands USGS uses (REQUIREMENTS.md R10, FR-SEL-010), in kilometres:
// shallow below the first edge, intermediate below the second, deep beyond.
const (
	IntermediateFromKm = 70
	DeepFromKm         = 300
)

// FixedDepthKm is the depth USGS assigns when the data are too poor to compute
// one (R11, FR-SEL-012). The summary feed does not say which quakes carry it.
const FixedDepthKm = 10

// The band words and the note that marks the fixed depth.
const (
	shallow        = "shallow"
	intermediate   = "intermediate"
	deep           = "deep"
	fixedDepthNote = " (often a fixed depth: USGS assigns %d km when it cannot compute one)"
)

// depthDecimals is how finely a depth is shown: one decimal, as the source
// gives it, so a value just under a band edge does not read as the edge.
const depthDecimals = 10

// DepthWording is the detail panel's Depth row (FR-SEL-010 to 012): the depth
// to one decimal with its band, measured from the value shown so the two never
// disagree; a negative depth reads as that far above sea level.
func DepthWording(km float64) string {
	shown := math.Round(km*depthDecimals) / depthDecimals
	if shown == 0 {
		shown = 0 // a depth rounding to -0.0 reads as 0.0
	}
	band := shallow
	switch {
	case shown >= DeepFromKm:
		band = deep
	case shown >= IntermediateFromKm:
		band = intermediate
	}
	var line string
	if shown < 0 {
		line = fmt.Sprintf("%.1f km above sea level, %s", -shown, band)
	} else {
		line = fmt.Sprintf("%.1f km, %s", shown, band)
	}
	if km == FixedDepthKm {
		line += fmt.Sprintf(fixedDepthNote, FixedDepthKm)
	}
	return line
}

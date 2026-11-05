package eonet

import (
	"encoding/json"
	"math"
)

// maxLatitude is the largest latitude there is. A vertex value beyond it can
// only be a longitude, which is how a ring proves its own order.
const maxLatitude = 90.0

// polygonOrder says how an event's polygon vertices are read.
type polygonOrder struct {
	latFirst bool // the reading for a ring that proves nothing
	proven   bool // a ring that proves its order is read by it instead
}

// ringOrder answers what a ring's own values prove about its order: latFirst
// when a second value lies beyond maxLatitude, longitude first when a first
// value does. ok is false when nothing does; it is false too when the ring
// says both.
func ringOrder(ring [][]float64) (latFirst, ok bool) {
	firstOver, secondOver := false, false
	for _, v := range ring {
		if len(v) < 2 {
			continue
		}
		firstOver = firstOver || math.Abs(v[0]) > maxLatitude
		secondOver = secondOver || math.Abs(v[1]) > maxLatitude
	}
	if firstOver == secondOver {
		return false, false
	}
	return secondOver, true
}

// feedOrder is the reading for this feed's latFirstPolygonSource polygons. Every
// ring that proves its order votes. Latitude first is read only while the
// proven rings are more often latitude first; a tie, including a feed where no
// ring proves anything, is GeoJSON. Measured over the 40 days to 2026-11-05:
// the last latitude-first rings were dated 2026-10-04 and all 12 proving rings
// since were longitude first, so GDACS corrected upstream. The tie went the
// other way until then and swapped every flood once the 7-day feed held no
// proving ring.
func feedOrder(events []wireEvent) polygonOrder {
	latVotes, lngVotes := 0, 0
	for _, w := range events {
		if !sourcedBy(w, latFirstPolygonSource) {
			continue
		}
		for _, g := range w.Geometry {
			var rings [][][]float64
			if g.Type != "Polygon" || json.Unmarshal(g.Coords, &rings) != nil || len(rings) == 0 {
				continue
			}
			if latFirst, ok := ringOrder(rings[0]); ok && latFirst {
				latVotes++
			} else if ok {
				lngVotes++
			}
		}
	}
	return polygonOrder{latFirst: latVotes > lngVotes, proven: true}
}

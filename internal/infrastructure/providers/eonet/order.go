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
// ring that proves its order votes. The measured latitude-first order stands
// unless the proven rings are more often longitude first, so a correction
// upstream is followed as soon as the data shows it rather than drawing every
// flood swapped. Measured over the 30 days to 2026-09-24: 15 of 57 GDACS rings
// proved their order, every one latitude first, with at least one in each week.
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
	return polygonOrder{latFirst: lngVotes <= latVotes, proven: true}
}

package eonet

import (
	"encoding/json"
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
)

// mapGeometry turns one EONET geometry into an observation; false when its
// date, type or coordinates are unusable. A polygon is placed at the mean of
// its outer ring's vertices (DATA-003); its shape is out of V1's scope.
func mapGeometry(g wireGeometry) (event.Observation, bool) {
	at, err := time.Parse(time.RFC3339, g.Date)
	if err != nil {
		return event.Observation{}, false
	}
	var where event.Point
	switch g.Type {
	case "Point":
		var c []float64
		if json.Unmarshal(g.Coords, &c) != nil || len(c) < 2 {
			return event.Observation{}, false
		}
		if where, err = event.NewPoint(c[1], c[0]); err != nil {
			return event.Observation{}, false
		}
	case "Polygon":
		var rings [][][]float64
		if json.Unmarshal(g.Coords, &rings) != nil || len(rings) == 0 {
			return event.Observation{}, false
		}
		p, ok := centroid(rings[0])
		if !ok {
			return event.Observation{}, false
		}
		where = p
	default:
		return event.Observation{}, false
	}
	o := event.Observation{At: at.UTC(), Where: where}
	if g.Magnitude != nil && g.Unit != nil {
		o.Measurement = &event.Measurement{Value: *g.Magnitude, Unit: *g.Unit}
	}
	return o, true
}

// centroid is the vertex mean of a ring, which is where the marker goes. The
// closing vertex repeats the first, so it is left out of the mean.
func centroid(ring [][]float64) (event.Point, bool) {
	vertices := ring
	if n := len(ring); n > 1 && len(ring[0]) >= 2 && len(ring[n-1]) >= 2 &&
		ring[0][0] == ring[n-1][0] && ring[0][1] == ring[n-1][1] {
		vertices = ring[:n-1]
	}
	if len(vertices) == 0 {
		return event.Point{}, false
	}
	var lat, lng float64
	for _, v := range vertices {
		if len(v) < 2 {
			return event.Point{}, false
		}
		lng += v[0]
		lat += v[1]
	}
	count := float64(len(vertices))
	p, err := event.NewPoint(lat/count, lng/count)
	return p, err == nil
}

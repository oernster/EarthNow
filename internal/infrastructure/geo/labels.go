package geo

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/oernster/EarthNow/internal/domain/event"
)

// labelsJSON is each country's ISO 3166-1 alpha-2 code to Natural Earth's label
// point for it, [latitude, longitude], written by `tools/geodata.py --labels`
// from the admin 0 countries layer (REQUIREMENTS.md FR-GLB-017).
//
//go:embed data/labels.json
var labelsJSON []byte

// ErrNoLabels refuses a label table holding no country.
var ErrNoLabels = errors.New("the label table holds no country")

// Labels answers the point the globe opens facing for a country (FR-GLB-015).
type Labels struct {
	points map[string][2]float64
}

// LoadLabels decodes the embedded label table.
func LoadLabels() (*Labels, error) {
	return loadLabels(labelsJSON)
}

func loadLabels(data []byte) (*Labels, error) {
	var points map[string][2]float64
	if err := json.Unmarshal(data, &points); err != nil {
		return nil, fmt.Errorf("labels: %w", err)
	}
	if len(points) == 0 {
		return nil, ErrNoLabels
	}
	return &Labels{points: points}, nil
}

// Label answers the country's label point; false when the table holds no such
// code or its point is not on the Earth.
func (l *Labels) Label(code string) (event.Point, bool) {
	p, ok := l.points[code]
	if !ok {
		return event.Point{}, false
	}
	at, err := event.NewPoint(p[0], p[1])
	return at, err == nil
}

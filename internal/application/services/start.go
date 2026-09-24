package services

import (
	"fmt"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/application/ports"
)

// StartView decides where the globe opens (FR-GLB-015, FR-GLB-016): facing the
// label point of the country the operating system's region names, else as it
// always has, with the reason worded for the log.
type StartView struct {
	region ports.RegionSource
	labels ports.LabelPoints
}

// NewStartView builds the use case over the region setting and the label table.
func NewStartView(region ports.RegionSource, labels ports.LabelPoints) *StartView {
	return &StartView{region: region, labels: labels}
}

// Answer answers the start view and the line the log records about it.
func (s *StartView) Answer() (dto.StartView, string) {
	code, err := s.region.Region()
	if err != nil {
		return dto.StartView{}, fmt.Sprintf("Start view: %v; opening as before", err)
	}
	at, ok := s.labels.Label(code)
	if !ok {
		return dto.StartView{}, fmt.Sprintf("Start view: no label point for region %s; opening as before", code)
	}
	return dto.StartView{Found: true, Lat: at.Lat, Lng: at.Lng},
		fmt.Sprintf("Start view: facing %s at %.4f, %.4f", code, at.Lat, at.Lng)
}

package services

import (
	"testing"

	"github.com/oernster/EarthNow/internal/domain/event"
)

func quakeAt(depth *float64) Shown {
	e := event.Event{Provider: event.USGS, ProviderEventID: "us1", Category: event.Earthquake}
	e.Extras.DepthKm = depth
	return Shown{Event: e, Observation: event.Observation{At: noon}, RetrievedAt: noon}
}

func TestFRSEL010_AQuakesDepthReachesTheWireWorded(t *testing.T) {
	t.Parallel()
	depth := 18.44
	if got := toDTO(quakeAt(&depth), noon).Depth; got != event.DepthWording(depth) {
		t.Errorf("Depth = %q, want the domain's wording", got)
	}
}

func TestFRSEL013_NoDepthLeavesTheRowEmpty(t *testing.T) {
	t.Parallel()
	if got := toDTO(quakeAt(nil), noon).Depth; got != "" {
		t.Errorf("Depth = %q, want none", got)
	}
}

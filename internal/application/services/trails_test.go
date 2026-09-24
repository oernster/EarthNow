package services

import (
	"reflect"
	"testing"

	"github.com/oernster/EarthNow/internal/domain/event"
)

func TestFRTRL001_AStormsTrailReachesTheWireAsPairs(t *testing.T) {
	t.Parallel()
	storm := Shown{Event: event.Event{Category: event.SevereStorm}, Observation: event.Observation{At: noon},
		RetrievedAt: noon, Trail: []event.Point{{Lat: 10, Lng: -60}, {Lat: 12, Lng: -62}}}
	if got, want := toDTO(storm, noon).Trail, [][2]float64{{10, -60}, {12, -62}}; !reflect.DeepEqual(got, want) {
		t.Errorf("Trail = %v, want %v", got, want)
	}
	if got := toDTO(Shown{Observation: event.Observation{At: noon}}, noon).Trail; got == nil || len(got) != 0 {
		t.Errorf("an event with no trail sent %v, want an empty list", got)
	}
}

func TestFRTRL005_TrailsStartOnAndTheChoiceIsKept(t *testing.T) {
	t.Parallel()
	if !Defaults().TrailsShown {
		t.Error("a first run hides storm trails")
	}
	store := &fakeSettings{}
	p := NewPreferences(store, (&minimums{}).apply)
	p.Load()
	chosen := p.Current()
	chosen.TrailsShown = false
	if held, _ := p.Update(chosen); held.TrailsShown || store.saved[0].TrailsShown {
		t.Errorf("switching trails off was not kept (FR-TRL-004): %+v", held)
	}
}

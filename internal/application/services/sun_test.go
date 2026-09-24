package services

import (
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/domain/sun"
)

// sunStep is FR-DAY-004's longest wait between two readings of the sun.
const sunStep = time.Minute

func TestFRDAY001_TheSunIsWhereTheDomainPutsItNow(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{now: noon}
	got := NewSun(clock).Position()
	at := sun.Subsolar(noon)
	if got.Lat != at.Lat || got.Lng != at.Lng {
		t.Errorf("Position() at %v = %+v, want %+v", noon, got, at)
	}
	if got.TwilightDegrees != sun.TwilightDegrees || got.CloudNightFloor != sun.CloudNightFloor {
		t.Errorf("Position() carries %+v, not the domain's light figures", got)
	}
}

func TestFRDAY004_TheSunIsReadFromTheClockEachTime(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{now: noon}
	s := NewSun(clock)
	before := s.Position()
	clock.now = noon.Add(sunStep)
	if after := s.Position(); after.Lng == before.Lng {
		t.Errorf("the sun stood still over %v: %+v", sunStep, after)
	}
}

func TestFRDAY007_TheLayerStartsShownAndItsChoiceIsKept(t *testing.T) {
	t.Parallel()
	if !Defaults().DayNightShown {
		t.Error("a first run hides the day and night layer")
	}
	store := &fakeSettings{}
	p := NewPreferences(store, (&minimums{}).apply)
	p.Load()
	if !p.Current().DayNightShown {
		t.Error("a run with no settings file hides the day and night layer")
	}
	chosen := p.Current()
	chosen.DayNightShown = false
	if held, _ := p.Update(chosen); held.DayNightShown || store.saved[0].DayNightShown {
		t.Errorf("hiding the layer was not kept: %+v, saved %+v", held, store.saved)
	}
}

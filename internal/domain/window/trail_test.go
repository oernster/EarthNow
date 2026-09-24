package window

import (
	"reflect"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
)

// fixes builds an event of the category with one observation per offset from
// noon, oldest first as events hold them, each at a distinct position.
func fixes(category event.Category, offsets ...time.Duration) event.Event {
	e := event.Event{Category: category}
	for i, d := range offsets {
		e.Observations = append(e.Observations, event.Observation{At: noon.Add(d), Where: event.Point{Lat: float64(i), Lng: float64(i)}})
	}
	return e
}

func TestFRTRL001_AStormsTrailIsItsFixesInsideTheWindow(t *testing.T) {
	t.Parallel()
	storm := fixes(event.SevereStorm, -30*time.Hour, -20*time.Hour, -2*time.Hour)
	day, _ := ByKey("24h")
	days, _ := ByKey("3d")
	if got, want := day.Trail(storm, noon), []event.Point{{Lat: 1, Lng: 1}, {Lat: 2, Lng: 2}}; !reflect.DeepEqual(got, want) {
		t.Errorf("24 h trail = %v, want %v", got, want)
	}
	if got := days.Trail(storm, noon); len(got) != 3 || got[2] != (event.Point{Lat: 2, Lng: 2}) {
		t.Errorf("3 day trail = %v, want all three ending at the newest", got)
	}
	marker, _ := days.Latest(storm, noon)
	if got := days.Trail(storm, noon); got[len(got)-1] != marker.Where {
		t.Errorf("the trail ends at %v, not the marker %v", got[len(got)-1], marker.Where)
	}
}

func TestFRTRL001_NoTrailForOneFixOrAnythingButAStorm(t *testing.T) {
	t.Parallel()
	w, _ := ByKey("7d")
	if got := w.Trail(fixes(event.SevereStorm, -time.Hour), noon); got != nil {
		t.Errorf("one fix gave %v", got)
	}
	if got := w.Trail(fixes(event.Ice, -48*time.Hour, -time.Hour), noon); got != nil {
		t.Errorf("an iceberg gave %v", got)
	}
}

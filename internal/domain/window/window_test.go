package window

import (
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
)

var noon = time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)

func TestFRTW001_OffersFiveWindowsInOrder(t *testing.T) {
	t.Parallel()
	keys := ""
	for _, w := range All() {
		keys += w.Key + " "
	}
	if keys != "1h 6h 24h 3d 7d " {
		t.Errorf("windows = %q", keys)
	}
	if Widest.Length != 7*24*time.Hour {
		t.Errorf("widest = %v", Widest.Length)
	}
}

func TestFRTW003_DefaultIsTwentyFourHours(t *testing.T) {
	t.Parallel()
	if Default.Key != "24h" {
		t.Errorf("default = %s", Default.Key)
	}
}

func TestByKey(t *testing.T) {
	t.Parallel()
	if w, ok := ByKey("3d"); !ok || w != ThreeDays {
		t.Errorf("ByKey(3d) = %v, %v", w, ok)
	}
	if _, ok := ByKey("2d"); ok {
		t.Error("ByKey(2d) found a window that does not exist")
	}
}

func TestFRTW002_ContainsOnlyTheSpanEndingNow(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		at   time.Time
		want bool
	}{
		{"half an hour ago", noon.Add(-30 * time.Minute), true},
		{"an hour and a half ago", noon.Add(-90 * time.Minute), false},
		{"exactly one hour ago", noon.Add(-time.Hour), false},
		{"a minute ahead of this clock", noon.Add(time.Minute), true},
		{"at the clock skew bound", noon.Add(ClockSkew), true},
		{"just past the clock skew bound", noon.Add(ClockSkew + time.Second), false},
		{"eleven days ahead, as a GDACS flood alert", noon.Add(11 * hoursPerDay * time.Hour), false},
	}
	for _, c := range cases {
		if got := OneHour.Contains(c.at, noon); got != c.want {
			t.Errorf("%s: Contains = %v, want %v", c.name, got, c.want)
		}
	}
}

func track(times ...time.Time) event.Event {
	e := event.Event{Provider: event.EONET, ProviderEventID: "EONET_1"}
	for i, at := range times {
		e.Observations = append(e.Observations, event.Observation{At: at, Where: event.Point{Lat: float64(i), Lng: float64(i)}})
	}
	return e
}

func TestDATA003_LatestObservationInsideTheWindowIsTheMarker(t *testing.T) {
	t.Parallel()
	storm := track(noon.Add(-30*time.Hour), noon.Add(-12*time.Hour), noon.Add(-6*time.Hour))
	o, ok := OneDay.Latest(storm, noon)
	if !ok || o.Where.Lat != 2 {
		t.Errorf("24 h: got %v, %v; want the newest point", o, ok)
	}
	o, ok = ThreeDays.Latest(storm, noon)
	if !ok || o.Where.Lat != 2 {
		t.Errorf("3 d: got %v, %v; want the newest point", o, ok)
	}
}

func TestFRTW002_EventWithNothingInsideIsNotShown(t *testing.T) {
	t.Parallel()
	storm := track(noon.Add(-30*time.Hour), noon.Add(-12*time.Hour))
	if _, ok := SixHours.Latest(storm, noon); ok {
		t.Error("an event with no observation in the last 6 h was shown")
	}
	if _, ok := OneDay.Latest(event.Event{}, noon); ok {
		t.Error("an event with no observations was shown")
	}
}

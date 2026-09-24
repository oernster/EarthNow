package sun

import (
	"math"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
)

// tolerance is FR-DAY-001's agreement with NOAA's calculator, in degrees.
const tolerance = 0.1

// noaa holds NOAA's solar calculator (gml.noaa.gov/grad/solcalc) read at
// latitude 0, longitude 0, GMT, 12:00:00 on each date, on 2026-09-24: the
// declination as shown and the equation of time in minutes, from which the
// subsolar longitude at 12:00 UTC is minus a quarter of a degree a minute.
var noaa = []struct {
	day         time.Time
	declination float64
	eot         float64
}{
	{time.Date(2026, time.March, 20, 12, 0, 0, 0, time.UTC), -0.04, -7.43},
	{time.Date(2026, time.June, 21, 12, 0, 0, 0, time.UTC), 23.44, -1.82},
	{time.Date(2026, time.September, 23, 12, 0, 0, 0, time.UTC), -0.19, 7.61},
	{time.Date(2026, time.December, 21, 12, 0, 0, 0, time.UTC), -23.44, 1.92},
}

func TestFRDAY001_SubsolarPointAgreesWithNOAA(t *testing.T) {
	t.Parallel()
	for _, c := range noaa {
		got := Subsolar(c.day)
		wantLng := -c.eot * degreesPerMinute
		if math.Abs(got.Lat-c.declination) > tolerance || math.Abs(got.Lng-wantLng) > tolerance {
			t.Errorf("Subsolar(%v) = %+v, want %.2f, %.4f", c.day, got, c.declination, wantLng)
		}
	}
}

func TestFRDAY001_SubsolarLongitudeWrapsAtTheDateLine(t *testing.T) {
	t.Parallel()
	// Past midnight the sun is over the Pacific; either side of the line the
	// longitude stays inside -180 to 180 and a minute moves it a quarter degree.
	for _, c := range noaa {
		for _, clock := range []time.Duration{0, 24*time.Hour - time.Minute} {
			at := c.day.Add(-12*time.Hour + clock)
			lng := Subsolar(at).Lng
			if lng < -halfTurn || lng >= halfTurn {
				t.Errorf("Subsolar(%v).Lng = %v, outside -180 to 180", at, lng)
			}
			step := math.Abs(Subsolar(at.Add(time.Minute)).Lng - lng)
			if step > halfTurn {
				step = fullTurn - step
			}
			if math.Abs(step-degreesPerMinute) > tolerance {
				t.Errorf("a minute from %v moved the sun %v degrees", at, step)
			}
		}
	}
}

func TestFRDAY001_ElevationAgreesWithNOAA(t *testing.T) {
	t.Parallel()
	// NOAA's elevation at 0, 0 on the same four instants, refraction included;
	// refraction adds under 0.01 degrees this high.
	want := []float64{88.14, 66.56, 88.09, 66.57}
	origin := event.Point{}
	for i, c := range noaa {
		if got := Elevation(origin, Subsolar(c.day)); math.Abs(got-want[i]) > tolerance {
			t.Errorf("Elevation at 0,0 on %v = %v, want %v", c.day, got, want[i])
		}
	}
}

func TestFRDAY002_ElevationSpansZenithToNadir(t *testing.T) {
	t.Parallel()
	overhead := event.Point{Lat: 10, Lng: 20}
	opposite := event.Point{Lat: -10, Lng: 20 - halfTurn}
	if got := Elevation(overhead, overhead); math.Abs(got-90) > tolerance {
		t.Errorf("Elevation under the sun = %v, want 90", got)
	}
	if got := Elevation(opposite, overhead); math.Abs(got+90) > tolerance {
		t.Errorf("Elevation opposite the sun = %v, want -90", got)
	}
}

func TestFRDAY002_LightRampsAcrossTwilight(t *testing.T) {
	t.Parallel()
	cases := []struct{ elevation, want float64 }{
		{-90, 0},
		{-TwilightDegrees, 0},
		{0, 0.5},
		{TwilightDegrees, 1},
		{90, 1},
	}
	for _, c := range cases {
		if got := Light(c.elevation); got != c.want {
			t.Errorf("Light(%v) = %v, want %v", c.elevation, got, c.want)
		}
	}
}

func TestFRDAY009_CloudsDimToTheNightFloor(t *testing.T) {
	t.Parallel()
	cases := []struct{ light, want float64 }{
		{0, 0.25},
		{0.5, 0.625},
		{1, 1},
	}
	for _, c := range cases {
		if got := CloudOpacity(c.light); got != c.want {
			t.Errorf("CloudOpacity(%v) = %v, want %v", c.light, got, c.want)
		}
	}
}

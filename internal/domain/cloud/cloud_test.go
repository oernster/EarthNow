package cloud

import (
	"math"
	"testing"
	"time"
)

func TestFRCLD006_BrightnessMapsToOpacity(t *testing.T) {
	t.Parallel()
	cases := []struct {
		brightness, want float64
	}{
		{0, 0},
		{ClearThreshold, 0},
		{(ClearThreshold + CloudThreshold) / 2.0, 0.5},
		{CloudThreshold, 1},
		{math.MaxUint8, 1},
	}
	for _, c := range cases {
		if got := Opacity(c.brightness); got != c.want {
			t.Errorf("Opacity(%v) = %v, want %v", c.brightness, got, c.want)
		}
	}
}

func TestFRCLD006_ACloudPixelIsWhiteAtItsOpacity(t *testing.T) {
	t.Parallel()
	cases := []struct {
		brightness uint8
		want       Pixel
	}{
		{ClearThreshold, Pixel{Cloud, Cloud, Cloud, 0}},
		{CloudThreshold, Pixel{Cloud, Cloud, Cloud, math.MaxUint8}},
		// 77 is 12/25 of the way up the ramp: 0.48 x 255 = 122.4, rounded to 122.
		{77, Pixel{Cloud, Cloud, Cloud, 122}},
	}
	for _, c := range cases {
		if got := Shade(c.brightness, true); got != c.want {
			t.Errorf("Shade(%d, true) = %+v, want %+v", c.brightness, got, c.want)
		}
	}
}

func TestFRCLD007_APixelWithNoDataIsTheVeil(t *testing.T) {
	t.Parallel()
	// 0.2 x 255 = 51.
	want := Pixel{VeilGrey, VeilGrey, VeilGrey, 51}
	for _, brightness := range []uint8{0, CloudThreshold, math.MaxUint8} {
		if got := Shade(brightness, false); got != want {
			t.Errorf("Shade(%d, false) = %+v, want %+v", brightness, got, want)
		}
	}
}

var validTime = time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC)

func TestFRCLD009_TheStatusGivesTheValidTimeAndAge(t *testing.T) {
	t.Parallel()
	now := validTime.Add(3 * time.Hour)
	if got, want := Status(validTime, now), "Clouds: image of 15:00 UTC, 3 h ago"; got != want {
		t.Errorf("Status = %q, want %q", got, want)
	}
}

func TestFRCLD010_AnImageOlderThanThreeIntervalsIsStale(t *testing.T) {
	t.Parallel()
	edge := validTime.Add(9 * time.Hour)
	if Stale(validTime, edge) {
		t.Error("an image exactly 9 h old is stale; FR-CLD-010 says older than")
	}
	late := edge.Add(time.Minute)
	if !Stale(validTime, late) {
		t.Error("an image 9 h 1 min old is not stale")
	}
	if got, want := Status(validTime, late), "Clouds: image of 15:00 UTC, 9 h ago (stale)"; got != want {
		t.Errorf("Status = %q, want %q", got, want)
	}
}

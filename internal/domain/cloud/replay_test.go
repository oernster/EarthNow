package cloud

import (
	"testing"
	"time"
)

func at(dayOfMonth, hour, minute int) time.Time {
	return time.Date(2026, 9, dayOfMonth, hour, minute, 0, 0, time.UTC)
}

func TestFRRPL015_TheSpanHoldsTheServicesThreeHourlyTimes(t *testing.T) {
	t.Parallel()
	got := ValidTimes(at(23, 12, 0), at(24, 12, 0))
	if len(got) != 9 || !got[0].Equal(at(23, 12, 0)) || !got[8].Equal(at(24, 12, 0)) {
		t.Fatalf("24 h from noon: %v", got)
	}
	got = ValidTimes(at(23, 13, 5), at(23, 20, 59))
	if len(got) != 2 || !got[0].Equal(at(23, 15, 0)) || !got[1].Equal(at(23, 18, 0)) {
		t.Fatalf("off the grid: %v", got)
	}
	if got = ValidTimes(at(23, 13, 0), at(23, 14, 0)); len(got) != 0 {
		t.Fatalf("an hour between two times holds none: %v", got)
	}
}

func TestFRRPL016_FRRPL017_TheReplayCloudWording(t *testing.T) {
	t.Parallel()
	if got := ReplayProgress(12, 56); got != "Clouds 12 of 56" {
		t.Errorf("progress %q", got)
	}
	if got := ReplayMissing([]time.Time{at(23, 15, 0), at(23, 18, 0)}); got != "Missing images: 23 Sep 15:00, 23 Sep 18:00 UTC" {
		t.Errorf("missing %q", got)
	}
}

func TestFRRPL014_TheImageDrawnIsTheLatestAtOrBeforeTheInstant(t *testing.T) {
	t.Parallel()
	times := []time.Time{at(23, 12, 0), at(23, 15, 0)}
	cases := []struct {
		at   time.Time
		want time.Time
		ok   bool
	}{
		{at(23, 14, 59), at(23, 12, 0), true},
		{at(23, 15, 0), at(23, 15, 0), true},
		{at(23, 11, 59), time.Time{}, false},
	}
	for _, c := range cases {
		if got, ok := AtOrBefore(times, c.at); ok != c.ok || !got.Equal(c.want) {
			t.Errorf("AtOrBefore(%v) = %v, %v", c.at, got, ok)
		}
	}
}

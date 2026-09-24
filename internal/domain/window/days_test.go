package window

import (
	"testing"
	"time"
)

func utc(dayOfMonth, hour int) time.Time {
	return time.Date(2026, time.September, dayOfMonth, hour, 0, 0, 0, time.UTC)
}

func TestFRBA001_DaysAreEveryUTCDayTheWindowOverlaps(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		w     Window
		now   time.Time
		first int
		last  int
	}{
		{"7 days at noon covers eight days", SevenDays, utc(24, 12), 17, 24},
		{"3 days at noon", ThreeDays, utc(24, 12), 21, 24},
		{"24 h at noon", OneDay, utc(24, 12), 23, 24},
		{"1 h at noon", OneHour, utc(24, 12), 24, 24},
		{"6 h at 03:00 reaches back a day", SixHours, utc(24, 3), 23, 24},
		{"24 h at midnight is the day before alone", OneDay, utc(24, 0), 23, 23},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := c.w.Days(c.now)
			want := c.last - c.first + 1
			if len(got) != want {
				t.Fatalf("got %d days %v, want %d", len(got), got, want)
			}
			for i, d := range got {
				if !d.Equal(utc(c.first+i, 0)) {
					t.Errorf("day %d is %v, want %v", i, d, utc(c.first+i, 0))
				}
			}
		})
	}
}

func TestFRBA001_DaysAreUTCWhateverTheClockZone(t *testing.T) {
	t.Parallel()
	bst := time.FixedZone("BST", int(time.Hour/time.Second))
	// 00:16 BST on 25 Sep is 23:16 UTC on 24 Sep: the UTC day is the 24th.
	got := OneHour.Days(time.Date(2026, time.September, 25, 0, 16, 0, 0, bst))
	if len(got) != 1 || !got[0].Equal(utc(24, 0)) || got[0].Location() != time.UTC {
		t.Fatalf("got %v, want 24 Sep 00:00 UTC alone", got)
	}
}

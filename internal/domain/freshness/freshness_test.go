package freshness

import (
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
)

var now = time.Date(2026, 9, 23, 11, 30, 0, 0, time.UTC)

func TestNFRFRESH002_AgeWording(t *testing.T) {
	t.Parallel()
	cases := []struct {
		ago  time.Duration
		want string
	}{
		{0, "under a minute ago"},
		{59 * time.Second, "under a minute ago"},
		{time.Minute, "1 min ago"},
		{14*time.Minute + 50*time.Second, "14 min ago"},
		{59 * time.Minute, "59 min ago"},
		{time.Hour, "1 h ago"},
		{47*time.Hour + 59*time.Minute, "47 h ago"},
		{48 * time.Hour, "2 days ago"},
		{6*24*time.Hour + 23*time.Hour, "6 days ago"},
		{-5 * time.Minute, "under a minute ago"},
	}
	for _, c := range cases {
		if got := Age(now.Add(-c.ago), now); got != c.want {
			t.Errorf("Age(%v ago) = %q, want %q", c.ago, got, c.want)
		}
	}
}

func TestFRSEL004_DayPrecisionReadsAsADate(t *testing.T) {
	t.Parallel()
	o := event.Observation{At: time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC), Precision: event.Day}
	if got := Reported(o, now); got != "Reported for 18 Sep 2026, 5 days ago" {
		t.Errorf("Reported = %q", got)
	}
}

func TestFRSEL004_DayAgeNearToday(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC)
	cases := map[time.Time]string{
		today:                   "today",
		today.AddDate(0, 0, 1):  "today",
		today.AddDate(0, 0, -1): "yesterday",
		today.AddDate(0, 0, -2): "2 days ago",
	}
	for date, want := range cases {
		if got := DayAge(date, now); got != want {
			t.Errorf("DayAge(%s) = %q, want %q", date.Format("2006-01-02"), got, want)
		}
	}
}

func TestReportedInstant(t *testing.T) {
	t.Parallel()
	o := event.Observation{At: now.Add(-14 * time.Minute), Precision: event.Instant}
	if got := Reported(o, now); got != "Observed 14 min ago" {
		t.Errorf("Reported = %q", got)
	}
}

func TestNFRFRESH001_StaleAfterThreeIntervals(t *testing.T) {
	t.Parallel()
	interval := time.Minute
	cases := []struct {
		name string
		last time.Time
		want bool
	}{
		{"never retrieved", time.Time{}, false},
		{"one interval old", now.Add(-interval), false},
		{"exactly three intervals", now.Add(-3 * interval), false},
		{"past three intervals", now.Add(-3*interval - time.Second), true},
	}
	for _, c := range cases {
		if got := Stale(c.last, now, interval); got != c.want {
			t.Errorf("%s: Stale = %v, want %v", c.name, got, c.want)
		}
	}
}

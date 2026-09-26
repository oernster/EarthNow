package event

import (
	"testing"
	"time"
)

func day(month time.Month, dayOfMonth int) time.Time {
	return time.Date(2026, month, dayOfMonth, 0, 0, 0, 0, time.UTC)
}

func TestAnEventIsOngoingOnlyWhenAReportCarriesIt(t *testing.T) {
	t.Parallel()
	if (Event{}).Ongoing() {
		t.Error("an event with no report reads as ongoing")
	}
	if !(Event{Report: Report{WeekFrom: day(9, 10), WeekTo: day(9, 16), Issued: day(9, 17)}}).Ongoing() {
		t.Error("a reported volcano does not read as ongoing")
	}
}

func TestFRPRV016_AReportIsCurrentForFourteenDaysAfterItsIssue(t *testing.T) {
	t.Parallel()
	r := Report{Issued: day(9, 17)}
	cases := []struct {
		now  time.Time
		want bool
	}{
		{day(9, 17), true},
		{day(10, 1), true},
		{day(10, 1).Add(time.Nanosecond), false},
		{day(10, 2), false},
	}
	for _, c := range cases {
		if got := r.Current(c.now); got != c.want {
			t.Errorf("current at %v = %v, want %v", c.now, got, c.want)
		}
	}
}

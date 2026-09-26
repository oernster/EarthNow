package freshness

import (
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
)

func date(year int, month time.Month, dayOfMonth int) time.Time {
	return time.Date(year, month, dayOfMonth, 0, 0, 0, 0, time.UTC)
}

func TestFRSEL004_AnOngoingEventReadsAsItsReportWeek(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		r    event.Report
		want string
	}{
		{"one month", event.Report{WeekFrom: date(2026, 9, 10), WeekTo: date(2026, 9, 16), Issued: date(2026, 9, 17)},
			"Continuing: report for 10 to 16 Sep 2026, issued 17 Sep 2026"},
		{"two months", event.Report{WeekFrom: date(2026, 8, 27), WeekTo: date(2026, 9, 2), Issued: date(2026, 9, 3)},
			"Continuing: report for 27 Aug to 2 Sep 2026, issued 3 Sep 2026"},
		{"two years", event.Report{WeekFrom: date(2026, 12, 31), WeekTo: date(2027, 1, 6), Issued: date(2027, 1, 7)},
			"Continuing: report for 31 Dec 2026 to 6 Jan 2027, issued 7 Jan 2027"},
	}
	for _, c := range cases {
		if got := Continuing(c.r); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}

func TestFRPRV016_AReportTooOldIsNamedByItsIssueDate(t *testing.T) {
	t.Parallel()
	got := TooOld(event.Report{Issued: date(2026, 9, 17)})
	if want := "The latest report, issued 17 Sep 2026, is too old to show"; got != want {
		t.Errorf("%q, want %q", got, want)
	}
}

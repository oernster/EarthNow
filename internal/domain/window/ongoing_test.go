package window

import (
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
)

// volcano is an ongoing volcano of the report week weekFrom to weekFrom + 6
// days, issued the day after, as the GVP adapter builds one.
func volcano(weekFrom time.Time) event.Event {
	weekTo := weekFrom.AddDate(0, 0, 6)
	issued := weekTo.AddDate(0, 0, 1)
	return event.Event{
		Provider: event.GVP, ProviderEventID: "262000", Category: event.Volcano,
		Observations: []event.Observation{{At: issued, Precision: event.Day}},
		Report:       event.Report{WeekFrom: weekFrom, WeekTo: weekTo, Issued: issued},
	}
}

func TestFRTW002_AnOngoingEventIsInEveryWindow(t *testing.T) {
	t.Parallel()
	v := volcano(utc(10, 0))
	now := utc(26, 12)
	for _, w := range All() {
		o, ok := w.Latest(v, now)
		if !ok {
			t.Errorf("%s: an ongoing volcano is not shown", w.Label)
			continue
		}
		if !o.At.Equal(utc(17, 0)) {
			t.Errorf("%s: placed by %v, want its issue day", w.Label, o.At)
		}
	}
}

func TestFRRPL009_AnOngoingEventJoinsTheReplayOnItsWeeksFirstDay(t *testing.T) {
	t.Parallel()
	v := volcano(utc(18, 0))
	span := Range{From: utc(17, 12).Add(-SevenDays.Length)}
	span.To = utc(17, 0).Add(24*time.Hour - time.Minute)
	if _, ok := span.Latest(v); ok {
		t.Error("shown at 17 Sep 23:59, before its week began")
	}
	span.To = utc(18, 0)
	if _, ok := span.Latest(v); !ok {
		t.Error("not shown at 18 Sep 00:00, its week's first day")
	}
}

func TestAnOngoingEventWithNoSightingIsNeverShown(t *testing.T) {
	t.Parallel()
	v := volcano(utc(10, 0))
	v.Observations = nil
	if _, ok := SevenDays.Latest(v, utc(26, 12)); ok {
		t.Error("an ongoing event with nowhere to be drawn was shown")
	}
}

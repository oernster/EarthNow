package event

import "time"

// ReportCurrency is FR-PRV-016's limit: a report issued longer ago than this
// shows none of its events. Two weekly cycles, so one late report never empties
// the globe while a feed that has stopped is not shown for ever.
const ReportCurrency = 14 * 24 * time.Hour

// Report is the weekly report an ongoing event comes from (FR-PRV-015): the
// first and last UTC days of the week it covers and the day it was issued. The
// zero value is no report, which is every event that is not ongoing.
type Report struct {
	WeekFrom time.Time
	WeekTo   time.Time
	Issued   time.Time
}

// Ongoing reports whether the event is in progress rather than a moment: one
// carried by a report, from its week's first day up to now.
func (e Event) Ongoing() bool { return !e.Report.Issued.IsZero() }

// Current reports whether the report is still recent enough at now to show its
// events: issued no more than ReportCurrency before now (FR-PRV-016).
func (r Report) Current(now time.Time) bool {
	return now.Sub(r.Issued) <= ReportCurrency
}

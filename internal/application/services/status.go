package services

import (
	"maps"
	"slices"
	"strconv"
	"time"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/domain/event"
	"github.com/oernster/EarthNow/internal/domain/freshness"
)

// statuses answers each provider's entry for the status area (FR-STS-001 to
// 003): loading, retrieved and stale, a failure's reason, a report too old to
// show and the unusable items its last answer held.
func (g *Globe) statuses(now time.Time) []dto.Provider {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make([]dto.Provider, 0, len(g.providers))
	for _, p := range g.providers {
		a := g.attempts[p.Name()]
		status := dto.Provider{Name: string(p.Name()), Dropped: droppedLine(a.dropped)}
		snap, held := g.store.Snapshot(p.Name())
		if held {
			status.Retrieved = freshness.Retrieved(snap.RetrievedAt, now)
			status.Stale = freshness.Stale(snap.RetrievedAt, now, p.Interval())
			status.Notice = reportNotice(snap.Events, now)
		} else {
			status.Loading = a.running || a.failed == nil
		}
		if a.failed != nil {
			status.Problem = a.failed.Error()
		}
		out = append(out, status)
	}
	return out
}

// noticeLine joins the standing cache notices: no cache at all, then each
// provider's, in a fixed order so the line does not shuffle between views.
func (g *Globe) noticeLine() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	notices := []string{g.noCache}
	for _, p := range slices.Sorted(maps.Keys(g.notices)) {
		notices = append(notices, g.notices[p])
	}
	return JoinNotices(notices...)
}

// droppedLine words FR-PRV-013's count of the items the last answer read held
// but could not use, so the status says what the globe is not showing; empty
// when there were none.
func droppedLine(dropped int) string {
	switch dropped {
	case 0:
		return ""
	case 1:
		return "1 item in the last answer could not be read and is not shown"
	}
	return strconv.Itoa(dropped) + " items in the last answer could not be read and are not shown"
}

// reportNotice is FR-PRV-016's notice when the newest report a provider holds
// is past its currency; empty when it holds none or it is current.
func reportNotice(events []event.Event, now time.Time) string {
	var newest event.Report
	for _, e := range events {
		if e.Ongoing() && e.Report.Issued.After(newest.Issued) {
			newest = e.Report
		}
	}
	if newest.Issued.IsZero() || newest.Current(now) {
		return ""
	}
	return freshness.TooOld(newest)
}

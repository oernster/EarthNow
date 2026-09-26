package services

import (
	"context"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
	"github.com/oernster/EarthNow/internal/domain/window"
)

// sepDay is midnight UTC on a day of September 2026.
func sepDay(dayOfMonth int) time.Time {
	return time.Date(2026, time.September, dayOfMonth, 0, 0, 0, 0, time.UTC)
}

// volcano is an ongoing volcano of the report week 10 to 16 Sep, issued 17 Sep,
// as the GVP adapter builds one from the feed captured on 2026-09-26.
func volcano(id string) event.Event {
	return event.Event{
		Provider: event.GVP, ProviderEventID: id, Category: event.Volcano, Title: id,
		Observations: []event.Observation{{At: sepDay(17), Precision: event.Day}},
		Report:       event.Report{WeekFrom: sepDay(10), WeekTo: sepDay(16), Issued: sepDay(17)},
	}
}

func TestFRTW002_AnOngoingVolcanoCountsInEveryWindow(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{sepDay(26).Add(12 * time.Hour)}
	st := NewStore(clock)
	st.Apply(event.GVP, ports.Fetched{Events: []event.Event{volcano("a"), volcano("b")}})
	for _, w := range window.All() {
		if got := Counts(st.Visible(w, Filter{}))[event.Volcano]; got != 2 {
			t.Errorf("%s: %d volcanoes, want 2", w.Label, got)
		}
	}
}

func TestFRPRV016_AReportPastFourteenDaysShowsNothing(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)}
	st := NewStore(clock)
	st.Apply(event.GVP, ports.Fetched{Events: []event.Event{volcano("a")}})
	if got := len(st.Visible(window.SevenDays, Filter{})); got != 1 {
		t.Errorf("1 Oct 00:00: %d shown, want 1", got)
	}
	clock.now = time.Date(2026, time.October, 2, 0, 0, 0, 0, time.UTC)
	if got := len(st.Visible(window.SevenDays, Filter{})); got != 0 {
		t.Errorf("2 Oct 00:00: %d shown, want 0", got)
	}
	if got := len(st.Within(window.SevenDays.Replay(clock.now, 1), Filter{})); got != 0 {
		t.Errorf("replay on 2 Oct: %d shown, want 0", got)
	}
}

func TestFRPRV016_TheStatusSaysWhenTheReportIsTooOld(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{sepDay(26)}
	g := NewGlobe(NewStore(clock), clock, nil)
	p := &fakeProvider{name: event.GVP, fetched: ports.Fetched{Events: []event.Event{volcano("a")}}}
	g.providers = []ports.Provider{p}
	_, _ = g.Refresh(context.Background(), p)
	if got := g.View("24h", Filter{}).Providers[0].Notice; got != "" {
		t.Errorf("current report: notice %q, want none", got)
	}
	clock.now = time.Date(2026, time.October, 2, 0, 0, 0, 0, time.UTC)
	view := g.View("24h", Filter{})
	if got := view.Providers[0].Notice; got != "The latest report, issued 17 Sep 2026, is too old to show" {
		t.Errorf("stale report: notice %q", got)
	}
	if len(view.Events) != 0 {
		t.Errorf("stale report still shows %d events", len(view.Events))
	}
}

func TestFRSEL004_AnOngoingVolcanoCrossesTheWireAsItsReportWeek(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{sepDay(26)}
	st := NewStore(clock)
	st.Apply(event.GVP, ports.Fetched{Events: []event.Event{volcano("a")}})
	st.Apply(event.USGS, ports.Fetched{Events: []event.Event{ev(event.USGS, "q", event.Earthquake, -sepDay(26).Sub(noon)+time.Hour)}})
	g := NewGlobe(st, clock, nil)
	var volcanoDTO, quakeDTO bool
	for _, e := range g.View("24h", Filter{}).Events {
		switch e.Provider {
		case "GVP":
			volcanoDTO = e.Ongoing && e.Reported == "Continuing: report for 10 to 16 Sep 2026, issued 17 Sep 2026"
		case "USGS":
			quakeDTO = !e.Ongoing
		}
	}
	if !volcanoDTO || !quakeDTO {
		t.Errorf("volcano worded as its week %v; quake not ongoing %v", volcanoDTO, quakeDTO)
	}
}

func TestDATA009_TheCacheKeepsAnOngoingVolcanoWhileItsReportIsCurrent(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{sepDay(26)}
	cache := &fakeCache{}
	g := NewGlobe(NewStore(clock), clock, nil)
	g.UseCache(cache)
	p := &fakeProvider{name: event.GVP, fetched: ports.Fetched{Events: []event.Event{volcano("a")}}}
	_, _ = g.Refresh(context.Background(), p)
	if got := len(cache.held[event.GVP].Events); got != 1 {
		t.Errorf("26 Sep: cached %d, want the volcano issued 9 days before", got)
	}
	clock.now = time.Date(2026, time.October, 2, 0, 0, 0, 0, time.UTC)
	_, _ = g.Refresh(context.Background(), p)
	if got := len(cache.held[event.GVP].Events); got != 0 {
		t.Errorf("2 Oct: cached %d, want none past the report's currency", got)
	}
}

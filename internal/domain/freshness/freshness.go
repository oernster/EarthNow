// Package freshness words how old a report is and decides when a provider's
// data has gone stale (REQUIREMENTS.md NFR-FRESH-001, NFR-FRESH-002, FR-SEL-004).
// Nothing here says "live" (FR-STS-006).
package freshness

import (
	"fmt"
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
)

// The boundaries of NFR-FRESH-002: minutes below an hour, hours below two days.
const (
	hoursBoundary = time.Hour
	daysBoundary  = 48 * time.Hour
	day           = 24 * time.Hour
)

// StaleAfterIntervals is NFR-FRESH-001: a provider is stale once its last
// success is older than this many refresh intervals.
const StaleAfterIntervals = 3

// dateLayout is how a day-precision date is shown: "18 Sep 2026".
const dateLayout = "2 Jan 2006"

// Age words the time between then and now, each unit rounded down. A time
// ahead of now reads as the newest wording rather than a negative age.
func Age(then, now time.Time) string {
	elapsed := now.Sub(then)
	switch {
	case elapsed < time.Minute:
		return "under a minute ago"
	case elapsed < hoursBoundary:
		return fmt.Sprintf("%d min ago", int(elapsed/time.Minute))
	case elapsed < daysBoundary:
		return fmt.Sprintf("%d h ago", int(elapsed/time.Hour))
	default:
		return fmt.Sprintf("%d days ago", int(elapsed/day))
	}
}

// DayAge words a date-only report in whole calendar days, since the source
// never said the time of day (FR-SEL-004).
func DayAge(date, now time.Time) string {
	then := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	today := now.UTC()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	switch days := int(today.Sub(then) / day); {
	case days <= 0:
		return "today"
	case days == 1:
		return "yesterday"
	default:
		return fmt.Sprintf("%d days ago", days)
	}
}

// Reported words an observation's time as its precision allows: an instant
// as "Observed 14 min ago", a date as "Reported for 18 Sep 2026, 5 days ago".
func Reported(o event.Observation, now time.Time) string {
	if o.Precision == event.Day {
		return fmt.Sprintf("Reported for %s, %s", o.At.UTC().Format(dateLayout), DayAge(o.At, now))
	}
	return "Observed " + Age(o.At, now)
}

// CountLine is FR-CNT-001: "3 events in the last 24 h".
func CountLine(n int, windowLabel string) string {
	return fmt.Sprintf("%s in the last %s", events(n), windowLabel)
}

// instantLayout words a replay instant: "20 Sep 14:00".
const instantLayout = "2 Jan 15:04"

// Instant words a replay instant in UTC: "20 Sep 14:00".
func Instant(at time.Time) string { return at.UTC().Format(instantLayout) }

// ReplayCountLine is FR-RPL-011: "3 events up to 20 Sep 14:00 UTC in the last 7 days".
func ReplayCountLine(n int, at time.Time, windowLabel string) string {
	return fmt.Sprintf("%s up to %s UTC in the last %s", events(n), Instant(at), windowLabel)
}

// ReplayLine is FR-RPL-019: "Replay: 20 Sep 14:00 UTC".
func ReplayLine(at time.Time) string {
	return "Replay: " + Instant(at) + " UTC"
}

// events words a count of events: "1 event", "3 events".
func events(n int) string {
	if n == 1 {
		return "1 event"
	}
	return fmt.Sprintf("%d events", n)
}

// Until words a coming time as "in 4 min"; a time already arrived reads "now".
func Until(then, now time.Time) string {
	wait := then.Sub(now)
	switch {
	case wait < time.Minute:
		return "now"
	case wait < hoursBoundary:
		return fmt.Sprintf("in %d min", int(wait/time.Minute))
	default:
		return fmt.Sprintf("in %d h", int(wait/time.Hour))
	}
}

// Retrieved words when a provider's data last arrived.
func Retrieved(at, now time.Time) string {
	return "Retrieved " + Age(at, now)
}

// Stale reports whether data last retrieved at lastSuccess has gone stale for
// a provider refreshed every interval. Data never retrieved is not stale; it
// is loading, which the status model says separately.
func Stale(lastSuccess, now time.Time, interval time.Duration) bool {
	if lastSuccess.IsZero() {
		return false
	}
	return now.Sub(lastSuccess) > StaleAfterIntervals*interval
}

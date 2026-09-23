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

// Stale reports whether data last retrieved at lastSuccess has gone stale for
// a provider refreshed every interval. Data never retrieved is not stale; it
// is loading, which the status model says separately.
func Stale(lastSuccess, now time.Time, interval time.Duration) bool {
	if lastSuccess.IsZero() {
		return false
	}
	return now.Sub(lastSuccess) > StaleAfterIntervals*interval
}

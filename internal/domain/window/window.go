// Package window holds the user's time window (REQUIREMENTS.md FR-TW-001 to 003)
// and the rule deciding which part of an event falls inside it (DATA-003, DATA-004).
package window

import (
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
)

// hoursPerDay converts the day-long windows to durations.
const hoursPerDay = 24

// Window is one selectable span ending at the current instant.
type Window struct {
	Key    string
	Label  string
	Length time.Duration
}

// The five windows of FR-TW-001.
var (
	OneHour   = Window{Key: "1h", Label: "1 h", Length: time.Hour}
	SixHours  = Window{Key: "6h", Label: "6 h", Length: 6 * time.Hour}
	OneDay    = Window{Key: "24h", Label: "24 h", Length: hoursPerDay * time.Hour}
	ThreeDays = Window{Key: "3d", Label: "3 days", Length: 3 * hoursPerDay * time.Hour}
	SevenDays = Window{Key: "7d", Label: "7 days", Length: 7 * hoursPerDay * time.Hour}
	Default   = OneDay
	Widest    = SevenDays
)

// All answers the windows in the order the control offers them.
func All() []Window {
	return []Window{OneHour, SixHours, OneDay, ThreeDays, SevenDays}
}

// ByKey finds a window by its key; false when no window has it.
func ByKey(key string) (Window, bool) {
	for _, w := range All() {
		if w.Key == key {
			return w, true
		}
	}
	return Window{}, false
}

// ClockSkew is how far ahead of this machine's clock an instant may lie and
// still count as now: a machine clock running slow must not hide the newest
// events. It is a bound, not an allowance for the future: GDACS published a flood
// alert dated days ahead (EONET_24511 on 2026-09-23 carried 2026-10-04), which
// is not an event that has happened.
const ClockSkew = 15 * time.Minute

// Contains reports whether an instant falls in the window ending at now, now
// stretched by ClockSkew (FR-TW-002).
func (w Window) Contains(at, now time.Time) bool {
	return at.After(now.Add(-w.Length)) && !at.After(now.Add(ClockSkew))
}

// Latest answers the event's newest observation inside the window; false when
// none falls inside it, which means the event is not shown (FR-TW-002). That
// observation's time is the event time and its position the marker position.
func (w Window) Latest(e event.Event, now time.Time) (event.Observation, bool) {
	for i := len(e.Observations) - 1; i >= 0; i-- {
		if o := e.Observations[i]; w.Contains(o.At, now) {
			return o, true
		}
	}
	return event.Observation{}, false
}

package window

import "time"

// day is the span of one UTC calendar day, the unit GWIS maps burnt areas in.
const day = hoursPerDay * time.Hour

// Days is FR-BA-001: every UTC calendar day that begins before now and ends
// after the window's start, oldest first, each as its midnight in UTC. The
// window is shown as it was chosen, so a day it only partly covers still counts.
func (w Window) Days(now time.Time) []time.Time {
	now = now.UTC()
	start := now.Add(-w.Length)
	var days []time.Time
	for d := start.Truncate(day); d.Before(now); d = d.Add(day) {
		if d.Add(day).After(start) {
			days = append(days, d)
		}
	}
	return days
}

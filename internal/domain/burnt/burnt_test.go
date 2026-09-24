package burnt

import (
	"testing"
	"time"
)

func at(month time.Month, dayOfMonth, hour, minute int) time.Time {
	return time.Date(2026, month, dayOfMonth, hour, minute, 0, 0, time.UTC)
}

func TestFRBA006_UnionKeepsTheHigherOpacity(t *testing.T) {
	t.Parallel()
	into := []uint8{0, 102, 230, 0}
	Union(into, []uint8{0, 230, 102, 50})
	want := []uint8{0, 230, 230, 50}
	for i := range want {
		if into[i] != want[i] {
			t.Fatalf("got %v, want %v", into, want)
		}
	}
}

func TestFRBA006_UnionOfUnequalPlanesJoinsWhatBothHold(t *testing.T) {
	t.Parallel()
	into := []uint8{1, 1, 1}
	Union(into, []uint8{9})
	if into[0] != 9 || into[1] != 1 || into[2] != 1 {
		t.Fatalf("got %v", into)
	}
	short := []uint8{1}
	Union(short, []uint8{0, 9})
	if short[0] != 1 {
		t.Fatalf("got %v", short)
	}
}

func TestFRBA008_StatusWordsTheSpanAndTheAge(t *testing.T) {
	t.Parallel()
	now := at(time.September, 24, 12, 0)
	retrieved := at(time.September, 24, 11, 48)
	cases := []struct {
		name        string
		first, last time.Time
		want        string
	}{
		{"a span in one month", at(time.September, 17, 0, 0), at(time.September, 23, 0, 0),
			"Burnt areas: 17 to 23 Sep (UTC), retrieved 12 min ago"},
		{"one day", at(time.September, 23, 0, 0), at(time.September, 23, 0, 0),
			"Burnt areas: 23 Sep (UTC), retrieved 12 min ago"},
		{"across a month", at(time.September, 30, 0, 0), at(time.October, 2, 0, 0),
			"Burnt areas: 30 Sep to 2 Oct (UTC), retrieved 12 min ago"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := Status(c.first, c.last, retrieved, now); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestFRBA008_SpanIsUTC(t *testing.T) {
	t.Parallel()
	bst := time.FixedZone("BST", int(time.Hour/time.Second))
	late := time.Date(2026, time.September, 24, 0, 30, 0, 0, bst) // 23:30 UTC on the 23rd
	if got := Span(late, late); got != "23 Sep" {
		t.Errorf("got %q, want 23 Sep", got)
	}
}

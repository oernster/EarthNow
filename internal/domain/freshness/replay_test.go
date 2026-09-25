package freshness

import (
	"testing"
	"time"
)

func TestFRRPL011_TheCountLineNamesTheReplayInstant(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, 9, 20, 14, 0, 0, 0, time.UTC)
	if got := ReplayCountLine(3, at, "7 days"); got != "3 events up to 20 Sep 14:00 UTC in the last 7 days" {
		t.Errorf("got %q", got)
	}
	if got := ReplayCountLine(1, at, "24 h"); got != "1 event up to 20 Sep 14:00 UTC in the last 24 h" {
		t.Errorf("got %q", got)
	}
}

func TestFRRPL019_TheStatusNamesTheReplayInstantInUTC(t *testing.T) {
	t.Parallel()
	bst := time.FixedZone("BST", int(time.Hour/time.Second))
	at := time.Date(2026, 9, 20, 15, 0, 0, 0, bst)
	if got := ReplayLine(at); got != "Replay: 20 Sep 14:00 UTC" {
		t.Errorf("got %q", got)
	}
}

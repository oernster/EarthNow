package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/cloud"
	"github.com/oernster/EarthNow/internal/domain/event"
	"github.com/oernster/EarthNow/internal/domain/sun"
	"github.com/oernster/EarthNow/internal/domain/window"
)

// sepAt is an instant in September 2026, UTC.
func sepAt(dayOfMonth, hour int) time.Time {
	return time.Date(2026, 9, dayOfMonth, hour, 0, 0, 0, time.UTC)
}

// replayCloudsFor follows a 24 h span ending 24 Sep 12:00, the service's
// newest image at 09:00 of that day: eight times from 23 Sep 12:00.
func replayCloudsFor(t *testing.T) (*ReplayClouds, *fakeClock, *fakeCloudSource) {
	t.Helper()
	clock := &fakeClock{now: sepAt(24, 12)}
	source := &fakeCloudSource{listed: sepAt(24, 9), replayFail: map[int64]error{}}
	r := NewReplayClouds(clock, source)
	if !r.Follow(sepAt(23, 12), sepAt(24, 12)) {
		t.Fatal("a new span should make a round due")
	}
	return r, clock, source
}

func replayRound(t *testing.T, r *ReplayClouds) error {
	t.Helper()
	if !r.Due() {
		t.Fatal("no round due")
	}
	return r.Refresh(context.Background())
}

func TestFRRPL015_TheSpansImagesAreFetchedOneARoundAtTheReplaySize(t *testing.T) {
	t.Parallel()
	r, _, source := replayCloudsFor(t)
	if line, p := r.Status(); line != cloud.ReplayListing || !p.Loading {
		t.Fatalf("before the listing: %q %+v", line, p)
	}
	if r.Follow(sepAt(23, 12), sepAt(24, 12)) {
		t.Fatal("the same span is not new")
	}
	_ = replayRound(t, r)
	if line, p := r.Status(); line != "Clouds 0 of 8" || !p.Refreshing {
		t.Fatalf("after the listing: %q %+v", line, p)
	}
	for range 8 {
		if err := replayRound(t, r); err != nil {
			t.Fatal(err)
		}
	}
	if len(source.replayFor) != 8 || !source.replayFor[0].Equal(sepAt(23, 12)) || !source.replayFor[7].Equal(sepAt(24, 9)) {
		t.Fatalf("fetched %v", source.replayFor)
	}
	if len(source.fetchedFor) != 0 {
		t.Fatal("a replay asked for full-size images")
	}
	if line, _ := r.Status(); line != "" || r.Due() || !r.NextWake().IsZero() {
		t.Fatalf("once all are in: %q", line)
	}
}

func TestFRRPL014_TheImageDrawnIsTheLatestHeldAtOrBeforeTheInstant(t *testing.T) {
	t.Parallel()
	r, _, _ := replayCloudsFor(t)
	for range 3 {
		_ = replayRound(t, r)
	}
	if got := r.Pick(sepAt(23, 14)); got != "2026-09-23T12:00:00Z" {
		t.Errorf("Pick = %q", got)
	}
	if got := r.Pick(sepAt(23, 11)); got != "" {
		t.Errorf("before the first image: %q", got)
	}
	if got := r.Image("2026-09-23T12:00:00Z"); !strings.HasPrefix(got, pngDataURL) {
		t.Errorf("Image = %q", got)
	}
	if r.Image("2026-09-23T18:00:00Z") != "" || r.Image("noon") != "" {
		t.Error("an image not held answers empty; so does a time that is not one")
	}
}

func TestFRRPL016_PlayDoesNotWaitWhileTheImagesArrive(t *testing.T) {
	t.Parallel()
	r, _, _ := replayCloudsFor(t)
	_ = replayRound(t, r)
	_ = replayRound(t, r)
	if line, _ := r.Status(); line != "Clouds 1 of 8" {
		t.Fatalf("line %q", line)
	}
	if r.NextWake().IsZero() {
		t.Fatal("the next image should be due")
	}
}

func TestFRRPL017_AMissingImageIsNamedAndTheNextFollows(t *testing.T) {
	t.Parallel()
	r, _, source := replayCloudsFor(t)
	source.replayFail[sepAt(23, 15).Unix()] = errors.New("timed out")
	_ = replayRound(t, r)
	_ = replayRound(t, r)
	if err := replayRound(t, r); err == nil {
		t.Fatal("the 15:00 image should fail")
	}
	_ = replayRound(t, r)
	if got := r.Pick(sepAt(23, 16)); got != "2026-09-23T12:00:00Z" {
		t.Errorf("16:00 draws %q, want the 12:00 image", got)
	}
	if _, p := r.Status(); p.Problem != "Missing images: 23 Sep 15:00 UTC" {
		t.Errorf("popover %+v", p)
	}
}

func TestFRRPL015_AFailedListingIsRetriedByTheBackoff(t *testing.T) {
	t.Parallel()
	r, clock, source := replayCloudsFor(t)
	source.latestErr = errors.New("offline")
	if err := replayRound(t, r); err == nil {
		t.Fatal("the listing should fail")
	}
	if _, p := r.Status(); p.Loading || p.Problem != "offline" || p.NextAttempt == "" {
		t.Fatalf("popover %+v", p)
	}
	if r.Due() || !r.NextWake().Equal(clock.now.Add(Backoff(CloudInterval, 1))) {
		t.Fatalf("retry at %v", r.NextWake())
	}
	source.latestErr = nil
	clock.now = r.NextWake()
	if err := replayRound(t, r); err != nil {
		t.Fatal(err)
	}
}

func TestFRRPL018_ReturningToTheEndReleasesTheImages(t *testing.T) {
	t.Parallel()
	r, _, _ := replayCloudsFor(t)
	_ = replayRound(t, r)
	_ = replayRound(t, r)
	r.Stop()
	if line, _ := r.Status(); line != "" || r.Pick(sepAt(24, 12)) != "" || r.Due() || !r.NextWake().IsZero() {
		t.Fatal("images held after the return")
	}
}

func TestFRRPL015_ARoundForASpanNoLongerFollowedIsDropped(t *testing.T) {
	t.Parallel()
	r, _, source := replayCloudsFor(t)
	source.onLatest = func() { r.Follow(sepAt(22, 12), sepAt(23, 12)) }
	_ = replayRound(t, r)
	if line, _ := r.Status(); line != cloud.ReplayListing {
		t.Fatalf("the old span's listing was kept: %q", line)
	}
	_ = replayRound(t, r)
	source.onReplay = func() { r.Follow(sepAt(21, 12), sepAt(22, 12)) }
	_ = replayRound(t, r)
	if r.Pick(sepAt(23, 12)) != "" {
		t.Fatal("the old span's image was kept")
	}
}

// replayFor builds the replay over a globe holding one quake at 20 Sep 10:00.
func replayFor(t *testing.T, cloudsShown bool) (*Replay, *fakeClock, *BurntAreas) {
	t.Helper()
	clock := &fakeClock{now: sepAt(24, 12)}
	store := NewStore(clock)
	store.Apply(event.USGS, ports.Fetched{Events: []event.Event{{
		Provider: event.USGS, ProviderEventID: "q", Category: event.Earthquake,
		Observations: []event.Observation{{At: sepAt(20, 10)}},
	}}})
	globe := NewGlobe(store, clock, nil)
	clouds := NewClouds(clock, &fakeCloudSource{listed: sepAt(24, 9)}, &fakeCloudCache{})
	clouds.SetShown(cloudsShown)
	burntLayer, _, _, _ := burntFor(window.SevenDays)
	burntLayer.clock = clock
	return NewReplay(globe, NewSun(clock), clouds, burntLayer, NewReplayClouds(clock, &fakeCloudSource{listed: sepAt(24, 9)})), clock, burntLayer
}

func TestFRRPL022_AnEventFromAfterTheSpansEndWaitsForTheReturn(t *testing.T) {
	t.Parallel()
	r, clock, _ := replayFor(t, false)
	clock.now = sepAt(24, 14)
	r.globe.store.Apply(event.EONET, ports.Fetched{Events: []event.Event{{
		Provider: event.EONET, ProviderEventID: "late", Category: event.Wildfire,
		Observations: []event.Observation{{At: sepAt(24, 13)}},
	}}})
	for _, position := range []float64{0.5, 1} {
		frame, _ := r.Frame(window.SevenDays.Key, Filter{}, sepAt(24, 12), position)
		for _, e := range frame.View.Events {
			if e.Provider == string(event.EONET) {
				t.Fatalf("at %v the replay drew an event from after its span", position)
			}
		}
	}
	if got := len(r.globe.View(window.SevenDays.Key, Filter{}).Events); got != 2 {
		t.Fatalf("the ordinary view holds %d events, want both", got)
	}
}

// The frame also carries the sun at the replay instant (FR-RPL-012).
func TestFRRPL009_AFrameShowsWhatHappenedByTheInstant(t *testing.T) {
	t.Parallel()
	r, _, _ := replayFor(t, true)
	end := sepAt(24, 12)
	before, due := r.Frame(window.SevenDays.Key, Filter{}, end, 0.3)
	if !due || len(before.View.Events) != 0 {
		t.Fatalf("at 0.3 (19 Sep 14:24): %d events, due %v", len(before.View.Events), due)
	}
	after, due := r.Frame(window.SevenDays.Key, Filter{}, end, 0.5)
	if due || len(after.View.Events) != 1 {
		t.Fatalf("at 0.5 (21 Sep 00:00): %d events, due %v", len(after.View.Events), due)
	}
	if after.View.CountLine != "1 event up to 21 Sep 00:00 UTC in the last 7 days" || after.Line != "Replay: 21 Sep 00:00 UTC" {
		t.Errorf("lines %q, %q", after.View.CountLine, after.Line)
	}
	want := sun.Subsolar(sepAt(21, 0))
	if after.At != "2026-09-21T00:00:00Z" || after.Sun.Lat != want.Lat || after.Sun.Lng != want.Lng {
		t.Errorf("at %s, sun %+v", after.At, after.Sun)
	}
	if after.CloudsLine != cloud.ReplayListing {
		t.Errorf("clouds line %q", after.CloudsLine)
	}
}

func TestFRRPL013_TheBurntDaysBuildUpWithTheReplay(t *testing.T) {
	t.Parallel()
	r, _, burntLayer := replayFor(t, false)
	burntLayer.SetShown(true)
	_ = runRound(t, burntLayer)
	end := sepAt(24, 12)
	frame, due := r.Frame(window.SevenDays.Key, Filter{}, end, 0.5)
	if due || frame.CloudsLine != "" {
		t.Fatal("no cloud images are followed while the cloud layer is hidden")
	}
	if !strings.Contains(frame.BurntKey, "2026-09-21") || strings.Contains(frame.BurntKey, "2026-09-22") {
		t.Errorf("key at 21 Sep 00:00: %s", frame.BurntKey)
	}
	if got := r.BurntImage(window.SevenDays.Key, end, 0.5); !strings.HasPrefix(got, pngDataURL) {
		t.Errorf("image %q", got)
	}
	r.End()
}

package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/burnt"
	"github.com/oernster/EarthNow/internal/domain/window"
)

// fakeBurntSource answers every day with an image naming it, drawn unless the
// day is listed empty, failing for the days listed.
type fakeBurntSource struct {
	asked      []string
	empty      map[string]bool
	fail       map[string]error
	composed   int
	composeErr error
	// during runs once, inside the first Day asked, as a change arriving mid-round.
	during func()
}

func (f *fakeBurntSource) Day(ctx context.Context, day time.Time) ([]byte, bool, error) {
	name := day.Format(dayKeyLayout)
	f.asked = append(f.asked, name)
	if f.during != nil {
		during := f.during
		f.during = nil
		during()
	}
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	if err := f.fail[name]; err != nil {
		return nil, false, err
	}
	return []byte(name), !f.empty[name], nil
}

func (f *fakeBurntSource) Compose(images [][]byte) ([]byte, error) {
	f.composed++
	if f.composeErr != nil {
		return nil, f.composeErr
	}
	parts := make([]string, len(images))
	for i, img := range images {
		parts[i] = string(img)
	}
	return []byte(strings.Join(parts, "+")), nil
}

type fakeBurntCache struct {
	days    []ports.BurntDay
	held    bool
	loadErr error
	saveErr error
	saved   []ports.BurntDay
}

func (f *fakeBurntCache) Load() ([]ports.BurntDay, bool, error) { return f.days, f.held, f.loadErr }
func (f *fakeBurntCache) Save(days []ports.BurntDay) error {
	f.saved = days
	return f.saveErr
}

// noonOn24 is 12:00 UTC on 24 Sep 2026, the instant of FR-BA-001's examples.
var noonOn24 = time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

func sep(dayOfMonth int) time.Time { return time.Date(2026, 9, dayOfMonth, 0, 0, 0, 0, time.UTC) }

func burntFor(w window.Window) (*BurntAreas, *fakeClock, *fakeBurntSource, *fakeBurntCache) {
	clock := &fakeClock{now: noonOn24}
	source := &fakeBurntSource{empty: map[string]bool{}, fail: map[string]error{}}
	cache := &fakeBurntCache{}
	b := NewBurntAreas(clock, source, cache)
	b.SetWindow(w.Key)
	return b, clock, source, cache
}

func runRound(t *testing.T, b *BurntAreas) error {
	t.Helper()
	if !b.Due() {
		t.Fatal("no round due")
	}
	_, err := b.Refresh(context.Background())
	return err
}

func TestFRBA002_ARoundAsksForEachDayOfTheWindowOnce(t *testing.T) {
	t.Parallel()
	b, _, source, cache := burntFor(window.SevenDays)
	b.SetShown(true)
	if err := runRound(t, b); err != nil {
		t.Fatal(err)
	}
	want := "2026-09-17 2026-09-18 2026-09-19 2026-09-20 2026-09-21 2026-09-22 2026-09-23 2026-09-24"
	if got := strings.Join(source.asked, " "); got != want {
		t.Fatalf("asked %s, want %s", got, want)
	}
	if len(cache.saved) != 8 || !cache.saved[0].Day.Equal(sep(17)) {
		t.Fatalf("saved %v", cache.saved)
	}
}

func TestFRBA003_EveryDayIsAskedAgainEachInterval(t *testing.T) {
	t.Parallel()
	b, clock, source, _ := burntFor(window.OneDay)
	b.SetShown(true)
	_ = runRound(t, b)
	if b.Due() {
		t.Fatal("due again before the interval")
	}
	if !b.NextWake().Equal(noonOn24.Add(BurntInterval)) {
		t.Fatalf("next wake %v", b.NextWake())
	}
	clock.now = clock.now.Add(BurntInterval)
	_ = runRound(t, b)
	if got := strings.Join(source.asked, " "); got != "2026-09-23 2026-09-24 2026-09-23 2026-09-24" {
		t.Fatalf("asked %s", got)
	}
}

func TestFRBA004_ANewWindowAsksOnlyForTheDaysNotHeld(t *testing.T) {
	t.Parallel()
	b, _, source, _ := burntFor(window.OneDay)
	b.SetShown(true)
	_ = runRound(t, b)
	source.asked = nil
	if !b.SetWindow(window.ThreeDays.Key) {
		t.Fatal("a wider window should make a round due")
	}
	_ = runRound(t, b)
	if got := strings.Join(source.asked, " "); got != "2026-09-21 2026-09-22" {
		t.Fatalf("asked %s, want 21 and 22 Sep alone", got)
	}
	if b.SetWindow(window.OneDay.Key) {
		t.Fatal("a narrower window needs nothing")
	}
}

func TestFRBA004_AWindowChangeWhileHiddenOrUnknownAsksNothing(t *testing.T) {
	t.Parallel()
	b, _, _, _ := burntFor(window.OneDay)
	if b.SetWindow(window.SevenDays.Key) {
		t.Fatal("hidden: nothing due")
	}
	if b.SetWindow("fortnight") || b.SetWindow(window.SevenDays.Key) {
		t.Fatal("an unknown or unchanged window changes nothing")
	}
}

func TestFRBA004_AWindowChangedDuringARoundIsFetchedWhenItEnds(t *testing.T) {
	t.Parallel()
	b, _, source, _ := burntFor(window.OneDay)
	b.SetShown(true)
	source.during = func() {
		if b.SetWindow(window.ThreeDays.Key) {
			t.Error("a running round takes the change when it ends")
		}
	}
	_ = runRound(t, b)
	source.asked = nil
	_ = runRound(t, b)
	if got := strings.Join(source.asked, " "); got != "2026-09-21 2026-09-22" {
		t.Fatalf("asked %s", got)
	}
}

func TestFRBA005_NothingIsAskedWhileHidden(t *testing.T) {
	t.Parallel()
	b, clock, source, _ := burntFor(window.SevenDays)
	for range 2 {
		if b.Due() || !b.NextWake().IsZero() {
			t.Fatal("due while hidden")
		}
		clock.now = clock.now.Add(BurntInterval)
	}
	if len(source.asked) != 0 {
		t.Fatalf("asked %v", source.asked)
	}
	if got := b.Status(); got.Line != "" || got.Shown {
		t.Fatalf("hidden status %+v", got)
	}
}

func TestFRBA008_StatusNamesTheDrawnSpanAndItsAge(t *testing.T) {
	t.Parallel()
	b, clock, source, _ := burntFor(window.SevenDays)
	source.empty["2026-09-24"] = true
	b.SetShown(true)
	_ = runRound(t, b)
	clock.now = clock.now.Add(12 * time.Minute)
	got := b.Status()
	if got.Line != "Burnt areas: 17 to 23 Sep (UTC), retrieved 12 min ago" {
		t.Fatalf("line %q", got.Line)
	}
	if got.Key == "" || got.Provider.Name != BurntProvider || got.Provider.Retrieved != "Retrieved 12 min ago" {
		t.Fatalf("status %+v", got)
	}
}

func TestFRBA009_AWindowWithNothingMappedSaysSo(t *testing.T) {
	t.Parallel()
	b, _, source, _ := burntFor(window.OneHour)
	source.empty["2026-09-24"] = true
	b.SetShown(true)
	if got := b.Status().Line; got != burnt.Retrieving {
		t.Fatalf("before the first round %q", got)
	}
	if !b.Status().Provider.Loading {
		t.Fatal("loading before the first round")
	}
	_ = runRound(t, b)
	got := b.Status()
	if got.Line != burnt.NoneYet || got.Key != "" || b.Image() != "" {
		t.Fatalf("status %+v", got)
	}
}

func TestFRBA012_ADayThatFailsLeavesTheOthersDrawn(t *testing.T) {
	t.Parallel()
	b, clock, source, _ := burntFor(window.OneDay)
	b.SetShown(true)
	_ = runRound(t, b)
	source.fail["2026-09-23"] = errors.New("timed out")
	clock.now = clock.now.Add(BurntInterval)
	err := runRound(t, b)
	if err == nil || !strings.Contains(err.Error(), "23 Sep: timed out") {
		t.Fatalf("err %v", err)
	}
	got := b.Status()
	if !strings.HasPrefix(got.Line, "Burnt areas: 23 to 24 Sep") {
		t.Fatalf("the held 23 Sep should still draw: %q", got.Line)
	}
	if got.Provider.Problem != "23 Sep: timed out" || got.Provider.NextAttempt == "" {
		t.Fatalf("provider %+v", got.Provider)
	}
	if !b.NextWake().Equal(clock.now.Add(Backoff(BurntInterval, 1))) {
		t.Fatalf("retry at %v", b.NextWake())
	}
	delete(source.fail, "2026-09-23")
	clock.now = b.NextWake()
	if err := runRound(t, b); err != nil || b.Status().Provider.Problem != "" {
		t.Fatalf("recovery: %v %+v", err, b.Status().Provider)
	}
}

func TestFRBA012_ARunningRoundIsRefreshingInThePopover(t *testing.T) {
	t.Parallel()
	b, clock, _, _ := burntFor(window.OneDay)
	b.SetShown(true)
	_ = runRound(t, b)
	clock.now = clock.now.Add(BurntInterval)
	if !b.Due() || !b.Status().Provider.Refreshing || !b.NextWake().IsZero() {
		t.Fatal("a running round with days held reads refreshing")
	}
}

func TestFRBA013_ARefusedAnswerIsAFailedDay(t *testing.T) {
	t.Parallel()
	b, _, source, _ := burntFor(window.OneHour)
	source.fail["2026-09-24"] = errors.New("the answer is not a PNG image")
	b.SetShown(true)
	if err := runRound(t, b); err == nil {
		t.Fatal("a refused answer must fail the round")
	}
}

func TestFRBA017_NothingHeldAndAFailureSaysTheMapsCouldNotBeRetrieved(t *testing.T) {
	t.Parallel()
	b, _, source, _ := burntFor(window.OneHour)
	source.fail["2026-09-24"] = errors.New("offline")
	b.SetShown(true)
	_ = runRound(t, b)
	got := b.Status()
	if got.Line != burnt.Unavailable || got.Provider.Loading {
		t.Fatalf("status %+v", got)
	}
}

func TestFRBA014_HeldDaysDrawOfflineAndOldOnesAreDiscarded(t *testing.T) {
	t.Parallel()
	b, _, _, cache := burntFor(window.ThreeDays)
	earlier := noonOn24.Add(-3 * time.Hour)
	cache.held = true
	cache.days = []ports.BurntDay{
		{Day: sep(10), PNG: []byte("old"), Drawn: true, RetrievedAt: earlier},
		{Day: sep(22), PNG: []byte("2026-09-22"), Drawn: true, RetrievedAt: earlier.Add(time.Hour)},
		{Day: sep(23), PNG: []byte("2026-09-23"), Drawn: true, RetrievedAt: earlier},
	}
	b.Restore()
	b.SetShown(true)
	got := b.Status()
	if got.Line != "Burnt areas: 22 to 23 Sep (UTC), retrieved 3 h ago" {
		t.Fatalf("line %q", got.Line)
	}
	b.mu.Lock()
	_, kept := b.held[sep(10).Unix()]
	b.mu.Unlock()
	if kept {
		t.Fatal("10 Sep is older than the widest window and should be gone")
	}
}

func TestFRBA014_ACacheProblemIsSaidInTheNotice(t *testing.T) {
	t.Parallel()
	b, _, _, cache := burntFor(window.OneHour)
	cache.loadErr = errors.New("damaged")
	b.Restore()
	if got := b.Status().Notice; !strings.Contains(got, "could not be read: damaged") {
		t.Fatalf("notice %q", got)
	}
	cache.saveErr = errors.New("disk full")
	b.SetShown(true)
	_ = runRound(t, b)
	if got := b.Status().Notice; !strings.Contains(got, "could not be saved: disk full") {
		t.Fatalf("notice %q", got)
	}
	cache.saveErr = nil
	b.SetShown(false)
	b.SetShown(true)
	_ = runRound(t, b)
	if got := b.Status().Notice; got != "" {
		t.Fatalf("a good save clears the notice: %q", got)
	}
}

func TestFRBA006_TheImageIsTheDrawnDaysComposedOnce(t *testing.T) {
	t.Parallel()
	b, _, source, _ := burntFor(window.ThreeDays)
	source.empty["2026-09-22"] = true
	b.SetShown(true)
	_ = runRound(t, b)
	url := b.Image()
	if !strings.HasPrefix(url, pngDataURL) {
		t.Fatalf("image %q", url)
	}
	if b.Image() != url || source.composed != 1 {
		t.Fatalf("composed %d times; the same key composes once", source.composed)
	}
}

func TestFRBA006_AComposeFailureIsSaidAndDrawsNothing(t *testing.T) {
	t.Parallel()
	b, _, source, _ := burntFor(window.OneHour)
	source.composeErr = errors.New("bad pixels")
	b.SetShown(true)
	_ = runRound(t, b)
	if b.Image() != "" || !strings.Contains(b.Status().Notice, "could not be drawn: bad pixels") {
		t.Fatalf("notice %q", b.Status().Notice)
	}
}

func TestFRBA012_ACanceledRoundRecordsAFailure(t *testing.T) {
	t.Parallel()
	b, _, _, _ := burntFor(window.OneHour)
	b.SetShown(true)
	if !b.Due() {
		t.Fatal("no round due")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := b.Refresh(ctx); err == nil {
		t.Fatal("a canceled round must fail")
	}
}

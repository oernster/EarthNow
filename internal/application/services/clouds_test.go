package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/cloud"
)

// fakeCloudSource lists one valid time and answers its image; else it fails.
type fakeCloudSource struct {
	listed     time.Time
	latestErr  error
	imageErr   error
	checks     int
	fetchedFor []time.Time
	replayFor  []time.Time
	replayFail map[int64]error
	// onLatest and onReplay run once, inside the next call, as a change
	// arriving mid-round.
	onLatest func()
	onReplay func()
}

func (f *fakeCloudSource) Latest(context.Context) (time.Time, error) {
	f.checks++
	once(&f.onLatest)
	return f.listed, f.latestErr
}

func (f *fakeCloudSource) Image(_ context.Context, validTime time.Time) ([]byte, error) {
	f.fetchedFor = append(f.fetchedFor, validTime)
	if f.imageErr != nil {
		return nil, f.imageErr
	}
	return []byte("png of " + validTime.Format(time.RFC3339)), nil
}

// ReplayImage answers the replay's image, failing for the times in replayFail.
func (f *fakeCloudSource) ReplayImage(_ context.Context, validTime time.Time) ([]byte, error) {
	f.replayFor = append(f.replayFor, validTime)
	once(&f.onReplay)
	if err := f.replayFail[validTime.Unix()]; err != nil {
		return nil, err
	}
	return []byte("small png of " + validTime.Format(time.RFC3339)), nil
}

type fakeCloudCache struct {
	held    *ports.CloudImage
	loadErr error
	saveErr error
	saved   []ports.CloudImage
}

func (f *fakeCloudCache) Load() (ports.CloudImage, bool, error) {
	if f.loadErr != nil || f.held == nil {
		return ports.CloudImage{}, false, f.loadErr
	}
	return *f.held, true, nil
}

func (f *fakeCloudCache) Save(img ports.CloudImage) error {
	f.saved = append(f.saved, img)
	return f.saveErr
}

var imageAt = time.Date(2026, 9, 23, 9, 0, 0, 0, time.UTC)

func cloudsFor(clock *fakeClock, source *fakeCloudSource, cache *fakeCloudCache) *Clouds {
	return NewClouds(clock, source, cache)
}

// runDue starts and finishes one check if one is due; it answers whether one ran.
func runDue(t *testing.T, c *Clouds) bool {
	t.Helper()
	if !c.Due() {
		return false
	}
	_, _ = c.Refresh(context.Background())
	return true
}

func cloudProvider(t *testing.T, c *Clouds) dto.Provider {
	t.Helper()
	p := c.Status().Provider
	if p.Name != CloudProvider {
		t.Fatalf("no EUMETSAT entry while shown: %+v", p)
	}
	return p
}

func TestFRCLD003_TheLayerStartsHiddenAndItsChoiceIsKept(t *testing.T) {
	t.Parallel()
	if Defaults().CloudsShown {
		t.Error("a first run shows the cloud layer")
	}
	store := &fakeSettings{}
	p := NewPreferences(store, (&minimums{}).apply)
	p.Load()
	chosen := p.Current()
	chosen.CloudsShown = true
	if held, _ := p.Update(chosen); !held.CloudsShown || !store.saved[0].CloudsShown {
		t.Errorf("showing the layer was not kept: %+v, saved %+v", held, store.saved)
	}
}

func TestFRCLD005_NoRequestWhileHidden(t *testing.T) {
	t.Parallel()
	clock, source := &fakeClock{now: noon}, &fakeCloudSource{listed: imageAt}
	c := cloudsFor(clock, source, &fakeCloudCache{})
	for range 2 {
		if runDue(t, c) {
			t.Error("a check is due while the layer is hidden")
		}
		if !c.NextWake().IsZero() {
			t.Error("the driver is woken for a hidden layer")
		}
		clock.now = clock.now.Add(CloudInterval)
	}
	if source.checks != 0 || len(source.fetchedFor) != 0 {
		t.Errorf("%d checks and %d fetches while hidden", source.checks, len(source.fetchedFor))
	}
	if got := c.Status(); got.Shown || got.Line != "" || got.Provider != (dto.Provider{}) {
		t.Errorf("hidden status = %+v", got)
	}
}

func TestFRCLD004_OneCheckPerInterval(t *testing.T) {
	t.Parallel()
	clock, source := &fakeClock{now: noon}, &fakeCloudSource{listed: imageAt}
	c := cloudsFor(clock, source, &fakeCloudCache{})
	c.SetShown(true)
	c.SetShown(true) // showing it again does not make a check due again
	if !runDue(t, c) {
		t.Fatal("showing the layer made no check due")
	}
	if runDue(t, c) {
		t.Error("a second check is due within the interval")
	}
	if got, want := c.NextWake(), noon.Add(CloudInterval); !got.Equal(want) {
		t.Errorf("NextWake = %v, want %v", got, want)
	}
	clock.now = clock.now.Add(CloudInterval)
	if !runDue(t, c) {
		t.Error("no check after one interval")
	}
	if source.checks != 2 {
		t.Errorf("%d capabilities requests over two intervals, want 2", source.checks)
	}
}

func TestFRCLD016_AnImageIsFetchedOnlyForANewTime(t *testing.T) {
	t.Parallel()
	clock, source, cache := &fakeClock{now: noon}, &fakeCloudSource{listed: imageAt}, &fakeCloudCache{}
	c := cloudsFor(clock, source, cache)
	if c.Image() != "" {
		t.Error("an image before any was held")
	}
	c.SetShown(true)
	c.Due()
	if fresh, err := c.Refresh(context.Background()); !fresh || err != nil {
		t.Fatalf("first check = %v, %v", fresh, err)
	}
	clock.now = clock.now.Add(CloudInterval)
	c.Due()
	if fresh, _ := c.Refresh(context.Background()); fresh {
		t.Error("a check listing the held time answered a new image")
	}
	if len(source.fetchedFor) != 1 || !source.fetchedFor[0].Equal(imageAt) {
		t.Errorf("images fetched for %v, want once for %v", source.fetchedFor, imageAt)
	}
	if len(cache.saved) != 1 {
		t.Errorf("%d images saved, want 1", len(cache.saved))
	}
	if !strings.HasPrefix(c.Image(), pngDataURL) {
		t.Errorf("the image is not handed over as a PNG data URL: %.40q", c.Image())
	}
	if got := c.Status().ValidTime; got != "2026-09-23T09:00:00Z" {
		t.Errorf("ValidTime = %q", got)
	}
}

func TestFRCLD011_AFailureKeepsTheHeldImageAndBacksOff(t *testing.T) {
	t.Parallel()
	for _, source := range []*fakeCloudSource{
		{listed: imageAt.Add(cloud.ImageInterval), latestErr: errors.New("capabilities: HTTP 503")},
		{listed: imageAt.Add(cloud.ImageInterval), imageErr: errors.New("GetMap: HTTP 503")},
	} {
		clock := &fakeClock{now: noon}
		held := ports.CloudImage{ValidTime: imageAt, PNG: []byte("held")}
		c := cloudsFor(clock, source, &fakeCloudCache{held: &held})
		c.Restore()
		c.SetShown(true)
		before := c.Image()
		c.Due()
		if fresh, err := c.Refresh(context.Background()); fresh || err == nil {
			t.Fatalf("a failed fetch answered %v, %v", fresh, err)
		}
		if c.Image() != before {
			t.Error("a failed fetch dropped the held image")
		}
		p := cloudProvider(t, c)
		if p.Name != CloudProvider || p.Problem == "" || p.NextAttempt != "next attempt in 30 min" {
			t.Errorf("popover entry = %+v", p)
		}
		if got, want := c.NextWake(), noon.Add(Backoff(CloudInterval, 1)); !got.Equal(want) {
			t.Errorf("retry at %v, want %v", got, want)
		}
	}
}

func TestFRCLD013_AFirstFetchThatFailsIsSaid(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{now: noon}
	c := cloudsFor(clock, &fakeCloudSource{latestErr: errors.New("no route to host")}, &fakeCloudCache{})
	c.SetShown(true)
	if got := c.Status().Line; got != cloudLoading {
		t.Errorf("before the first fetch the line reads %q", got)
	}
	if p := cloudProvider(t, c); !p.Loading || p.Retrieved != "" {
		t.Errorf("before the first fetch the entry is %+v", p)
	}
	runDue(t, c)
	if got := c.Status().Line; got != cloud.Unavailable {
		t.Errorf("after a failed first fetch the line reads %q", got)
	}
	if p := cloudProvider(t, c); p.Loading || p.Problem != "no route to host" {
		t.Errorf("after a failed first fetch the entry is %+v", p)
	}
	if c.Image() != "" {
		t.Error("an image is drawn after a failed first fetch")
	}
}

func TestFRCLD014_TheHeldImageDrawsAtStartWithItsAge(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{now: imageAt.Add(3 * time.Hour)}
	held := ports.CloudImage{ValidTime: imageAt, PNG: []byte("held")}
	c := cloudsFor(clock, &fakeCloudSource{}, &fakeCloudCache{held: &held})
	c.Restore()
	c.SetShown(true)
	if got, want := c.Status().Line, "Clouds: image of 09:00 UTC, 3 h ago"; got != want {
		t.Errorf("line = %q, want %q", got, want)
	}
	if c.Image() == "" {
		t.Error("the held image is not drawn at start")
	}
	c.Due()
	if p := cloudProvider(t, c); !p.Refreshing || p.Loading || p.Stale {
		t.Errorf("entry while checking over a held image = %+v", p)
	}
	clock.now = imageAt.Add(3*cloud.ImageInterval + time.Minute)
	if p := cloudProvider(t, c); !p.Stale {
		t.Errorf("entry over a stale image = %+v", p)
	}
}

func TestClouds_CacheProblemsAreNotices(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{now: noon}
	cache := &fakeCloudCache{loadErr: errors.New("corrupt"), saveErr: errors.New("disk full")}
	c := cloudsFor(clock, &fakeCloudSource{listed: imageAt}, cache)
	c.Restore()
	if got := c.Status().Notice; !strings.Contains(got, "corrupt") {
		t.Errorf("unreadable cache notice = %q", got)
	}
	c.SetShown(true)
	runDue(t, c)
	if got := c.Status().Notice; !strings.Contains(got, "disk full") {
		t.Errorf("failed save notice = %q", got)
	}
	if p := cloudProvider(t, c); p.Retrieved == "" || p.Problem != "" {
		t.Errorf("a failed save is a fetch problem: %+v", p)
	}
	cache.saveErr = nil
	source := &fakeCloudSource{listed: imageAt.Add(cloud.ImageInterval)}
	c.source = source
	clock.now = clock.now.Add(CloudInterval)
	runDue(t, c)
	if got := c.Status().Notice; got != "" {
		t.Errorf("a good save left the notice %q", got)
	}
}

// once runs a hook and clears it, so it fires on the next call alone.
func once(hook *func()) {
	if h := *hook; h != nil {
		*hook = nil
		h()
	}
}

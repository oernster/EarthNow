package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
)

type fakeCache struct {
	held    map[event.Provider]ports.Snapshot
	loadErr error
	saveErr error
	saved   int
}

func (c *fakeCache) Load(p event.Provider) (ports.Snapshot, bool, error) {
	s, ok := c.held[p]
	return s, ok, c.loadErr
}

func (c *fakeCache) Save(p event.Provider, s ports.Snapshot) error {
	c.saved++
	if c.held == nil {
		c.held = map[event.Provider]ports.Snapshot{}
	}
	c.held[p] = s
	return c.saveErr
}

func TestFRSTS005_NoCacheIsANotice(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	g := NewGlobe(NewStore(clock), clock, nil)
	g.RestoreCached()
	if got := g.View("24h", Filter{}).Notice; got != NoCache {
		t.Errorf("notice = %q", got)
	}
}

func TestFRSTS004_CachedEventsShowBeforeTheFirstFetch(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	usgs := &fakeProvider{name: event.USGS, err: errors.New("offline")}
	cache := &fakeCache{held: map[event.Provider]ports.Snapshot{
		event.USGS: {Events: []event.Event{quake("a", 3.2, "")}, RetrievedAt: noon.Add(-3 * time.Hour)},
	}}
	g := NewGlobe(NewStore(clock), clock, []ports.Provider{usgs})
	g.UseCache(cache)
	g.RestoreCached()
	refreshEach(g)
	view := g.View("24h", Filter{})
	if len(view.Events) != 1 || view.Events[0].Retrieved != "Retrieved 3 h ago" || view.Notice != "" {
		t.Fatalf("view = %+v", view)
	}
	if p := view.Providers[0]; !p.Stale || p.Problem == "" {
		t.Errorf("provider = %+v, want stale with the failure", p)
	}
}

func TestSuccessfulFetchIsCached(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	cache := &fakeCache{}
	g := NewGlobe(NewStore(clock), clock, nil)
	g.UseCache(cache)
	p := &fakeProvider{name: event.USGS, fetched: ports.Fetched{Events: []event.Event{quake("a", 3, "")}}}
	_, _ = g.Refresh(context.Background(), p)
	if cache.saved != 1 || len(cache.held[event.USGS].Events) != 1 {
		t.Errorf("saved %d, held %+v", cache.saved, cache.held)
	}
	cache.saveErr = errors.New("disk full")
	_, _ = g.Refresh(context.Background(), p)
	if got := g.View("24h", Filter{}).Notice; got != "Events could not be cached: disk full" {
		t.Errorf("notice = %q", got)
	}
}

// DATA-009: the cache holds the set, its retrieved-at and its validator, less any
// event with nothing inside the widest window: eight days old is discarded, an
// event with one sighting inside the week is kept whole.
func TestDATA009_TheCacheDiscardsWhatTheWidestWindowCannotShow(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	cache := &fakeCache{}
	g := NewGlobe(NewStore(clock), clock, nil)
	g.UseCache(cache)
	week := 7 * 24 * time.Hour
	old := ev(event.USGS, "old", event.Earthquake, week+24*time.Hour)
	track := ev(event.USGS, "track", event.Earthquake, week+time.Hour)
	track.Observations = append(track.Observations, event.Observation{At: noon.Add(-time.Hour)})
	fresh := quake("fresh", 3, "")
	p := &fakeProvider{name: event.USGS, fetched: ports.Fetched{Events: []event.Event{old, track, fresh}, Validator: "v1"}}
	_, _ = g.Refresh(context.Background(), p)
	held := cache.held[event.USGS]
	var ids []string
	for _, e := range held.Events {
		ids = append(ids, e.ProviderEventID)
	}
	if len(ids) != 2 || ids[0] == "old" || ids[1] == "old" || held.Validator != "v1" || !held.RetrievedAt.Equal(noon) {
		t.Errorf("cached %v, validator %q, retrieved %v; want track and fresh", ids, held.Validator, held.RetrievedAt)
	}
	if len(held.Events[0].Observations)+len(held.Events[1].Observations) != 3 {
		t.Error("a kept event lost its older sightings")
	}
}

// FR-STS-005: a refused save's notice goes once that provider saves again, as
// the image layers' do. Another provider saving does not clear it; two
// providers' notices are both kept (audit round 2, E-2).
func TestFRSTS005_ACacheNoticeClearsOnTheNextGoodSave(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	usgsCache := &fakeCache{saveErr: errors.New("the file is locked by another process")}
	g := NewGlobe(NewStore(clock), clock, nil)
	g.UseCache(usgsCache)
	usgs := &fakeProvider{name: event.USGS, fetched: ports.Fetched{Events: []event.Event{quake("a", 3, "")}}}
	eonet := &fakeProvider{name: event.EONET, fetched: ports.Fetched{Events: []event.Event{ev(event.EONET, "e", event.Wildfire, time.Hour)}}}
	_, _ = g.Refresh(context.Background(), usgs)
	_, _ = g.Refresh(context.Background(), eonet)
	_, _ = g.Refresh(context.Background(), usgs)
	usgsCache.saveErr = nil
	_, _ = g.Refresh(context.Background(), eonet)
	if got := g.View("24h", Filter{}).Notice; got != "Events could not be cached: the file is locked by another process" {
		t.Errorf("after another provider saved, notice = %q; want USGS's kept", got)
	}
	_, _ = g.Refresh(context.Background(), usgs)
	_, _ = g.Refresh(context.Background(), usgs)
	if got := g.View("24h", Filter{}).Notice; got != "" {
		t.Errorf("after two good saves the notice still reads %q", got)
	}
}

// A cache that could not be read at start is mended by that provider's next
// good save, so its notice goes then; another provider's stays.
func TestFRSTS005_AReadNoticeClearsWhenThatProviderSaves(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	eonet := &fakeProvider{name: event.EONET, fetched: ports.Fetched{Events: []event.Event{ev(event.EONET, "e", event.Wildfire, time.Hour)}}}
	usgs := &fakeProvider{name: event.USGS, fetched: ports.Fetched{Events: []event.Event{quake("a", 3, "")}}}
	g := NewGlobe(NewStore(clock), clock, []ports.Provider{eonet, usgs})
	broken := &fakeCache{loadErr: errors.New("damaged")}
	g.UseCache(broken)
	g.RestoreCached()
	want := "Cached events could not be read: damaged" + noticeSeparator + "Cached events could not be read: damaged"
	if got := g.View("24h", Filter{}).Notice; got != want {
		t.Fatalf("notice = %q, want both providers' %q", got, want)
	}
	_, _ = g.Refresh(context.Background(), eonet)
	if got := g.View("24h", Filter{}).Notice; got != "Cached events could not be read: damaged" {
		t.Errorf("after EONET saved, notice = %q; want USGS's alone", got)
	}
}

// NFR-REL-001: an unreadable cache at startup is a stated notice in the window.
func TestUnreadableCacheIsANotice(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	g := NewGlobe(NewStore(clock), clock, []ports.Provider{&fakeProvider{name: event.EONET}})
	g.UseCache(&fakeCache{loadErr: errors.New("the EONET cache is damaged")})
	g.RestoreCached()
	if got := g.View("24h", Filter{}).Notice; got != "Cached events could not be read: the EONET cache is damaged" {
		t.Errorf("notice = %q", got)
	}
}

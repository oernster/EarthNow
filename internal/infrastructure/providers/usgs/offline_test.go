package usgs

// The offline promise end to end: the real Globe, Store, USGS adapter and disk
// cache, with only the network replaced by a fake fetcher (audit round 2, E-1
// and E-3).

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/application/services"
	"github.com/oernster/EarthNow/internal/infrastructure/cache"
	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

// offlineNow is the instant every offline test runs at.
var offlineNow = time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

// cacheCap is far above any body these tests write.
const cacheCap = 1 << 20

// offlineWindow is the window every offline view is asked for.
const offlineWindow = "24h"

// feature words one usable quake an hour before offlineNow.
func feature(id string, mag float64) string {
	at := offlineNow.Add(-time.Hour).UnixMilli()
	return fmt.Sprintf(`{"id":%q,"properties":{"mag":%v,"time":%d,"type":"earthquake"},"geometry":{"coordinates":[1,2,3]}}`, id, mag, at)
}

func collection(features ...string) []byte {
	return []byte(`{"type":"FeatureCollection","features":[` + strings.Join(features, ",") + `]}`)
}

// offlineRig is the composition root's wiring of one provider, its cache and
// the minimum's appliers, over a fetcher the test answers.
type offlineRig struct {
	fetcher *fakeFetcher
	quakes  *Adapter
	globe   *services.Globe
	prefs   *services.Preferences
	dir     string
}

// memorySettings is a settings store holding its file in memory.
type memorySettings struct{ held *ports.Settings }

func (m *memorySettings) Load(defaults ports.Settings) (ports.Settings, bool, error) {
	if m.held == nil {
		return defaults, false, nil
	}
	return *m.held, true, nil
}

func (m *memorySettings) Save(s ports.Settings) error { m.held = &s; return nil }
func (m *memorySettings) Path() string                { return "memory" }

// newOfflineRig wires as main.go does, starting under the magnitude key.
func newOfflineRig(t *testing.T, magnitude string) *offlineRig {
	t.Helper()
	clock := fixedClock{offlineNow}
	f := &fakeFetcher{}
	quakes := New(f, AllMagnitudes)
	globe := services.NewGlobe(services.NewStore(clock), clock, []ports.Provider{quakes})
	dir := t.TempDir()
	globe.UseCache(cache.New(dir, cacheCap))
	settings := services.Defaults()
	settings.Magnitude = magnitude
	prefs := services.NewPreferences(&memorySettings{held: &settings}, quakes.SetMinimum, globe.SetMinimum)
	prefs.Load()
	return &offlineRig{fetcher: f, quakes: quakes, globe: globe, prefs: prefs, dir: dir}
}

// choose is the reader picking a minimum magnitude in Settings.
func (r *offlineRig) choose(magnitude string) {
	chosen := r.prefs.Current()
	chosen.Magnitude = magnitude
	r.prefs.Update(chosen)
}

func (r *offlineRig) answer(body []byte) (services.Outcome, error) {
	r.fetcher.resp, r.fetcher.err = httpfetch.Response{Body: body}, nil
	return r.globe.Refresh(context.Background(), r.quakes)
}

func (r *offlineRig) fail() {
	r.fetcher.resp, r.fetcher.err = httpfetch.Response{}, errors.New("offline")
	_, _ = r.globe.Refresh(context.Background(), r.quakes)
}

// FR-PRV-012: a valid collection whose every feature is unusable (a schema
// drift: none carries properties.time) is a parse failure. The stored set and
// the cache are kept and the status names the reason.
func TestFRPRV012_AnAnswerWithNoUsableItemKeepsTheOfflineSet(t *testing.T) {
	t.Parallel()
	r := newOfflineRig(t, "all")
	if _, err := r.answer(collection(feature("a", 3), feature("b", 4))); err != nil {
		t.Fatal(err)
	}
	drifted := `{"id":"%s","properties":{"mag":3},"geometry":{"coordinates":[1,2,3]}}`
	out, err := r.answer(collection(fmt.Sprintf(drifted, "x"), fmt.Sprintf(drifted, "y"), fmt.Sprintf(drifted, "z")))
	if err == nil {
		t.Errorf("an answer with no usable item was taken as a success: %+v", out)
	}
	view := r.globe.View(offlineWindow, services.Filter{})
	if len(view.Events) != 2 || view.Providers[0].Problem == "" {
		t.Errorf("after the drifted answer: %d events, problem %q; want the 2 held and a reason", len(view.Events), view.Providers[0].Problem)
	}
	snap, held, err := cache.New(r.dir, cacheCap).Load(r.quakes.Name())
	if err != nil || !held || len(snap.Events) != 2 {
		t.Errorf("the cache holds %d events (held %v, %v); want the 2 kept", len(snap.Events), held, err)
	}
}

// FR-SET-002: a minimum raised while USGS cannot be reached hides the held
// quakes below it at once, keeping them held (the offline set is not cut).
func TestFRSET002_ARaisedMinimumHidesHeldQuakesBelowIt(t *testing.T) {
	t.Parallel()
	r := newOfflineRig(t, "2.5")
	if _, err := r.answer(collection(feature("small", 3), feature("large", 5))); err != nil {
		t.Fatal(err)
	}
	r.choose("4.5")
	r.fail()
	view := r.globe.View(offlineWindow, services.Filter{})
	if len(view.Events) != 1 || !strings.HasSuffix(view.Events[0].ID, "large") {
		t.Errorf("under a 4.5 minimum the view shows %+v; want only the M5", view.Events)
	}
}

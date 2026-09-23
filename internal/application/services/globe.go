package services

import (
	"context"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
	"github.com/oernster/EarthNow/internal/domain/freshness"
	"github.com/oernster/EarthNow/internal/domain/window"
)

// timeLayout is how instants cross the wire; the page shows UTC and local time.
const timeLayout = time.RFC3339

// secureScheme is the only scheme a source link may carry to be a link (FR-SEL-006).
const secureScheme = "https"

// attempt is what the last fetch of one provider did.
type attempt struct {
	running bool
	failed  error
}

// Globe is the use case the window drives: refresh the providers, then answer
// what to show for a window and filter.
type Globe struct {
	store     *Store
	clock     ports.Clock
	providers []ports.Provider
	mu        sync.Mutex
	attempts  map[event.Provider]attempt
	geocoder  ports.Geocoder
	cache     ports.SnapshotCache
	notice    string
}

// UseCache keeps each provider's set across runs. Without one the events live
// in memory only, which the view's notice says (FR-STS-005).
func (g *Globe) UseCache(cache ports.SnapshotCache) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.cache = cache
}

// NoCache is the notice given when events cannot be kept across a restart.
const NoCache = "Events are held in memory only and will not survive a restart"

// RestoreCached puts each provider's cached set back before any fetch, so
// the globe shows the last known events at once, marked with their retrieval
// time (FR-STS-004). An unreadable cache is a notice, never a stop.
func (g *Globe) RestoreCached() {
	g.mu.Lock()
	cache := g.cache
	g.mu.Unlock()
	if cache == nil {
		g.setNotice(NoCache)
		return
	}
	for _, p := range g.providers {
		snap, held, err := cache.Load(p.Name())
		if err != nil {
			g.setNotice("Cached events could not be read: " + err.Error())
			continue
		}
		if held {
			g.store.Restore(p.Name(), snap)
		}
	}
}

func (g *Globe) setNotice(n string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.notice = n
}

// save writes a provider's set to the cache after a successful fetch.
func (g *Globe) save(p event.Provider) {
	g.mu.Lock()
	cache := g.cache
	g.mu.Unlock()
	if cache == nil {
		return
	}
	snap, _ := g.store.Snapshot(p)
	if err := cache.Save(p, snap); err != nil {
		g.setNotice("Events could not be cached: " + err.Error())
	}
}

// NewGlobe builds the use case over its collaborators.
func NewGlobe(store *Store, clock ports.Clock, providers []ports.Provider) *Globe {
	return &Globe{store: store, clock: clock, providers: providers, attempts: map[event.Provider]attempt{}}
}

// SetGeocoder supplies the geocoder once its data has loaded; until then place
// lines are empty rather than blocking the window.
func (g *Globe) SetGeocoder(geocoder ports.Geocoder) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.geocoder = geocoder
}

// Place answers the place line for a point; empty while no geocoder is loaded.
func (g *Globe) Place(lat, lng float64) string {
	g.mu.Lock()
	geocoder := g.geocoder
	g.mu.Unlock()
	if geocoder == nil {
		return ""
	}
	return geocoder.Describe(lat, lng)
}

// RefreshAll fetches every provider at once. Each result lands in the store
// on its own, so one provider failing never holds up or spoils another
// (FR-PRV-008). It returns once every provider has answered.
func (g *Globe) RefreshAll(ctx context.Context) {
	var wg sync.WaitGroup
	for _, p := range g.providers {
		wg.Add(1)
		go func(p ports.Provider) {
			defer wg.Done()
			_, _ = g.Refresh(ctx, p)
		}(p)
	}
	wg.Wait()
}

// Outcome is what one successful fetch did, for the log (NFR-OBS-001).
type Outcome struct {
	NotModified bool
	Count       int
	Dropped     int
}

// Refresh fetches one provider, conditional on its stored validator.
func (g *Globe) Refresh(ctx context.Context, p ports.Provider) (Outcome, error) {
	g.record(p.Name(), attempt{running: true})
	validator := ""
	if snap, ok := g.store.Snapshot(p.Name()); ok {
		validator = snap.Validator
	}
	fetched, err := p.Fetch(ctx, validator)
	if err != nil {
		g.record(p.Name(), attempt{failed: err})
		return Outcome{}, err
	}
	g.store.Apply(p.Name(), fetched)
	g.record(p.Name(), attempt{})
	g.save(p.Name())
	snap, _ := g.store.Snapshot(p.Name())
	return Outcome{NotModified: fetched.NotModified, Count: len(snap.Events), Dropped: fetched.Dropped}, nil
}

func (g *Globe) record(p event.Provider, a attempt) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.attempts[p] = a
}

// Windows answers the time window choices (FR-TW-001).
func (g *Globe) Windows() []dto.Choice {
	var out []dto.Choice
	for _, w := range window.All() {
		out = append(out, dto.Choice{Key: w.Key, Label: w.Label})
	}
	return out
}

// View answers what to show for a window key and filter. An unknown key falls
// back to the default window rather than showing nothing.
func (g *Globe) View(windowKey string, f Filter) dto.View {
	w, ok := window.ByKey(windowKey)
	if !ok {
		w = window.Default
	}
	now := g.clock.Now()
	shown := g.store.Visible(w, f)
	view := dto.View{
		WindowKey: w.Key,
		CountLine: freshness.CountLine(len(shown), w.Label),
		Events:    make([]dto.Event, 0, len(shown)),
		Counts:    map[string]int{},
	}
	for c, n := range Counts(shown) {
		view.Counts[string(c)] = n
	}
	for _, s := range shown {
		view.Events = append(view.Events, toDTO(s, now))
	}
	view.Providers = g.statuses(now)
	g.mu.Lock()
	view.Notice = g.notice
	g.mu.Unlock()
	return view
}

func (g *Globe) statuses(now time.Time) []dto.Provider {
	g.mu.Lock()
	defer g.mu.Unlock()
	out := make([]dto.Provider, 0, len(g.providers))
	for _, p := range g.providers {
		a := g.attempts[p.Name()]
		status := dto.Provider{Name: string(p.Name())}
		snap, held := g.store.Snapshot(p.Name())
		if held {
			status.Retrieved = freshness.Retrieved(snap.RetrievedAt, now)
			status.Stale = freshness.Stale(snap.RetrievedAt, now, p.Interval())
		} else {
			status.Loading = a.running || a.failed == nil
		}
		if a.failed != nil {
			status.Problem = a.failed.Error()
		}
		out = append(out, status)
	}
	return out
}

func toDTO(s Shown, now time.Time) dto.Event {
	o := s.Observation
	out := dto.Event{
		ID:          s.Event.ID(),
		Provider:    string(s.Event.Provider),
		Category:    string(s.Event.Category),
		Title:       s.Event.Title,
		Description: s.Event.Description,
		Lat:         o.Where.Lat,
		Lng:         o.Where.Lng,
		At:          o.At.UTC().Format(timeLayout),
		DayOnly:     o.Precision == event.Day,
		Reported:    freshness.Reported(o, now),
		RetrievedAt: s.RetrievedAt.UTC().Format(timeLayout),
		Retrieved:   freshness.Retrieved(s.RetrievedAt, now),
		Band:        event.Band(s.Event, o),
		SourceURL:   SafeURL(s.Event.SourceURL),
		Ended:       s.Event.Ended(),
	}
	if out.SourceURL == "" {
		out.SourceText = s.Event.SourceURL
	}
	if m := o.Measurement; m != nil {
		out.Measurement = strconv.FormatFloat(m.Value, 'f', -1, 64)
		if m.Unit != "" {
			out.Measurement += " " + m.Unit
		}
	}
	return out
}

// SafeURL keeps a source link only when it is an absolute https URL with a
// host (FR-SEL-006) naming a page rather than a data file (FR-SEL-009); anything
// else is dropped rather than offered as a link, so the page shows it as text.
func SafeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != secureScheme || u.Host == "" || !event.IsPage(raw) {
		return ""
	}
	return raw
}

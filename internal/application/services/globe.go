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

// attempt is what the last fetch of one provider did. dropped is the count of
// unusable items in the last answer read (FR-PRV-013), kept across a 304 and a
// failure, since the set held is still the one that answer gave.
type attempt struct {
	running bool
	failed  error
	dropped int
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
	// noCache is NoCache while there is no cache at all; empty otherwise.
	noCache string
	// notices holds each provider's cache problem until its next good save.
	notices map[event.Provider]string
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
		g.mu.Lock()
		g.noCache = NoCache
		g.mu.Unlock()
		return
	}
	for _, p := range g.providers {
		snap, held, err := cache.Load(p.Name())
		if err != nil {
			g.setNotice(p.Name(), "Cached events could not be read: "+err.Error())
			continue
		}
		if held {
			g.store.Restore(p.Name(), snap)
		}
	}
}

// setNotice records a provider's cache problem; an empty one clears it.
func (g *Globe) setNotice(p event.Provider, n string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if n == "" {
		delete(g.notices, p)
		return
	}
	g.notices[p] = n
}

// save writes a provider's set to the cache after a successful fetch, less any
// event with no sighting inside the widest window, which no window can show
// (DATA-009). A good save clears that provider's notice, as the image layers'
// do: its file is whole again, whatever was wrong with it before.
func (g *Globe) save(p event.Provider) {
	g.mu.Lock()
	cache := g.cache
	g.mu.Unlock()
	if cache == nil {
		return
	}
	snap, _ := g.store.Snapshot(p)
	snap.Events = showable(snap.Events, g.clock.Now())
	if err := cache.Save(p, snap); err != nil {
		g.setNotice(p, "Events could not be cached: "+err.Error())
		return
	}
	g.setNotice(p, "")
}

// showable keeps each event some window can still show at now, whole: one
// with a sighting inside the widest window (sightings before it included) or
// an ongoing one whose report is current (FR-PRV-016).
func showable(events []event.Event, now time.Time) []event.Event {
	cutoff := now.Add(-window.Widest.Length)
	kept := make([]event.Event, 0, len(events))
	for _, e := range events {
		if e.Ongoing() {
			if e.Report.Current(now) {
				kept = append(kept, e)
			}
			continue
		}
		for _, o := range e.Observations {
			if !o.At.Before(cutoff) {
				kept = append(kept, e)
				break
			}
		}
	}
	return kept
}

// NewGlobe builds the use case over its collaborators.
func NewGlobe(store *Store, clock ports.Clock, providers []ports.Provider) *Globe {
	return &Globe{store: store, clock: clock, providers: providers, attempts: map[event.Provider]attempt{}, notices: map[event.Provider]string{}}
}

// SetMinimum is the minimum magnitude the globe shows from now on (FR-SET-002),
// at once and whether or not USGS can be reached; see Store.SetMinimum.
func (g *Globe) SetMinimum(minimum float64) { g.store.SetMinimum(minimum) }

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

// Outcome is what one successful fetch did, for the log (NFR-OBS-001).
type Outcome struct {
	NotModified bool
	Count       int
	Dropped     int
}

// Refresh fetches one provider, conditional on its stored validator. A failed
// fetch, a parse failure among them (FR-PRV-012), keeps the stored set and the
// cache as they were.
func (g *Globe) Refresh(ctx context.Context, p ports.Provider) (Outcome, error) {
	name := p.Name()
	dropped := g.begin(name)
	validator := ""
	if snap, ok := g.store.Snapshot(name); ok {
		validator = snap.Validator
	}
	fetched, err := p.Fetch(ctx, validator)
	if err != nil {
		g.record(name, attempt{failed: err, dropped: dropped})
		return Outcome{}, err
	}
	if !fetched.NotModified {
		dropped = fetched.Dropped
	}
	g.store.Apply(name, fetched)
	g.record(name, attempt{dropped: dropped})
	g.save(name)
	snap, _ := g.store.Snapshot(name)
	return Outcome{NotModified: fetched.NotModified, Count: len(snap.Events), Dropped: fetched.Dropped}, nil
}

// begin marks a provider's fetch as running and answers the dropped count of
// the last answer read, which stands until another answer is read.
func (g *Globe) begin(p event.Provider) int {
	g.mu.Lock()
	defer g.mu.Unlock()
	dropped := g.attempts[p].dropped
	g.attempts[p] = attempt{running: true, dropped: dropped}
	return dropped
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

// windowOf answers the window a key names; an unknown key falls back to the
// default window rather than showing nothing.
func windowOf(key string) window.Window {
	if w, ok := window.ByKey(key); ok {
		return w
	}
	return window.Default
}

// View answers what to show for a window key and filter.
func (g *Globe) View(windowKey string, f Filter) dto.View {
	w := windowOf(windowKey)
	shown := g.store.Visible(w, f)
	return g.view(w, shown, freshness.CountLine(len(shown), w.Label))
}

// ReplayView answers what a replay shows over its range (FR-RPL-009), counted
// up to its instant (FR-RPL-011).
func (g *Globe) ReplayView(w window.Window, r window.Range, f Filter) dto.View {
	shown := g.store.Within(r, f)
	return g.view(w, shown, freshness.ReplayCountLine(len(shown), r.To, w.Label))
}

func (g *Globe) view(w window.Window, shown []Shown, countLine string) dto.View {
	now := g.clock.Now()
	view := dto.View{
		WindowKey: w.Key,
		CountLine: countLine,
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
	view.Notice = g.noticeLine()
	return view
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
	if s.Event.Ongoing() {
		out.Ongoing = true
		out.Reported = freshness.Continuing(s.Event.Report)
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
	out.Trail = [][2]float64{}
	for _, p := range s.Trail {
		out.Trail = append(out.Trail, [2]float64{p.Lat, p.Lng})
	}
	if d := s.Event.Extras.DepthKm; d != nil {
		out.Depth = event.DepthWording(*d)
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

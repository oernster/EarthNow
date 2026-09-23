package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/application/services"
	"github.com/oernster/EarthNow/internal/domain/event"
	"github.com/oernster/EarthNow/internal/domain/freshness"
	"github.com/oernster/EarthNow/internal/infrastructure/geo"
)

// changedEvent tells the page to ask for a fresh view.
const changedEvent = "events-changed"

// ErrUnsafeURL refuses a link that is not an absolute https URL (FR-SEL-006).
var ErrUnsafeURL = errors.New("not an https link")

// App is the facade the page calls. It owns no rules: it forwards to the
// application and runs the background work.
type App struct {
	ctx    context.Context
	globe  *services.Globe
	sched  *services.Scheduler
	prefs  *services.Preferences
	byName map[event.Provider]ports.Provider
	wake   chan struct{}
}

// NewApp builds the facade.
func NewApp(globe *services.Globe, sched *services.Scheduler, prefs *services.Preferences, providers []ports.Provider) *App {
	byName := map[event.Provider]ports.Provider{}
	for _, p := range providers {
		byName[p.Name()] = p
	}
	return &App{globe: globe, sched: sched, prefs: prefs, byName: byName, wake: make(chan struct{}, 1)}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Printf("startup")
	go a.guard("gazetteer", a.loadGazetteer)
	go a.guard("scheduler", a.drive)
}

// domReady hands the webview the keyboard, which it cannot take for itself
// (keeb: hosted webview rule, measured in SymChit).
func (a *App) domReady(ctx context.Context) {
	if ctx != nil {
		wruntime.Show(ctx)
	}
}

// guard runs background work with a recover at its top: a panic is logged
// and told to the page instead of ending the run (NFR-REL-003).
func (a *App) guard(name string, work func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("%s panicked: %v\n%s", name, r, debug.Stack())
			wruntime.EventsEmit(a.ctx, "problem", fmt.Sprintf("%s stopped unexpectedly; details are in the log", name))
		}
	}()
	work()
}

func (a *App) loadGazetteer() {
	t0 := time.Now()
	g, err := geo.Load()
	if err != nil {
		log.Printf("gazetteer: %v", err)
		wruntime.EventsEmit(a.ctx, "problem", "Place names are unavailable: "+err.Error())
		return
	}
	a.globe.SetGeocoder(g)
	log.Printf("gazetteer loaded in %v", time.Since(t0))
}

// drive starts whatever the scheduler says is due, then sleeps until the next
// provider falls due, a fetch finishes or a manual refresh arrives. Every
// timing rule lives in the scheduler (FR-PRV-005 to 010).
func (a *App) drive() {
	for {
		for _, name := range a.sched.Due() {
			p := a.byName[name]
			go a.guard("refresh "+string(name), func() { a.fetch(p) })
		}
		var timer *time.Timer
		var fire <-chan time.Time
		if next := a.sched.NextWake(); !next.IsZero() {
			timer = time.NewTimer(max(time.Until(next), 0))
			fire = timer.C
		}
		select {
		case <-a.ctx.Done():
			return
		case <-fire:
		case <-a.wake:
		}
		if timer != nil {
			timer.Stop()
		}
	}
}

// fetch runs one provider and logs what happened (NFR-OBS-001). A panic is
// recorded as a failure before the guard reports it, so the provider is still
// retried rather than left marked as running for ever.
func (a *App) fetch(p ports.Provider) {
	name := p.Name()
	defer func() {
		if r := recover(); r != nil {
			a.sched.Failed(name)
			a.nudge()
			panic(r)
		}
	}()
	log.Printf("%s: fetch start", name)
	t0 := time.Now()
	out, err := a.globe.Refresh(a.ctx, p)
	if err != nil {
		delay := a.sched.Failed(name)
		log.Printf("%s: fetch failed after %v: %v; retrying in %v", name, time.Since(t0), err, delay)
	} else {
		a.sched.Succeeded(name)
		status := "200"
		if out.NotModified {
			status = "304 not modified"
		}
		log.Printf("%s: fetch done in %v: %s, %d events, %d dropped", name, time.Since(t0), status, out.Count, out.Dropped)
	}
	wruntime.EventsEmit(a.ctx, changedEvent)
	a.nudge()
}

// nudge wakes the driver without ever blocking on it.
func (a *App) nudge() {
	select {
	case a.wake <- struct{}{}:
	default:
	}
}

// RefreshNow is the manual refresh (FR-PRV-009, FR-PRV-010). It answers an
// empty string when started, else when a refresh becomes available.
func (a *App) RefreshNow() string {
	ok, available := a.sched.Manual()
	if !ok {
		return "Refresh available " + freshness.Until(available, time.Now())
	}
	log.Printf("manual refresh")
	a.nudge()
	return ""
}

// View answers what to show for a window and the categories and providers
// switched off.
func (a *App) View(windowKey string, hiddenCategories, hiddenProviders []string) dto.View {
	f := services.Filter{HiddenCategories: map[event.Category]bool{}, HiddenProviders: map[event.Provider]bool{}}
	for _, c := range hiddenCategories {
		f.HiddenCategories[event.Category(c)] = true
	}
	for _, p := range hiddenProviders {
		f.HiddenProviders[event.Provider(p)] = true
	}
	view := a.globe.View(windowKey, f)
	now := time.Now()
	for i, p := range view.Providers {
		if p.Problem != "" {
			view.Providers[i].NextAttempt = "next attempt " + freshness.Until(a.sched.NextAttempt(event.Provider(p.Name)), now)
		}
	}
	view.Notice = services.JoinNotices(view.Notice, a.prefs.Notice())
	return view
}

// Windows answers the time window choices.
func (a *App) Windows() []dto.Choice { return a.globe.Windows() }

// Settings answers the reader's settings as held.
func (a *App) Settings() dto.Settings { return a.prefs.Current() }

// SettingChoices answers what the settings dialog offers.
func (a *App) SettingChoices() dto.SettingChoices { return a.prefs.Choices() }

// SaveSettings keeps the page's settings and answers them as now held. A
// provider whose query changed is fetched again at once (FR-SET-002).
func (a *App) SaveSettings(chosen dto.Settings) dto.Settings {
	held, refetch := a.prefs.Update(chosen)
	for _, p := range refetch {
		log.Printf("%s: query changed; fetching again", p)
		a.sched.Expedite(p)
	}
	if len(refetch) > 0 {
		a.nudge()
	}
	return held
}

// Place answers the nearest-place line for a point (FR-GEO-001).
func (a *App) Place(lat, lng float64) string { return a.globe.Place(lat, lng) }

// OpenSource opens a source page in the system browser, never in the app
// (FR-SEL-005). Only an https link is opened.
func (a *App) OpenSource(link string) error {
	if services.SafeURL(link) == "" {
		return fmt.Errorf("%w: %q", ErrUnsafeURL, link)
	}
	wruntime.BrowserOpenURL(a.ctx, link)
	return nil
}

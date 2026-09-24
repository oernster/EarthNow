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
	"github.com/oernster/EarthNow/internal/infrastructure/window"
	"github.com/oernster/EarthNow/internal/product"
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
	help   Help
	byName map[event.Provider]ports.Provider
	wake   chan struct{}
}

// Help is what the help dialogs read (FR-HLP-001 to 003), gathered at the
// composition root: the texts are embedded there and the version is stamped there.
type Help struct {
	About   dto.About
	Licence string
	Notices string
}

// NewApp builds the facade.
func NewApp(globe *services.Globe, sched *services.Scheduler, prefs *services.Preferences, help Help, providers []ports.Provider) *App {
	byName := map[event.Provider]ports.Provider{}
	for _, p := range providers {
		byName[p.Name()] = p
	}
	return &App{globe: globe, sched: sched, prefs: prefs, help: help, byName: byName, wake: make(chan struct{}, 1)}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Printf("startup")
	go a.guard("gazetteer", a.loadGazetteer)
	go a.guard("scheduler", a.drive)
}

// domReady hands the webview the keyboard, which it cannot take for itself
// (keeb: hosted webview rule). Ported from the setup program's installer/app.go:
// raising the window alone loses a race inside Wails (see
// internal/infrastructure/window), so the WebView2 child is focused directly, as
// a click would.
func (a *App) domReady(context.Context) { a.show() }

// show gives the webview the keyboard, falling back to asking Wails for the window.
func (a *App) show() {
	focused := window.TakeFocus()
	log.Printf("keyboard: webview child focused %v", focused)
	if focused {
		return
	}
	if a.ctx != nil {
		wruntime.Show(a.ctx)
	}
}

// TakeKeyboard is called by the page when it finds it has no keyboard. The page is
// the only thing that can tell: from Go the window looks focused either way. It is
// the second half of the repair, since domReady runs before the webview is
// necessarily ready to keep what it is given.
func (a *App) TakeKeyboard() { a.show() }

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
		due := a.sched.Due()
		for _, name := range due {
			p := a.byName[name]
			go a.guard("refresh "+string(name), func() { a.fetch(p) })
		}
		// FR-STS-007: the page hears that fetches have started, not only that
		// they have finished, so it can say they are under way.
		if len(due) > 0 {
			wruntime.EventsEmit(a.ctx, changedEvent)
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

// RefreshNow is the manual refresh (FR-PRV-009, FR-PRV-010). It answers when
// the last manual refresh was made, in Unix milliseconds: this one when it
// starts, the earlier one when the cooldown refuses it.
func (a *App) RefreshNow() int64 {
	ok, last := a.sched.Manual()
	if ok {
		log.Printf("manual refresh")
		a.nudge()
	}
	return last.UnixMilli()
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
	a.sched.MarkRefreshing(view.Providers)
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

// About answers what the About dialog shows (FR-HLP-001).
func (a *App) About() dto.About { return a.help.About }

// Licence answers the full text of EarthNow's licence (FR-HLP-002).
func (a *App) Licence() string { return a.help.Licence }

// Notices answers the third-party notices (FR-HLP-003).
func (a *App) Notices() string { return a.help.Notices }

// Place answers the nearest-place line for a point (FR-GEO-001).
func (a *App) Place(lat, lng float64) string { return a.globe.Place(lat, lng) }

// OpenSource opens a source page in the system browser, never in the app
// (FR-SEL-005). Only an https link is opened.
func (a *App) OpenSource(link string) error { return a.openExternal(link) }

// Donate opens the donation page in the system browser (FR-DON-003). The page
// never holds the address, whose one home is internal/product. It goes through
// the same https allowlist as a source link. Nothing is fetched here.
func (a *App) Donate() error { return a.openExternal(product.DonateURL) }

// openExternal hands an https link to the desktop; anything else is refused.
func (a *App) openExternal(link string) error {
	if services.SafeURL(link) == "" {
		return fmt.Errorf("%w: %q", ErrUnsafeURL, link)
	}
	wruntime.BrowserOpenURL(a.ctx, link)
	return nil
}

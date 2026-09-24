package services

import (
	"context"
	"encoding/base64"
	"errors"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/burnt"
	"github.com/oernster/EarthNow/internal/domain/freshness"
	"github.com/oernster/EarthNow/internal/domain/window"
)

// BurntInterval is FR-BA-003's gap between rounds: every day of the window is
// asked for again, since a day's image is not known never to change.
const BurntInterval = time.Hour

// BurntProvider names the burnt-area service in the provider status popover
// (FR-BA-012).
const BurntProvider = "GWIS"

// dayKeyLayout names a day in the image key.
const dayKeyLayout = "2006-01-02"

// BurntAreas is the burnt-area layer's use case: it asks only while the layer
// is shown (FR-BA-005), a day at a time for every day the window overlaps
// (FR-BA-001 to 004), keeps each day's image through failures (FR-BA-012) and
// words its state (FR-BA-008, FR-BA-009, FR-BA-017). Like the scheduler it
// holds no timer: the driver asks whether a round is due and when to wake.
type BurntAreas struct {
	mu          sync.Mutex
	clock       ports.Clock
	source      ports.BurntSource
	cache       ports.BurntCache
	shown       bool
	win         window.Window
	held        map[int64]ports.BurntDay
	running     bool
	onlyMissing bool
	failures    int
	nextDue     time.Time
	problem     string
	notice      string
	composedKey string
	composed    string
}

// NewBurntAreas builds the use case over the default window, hidden until the
// saved or default setting shows it (FR-BA-011); SetWindow follows the reader's.
func NewBurntAreas(clock ports.Clock, source ports.BurntSource, cache ports.BurntCache) *BurntAreas {
	return &BurntAreas{clock: clock, source: source, cache: cache, win: window.Default, held: map[int64]ports.BurntDay{}}
}

// Restore reads the held days from the cache, so the layer draws offline at
// start (FR-BA-014). Nothing held is not a fault; a cache that cannot be read
// is one, said in the notice.
func (b *BurntAreas) Restore() {
	days, held, err := b.cache.Load()
	b.mu.Lock()
	defer b.mu.Unlock()
	switch {
	case err != nil:
		b.notice = "The saved burnt-area maps could not be read: " + err.Error()
	case held:
		for _, d := range days {
			b.held[d.Day.Unix()] = d
		}
		b.prune(b.clock.Now())
	}
}

// SetShown shows or hides the layer. Showing it makes a whole round due at once.
func (b *BurntAreas) SetShown(shown bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if shown && !b.shown {
		b.nextDue, b.onlyMissing = b.clock.Now(), false
	}
	b.shown = shown
}

// SetWindow follows the reader's time window. It answers true when days the
// new window needs are not held and a round for them is now due (FR-BA-004).
func (b *BurntAreas) SetWindow(key string) bool {
	w, ok := window.ByKey(key)
	b.mu.Lock()
	defer b.mu.Unlock()
	if !ok || w == b.win {
		return false
	}
	b.win = w
	if !b.shown || b.running || len(b.missing(b.clock.Now())) == 0 {
		return false
	}
	b.onlyMissing, b.nextDue = true, b.clock.Now()
	return true
}

// Due answers whether a round should start now, marking it running when one
// should. It never answers true while the layer is hidden (FR-BA-005).
func (b *BurntAreas) Due() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.shown || b.running || b.clock.Now().Before(b.nextDue) {
		return false
	}
	b.running = true
	return true
}

// NextWake answers when the next round falls due; zero while hidden or while
// a round is running.
func (b *BurntAreas) NextWake() time.Time {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.shown || b.running {
		return time.Time{}
	}
	return b.nextDue
}

// Refresh runs one round that Due started and answers how many days arrived.
// A day that fails keeps its held image while the others are kept as they
// arrive; the round is then retried by FR-PRV-006's backoff (FR-BA-012).
func (b *BurntAreas) Refresh(ctx context.Context) (int, error) {
	b.mu.Lock()
	days := b.roundDays(b.clock.Now())
	b.mu.Unlock()
	var got []ports.BurntDay
	var failed []string
	for _, d := range days {
		img, anything, err := b.source.Day(ctx, d)
		if err != nil {
			failed = append(failed, burnt.Span(d, d)+": "+err.Error())
			continue
		}
		got = append(got, ports.BurntDay{Day: d, PNG: img, Drawn: anything, RetrievedAt: b.clock.Now()})
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.running = false
	now := b.clock.Now()
	for _, g := range got {
		b.held[g.Day.Unix()] = g
	}
	b.prune(now)
	if len(got) > 0 {
		b.save()
	}
	if len(failed) > 0 {
		b.failures++
		b.nextDue = now.Add(Backoff(BurntInterval, b.failures))
		b.problem = strings.Join(failed, noticeSeparator)
		return len(got), errors.New(b.problem)
	}
	b.failures, b.problem = 0, ""
	b.nextDue = now.Add(BurntInterval)
	// The window may have changed while this round ran (FR-BA-004).
	if len(b.missing(now)) > 0 {
		b.onlyMissing, b.nextDue = true, now
	}
	return len(got), nil
}

// roundDays answers the days this round asks for: those the window lacks
// after a change of window, else every day of it. The caller holds the lock.
func (b *BurntAreas) roundDays(now time.Time) []time.Time {
	if b.onlyMissing {
		b.onlyMissing = false
		return b.missing(now)
	}
	return b.win.Days(now)
}

// missing answers the window's days not held. The caller holds the lock.
func (b *BurntAreas) missing(now time.Time) []time.Time {
	var out []time.Time
	for _, d := range b.win.Days(now) {
		if _, ok := b.held[d.Unix()]; !ok {
			out = append(out, d)
		}
	}
	return out
}

// prune discards days older than the widest window reaches (FR-BA-014). The
// caller holds the lock.
func (b *BurntAreas) prune(now time.Time) {
	oldest := window.Widest.Days(now)[0]
	for k, d := range b.held {
		if d.Day.Before(oldest) {
			delete(b.held, k)
		}
	}
}

// save keeps every held day, oldest first. The caller holds the lock.
func (b *BurntAreas) save() {
	days := make([]ports.BurntDay, 0, len(b.held))
	for _, d := range b.held {
		days = append(days, d)
	}
	slices.SortFunc(days, func(x, y ports.BurntDay) int { return x.Day.Compare(y.Day) })
	if err := b.cache.Save(days); err != nil {
		b.notice = "The burnt-area maps could not be saved: " + err.Error()
	} else {
		b.notice = ""
	}
}

// inWindow answers the held days of the window, oldest first, with whether
// every day of it is held. The caller holds the lock.
func (b *BurntAreas) inWindow(now time.Time) ([]ports.BurntDay, bool) {
	var out []ports.BurntDay
	all := true
	for _, d := range b.win.Days(now) {
		if h, ok := b.held[d.Unix()]; ok {
			out = append(out, h)
		} else {
			all = false
		}
	}
	return out, all
}

// drawn answers the days that drew anything.
func drawn(days []ports.BurntDay) []ports.BurntDay {
	var out []ports.BurntDay
	for _, d := range days {
		if d.Drawn {
			out = append(out, d)
		}
	}
	return out
}

// imageKey names the days drawn and when each arrived, so a new key means a
// new image.
func imageKey(days []ports.BurntDay) string {
	parts := make([]string, len(days))
	for i, d := range days {
		parts[i] = d.Day.UTC().Format(dayKeyLayout) + "@" + d.RetrievedAt.UTC().Format(time.RFC3339Nano)
	}
	return strings.Join(parts, ",")
}

// Status answers the layer's state for the page; the line is empty while the
// layer is hidden.
func (b *BurntAreas) Status() dto.BurntAreas {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := dto.BurntAreas{Shown: b.shown, Notice: b.notice}
	if !b.shown {
		return out
	}
	now := b.clock.Now()
	held, all := b.inWindow(now)
	shown := drawn(held)
	out.Key = imageKey(shown)
	switch {
	case len(shown) > 0:
		out.Line = burnt.Status(shown[0].Day, shown[len(shown)-1].Day, oldest(held), now)
	case all:
		out.Line = burnt.NoneYet
	case b.problem != "":
		out.Line = burnt.Unavailable
	default:
		out.Line = burnt.Retrieving
	}
	out.Provider = b.provider(now, held)
	return out
}

// oldest answers the earliest arrival among days, which the line's age is
// worded from, so it never claims more freshness than the stalest day has.
func oldest(days []ports.BurntDay) time.Time {
	at := days[0].RetrievedAt
	for _, d := range days[1:] {
		if d.RetrievedAt.Before(at) {
			at = d.RetrievedAt
		}
	}
	return at
}

// provider is GWIS's entry for the provider status popover (FR-BA-012). The
// caller holds the lock.
func (b *BurntAreas) provider(now time.Time, held []ports.BurntDay) dto.Provider {
	out := dto.Provider{Name: BurntProvider, Problem: b.problem}
	switch {
	case len(held) > 0:
		out.Refreshing = b.running
		newest := held[0].RetrievedAt
		for _, d := range held[1:] {
			newest = latest(newest, d.RetrievedAt)
		}
		out.Retrieved = freshness.Retrieved(newest, now)
	case b.problem == "":
		out.Loading = true
	}
	if b.problem != "" {
		out.NextAttempt = "next attempt " + freshness.Until(b.nextDue, now)
	}
	return out
}

// latest answers the later of two instants.
func latest(x, y time.Time) time.Time {
	if y.After(x) {
		return y
	}
	return x
}

// Image answers the window's drawn days as one image, a data URL the page can
// draw; empty when none draws. The last one composed is kept until its key
// changes, since composing decodes every day.
func (b *BurntAreas) Image() string {
	b.mu.Lock()
	held, _ := b.inWindow(b.clock.Now())
	shown := drawn(held)
	key := imageKey(shown)
	if key == b.composedKey {
		defer b.mu.Unlock()
		return b.composed
	}
	b.mu.Unlock()
	images := make([][]byte, len(shown))
	for i, d := range shown {
		images[i] = d.PNG
	}
	url := ""
	if len(images) > 0 {
		img, err := b.source.Compose(images)
		if err != nil {
			b.mu.Lock()
			defer b.mu.Unlock()
			b.notice = "The burnt-area maps could not be drawn: " + err.Error()
			return ""
		}
		url = pngDataURL + base64.StdEncoding.EncodeToString(img)
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.composedKey, b.composed = key, url
	return url
}

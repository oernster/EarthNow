package services

import (
	"context"
	"encoding/base64"
	"sync"
	"time"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/cloud"
	"github.com/oernster/EarthNow/internal/domain/freshness"
)

// CloudInterval is FR-CLD-004's gap between checks for a newer image.
const CloudInterval = time.Hour

// CloudProvider names the cloud service in the provider status popover
// (FR-CLD-011).
const CloudProvider = "EUMETSAT"

// cloudLoading is the status line while the first image is on its way.
const cloudLoading = "Clouds: retrieving the image"

// pngDataURL prefixes the image handed to the page, which draws it without
// reaching the network (NFR-SEC-002).
const pngDataURL = "data:image/png;base64,"

// Clouds is the cloud layer's use case: it asks only while the layer is shown
// (FR-CLD-005), fetches an image only when a newer one is listed (FR-CLD-016),
// keeps the held image through failures (FR-CLD-011) and words its status
// (FR-CLD-009, FR-CLD-013). Like the scheduler it holds no timer: the driver
// asks whether a check is due and when to wake.
type Clouds struct {
	mu          sync.Mutex
	clock       ports.Clock
	source      ports.CloudSource
	cache       ports.CloudCache
	shown       bool
	held        ports.CloudImage
	retrievedAt time.Time
	running     bool
	failures    int
	nextDue     time.Time
	problem     string
	notice      string
}

// NewClouds builds the use case, hidden until told otherwise (FR-CLD-003).
func NewClouds(clock ports.Clock, source ports.CloudSource, cache ports.CloudCache) *Clouds {
	return &Clouds{clock: clock, source: source, cache: cache}
}

// Restore reads the held image from the cache, so the layer draws offline at
// start (FR-CLD-014). Nothing held is not a fault; a cache that cannot be read
// is one, said in the notice.
func (c *Clouds) Restore() {
	img, held, err := c.cache.Load()
	c.mu.Lock()
	defer c.mu.Unlock()
	switch {
	case err != nil:
		c.notice = "The saved cloud image could not be read: " + err.Error()
	case held:
		c.held = img
	}
}

// SetShown shows or hides the layer. Showing it makes a check due at once.
func (c *Clouds) SetShown(shown bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if shown && !c.shown {
		c.nextDue = c.clock.Now()
	}
	c.shown = shown
}

// Due answers whether a check should start now. When one should, it marks it
// running, so a check is never started twice. It never answers true while the layer is
// hidden (FR-CLD-005).
func (c *Clouds) Due() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.shown || c.running || c.clock.Now().Before(c.nextDue) {
		return false
	}
	c.running = true
	return true
}

// NextWake answers when the next check falls due; zero while hidden or while
// a check is running.
func (c *Clouds) NextWake() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.shown || c.running {
		return time.Time{}
	}
	return c.nextDue
}

// Refresh runs one check that Due started. It answers whether a new image is
// now held; a failure keeps the held image and schedules the retry by
// FR-PRV-006's backoff (FR-CLD-011).
func (c *Clouds) Refresh(ctx context.Context) (bool, error) {
	fresh, err := c.check(ctx)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.running = false
	now := c.clock.Now()
	if err != nil {
		c.failures++
		c.nextDue = now.Add(Backoff(CloudInterval, c.failures))
		c.problem = err.Error()
		return false, err
	}
	c.failures, c.problem, c.retrievedAt = 0, "", now
	c.nextDue = now.Add(CloudInterval)
	if fresh == nil {
		return false, nil
	}
	c.held = *fresh
	if err := c.cache.Save(*fresh); err != nil {
		c.notice = "The cloud image could not be saved: " + err.Error()
	} else {
		c.notice = ""
	}
	return true, nil
}

// check asks for the newest valid time and fetches its image when it is not
// the one held; it answers nil when nothing new is listed.
func (c *Clouds) check(ctx context.Context) (*ports.CloudImage, error) {
	latest, err := c.source.Latest(ctx)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	unchanged := c.held.ValidTime.Equal(latest)
	c.mu.Unlock()
	if unchanged {
		return nil, nil
	}
	png, err := c.source.Image(ctx, latest)
	if err != nil {
		return nil, err
	}
	return &ports.CloudImage{ValidTime: latest, PNG: png}, nil
}

// Status answers the cloud layer's state for the page. The line is empty
// while the layer is hidden.
func (c *Clouds) Status() dto.Clouds {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := dto.Clouds{Shown: c.shown, Notice: c.notice}
	if !c.held.ValidTime.IsZero() {
		out.ValidTime = c.held.ValidTime.UTC().Format(time.RFC3339)
	}
	if !c.shown {
		return out
	}
	now := c.clock.Now()
	out.Provider = c.provider(now)
	switch {
	case out.ValidTime != "":
		out.Line = cloud.Status(c.held.ValidTime, now)
	case c.problem != "":
		out.Line = cloud.Unavailable
	default:
		out.Line = cloudLoading
	}
	return out
}

// provider is the cloud service's entry for the provider status popover
// (FR-CLD-011). The caller holds the lock.
func (c *Clouds) provider(now time.Time) dto.Provider {
	out := dto.Provider{Name: CloudProvider, Problem: c.problem}
	held := !c.held.ValidTime.IsZero()
	switch {
	case held:
		out.Refreshing = c.running
		out.Stale = cloud.Stale(c.held.ValidTime, now)
	case c.problem == "":
		out.Loading = true
	}
	if !c.retrievedAt.IsZero() {
		out.Retrieved = freshness.Retrieved(c.retrievedAt, now)
	}
	if c.problem != "" {
		out.NextAttempt = "next attempt " + freshness.Until(c.nextDue, now)
	}
	return out
}

// Image answers the held image as a data URL the page can draw; empty when
// none is held.
func (c *Clouds) Image() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.held.PNG) == 0 {
		return ""
	}
	return pngDataURL + base64.StdEncoding.EncodeToString(c.held.PNG)
}

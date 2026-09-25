package services

import (
	"context"
	"encoding/base64"
	"slices"
	"sync"
	"time"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/cloud"
	"github.com/oernster/EarthNow/internal/domain/freshness"
)

// ReplayCloudProvider names the replay's cloud images in the provider status
// popover (FR-RPL-017), apart from the ordinary cloud image's entry.
const ReplayCloudProvider = CloudProvider + " (replay)"

// ReplayClouds holds a replay's cloud images (FR-RPL-014 to 018): for a span it
// lists the valid times up to the newest the service has, then fetches one
// image a round at the replay's size, in memory only. Like the others it holds
// no timer: the driver asks whether a round is due.
type ReplayClouds struct {
	mu       sync.Mutex
	clock    ports.Clock
	source   ports.CloudSource
	from, to time.Time
	active   bool
	listed   bool
	running  bool
	span     int
	pending  []time.Time
	held     map[int64][]byte
	times    []time.Time
	missing  []time.Time
	total    int
	failures int
	nextDue  time.Time
	problem  string
}

// NewReplayClouds builds the use case, holding nothing until a span is followed.
func NewReplayClouds(clock ports.Clock, source ports.CloudSource) *ReplayClouds {
	return &ReplayClouds{clock: clock, source: source, held: map[int64][]byte{}}
}

// Follow makes from to to the span whose images are held. A new span drops the
// last one's images and answers true: a round is due at once (FR-RPL-015).
func (r *ReplayClouds) Follow(from, to time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.active && r.from.Equal(from) && r.to.Equal(to) {
		return false
	}
	r.release()
	r.from, r.to, r.active = from, to, true
	r.nextDue = r.clock.Now()
	return true
}

// Stop releases every replay image (FR-RPL-018).
func (r *ReplayClouds) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.release()
}

// release forgets the span and its images. The caller holds the lock.
func (r *ReplayClouds) release() {
	r.active, r.listed = false, false
	r.span++
	r.pending, r.times, r.missing = nil, nil, nil
	r.held = map[int64][]byte{}
	r.total, r.failures, r.problem = 0, 0, ""
}

// Due answers whether a round should start now, marking it running.
func (r *ReplayClouds) Due() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.active || r.running || (r.listed && len(r.pending) == 0) || r.clock.Now().Before(r.nextDue) {
		return false
	}
	r.running = true
	return true
}

// NextWake answers when the next round falls due; zero when none will.
func (r *ReplayClouds) NextWake() time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.active || r.running || (r.listed && len(r.pending) == 0) {
		return time.Time{}
	}
	return r.nextDue
}

// Refresh runs one round that Due started: the listing first, then one image.
// A listing that fails is retried by FR-PRV-006's backoff; an image that fails
// is recorded as missing and the next one follows at once (FR-RPL-017).
func (r *ReplayClouds) Refresh(ctx context.Context) error {
	r.mu.Lock()
	span, listed, from, to := r.span, r.listed, r.from, r.to
	var next time.Time
	if listed {
		next = r.pending[0]
	}
	r.mu.Unlock()
	if !listed {
		return r.list(ctx, span, from, to)
	}
	img, err := r.source.ReplayImage(ctx, next)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.running = false
	if span != r.span {
		return nil
	}
	r.pending = r.pending[1:]
	if err != nil {
		r.missing = append(r.missing, next)
		return err
	}
	r.held[next.Unix()] = img
	r.times = append(r.times, next)
	return nil
}

// list reads the newest valid time and lists the span's times up to it.
func (r *ReplayClouds) list(ctx context.Context, span int, from, to time.Time) error {
	newest, err := r.source.Latest(ctx)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.running = false
	if span != r.span {
		return nil
	}
	if err != nil {
		r.failures++
		r.nextDue = r.clock.Now().Add(Backoff(CloudInterval, r.failures))
		r.problem = err.Error()
		return err
	}
	if newest.Before(to) {
		to = newest
	}
	r.pending = cloud.ValidTimes(from, to)
	r.total, r.listed, r.failures, r.problem = len(r.pending), true, 0, ""
	return nil
}

// Pick answers the valid time of the image to draw at an instant (FR-RPL-014),
// RFC 3339; empty when none is held at or before it.
func (r *ReplayClouds) Pick(at time.Time) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := cloud.AtOrBefore(r.times, at)
	if !ok {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// Image answers a held replay image as a data URL; empty when not held.
func (r *ReplayClouds) Image(validTime string) string {
	at, err := time.Parse(time.RFC3339, validTime)
	if err != nil {
		return ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	img, ok := r.held[at.Unix()]
	if !ok {
		return ""
	}
	return pngDataURL + base64.StdEncoding.EncodeToString(img)
}

// Status answers the replay's cloud line (FR-RPL-016), empty once every image
// is in or when no span is followed, with its popover entry (FR-RPL-017).
func (r *ReplayClouds) Status() (string, dto.Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := dto.Provider{Name: ReplayCloudProvider, Problem: r.problem}
	if !r.active {
		return "", out
	}
	now := r.clock.Now()
	line := ""
	switch {
	case !r.listed:
		line, out.Loading = cloud.ReplayListing, r.problem == ""
	case len(r.pending) > 0:
		line, out.Refreshing = cloud.ReplayProgress(len(r.times), r.total), true
	}
	if len(r.missing) > 0 {
		out.Problem = cloud.ReplayMissing(slices.Clone(r.missing))
	}
	if r.problem != "" {
		out.NextAttempt = "next attempt " + freshness.Until(r.nextDue, now)
	}
	return line, out
}

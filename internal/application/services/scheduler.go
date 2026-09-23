package services

import (
	"sync"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
)

// BackoffCeiling is FR-PRV-006's longest wait between retries.
const BackoffCeiling = 30 * time.Minute

// ManualCooldown is FR-PRV-010's shortest gap between manual refreshes.
const ManualCooldown = 30 * time.Second

type schedule struct {
	interval time.Duration
	failures int
	nextDue  time.Time
	running  bool
}

// Scheduler decides when each provider is next fetched. It holds no timer:
// the driver asks what is due and when to wake, so the rules are testable on
// a fake clock and no business logic sleeps (Plan 11).
type Scheduler struct {
	mu         sync.Mutex
	clock      ports.Clock
	order      []event.Provider
	states     map[event.Provider]*schedule
	lastManual time.Time
}

// NewScheduler makes every provider due at once.
func NewScheduler(clock ports.Clock, providers []ports.Provider) *Scheduler {
	s := &Scheduler{clock: clock, states: map[event.Provider]*schedule{}}
	now := clock.Now()
	for _, p := range providers {
		s.order = append(s.order, p.Name())
		s.states[p.Name()] = &schedule{interval: p.Interval(), nextDue: now}
	}
	return s
}

// Due answers the providers whose time has come and marks them running, so a
// provider still fetching is never started twice.
func (s *Scheduler) Due() []event.Provider {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock.Now()
	var due []event.Provider
	for _, p := range s.order {
		st := s.states[p]
		if !st.running && !now.Before(st.nextDue) {
			st.running = true
			due = append(due, p)
		}
	}
	return due
}

// NextWake answers the earliest time any idle provider falls due; zero when
// every provider is running.
func (s *Scheduler) NextWake() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	var wake time.Time
	for _, p := range s.order {
		st := s.states[p]
		if !st.running && (wake.IsZero() || st.nextDue.Before(wake)) {
			wake = st.nextDue
		}
	}
	return wake
}

// Succeeded restores the provider's normal interval (FR-PRV-007).
func (s *Scheduler) Succeeded(p event.Provider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.states[p]
	st.running, st.failures = false, 0
	st.nextDue = s.clock.Now().Add(st.interval)
}

// Failed schedules the retry: the delay doubles from the interval with each
// consecutive failure, capped at BackoffCeiling (FR-PRV-006). It answers the
// delay chosen.
func (s *Scheduler) Failed(p event.Provider) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.states[p]
	st.running = false
	st.failures++
	delay := st.interval
	for i := 1; i < st.failures && delay < BackoffCeiling; i++ {
		delay *= 2
	}
	if delay > BackoffCeiling {
		delay = BackoffCeiling
	}
	st.nextDue = s.clock.Now().Add(delay)
	return delay
}

// NextAttempt answers when a provider is next fetched (FR-STS-003).
func (s *Scheduler) NextAttempt(p event.Provider) time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.states[p]; ok {
		return st.nextDue
	}
	return time.Time{}
}

// Manual makes every idle provider due now (FR-PRV-009), unless the last
// manual refresh was within ManualCooldown; then it answers false and when a
// refresh becomes available (FR-PRV-010).
func (s *Scheduler) Manual() (bool, time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock.Now()
	if available := s.lastManual.Add(ManualCooldown); !s.lastManual.IsZero() && now.Before(available) {
		return false, available
	}
	s.lastManual = now
	for _, st := range s.states {
		if !st.running {
			st.nextDue = now
		}
	}
	return true, now
}

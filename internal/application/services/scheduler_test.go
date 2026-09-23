package services

import (
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
)

type intervalProvider struct {
	fakeProvider
	every time.Duration
}

func (p *intervalProvider) Interval() time.Duration { return p.every }

func schedulerFor(clock *fakeClock) *Scheduler {
	return NewScheduler(clock, []ports.Provider{
		&intervalProvider{fakeProvider{name: event.USGS}, time.Minute},
		&intervalProvider{fakeProvider{name: event.EONET}, 10 * time.Minute},
	})
}

func names(ps []event.Provider) string {
	s := ""
	for _, p := range ps {
		s += string(p) + " "
	}
	return s
}

func TestFRPRV005_EachProviderOnItsOwnInterval(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	s := schedulerFor(clock)
	if got := names(s.Due()); got != "USGS EONET " {
		t.Fatalf("at start due = %q", got)
	}
	if got := s.Due(); len(got) != 0 {
		t.Errorf("running providers were due again: %v", got)
	}
	if !s.NextWake().IsZero() {
		t.Error("NextWake with everything running should be zero")
	}
	s.Succeeded(event.USGS)
	s.Succeeded(event.EONET)
	if w := s.NextWake(); !w.Equal(noon.Add(time.Minute)) {
		t.Errorf("NextWake = %v", w)
	}
	clock.now = noon.Add(time.Minute)
	if got := names(s.Due()); got != "USGS " {
		t.Errorf("after 60 s due = %q, want exactly one USGS fetch", got)
	}
	if got := s.NextAttempt(event.EONET); !got.Equal(noon.Add(10 * time.Minute)) {
		t.Errorf("EONET next attempt = %v", got)
	}
	if !s.NextAttempt(event.Provider("NONE")).IsZero() {
		t.Error("an unknown provider has a next attempt")
	}
}

func TestFRPRV006_FRPRV007_BackoffDoublesToTheCeilingThenResets(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	s := schedulerFor(clock)
	s.Due()
	want := []time.Duration{1, 2, 4, 8, 16, 30, 30}
	for i, minutes := range want {
		if got := s.Failed(event.USGS); got != minutes*time.Minute {
			t.Errorf("failure %d delay = %v, want %v", i+1, got, minutes*time.Minute)
		}
	}
	s.Succeeded(event.USGS)
	if got := s.NextAttempt(event.USGS); !got.Equal(noon.Add(time.Minute)) {
		t.Errorf("after success next = %v, want the normal interval", got)
	}
	if got := s.Failed(event.USGS); got != time.Minute {
		t.Errorf("first failure after success = %v", got)
	}
}

func TestFRPRV009_FRPRV010_ManualRefreshWithCooldown(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	s := schedulerFor(clock)
	s.Due()
	s.Succeeded(event.USGS)
	s.Succeeded(event.EONET)
	clock.now = noon.Add(5 * time.Second)
	if ok, _ := s.Manual(); !ok {
		t.Fatal("first manual refresh refused")
	}
	if got := names(s.Due()); got != "USGS EONET " {
		t.Errorf("after manual due = %q", got)
	}
	clock.now = noon.Add(20 * time.Second)
	ok, available := s.Manual()
	if ok || !available.Equal(noon.Add(35*time.Second)) {
		t.Errorf("manual within cooldown = %v, available %v", ok, available)
	}
	s.Succeeded(event.USGS)
	clock.now = noon.Add(36 * time.Second)
	if ok, _ := s.Manual(); !ok {
		t.Error("manual after cooldown refused")
	}
	if got := names(s.Due()); got != "USGS " {
		t.Errorf("a still-running EONET was made due: %q", got)
	}
}

func TestFRSET002_ExpediteFetchesTheChangedProviderNow(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{noon}
	s := schedulerFor(clock)
	s.Due()
	s.Succeeded(event.USGS)
	s.Succeeded(event.EONET)
	clock.now = noon.Add(10 * time.Second)
	s.Expedite(event.USGS)
	s.Expedite(event.Provider("NONE"))
	if got := names(s.Due()); got != "USGS " {
		t.Fatalf("idle expedite due = %q", got)
	}
	// Mid-fetch: the running fetch asks the old question, so the provider is
	// due again as soon as it ends, success or failure.
	s.Expedite(event.USGS)
	s.Succeeded(event.USGS)
	if got := s.NextAttempt(event.USGS); !got.Equal(clock.now) {
		t.Errorf("after success next = %v, want now", got)
	}
	s.Due()
	s.Expedite(event.USGS)
	if got := s.Failed(event.USGS); got != 0 || !s.NextAttempt(event.USGS).Equal(clock.now) {
		t.Errorf("after failure delay = %v, next %v", got, s.NextAttempt(event.USGS))
	}
	s.Due()
	s.Succeeded(event.USGS)
	if got := s.NextAttempt(event.USGS); !got.Equal(clock.now.Add(time.Minute)) {
		t.Errorf("an honoured expedite lingered: next %v", got)
	}
}

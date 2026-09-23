package services

import (
	"errors"
	"math"
	"testing"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
)

type fakeSettings struct {
	held    *ports.Settings
	loadErr error
	saveErr error
	saved   []ports.Settings
}

func (f *fakeSettings) Load(defaults ports.Settings) (ports.Settings, bool, error) {
	if f.loadErr != nil {
		return ports.Settings{}, false, f.loadErr
	}
	if f.held == nil {
		return defaults, false, nil
	}
	return *f.held, true, nil
}

func (f *fakeSettings) Save(s ports.Settings) error {
	f.saved = append(f.saved, s)
	return f.saveErr
}

func (f *fakeSettings) Path() string { return `C:\data\settings.json` }

// minimums records every minimum handed to the USGS adapter.
type minimums []float64

func (m *minimums) apply(v float64) { *m = append(*m, v) }

func TestFRSET004_MissingFileStartsOnDefaultsAndSaysSo(t *testing.T) {
	t.Parallel()
	var applied minimums
	p := NewPreferences(&fakeSettings{}, applied.apply)
	p.Load()
	got := p.Current()
	if !got.AutoRotate || got.Magnitude != DefaultMagnitude || got.Speed != DefaultSpeed || got.WindowKey != "24h" {
		t.Errorf("defaults = %+v", got)
	}
	if got.HiddenCategories == nil || got.HiddenProviders == nil {
		t.Error("an empty list crossed the wire as null")
	}
	if p.Notice() != `No settings file at C:\data\settings.json yet; using defaults` {
		t.Errorf("notice = %q", p.Notice())
	}
	if len(applied) != 1 || applied[0] != 3.0 {
		t.Errorf("minimums applied = %v", applied)
	}
}

func TestFRSET004_UnreadableFileStartsOnDefaultsAndNamesIt(t *testing.T) {
	t.Parallel()
	var applied minimums
	p := NewPreferences(&fakeSettings{loadErr: errors.New("damaged")}, applied.apply)
	p.Load()
	if p.Current().Magnitude != DefaultMagnitude {
		t.Errorf("settings = %+v", p.Current())
	}
	if p.Notice() != `Settings could not be read from C:\data\settings.json: damaged; using defaults` {
		t.Errorf("notice = %q", p.Notice())
	}
}

func TestStoredSettingsAreRestoredAndUnknownKeysDefaulted(t *testing.T) {
	t.Parallel()
	var applied minimums
	held := ports.Settings{AutoRotate: false, Magnitude: "all", Speed: "warp", Window: "1y", HiddenCategories: []string{"ICE"}}
	p := NewPreferences(&fakeSettings{held: &held}, applied.apply)
	p.Load()
	got := p.Current()
	if got.AutoRotate || got.Magnitude != "all" || got.Speed != DefaultSpeed || got.WindowKey != "24h" || len(got.HiddenCategories) != 1 {
		t.Errorf("restored = %+v", got)
	}
	if p.Notice() != "" || len(applied) != 1 || !math.IsInf(applied[0], -1) {
		t.Errorf("notice %q, minimums %v", p.Notice(), applied)
	}
}

func TestFRSET002_ChangingTheMinimumRefetchesUSGS(t *testing.T) {
	t.Parallel()
	var applied minimums
	store := &fakeSettings{}
	p := NewPreferences(store, applied.apply)
	p.Load()
	chosen := p.Current()
	chosen.Magnitude = "4.5"
	got, refetch := p.Update(chosen)
	if got.Magnitude != "4.5" || len(refetch) != 1 || refetch[0] != event.USGS {
		t.Errorf("update = %+v, refetch %v", got, refetch)
	}
	if applied[len(applied)-1] != 4.5 || len(store.saved) != 1 || p.Notice() != "" {
		t.Errorf("minimums %v, saved %d, notice %q", applied, len(store.saved), p.Notice())
	}
}

func TestFRGLB011_FRFLT005_OtherChangesSaveWithoutARefetch(t *testing.T) {
	t.Parallel()
	var applied minimums
	store := &fakeSettings{}
	p := NewPreferences(store, applied.apply)
	p.Load()
	got, refetch := p.Update(dto.Settings{AutoRotate: false, Magnitude: "", Speed: "fast", WindowKey: "7d", HiddenProviders: []string{"EONET"}})
	if got.AutoRotate || got.Magnitude != DefaultMagnitude || got.Speed != "fast" || got.WindowKey != "7d" || got.HiddenProviders[0] != "EONET" || got.HiddenCategories == nil {
		t.Errorf("update = %+v", got)
	}
	if refetch != nil || len(applied) != 1 || store.saved[0].Window != "7d" {
		t.Errorf("refetch %v, minimums %v, saved %+v", refetch, applied, store.saved)
	}
}

func TestAFailedSaveKeepsTheChangeAndSaysSo(t *testing.T) {
	t.Parallel()
	var applied minimums
	p := NewPreferences(&fakeSettings{saveErr: errors.New("read-only")}, applied.apply)
	p.Load()
	got, _ := p.Update(dto.Settings{Magnitude: "1.0"})
	if got.Magnitude != "1.0" || p.Current().Magnitude != "1.0" {
		t.Errorf("update = %+v", got)
	}
	if p.Notice() != `Settings could not be saved to C:\data\settings.json: read-only` {
		t.Errorf("notice = %q", p.Notice())
	}
}

func TestChoicesOfferEveryMagnitudeAndSpeed(t *testing.T) {
	t.Parallel()
	c := NewPreferences(&fakeSettings{}, func(float64) {}).Choices()
	if len(c.Magnitudes) != len(Magnitudes) || c.Magnitudes[0].Key != "all" || len(c.Speeds) != len(Speeds) {
		t.Errorf("choices = %+v", c)
	}
	if c.Speeds[1].SecondsPerRevolution != 240 {
		t.Errorf("normal speed = %+v, FR-GLB-002 says 240 s", c.Speeds[1])
	}
}

func TestJoinNoticesSkipsTheEmpty(t *testing.T) {
	t.Parallel()
	if got := JoinNotices("", "a", "", "b"); got != "a; b" {
		t.Errorf("JoinNotices = %q", got)
	}
	if got := JoinNotices("", ""); got != "" {
		t.Errorf("JoinNotices of nothing = %q", got)
	}
}

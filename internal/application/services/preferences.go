package services

import (
	"math"
	"slices"
	"strings"
	"sync"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
	"github.com/oernster/EarthNow/internal/domain/window"
)

// Magnitude is one USGS minimum on offer (FR-SET-002).
type Magnitude struct {
	Key     string
	Label   string
	Minimum float64
}

// Magnitudes are FR-SET-002's choices, lowest first. "All" keeps every event,
// those with no magnitude included.
var Magnitudes = []Magnitude{
	{Key: "all", Label: "All", Minimum: math.Inf(-1)},
	{Key: "1.0", Label: "1.0 and above", Minimum: 1.0},
	{Key: "2.5", Label: "2.5 and above", Minimum: 2.5},
	{Key: "3.0", Label: "3.0 and above", Minimum: 3.0},
	{Key: "4.5", Label: "4.5 and above", Minimum: 4.5},
}

// DefaultMagnitude is the owner's default of 2.5 (OQ-003, amended 2026-09-23):
// 362 earthquakes that week against 243 at 3.0.
const DefaultMagnitude = "2.5"

// Speeds are FR-SET-003's idle rotation presets. Normal is FR-GLB-002's one
// revolution per 240 s; the other two halve and double it.
var Speeds = []dto.Speed{
	{Key: "slow", Label: "Slow", SecondsPerRevolution: 480},
	{Key: "normal", Label: "Normal", SecondsPerRevolution: 240},
	{Key: "fast", Label: "Fast", SecondsPerRevolution: 120},
}

// DefaultSpeed is FR-GLB-002's idle speed.
const DefaultSpeed = "normal"

// Preferences holds the reader's settings: loads them at start, answers them
// to the page, saves every change and tells the USGS adapter its minimum.
type Preferences struct {
	mu           sync.Mutex
	store        ports.SettingsStore
	applyMinimum func(float64)
	current      ports.Settings
	notice       string
}

// NewPreferences builds the use case over its store. applyMinimum is handed
// the USGS minimum at load and on every change of it.
func NewPreferences(store ports.SettingsStore, applyMinimum func(float64)) *Preferences {
	return &Preferences{store: store, applyMinimum: applyMinimum, current: Defaults()}
}

// Defaults are the settings of a first run: rotating at the normal speed,
// minimum 2.5, the 24 h window, nothing filtered out and the day and night
// layer shown (FR-DAY-007).
func Defaults() ports.Settings {
	return ports.Settings{
		AutoRotate:    true,
		Magnitude:     DefaultMagnitude,
		Speed:         DefaultSpeed,
		Window:        window.Default.Key,
		DayNightShown: true,
	}
}

// Load reads the stored settings. A missing file and an unreadable one both
// start the run on defaults; each is said in the notice, worded apart, since
// only the second is a fault (FR-SET-004).
func (p *Preferences) Load() {
	loaded, held, err := p.store.Load(Defaults())
	p.mu.Lock()
	switch {
	case err != nil:
		p.current = Defaults()
		p.notice = "Settings could not be read from " + p.store.Path() + ": " + err.Error() + "; using defaults"
	case !held:
		p.current = Defaults()
		p.notice = "No settings file at " + p.store.Path() + " yet; using defaults"
	default:
		p.current = normalised(loaded)
	}
	minimum := minimumOf(p.current.Magnitude)
	p.mu.Unlock()
	p.applyMinimum(minimum)
}

// Notice is the settings problem the reader should know about; empty when none.
func (p *Preferences) Notice() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.notice
}

// Current answers the settings as the page reads them.
func (p *Preferences) Current() dto.Settings {
	p.mu.Lock()
	defer p.mu.Unlock()
	return toSettingsDTO(p.current)
}

// Choices answers what the settings dialog offers.
func (p *Preferences) Choices() dto.SettingChoices {
	out := dto.SettingChoices{Speeds: slices.Clone(Speeds)}
	for _, m := range Magnitudes {
		out.Magnitudes = append(out.Magnitudes, dto.Choice{Key: m.Key, Label: m.Label})
	}
	return out
}

// Update takes the page's settings, keeps what is valid and saves them. It
// answers the settings now held and the providers to fetch again because their
// query changed. A save that fails keeps the change for this run and says so.
func (p *Preferences) Update(chosen dto.Settings) (dto.Settings, []event.Provider) {
	next := normalised(ports.Settings{
		AutoRotate:       chosen.AutoRotate,
		Magnitude:        chosen.Magnitude,
		Speed:            chosen.Speed,
		Window:           chosen.WindowKey,
		HiddenCategories: chosen.HiddenCategories,
		HiddenProviders:  chosen.HiddenProviders,
		CloudsShown:      chosen.CloudsShown,
		DayNightShown:    chosen.DayNightShown,
	})
	err := p.store.Save(next)
	p.mu.Lock()
	changed := next.Magnitude != p.current.Magnitude
	p.current = next
	if err != nil {
		p.notice = "Settings could not be saved to " + p.store.Path() + ": " + err.Error()
	} else {
		p.notice = ""
	}
	p.mu.Unlock()
	if !changed {
		return toSettingsDTO(next), nil
	}
	p.applyMinimum(minimumOf(next.Magnitude))
	return toSettingsDTO(next), []event.Provider{event.USGS}
}

// normalised replaces any key nobody offers with its default, so a hand-edited
// or older file cannot leave a control showing nothing chosen.
func normalised(s ports.Settings) ports.Settings {
	defaults := Defaults()
	if !slices.ContainsFunc(Magnitudes, func(m Magnitude) bool { return m.Key == s.Magnitude }) {
		s.Magnitude = defaults.Magnitude
	}
	if !slices.ContainsFunc(Speeds, func(v dto.Speed) bool { return v.Key == s.Speed }) {
		s.Speed = defaults.Speed
	}
	if _, ok := window.ByKey(s.Window); !ok {
		s.Window = defaults.Window
	}
	s.HiddenCategories = nonNil(s.HiddenCategories)
	s.HiddenProviders = nonNil(s.HiddenProviders)
	return s
}

// minimumOf answers a normalised key's minimum magnitude.
func minimumOf(key string) float64 {
	i := slices.IndexFunc(Magnitudes, func(m Magnitude) bool { return m.Key == key })
	return Magnitudes[i].Minimum
}

// noticeSeparator joins standing notices from different use cases.
const noticeSeparator = "; "

// JoinNotices words several standing notices as one line, skipping the empty.
func JoinNotices(notices ...string) string {
	var kept []string
	for _, n := range notices {
		if n != "" {
			kept = append(kept, n)
		}
	}
	return strings.Join(kept, noticeSeparator)
}

func nonNil(list []string) []string {
	if list == nil {
		return []string{}
	}
	return slices.Clone(list)
}

func toSettingsDTO(s ports.Settings) dto.Settings {
	return dto.Settings{
		AutoRotate:       s.AutoRotate,
		Magnitude:        s.Magnitude,
		Speed:            s.Speed,
		WindowKey:        s.Window,
		HiddenCategories: nonNil(s.HiddenCategories),
		HiddenProviders:  nonNil(s.HiddenProviders),
		CloudsShown:      s.CloudsShown,
		DayNightShown:    s.DayNightShown,
	}
}

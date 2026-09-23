package cache

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
)

const capBytes = 1 << 20

func sample() ports.Snapshot {
	depth := 68.6
	return ports.Snapshot{
		RetrievedAt: time.Date(2026, 9, 23, 11, 52, 4, 0, time.UTC),
		Validator:   "Wed, 23 Sep 2026 11:52:04 GMT",
		Events: []event.Event{{
			Provider: event.USGS, ProviderEventID: "ak1", Category: event.Earthquake, Title: "M 3.2",
			Extras: event.Extras{DepthKm: &depth, MagnitudeType: "ml"},
			Observations: []event.Observation{{
				At: time.Date(2026, 9, 23, 11, 40, 0, 0, time.UTC), Where: event.Point{Lat: 61.899, Lng: -150.919},
				Measurement: &event.Measurement{Value: 3.21, Unit: "ml"},
			}},
		}},
	}
}

func TestFRSTS004_SaveThenLoadRoundTrips(t *testing.T) {
	t.Parallel()
	c := New(filepath.Join(t.TempDir(), "cache"), capBytes)
	if _, held, err := c.Load(event.USGS); held || err != nil {
		t.Fatalf("empty cache: held %v, %v", held, err)
	}
	want := sample()
	if err := c.Save(event.USGS, want); err != nil {
		t.Fatal(err)
	}
	got, held, err := c.Load(event.USGS)
	if err != nil || !held || !got.RetrievedAt.Equal(want.RetrievedAt) || got.Validator != want.Validator {
		t.Fatalf("loaded %+v, %v, %v", got, held, err)
	}
	e := got.Events[0]
	if e.ID() != "USGS:ak1" || *e.Extras.DepthKm != 68.6 || e.Observations[0].Measurement.Value != 3.21 {
		t.Errorf("event = %+v", e)
	}
	entries, _ := os.ReadDir(c.dir)
	if len(entries) != 1 {
		t.Errorf("NFR-REL-005: %d files left in the cache folder, want the one", len(entries))
	}
}

func TestLoadTreatsAnotherSchemaAsAbsent(t *testing.T) {
	t.Parallel()
	c := New(t.TempDir(), capBytes)
	_ = os.WriteFile(c.path(event.EONET), []byte(`{"schema":99,"snapshot":{}}`), 0o644)
	if _, held, err := c.Load(event.EONET); held || err != nil {
		t.Errorf("held %v, %v", held, err)
	}
}

func TestLoadReportsFaults(t *testing.T) {
	t.Parallel()
	c := New(t.TempDir(), 64)
	_ = os.WriteFile(c.path(event.EONET), []byte(`{broken`), 0o644)
	if _, _, err := c.Load(event.EONET); err == nil {
		t.Error("a damaged file loaded")
	}
	_ = os.WriteFile(c.path(event.USGS), []byte(strings.Repeat(" ", 65)), 0o644)
	if _, _, err := c.Load(event.USGS); !errors.Is(err, ErrTooLarge) {
		t.Errorf("oversized file: %v", err)
	}
	dirAsFile := New(t.TempDir(), capBytes)
	_ = os.Mkdir(dirAsFile.path(event.USGS), 0o755)
	if _, _, err := dirAsFile.Load(event.USGS); err == nil {
		t.Error("a folder in place of the file loaded")
	}
}

func TestSaveReportsFaults(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	_ = os.WriteFile(blocker, nil, 0o644)
	if err := New(filepath.Join(blocker, "under"), capBytes).Save(event.USGS, sample()); err == nil {
		t.Error("saved beneath a file")
	}
	c := New(t.TempDir(), capBytes)
	_ = os.Mkdir(c.path(event.USGS), 0o755)
	_ = os.WriteFile(filepath.Join(c.path(event.USGS), "keep"), nil, 0o644)
	if err := c.Save(event.USGS, sample()); err == nil {
		t.Error("replaced a non-empty folder")
	}
}

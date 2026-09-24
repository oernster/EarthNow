package cache

import (
	"bytes"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
)

func TestFRBA014_TheBurntDaysRoundTrip(t *testing.T) {
	t.Parallel()
	b := NewBurnt(t.TempDir(), capBytes)
	if _, held, err := b.Load(); held || err != nil {
		t.Fatalf("empty cache: held %v, %v", held, err)
	}
	at := time.Date(2026, 9, 24, 11, 48, 0, 0, time.UTC)
	want := []ports.BurntDay{
		{Day: time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC), PNG: []byte("\x89PNG 23"), Drawn: true, RetrievedAt: at},
		{Day: time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC), PNG: []byte("\x89PNG 24"), RetrievedAt: at},
	}
	if err := b.Save(want); err != nil {
		t.Fatal(err)
	}
	got, held, err := b.Load()
	if err != nil || !held || len(got) != len(want) {
		t.Fatalf("loaded %+v, %v, %v", got, held, err)
	}
	for i := range want {
		if !got[i].Day.Equal(want[i].Day) || !bytes.Equal(got[i].PNG, want[i].PNG) ||
			got[i].Drawn != want[i].Drawn || !got[i].RetrievedAt.Equal(want[i].RetrievedAt) {
			t.Errorf("day %d: %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestBurntLoadTreatsAnotherSchemaAsAbsentAndDamageAsAFault(t *testing.T) {
	t.Parallel()
	b := NewBurnt(t.TempDir(), capBytes)
	_ = os.WriteFile(b.path(), []byte(`{"schema":99}`), 0o644)
	if _, held, err := b.Load(); held || err != nil {
		t.Errorf("another schema: held %v, %v", held, err)
	}
	_ = os.WriteFile(b.path(), []byte(`{broken`), 0o644)
	if _, _, err := b.Load(); err == nil {
		t.Error("a damaged file loaded")
	}
}

func TestBurntWithNoFolderHoldsNothingAndRefusesToSave(t *testing.T) {
	t.Parallel()
	b := NewBurnt("", capBytes)
	if _, held, err := b.Load(); held || err != nil {
		t.Errorf("no folder: held %v, %v", held, err)
	}
	if err := b.Save(nil); !errors.Is(err, ErrNoFolder) {
		t.Errorf("Save = %v, want ErrNoFolder", err)
	}
}

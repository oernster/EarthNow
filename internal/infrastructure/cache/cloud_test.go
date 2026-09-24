package cache

import (
	"bytes"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
)

func TestFRCLD014_TheCloudImageRoundTrips(t *testing.T) {
	t.Parallel()
	c := NewCloud(t.TempDir(), capBytes)
	if _, held, err := c.Load(); held || err != nil {
		t.Fatalf("empty cache: held %v, %v", held, err)
	}
	want := ports.CloudImage{ValidTime: time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC), PNG: []byte("\x89PNG drawn")}
	if err := c.Save(want); err != nil {
		t.Fatal(err)
	}
	got, held, err := c.Load()
	if err != nil || !held || !got.ValidTime.Equal(want.ValidTime) || !bytes.Equal(got.PNG, want.PNG) {
		t.Fatalf("loaded %+v, %v, %v", got, held, err)
	}
}

func TestCloudLoadTreatsAnotherSchemaAsAbsentAndDamageAsAFault(t *testing.T) {
	t.Parallel()
	c := NewCloud(t.TempDir(), capBytes)
	_ = os.WriteFile(c.path(), []byte(`{"schema":99}`), 0o644)
	if _, held, err := c.Load(); held || err != nil {
		t.Errorf("another schema: held %v, %v", held, err)
	}
	_ = os.WriteFile(c.path(), []byte(`{broken`), 0o644)
	if _, _, err := c.Load(); err == nil {
		t.Error("a damaged file loaded")
	}
}

func TestCloudWithNoFolderHoldsNothingAndRefusesToSave(t *testing.T) {
	t.Parallel()
	c := NewCloud("", capBytes)
	if _, held, err := c.Load(); held || err != nil {
		t.Errorf("no folder: held %v, %v", held, err)
	}
	if err := c.Save(ports.CloudImage{PNG: []byte("x")}); !errors.Is(err, ErrNoFolder) {
		t.Errorf("Save = %v, want ErrNoFolder", err)
	}
}

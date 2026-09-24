package settings

// Ported from ED Voyage Companion's config/settings_test.go. Each store writes
// inside a temporary folder that does not exist yet, so every save also
// exercises the folder being created.

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/oernster/EarthNow/internal/application/ports"
)

var defaults = ports.Settings{AutoRotate: true, Magnitude: "2.5", Speed: "normal", Window: "24h", DayNightShown: true, TrailsShown: true}

func storeIn(t *testing.T) *File {
	t.Helper()
	return New(filepath.Join(t.TempDir(), "nested"))
}

func plant(t *testing.T, f *File, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(f.path), dirPerm); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f.path, []byte(body), filePerm); err != nil {
		t.Fatal(err)
	}
}

func TestFRSET004_NoFileIsAbsenceNotAFault(t *testing.T) {
	t.Parallel()
	got, held, err := storeIn(t).Load(defaults)
	if held || err != nil || !reflect.DeepEqual(got, defaults) {
		t.Errorf("Load = %+v, %v, %v", got, held, err)
	}
}

func TestChoicesSurviveASave(t *testing.T) {
	t.Parallel()
	store := storeIn(t)
	want := ports.Settings{AutoRotate: false, Magnitude: "4.5", Speed: "fast", Window: "7d", HiddenCategories: []string{"ICE"}, HiddenProviders: []string{"EONET"}, CloudsShown: true}
	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	got, held, err := store.Load(defaults)
	if !held || err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("Load = %+v, %v, %v; want %+v", got, held, err, want)
	}
	entries, _ := os.ReadDir(filepath.Dir(store.path))
	for _, e := range entries {
		if e.Name() != FileName {
			t.Errorf("a save left %q behind", e.Name())
		}
	}
}

func TestAFieldTheFileLacksKeepsItsDefault(t *testing.T) {
	t.Parallel()
	store := storeIn(t)
	plant(t, store, `{"magnitude":"all"}`)
	got, held, err := store.Load(defaults)
	if !held || err != nil || !got.AutoRotate || got.Magnitude != "all" || got.Window != "24h" {
		t.Errorf("Load = %+v, %v, %v", got, held, err)
	}
}

func TestFRDAY007_AFileFromBeforeTheLayerStartsItShown(t *testing.T) {
	t.Parallel()
	store := storeIn(t)
	// A 1.0.0 file: every field of its day, none for the day and night layer.
	plant(t, store, `{"autoRotate":true,"magnitude":"2.5","speed":"normal","window":"24h","cloudsShown":true}`)
	got, held, err := store.Load(defaults)
	if !held || err != nil || !got.DayNightShown {
		t.Errorf("Load = %+v, %v, %v; want the layer shown", got, held, err)
	}
	// FR-TRL-005: the same file holds no trails field either, so trails start on.
	if !got.TrailsShown {
		t.Errorf("Load = %+v; want storm trails on", got)
	}
}

func TestFRDAY007_HidingTheLayerSurvivesASave(t *testing.T) {
	t.Parallel()
	store := storeIn(t)
	want := defaults
	want.DayNightShown = false
	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	if got, _, err := store.Load(defaults); err != nil || got.DayNightShown {
		t.Errorf("Load = %+v, %v; want the layer hidden", got, err)
	}
}

func TestFRSET004_ADamagedFileIsAFault(t *testing.T) {
	t.Parallel()
	store := storeIn(t)
	plant(t, store, "{ this is not json")
	if got, held, err := store.Load(defaults); held || err == nil || !reflect.DeepEqual(got, defaults) {
		t.Errorf("Load = %+v, %v, %v", got, held, err)
	}
}

func TestAnOversizedFileIsRefused(t *testing.T) {
	t.Parallel()
	store := storeIn(t)
	plant(t, store, `{"window":"`+strings.Repeat("x", maxBytes)+`"}`)
	if _, _, err := store.Load(defaults); !errors.Is(err, ErrTooLarge) {
		t.Errorf("Load error = %v, want ErrTooLarge", err)
	}
}

func TestAFolderWhereTheFileShouldBeIsAFault(t *testing.T) {
	t.Parallel()
	store := storeIn(t)
	if err := os.MkdirAll(store.path, dirPerm); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Load(defaults); err == nil {
		t.Error("a folder in place of the file loaded")
	}
	if err := store.Save(defaults); err == nil {
		t.Error("a save over a folder reported success")
	}
}

func TestASaveWhoseWorkingFileCannotBeWrittenFails(t *testing.T) {
	t.Parallel()
	store := storeIn(t)
	if err := os.MkdirAll(store.path+".writing", dirPerm); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(defaults); err == nil {
		t.Error("a save whose working file is a folder reported success")
	}
}

func TestASaveWhoseFolderCannotBeMadeFails(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, filePerm); err != nil {
		t.Fatal(err)
	}
	if err := New(filepath.Join(blocker, "below")).Save(defaults); err == nil {
		t.Error("a save beneath a file reported success")
	}
}

// NFR-REL-001: a startup failure is carried on, never the end of the run.
func TestAMachineWithNoDataFolderStillStarts(t *testing.T) {
	t.Parallel()
	store := New("")
	if got, held, err := store.Load(defaults); held || !errors.Is(err, ErrNoFolder) || !reflect.DeepEqual(got, defaults) {
		t.Errorf("Load = %+v, %v, %v", got, held, err)
	}
	if err := store.Save(defaults); !errors.Is(err, ErrNoFolder) {
		t.Errorf("Save = %v", err)
	}
	if store.Path() != FileName {
		t.Errorf("Path = %q", store.Path())
	}
	if p := storeIn(t).Path(); !strings.HasSuffix(p, FileName) {
		t.Errorf("Path = %q", p)
	}
}

// Package settings keeps the reader's settings in one JSON file under
// %LOCALAPPDATA%\EarthNow (REQUIREMENTS.md FR-SET-004, NFR-PRIV-002). Ported
// from ED Voyage Companion's internal/infrastructure/config/settings.go.
package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/oernster/EarthNow/internal/application/ports"
)

// FileName is the settings file inside the data folder.
const FileName = "settings.json"

// maxBytes caps what Load will read. A real file is a few hundred bytes; the
// cap exists so a file that is not ours cannot ask for unbounded memory.
const maxBytes = 64 << 10

// dirPerm and filePerm are the ordinary permissions for a per-user file.
const (
	dirPerm  = 0o755
	filePerm = 0o644
)

// ErrNoFolder is the fault of a machine with no data folder to keep settings in.
var ErrNoFolder = errors.New("no data folder on this machine")

// ErrTooLarge refuses a settings file over the size cap before reading it.
var ErrTooLarge = errors.New("settings file larger than the size cap")

// stored is the file's shape, kept apart from ports.Settings on purpose: the
// file is a format and the port is a type, so a field rename in the
// application must not silently change what is on disk. The json names are
// the contract and they are here.
type stored struct {
	AutoRotate       bool     `json:"autoRotate"`
	Magnitude        string   `json:"magnitude"`
	Speed            string   `json:"speed"`
	Window           string   `json:"window"`
	HiddenCategories []string `json:"hiddenCategories"`
	HiddenProviders  []string `json:"hiddenProviders"`
	CloudsShown      bool     `json:"cloudsShown"`
	DayNightShown    bool     `json:"dayNightShown"`
	TrailsShown      bool     `json:"trailsShown"`
	BurntShown       bool     `json:"burntShown"`
}

// File reads and writes the settings file.
type File struct{ path string }

// New keeps the settings in dir. An empty dir is a machine with no data
// folder: the store still exists, loads as a fault and refuses to save, so the
// run starts on defaults and says why rather than refusing to start.
func New(dir string) *File {
	if dir == "" {
		return &File{}
	}
	return &File{path: filepath.Join(dir, FileName)}
}

// Path implements ports.SettingsStore.
func (f *File) Path() string {
	if f.path == "" {
		return FileName
	}
	return f.path
}

// Load implements ports.SettingsStore. A missing file is absence; one that
// cannot be read, is over the cap or does not parse is a fault. Fields the
// file does not hold keep the defaults they started from.
func (f *File) Load(defaults ports.Settings) (ports.Settings, bool, error) {
	if f.path == "" {
		return defaults, false, ErrNoFolder
	}
	fh, err := os.Open(f.path)
	if errors.Is(err, fs.ErrNotExist) {
		return defaults, false, nil
	}
	if err != nil {
		return defaults, false, fmt.Errorf("opening: %w", err)
	}
	defer func() { _ = fh.Close() }()
	raw, err := io.ReadAll(io.LimitReader(fh, maxBytes+1))
	if err != nil {
		return defaults, false, fmt.Errorf("reading: %w", err)
	}
	if len(raw) > maxBytes {
		return defaults, false, ErrTooLarge
	}
	held := toStored(defaults)
	if err := json.Unmarshal(raw, &held); err != nil {
		return defaults, false, fmt.Errorf("the file is damaged: %w", err)
	}
	return ports.Settings{
		AutoRotate:       held.AutoRotate,
		Magnitude:        held.Magnitude,
		Speed:            held.Speed,
		Window:           held.Window,
		HiddenCategories: held.HiddenCategories,
		HiddenProviders:  held.HiddenProviders,
		CloudsShown:      held.CloudsShown,
		DayNightShown:    held.DayNightShown,
		TrailsShown:      held.TrailsShown,
		BurntShown:       held.BurntShown,
	}, true, nil
}

// Save implements ports.SettingsStore. It writes through a temporary file in
// the same folder and renames over the target, so an interrupted write leaves
// the previous settings rather than a truncated file.
func (f *File) Save(chosen ports.Settings) error {
	if f.path == "" {
		return ErrNoFolder
	}
	// The encode cannot fail: stored holds only bools, strings and string
	// lists. The error is discarded rather than checked, because a branch
	// nothing can reach is a branch nothing can test.
	raw, _ := json.MarshalIndent(toStored(chosen), "", "  ")
	dir := filepath.Dir(f.path)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}
	temporary := f.path + ".writing"
	if err := os.WriteFile(temporary, raw, filePerm); err != nil {
		return fmt.Errorf("writing: %w", err)
	}
	if err := os.Rename(temporary, f.path); err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("replacing: %w", err)
	}
	return nil
}

func toStored(s ports.Settings) stored {
	return stored{
		AutoRotate:       s.AutoRotate,
		Magnitude:        s.Magnitude,
		Speed:            s.Speed,
		Window:           s.Window,
		HiddenCategories: s.HiddenCategories,
		HiddenProviders:  s.HiddenProviders,
		CloudsShown:      s.CloudsShown,
		DayNightShown:    s.DayNightShown,
		TrailsShown:      s.TrailsShown,
		BurntShown:       s.BurntShown,
	}
}

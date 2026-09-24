// Package cache keeps each provider's last successful set on disk, one JSON
// file per provider, written atomically (REQUIREMENTS.md FR-STS-004, DATA-009,
// NFR-REL-005): a new file is written beside the old one and renamed over it,
// so an interrupted write leaves the previous set intact. The cloud layer's
// last image is kept the same way (FR-CLD-014).
package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
)

// schemaVersion is written into every file; a file of another version is
// treated as absent rather than misread.
const schemaVersion = 1

// ErrTooLarge refuses a cache file over the size cap before reading it.
var ErrTooLarge = errors.New("cache file larger than the size cap")

type file struct {
	Schema   int            `json:"schema"`
	Snapshot ports.Snapshot `json:"snapshot"`
}

// Files is a directory of per-provider cache files.
type Files struct {
	dir      string
	maxBytes int64
}

// New keeps the cache in dir, refusing any file over maxBytes.
func New(dir string, maxBytes int64) *Files {
	return &Files{dir: dir, maxBytes: maxBytes}
}

func (f *Files) path(p event.Provider) string {
	return filepath.Join(f.dir, strings.ToLower(string(p))+".json")
}

// Load implements ports.SnapshotCache. A missing file is absence, as is one
// written by another schema version; a file that cannot be read is a fault.
func (f *Files) Load(p event.Provider) (ports.Snapshot, bool, error) {
	raw, found, err := readCapped(f.path(p), f.maxBytes, string(p))
	if !found || err != nil {
		return ports.Snapshot{}, false, err
	}
	var stored file
	if err := json.Unmarshal(raw, &stored); err != nil {
		return ports.Snapshot{}, false, fmt.Errorf("the %s cache is damaged: %w", p, err)
	}
	if stored.Schema != schemaVersion {
		return ports.Snapshot{}, false, nil
	}
	return stored.Snapshot, true, nil
}

// Save implements ports.SnapshotCache atomically.
func (f *Files) Save(p event.Provider, snap ports.Snapshot) error {
	return writeJSON(f.dir, f.path(p), string(p), file{Schema: schemaVersion, Snapshot: snap})
}

// readCapped reads the file at path, refusing one over maxBytes without
// reading past the cap. A missing file answers false with no error; what
// names the cache in any error.
func readCapped(path string, maxBytes int64, what string) ([]byte, bool, error) {
	fh, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("opening the %s cache: %w", what, err)
	}
	defer func() { _ = fh.Close() }()
	raw, err := io.ReadAll(io.LimitReader(fh, maxBytes+1))
	if err != nil {
		return nil, false, fmt.Errorf("reading the %s cache: %w", what, err)
	}
	if int64(len(raw)) > maxBytes {
		return nil, false, fmt.Errorf("%w: %s", ErrTooLarge, path)
	}
	return raw, true, nil
}

// writeJSON encodes v and writes it to path atomically, creating dir first.
func writeJSON(dir, path, what string, v any) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating the cache folder: %w", err)
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encoding the %s cache: %w", what, err)
	}
	tmp, err := os.CreateTemp(dir, "."+strings.ToLower(what)+"-*.tmp")
	if err != nil {
		return fmt.Errorf("writing the %s cache: %w", what, err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("writing the %s cache: %w", what, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("writing the %s cache: %w", what, err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("replacing the %s cache: %w", what, err)
	}
	return nil
}

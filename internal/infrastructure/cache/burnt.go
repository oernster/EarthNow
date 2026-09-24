package cache

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
)

// burntName names the burnt-area file and its errors.
const burntName = "burnt"

type burntDay struct {
	Day         time.Time `json:"day"`
	PNG         []byte    `json:"png"`
	Drawn       bool      `json:"drawn"`
	RetrievedAt time.Time `json:"retrievedAt"`
}

type burntFile struct {
	Schema int        `json:"schema"`
	Days   []burntDay `json:"days"`
}

// Burnt keeps the last good image of each burnt-area day (FR-BA-014) in one
// file beside the providers' own. Which days to keep is the application's
// rule; this keeps what it is handed.
type Burnt struct {
	dir      string
	maxBytes int64
}

// NewBurnt keeps the days in dir, refusing a file over maxBytes. An empty dir
// is a machine with no data folder: nothing is held and nothing can be saved.
func NewBurnt(dir string, maxBytes int64) *Burnt {
	return &Burnt{dir: dir, maxBytes: maxBytes}
}

func (b *Burnt) path() string { return filepath.Join(b.dir, burntName+".json") }

// Load implements ports.BurntCache. A missing file is absence, as is one
// written by another schema version; a file that cannot be read is a fault.
func (b *Burnt) Load() ([]ports.BurntDay, bool, error) {
	if b.dir == "" {
		return nil, false, nil
	}
	raw, found, err := readCapped(b.path(), b.maxBytes, burntName)
	if !found || err != nil {
		return nil, false, err
	}
	var stored burntFile
	if err := json.Unmarshal(raw, &stored); err != nil {
		return nil, false, fmt.Errorf("the %s cache is damaged: %w", burntName, err)
	}
	if stored.Schema != schemaVersion {
		return nil, false, nil
	}
	days := make([]ports.BurntDay, len(stored.Days))
	for i, d := range stored.Days {
		days[i] = ports.BurntDay{Day: d.Day, PNG: d.PNG, Drawn: d.Drawn, RetrievedAt: d.RetrievedAt}
	}
	return days, true, nil
}

// Save implements ports.BurntCache atomically.
func (b *Burnt) Save(days []ports.BurntDay) error {
	if b.dir == "" {
		return ErrNoFolder
	}
	stored := burntFile{Schema: schemaVersion, Days: make([]burntDay, len(days))}
	for i, d := range days {
		stored.Days[i] = burntDay{Day: d.Day, PNG: d.PNG, Drawn: d.Drawn, RetrievedAt: d.RetrievedAt}
	}
	return writeJSON(b.dir, b.path(), burntName, stored)
}

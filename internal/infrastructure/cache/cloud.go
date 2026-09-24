package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
)

// cloudName names the cloud image's file and its errors.
const cloudName = "cloud"

// ErrNoFolder refuses to save the cloud image on a machine with no data folder.
var ErrNoFolder = errors.New("there is no data folder to keep the cloud image in")

type cloudFile struct {
	Schema    int       `json:"schema"`
	ValidTime time.Time `json:"validTime"`
	PNG       []byte    `json:"png"`
}

// Cloud keeps the cloud layer's last good image and its valid time
// (FR-CLD-014) in one file beside the providers' own.
type Cloud struct {
	dir      string
	maxBytes int64
}

// NewCloud keeps the image in dir, refusing a file over maxBytes. An empty dir
// is a machine with no data folder: nothing is held and nothing can be saved.
func NewCloud(dir string, maxBytes int64) *Cloud {
	return &Cloud{dir: dir, maxBytes: maxBytes}
}

func (c *Cloud) path() string { return filepath.Join(c.dir, cloudName+".json") }

// Load implements ports.CloudCache. A missing file is absence, as is one
// written by another schema version; a file that cannot be read is a fault.
func (c *Cloud) Load() (ports.CloudImage, bool, error) {
	if c.dir == "" {
		return ports.CloudImage{}, false, nil
	}
	raw, found, err := readCapped(c.path(), c.maxBytes, cloudName)
	if !found || err != nil {
		return ports.CloudImage{}, false, err
	}
	var stored cloudFile
	if err := json.Unmarshal(raw, &stored); err != nil {
		return ports.CloudImage{}, false, fmt.Errorf("the %s cache is damaged: %w", cloudName, err)
	}
	if stored.Schema != schemaVersion {
		return ports.CloudImage{}, false, nil
	}
	return ports.CloudImage{ValidTime: stored.ValidTime, PNG: stored.PNG}, true, nil
}

// Save implements ports.CloudCache atomically.
func (c *Cloud) Save(img ports.CloudImage) error {
	if c.dir == "" {
		return ErrNoFolder
	}
	return writeJSON(c.dir, c.path(), cloudName, cloudFile{Schema: schemaVersion, ValidTime: img.ValidTime, PNG: img.PNG})
}

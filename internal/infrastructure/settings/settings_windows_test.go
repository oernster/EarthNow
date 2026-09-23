//go:build windows

package settings

import (
	"path/filepath"
	"testing"
)

// An open that fails for a reason other than absence is a fault, not a first
// run. A name Windows refuses outright is one such reason; other platforms
// accept the character, so this lives behind the build tag.
func TestAnUnopenableFileIsAFault(t *testing.T) {
	t.Parallel()
	store := New(filepath.Join(t.TempDir(), "bad<name"))
	if _, held, err := store.Load(defaults); held || err == nil {
		t.Errorf("Load = %v, %v; want a fault", held, err)
	}
}

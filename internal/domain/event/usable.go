package event

import (
	"errors"
	"fmt"
)

// ErrNoneUsable is FR-PRV-012's parse failure for an answer that lists items
// yet yields none: what a change to a source's format looks like. Taken as an
// empty set it would wipe the source's offline copy and report it healthy.
var ErrNoneUsable = errors.New("none of its items could be used")

// CheckUsable is the one rule every provider adapter applies to its answer.
// usable counts the items it mapped before any filter of the reader's; dropped
// counts those it refused as malformed (FR-PRV-013). An answer whose every item
// was refused is an error. One with no items is not; nor is one with a usable.
func CheckUsable(usable, dropped int) error {
	if usable == 0 && dropped > 0 {
		return fmt.Errorf("%w: %d dropped as malformed", ErrNoneUsable, dropped)
	}
	return nil
}

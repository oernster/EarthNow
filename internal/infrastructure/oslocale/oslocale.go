// Package oslocale reads the operating system's country or region setting
// (REQUIREMENTS.md FR-GLB-014): the user's home location on Windows, the region
// of AppleLocale on macOS, the territory of LC_ALL or LANG elsewhere. All three
// readings are measured (ASM-010). Nothing read here leaves the machine.
package oslocale

import (
	"errors"
	"fmt"
)

// ErrNoRegion is a setting that names no country: absent, C, POSIX or a UN
// area such as 001. It is an answer, not a fault; the globe opens as before.
var ErrNoRegion = errors.New("the region setting names no country")

// Setting is the operating system's country or region setting.
type Setting struct{}

// Region implements ports.RegionSource: the setting as an ISO 3166-1 alpha-2 code.
func (Setting) Region() (string, error) {
	return decide(read())
}

// decide turns what the platform answered into a code or a worded reason.
func decide(raw string, parse func(string) (string, bool), err error) (string, error) {
	if err != nil {
		return "", fmt.Errorf("reading the region setting: %w", err)
	}
	code, ok := parse(raw)
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrNoRegion, raw)
	}
	return code, nil
}

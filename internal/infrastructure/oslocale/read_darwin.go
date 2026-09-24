//go:build darwin

package oslocale

import (
	"os/exec"

	"github.com/oernster/EarthNow/internal/domain/region"
)

// read asks macOS for AppleLocale, whose region part is the Region setting
// (en_GB; en_US@rg=gbzzzz where the region differs from the language's own).
// Not yet measured on a Mac (ASM-010).
func read() (string, func(string) (string, bool), error) {
	out, err := exec.Command("defaults", "read", "-g", "AppleLocale").Output()
	return string(out), region.FromLocale, err
}

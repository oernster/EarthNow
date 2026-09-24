//go:build !windows && !darwin

package oslocale

import (
	"os"

	"github.com/oernster/EarthNow/internal/domain/region"
)

// read takes the territory from LC_ALL, which overrides every category, else
// from LANG, the desktop's own locale (ASM-010). The Flatpak passes both through.
func read() (string, func(string) (string, bool), error) {
	if all := os.Getenv("LC_ALL"); all != "" {
		return all, region.FromLocale, nil
	}
	return os.Getenv("LANG"), region.FromLocale, nil
}

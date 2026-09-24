package oslocale

import (
	"errors"
	"testing"

	"github.com/oernster/EarthNow/internal/domain/region"
)

func TestFRGLB014_TheSettingBecomesACodeOrAReason(t *testing.T) {
	t.Parallel()
	if code, err := decide("en_GB.UTF-8", region.FromLocale, nil); code != "GB" || err != nil {
		t.Errorf("a locale with a territory = %q, %v", code, err)
	}
	if code, err := decide("001", region.Code, nil); code != "" || !errors.Is(err, ErrNoRegion) {
		t.Errorf("the world = %q, %v; want ErrNoRegion", code, err)
	}
	failed := errors.New("no such procedure")
	if code, err := decide("", region.Code, failed); code != "" || !errors.Is(err, failed) {
		t.Errorf("a failed read = %q, %v; want the failure", code, err)
	}
}

func TestFRGLB014_ThisMachineAnswersACodeOrNone(t *testing.T) {
	t.Parallel()
	code, err := Setting{}.Region()
	switch {
	case err == nil:
		if _, ok := region.Code(code); !ok {
			t.Errorf("Region() = %q, not a code", code)
		}
		t.Logf("this machine's region: %s", code)
	case errors.Is(err, ErrNoRegion):
		t.Logf("this machine sets no country: %v", err)
	default:
		t.Errorf("Region() failed: %v", err)
	}
}

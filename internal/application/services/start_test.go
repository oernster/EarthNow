package services

import (
	"errors"
	"strings"
	"testing"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/domain/event"
)

type fakeRegion struct {
	code string
	err  error
}

func (f fakeRegion) Region() (string, error) { return f.code, f.err }

type fakeLabels map[string]event.Point

func (f fakeLabels) Label(code string) (event.Point, bool) {
	p, ok := f[code]
	return p, ok
}

var britain = fakeLabels{"GB": {Lat: 54.4027, Lng: -2.1163}}

func TestFRGLB015_TheGlobeOpensFacingTheRegionsLabelPoint(t *testing.T) {
	t.Parallel()
	got, line := NewStartView(fakeRegion{code: "GB"}, britain).Answer()
	want := dto.StartView{Found: true, Lat: 54.4027, Lng: -2.1163}
	if got != want {
		t.Errorf("Answer() = %+v, want %+v", got, want)
	}
	if !strings.Contains(line, "facing GB") {
		t.Errorf("log line %q does not name the region", line)
	}
}

func TestFRGLB016_NoUsableRegionOpensAsBeforeAndSaysWhy(t *testing.T) {
	t.Parallel()
	unset := errors.New("the region setting names no country: \"001\"")
	cases := []struct {
		region fakeRegion
		reason string
	}{
		{fakeRegion{err: unset}, "001"},
		{fakeRegion{code: "ZZ"}, "no label point for region ZZ"},
	}
	for _, c := range cases {
		got, line := NewStartView(c.region, britain).Answer()
		if got.Found {
			t.Errorf("%+v opened facing %+v", c.region, got)
		}
		if !strings.Contains(line, c.reason) || !strings.Contains(line, "opening as before") {
			t.Errorf("log line %q does not say why", line)
		}
	}
}

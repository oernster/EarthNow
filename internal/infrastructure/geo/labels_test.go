package geo

import (
	"errors"
	"testing"
)

func TestFRGLB017_TheEmbeddedTableHoldsOneLabelPerCountry(t *testing.T) {
	t.Parallel()
	labels, err := LoadLabels()
	if err != nil {
		t.Fatal(err)
	}
	// Natural Earth 5.1.1's label points (measured), including the codes whose
	// dependencies share them: the kept row is the country itself.
	want := map[string][2]float64{
		"GB": {54.4027, -2.1163},
		"FR": {46.6961, 2.5523},
		"NO": {61.3571, 9.68},
		"AU": {-24.1295, 134.0497},
		"BR": {-12.0987, -49.5594},
		"KZ": {49.0541, 68.6855},
		"US": {39.5385, -97.4826},
	}
	for code, p := range want {
		got, ok := labels.Label(code)
		if !ok || got.Lat != p[0] || got.Lng != p[1] {
			t.Errorf("Label(%s) = %+v, %v; want %v", code, got, ok, p)
		}
	}
}

func TestFRGLB016_ACodeTheTableLacksHasNoLabel(t *testing.T) {
	t.Parallel()
	labels, err := loadLabels([]byte(`{"GB":[54.4,-2.1],"XX":[91,0]}`))
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{"ZZ", "001", "", "XX"} {
		if at, ok := labels.Label(code); ok {
			t.Errorf("Label(%q) = %+v, want none", code, at)
		}
	}
}

func TestLabelsRefuseADamagedOrEmptyTable(t *testing.T) {
	t.Parallel()
	if _, err := loadLabels([]byte(`{`)); err == nil {
		t.Error("a damaged table loaded")
	}
	if _, err := loadLabels([]byte(`{}`)); !errors.Is(err, ErrNoLabels) {
		t.Errorf("an empty table: %v", err)
	}
}

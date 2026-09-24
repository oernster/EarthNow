package event

import "testing"

func TestFRSEL010_DepthReadsInKilometresWithItsBand(t *testing.T) {
	t.Parallel()
	cases := []struct {
		km   float64
		want string
	}{
		{18.44, "18.4 km, shallow"},
		{69.94, "69.9 km, shallow"},
		{69.96, "70.0 km, intermediate"},
		{70, "70.0 km, intermediate"},
		{121.534, "121.5 km, intermediate"},
		{300, "300.0 km, deep"},
		{583.682, "583.7 km, deep"},
		{0, "0.0 km, shallow"},
		{-0.04, "0.0 km, shallow"},
	}
	for _, c := range cases {
		if got := DepthWording(c.km); got != c.want {
			t.Errorf("DepthWording(%v) = %q, want %q", c.km, got, c.want)
		}
	}
}

func TestFRSEL011_ADepthAboveSeaLevelSaysSo(t *testing.T) {
	t.Parallel()
	if got, want := DepthWording(-1.2), "1.2 km above sea level, shallow"; got != want {
		t.Errorf("DepthWording(-1.2) = %q, want %q", got, want)
	}
}

func TestFRSEL012_ExactlyTenKilometresIsMarkedAsOftenFixed(t *testing.T) {
	t.Parallel()
	want := "10.0 km, shallow (often a fixed depth: USGS assigns 10 km when it cannot compute one)"
	if got := DepthWording(10); got != want {
		t.Errorf("DepthWording(10) = %q, want %q", got, want)
	}
	if got, want := DepthWording(10.04), "10.0 km, shallow"; got != want {
		t.Errorf("DepthWording(10.04) = %q, want %q", got, want)
	}
}

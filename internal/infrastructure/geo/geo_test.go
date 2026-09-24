package geo

import (
	"bytes"
	"compress/gzip"
	"strings"
	"sync"
	"testing"
)

var (
	loaded     *Gazetteer
	loadErr    error
	loadedOnce sync.Once
)

func gazetteer(t *testing.T) *Gazetteer {
	t.Helper()
	loadedOnce.Do(func() { loaded, loadErr = Load() })
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	return loaded
}

// The expected lines are the ones measured in the Phase 0 spike.
func TestFRGEO_KnownPoints(t *testing.T) {
	t.Parallel()
	g := gazetteer(t)
	cases := []struct {
		name     string
		lat, lng float64
		want     string
	}{
		{"FR-GEO-001 near Tromso", 69.70, 19.10, "8 km NE of Tromsø, Troms, Norway"},
		{"FR-GEO-002 Kehl across the Rhine", 48.5717, 7.8156, "In Germany; 5 km E of Strasbourg, Alsace, France"},
		{"FR-GEO-003 mid-Atlantic", 30.0, -40.0, "At sea; 1,408 km SW of Horta, Azores, Portugal"},
		{"region disambiguates Montana", 61.899, -150.919, "49 km SW of Montana, Alaska, United States of America"},
		{"antimeridian", -17.0, 179.9, "At sea; 85 km SE of Labasa, Western, Fiji"},
		{"FR-GEO-002 the Ross Ice Shelf is Antarctica", -81.0, 180.0, "441 km SE of Scott Base, Antarctica"},
		{"FR-GEO-002 the Ronne Ice Shelf is Antarctica", -78.0, -60.0, "523 km NW of Sobral Base, Antarctica"},
		{"FR-GEO-003 the Weddell Sea is still sea", -70.0, -30.0, "At sea; 672 km W of Wasa Station, Antarctica"},
	}
	for _, c := range cases {
		if got := g.Describe(c.lat, c.lng); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}

func TestSourceErrorsAreCorrected(t *testing.T) {
	t.Parallel()
	got := gazetteer(t).Describe(4.9372, -52.326)
	if !strings.Contains(got, "Cayenne, Guyane, France") {
		t.Errorf("Cayenne line = %q", got)
	}
}

func gz(t *testing.T, s string) []byte {
	t.Helper()
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, _ = w.Write([]byte(s))
	_ = w.Close()
	return b.Bytes()
}

func TestWordingAtZeroKilometres(t *testing.T) {
	t.Parallel()
	places := gz(t, `[{"n":"Here","r":"Here","c":"Land","lat":10,"lng":10}]`)
	square := `[{"n":"Land","r":[[[9,9],[11,9],[11,11],[9,11],[9,9]]]}]`
	g, err := load(places, gz(t, square), gz(t, `[]`))
	if err != nil {
		t.Fatal(err)
	}
	if got := g.Describe(10, 10); got != "At Here, Land" {
		t.Errorf("on land = %q", got)
	}
	g, _ = load(places, gz(t, `[]`), gz(t, `[]`))
	if got := g.Describe(10, 10); got != "At sea; at Here, Land" {
		t.Errorf("at sea = %q", got)
	}
	if lowerFirst("") != "" || lowerFirst("12 km N") != "12 km N" {
		t.Error("lowerFirst changed a line with no leading At")
	}
}

// FR-GEO-002: a point on an ice shelf lies in Antarctica; a point off the shelf
// but inside a hole in it (an island's ground) is not on the shelf.
func TestFRGEO002_AnIceShelfCountsAsAntarctica(t *testing.T) {
	t.Parallel()
	places := gz(t, `[{"n":"Base","r":"","c":"Antarctica","lat":-70,"lng":0}]`)
	pole := gz(t, `[{"n":"Antarctica","r":[[[-10,-90],[10,-90],[10,-85],[-10,-85],[-10,-90]]]}]`)
	shelf := gz(t, `[{"n":"Shelf","r":[[[-10,-80],[10,-80],[10,-75],[-10,-75],[-10,-80]],[[-1,-78],[1,-78],[1,-77],[-1,-77],[-1,-78]]]}]`)
	g, err := load(places, pole, shelf)
	if err != nil {
		t.Fatal(err)
	}
	if got := g.countryAt(-79, 5); got != shelfCountry {
		t.Errorf("on the shelf = %q, want %s", got, shelfCountry)
	}
	if got := g.countryAt(-77.5, 0); got != "" {
		t.Errorf("in the shelf's hole = %q, want none", got)
	}
}

func TestLoadRefusesUnusableData(t *testing.T) {
	t.Parallel()
	good := gz(t, `[{"n":"A","c":"B","lat":1,"lng":1}]`)
	none := gz(t, `[]`)
	shelf := gz(t, `[{"n":"S","r":[[[0,0],[1,0],[1,1],[0,0]]]}]`)
	cases := map[string][3][]byte{
		"places not gzip":            {[]byte("plain"), none, none},
		"places not JSON":            {gz(t, `{`), none, none},
		"countries not gzip":         {good, []byte("plain"), none},
		"shelves not gzip":           {good, none, []byte("plain")},
		"no places":                  {none, none, none},
		"truncated gzip":             {gz(t, `[1,2,3]`)[:12], none, none},
		"shelves with no Antarctica": {good, gz(t, `[{"n":"B","r":[]}]`), shelf},
	}
	for name, data := range cases {
		if _, err := load(data[0], data[1], data[2]); err == nil {
			t.Errorf("%s: loaded", name)
		}
	}
}

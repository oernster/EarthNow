package gwis

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"net/url"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/domain/burnt"
	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch/fetchtest"
	"github.com/oernster/EarthNow/internal/infrastructure/pngcheck"
)

// mapImage is an image as the service sends it: transparent except the
// pixels given, each red at its opacity.
func mapImage(t *testing.T, w, h int, pixels map[int]uint8) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i, a := range pixels {
		img.SetNRGBA(i%w, i/w, color.NRGBA{R: burnt.Red, A: a})
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

var sep23 = time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC)

func TestFRBA002_OneDayIsAskedForAtTheSizeAndDay(t *testing.T) {
	t.Parallel()
	g := fetchtest.Answering(mapImage(t, Width, Height, nil))
	if _, _, err := New(g, BaseURL).Day(context.Background(), sep23); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(g.Asked[0])
	if err != nil || u.Host != Host {
		t.Fatalf("asked %s", g.Asked[0])
	}
	q := u.Query()
	if q.Get("LAYERS") != "nrt.ba" || q.Get("SRS") != "EPSG:4326" || q.Get("WIDTH") != "2048" ||
		q.Get("HEIGHT") != "1024" || q.Get("time") != "2026-09-23" || g.Accepts[0] != "image/png" {
		t.Fatalf("query %v accepting %q", q, g.Accepts[0])
	}
}

func TestFRBA009_ADayIsDrawnOnlyWhenAPixelIsBurnt(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name   string
		pixels map[int]uint8
		want   bool
	}{
		{"an empty day", nil, false},
		{"one faint pixel", map[int]uint8{Width + 3: 1}, true},
	} {
		img := mapImage(t, Width, Height, c.pixels)
		body, drawn, err := New(fetchtest.Answering(img), BaseURL).Day(context.Background(), sep23)
		if err != nil || drawn != c.want || !bytes.Equal(body, img) {
			t.Errorf("%s: drawn %v, err %v", c.name, drawn, err)
		}
	}
}

func TestFRBA013_AnythingButThePNGAskedForIsRefused(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name string
		body []byte
		want error
	}{
		{"an empty body, as a range answers", nil, pngcheck.ErrNotAnImage},
		{"a PNG of another size", mapImage(t, Width/2, Height/2, nil), pngcheck.ErrWrongSize},
	} {
		if _, _, err := New(fetchtest.Answering(c.body), BaseURL).Day(context.Background(), sep23); !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
	}
	down := errors.New("offline")
	if _, _, err := New(&fetchtest.Getter{Err: down}, BaseURL).Day(context.Background(), sep23); !errors.Is(err, down) {
		t.Errorf("err = %v, want the getter's", err)
	}
}

func TestFRBA006_ComposeKeepsTheHighestOpacityInRed(t *testing.T) {
	t.Parallel()
	const shared, onlyFirst, onlySecond = 5, Width*7 + 1, Width*Height - 1
	first := mapImage(t, Width, Height, map[int]uint8{shared: 102, onlyFirst: 255})
	second := mapImage(t, Width, Height, map[int]uint8{shared: 230, onlySecond: 60})
	out, err := New(nil, BaseURL).Compose([][]byte{first, second})
	if err != nil {
		t.Fatal(err)
	}
	img, err := pngcheck.Decode(out, Width, Height)
	if err != nil {
		t.Fatal(err)
	}
	want := map[int]uint8{shared: 230, onlyFirst: 255, onlySecond: 60, 0: 0}
	for i, a := range want {
		c := color.NRGBAModel.Convert(img.At(i%Width, i/Width)).(color.NRGBA)
		if c.A != a || (a > 0 && (c.R != burnt.Red || c.G != burnt.Green || c.B != burnt.Blue)) {
			t.Errorf("pixel %d is %+v, want red at %d", i, c, a)
		}
	}
}

func TestFRBA006_ComposeRefusesADamagedImage(t *testing.T) {
	t.Parallel()
	if _, err := New(nil, BaseURL).Compose([][]byte{[]byte("not a png")}); !errors.Is(err, pngcheck.ErrNotAnImage) {
		t.Fatalf("err = %v", err)
	}
}

package clouds

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/domain/cloud"
	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch"
)

// fakeGetter answers one body and records what was asked.
type fakeGetter struct {
	resp    httpfetch.Response
	err     error
	asked   []string
	accepts []string
}

func (f *fakeGetter) GetAccepting(_ context.Context, rawURL, _, accept string) (httpfetch.Response, error) {
	f.asked = append(f.asked, rawURL)
	f.accepts = append(f.accepts, accept)
	return f.resp, f.err
}

func sourceAnswering(body []byte) (*Source, *fakeGetter) {
	g := &fakeGetter{resp: httpfetch.Response{Body: body}}
	return New(g, BaseURL), g
}

// sourceImage is a grey image as the service sends it: row 0 holds a clear, a
// half and a full cloud pixel, then one with no data.
func sourceImage(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	grey := func(x int, v, a uint8) { img.SetNRGBA(x, 0, color.NRGBA{R: v, G: v, B: v, A: a}) }
	grey(0, cloud.ClearThreshold, 255)
	grey(1, 77, 255)
	grey(2, cloud.CloudThreshold, 255)
	grey(3, cloud.CloudThreshold, 0)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestFRCLD004_TheNewestValidTimeIsTheTimeDimensionsDefault(t *testing.T) {
	t.Parallel()
	doc, err := os.ReadFile("testdata/capabilities.xml")
	if err != nil {
		t.Fatal(err)
	}
	s, g := sourceAnswering(doc)
	got, err := s.Latest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("Latest = %v, want %v", got, want)
	}
	u, _ := url.Parse(g.asked[0])
	if u.Host != Host || u.Path != "/geoserver/mumi/worldcloudmap_ir108/ows" || u.Query().Get("request") != "GetCapabilities" {
		t.Errorf("asked %s", g.asked[0])
	}
	if g.accepts[0] != httpfetch.AcceptXML {
		t.Errorf("capabilities asked for %q", g.accepts[0])
	}
}

func TestLatest_RefusesADocumentWithNoUsableTime(t *testing.T) {
	t.Parallel()
	for _, doc := range []string{
		`<WMS_Capabilities><Dimension name="elevation" default="0"/></WMS_Capabilities>`,
		`<WMS_Capabilities><Dimension name="time" default="yesterday"/></WMS_Capabilities>`,
	} {
		s, _ := sourceAnswering([]byte(doc))
		if _, err := s.Latest(context.Background()); !errors.Is(err, ErrNoTime) {
			t.Errorf("%s: err = %v, want ErrNoTime", doc, err)
		}
	}
	s, _ := sourceAnswering([]byte(`<WMS_Capabilities><Layer>`))
	if _, err := s.Latest(context.Background()); err == nil {
		t.Error("a truncated document was read")
	}
}

func TestFRCLD016_TheImageIsAskedForAtTheListedTime(t *testing.T) {
	t.Parallel()
	s, g := sourceAnswering(sourceImage(t, Width, Height))
	at := time.Date(2026, 9, 24, 15, 0, 0, 0, time.UTC)
	if _, err := s.Image(context.Background(), at); err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(g.asked[0])
	q := u.Query()
	want := map[string]string{
		"request": "GetMap", "layers": "mumi:worldcloudmap_ir108", "crs": "CRS:84",
		"bbox": "-180,-90,180,90", "width": "2048", "height": "1024",
		"format": "image/png", "time": "2026-09-24T15:00:00Z",
	}
	for k, v := range want {
		if q.Get(k) != v {
			t.Errorf("%s = %q, want %q", k, q.Get(k), v)
		}
	}
}

func TestFRCLD006_TheDrawnImageFollowsTheRamp(t *testing.T) {
	t.Parallel()
	out, err := Draw(sourceImage(t, Width, Height))
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	for x, v := range []uint8{cloud.ClearThreshold, 77, cloud.CloudThreshold} {
		want := cloud.Shade(v, true)
		got := color.NRGBAModel.Convert(img.At(x, 0)).(color.NRGBA)
		if got.A != want.A || (want.A != 0 && got.R != want.R) {
			t.Errorf("pixel %d = %+v, want %+v", x, got, want)
		}
	}
	veil := cloud.Shade(0, false)
	if got := color.NRGBAModel.Convert(img.At(3, 0)).(color.NRGBA); got != (color.NRGBA{R: veil.R, G: veil.G, B: veil.B, A: veil.A}) {
		t.Errorf("no-data pixel = %+v, want the veil", got)
	}
}

func TestFRCLD012_AnythingButThePNGAskedForIsRefused(t *testing.T) {
	t.Parallel()
	exception := []byte(`<?xml version="1.0"?><ServiceExceptionReport><ServiceException>bad time</ServiceException></ServiceExceptionReport>`)
	good := sourceImage(t, Width, Height)
	cases := []struct {
		name string
		body []byte
		want error
	}{
		{"an XML exception served with 200", exception, ErrNotAnImage},
		{"a PNG of another size", sourceImage(t, Width/2, Height/2), ErrWrongSize},
		{"a PNG cut short", good[:len(good)/2], ErrNotAnImage},
	}
	for _, c := range cases {
		s, _ := sourceAnswering(c.body)
		if _, err := s.Image(context.Background(), time.Now()); !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
	}
}

func TestFetch_FailuresReachTheCaller(t *testing.T) {
	t.Parallel()
	down := errors.New("no route to host")
	s := New(&fakeGetter{err: down}, BaseURL)
	if _, err := s.Latest(context.Background()); !errors.Is(err, down) {
		t.Errorf("err = %v, want the fetch error", err)
	}
	s = New(&fakeGetter{resp: httpfetch.Response{NotModified: true}}, BaseURL)
	if _, err := s.Image(context.Background(), time.Now()); !errors.Is(err, ErrNotChanged) {
		t.Errorf("err = %v, want ErrNotChanged", err)
	}
}

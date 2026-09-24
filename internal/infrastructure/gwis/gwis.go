// Package gwis reaches the Global Wildfire Information System's burnt-area map
// (REQUIREMENTS.md 3.2.13): it asks for one UTC day's image (FR-BA-002),
// refuses anything that is not the image asked for (FR-BA-013) and draws the
// days of a window together by the domain's rule (FR-BA-006), so the page is
// handed finished pixels.
package gwis

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/url"
	"time"

	"github.com/oernster/EarthNow/internal/domain/burnt"
	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch"
	"github.com/oernster/EarthNow/internal/infrastructure/pngcheck"
)

// Host is the one host the burnt-area layer reaches (NFR-PRIV-001).
const Host = "maps.effis.emergency.copernicus.eu"

// BaseURL is the service's WMS endpoint (R12).
const BaseURL = "https://" + Host + "/gwis"

// The layer and the image FR-BA-002 asks for: the whole world in EPSG:4326 at
// 2048 x 1024 (the width the burnt-area spike, FR-BA-015, confirms), one day
// per request, since a range answers an empty body (measured 2026-09-24).
const (
	layer      = "nrt.ba"
	srs        = "EPSG:4326"
	bbox       = "-180,-90,180,90"
	Width      = 2048
	Height     = 1024
	wmsVersion = "1.1.1"
	pngType    = "image/png"
	dayLayout  = "2006-01-02"
)

// alphaShift turns a 16-bit colour channel into an 8-bit one.
const alphaShift = 8

// Source implements ports.BurntSource over the service at base.
type Source struct {
	get  httpfetch.Getter
	base string
}

// New builds the adapter; base is BaseURL outside tests.
func New(get httpfetch.Getter, base string) *Source {
	return &Source{get: get, base: base}
}

// Day implements ports.BurntSource: the image of one UTC day, checked, with
// whether it drew anything.
func (s *Source) Day(ctx context.Context, day time.Time) ([]byte, bool, error) {
	q := url.Values{
		"SERVICE": {"WMS"}, "VERSION": {wmsVersion}, "REQUEST": {"GetMap"},
		"LAYERS": {layer}, "STYLES": {""}, "SRS": {srs}, "BBOX": {bbox},
		"WIDTH": {fmt.Sprint(Width)}, "HEIGHT": {fmt.Sprint(Height)},
		"FORMAT": {pngType}, "TRANSPARENT": {"true"},
		"time": {day.UTC().Format(dayLayout)},
	}
	body, err := httpfetch.Fresh(ctx, s.get, s.base+"?"+q.Encode(), pngType)
	if err != nil {
		return nil, false, err
	}
	img, err := pngcheck.Decode(body, Width, Height)
	if err != nil {
		return nil, false, err
	}
	for _, a := range opacities(img) {
		if a > 0 {
			return body, true, nil
		}
	}
	return body, false, nil
}

// Compose implements ports.BurntSource: every pixel at the highest opacity any
// of the images gives it, in the source's own red (FR-BA-006). Each image is
// checked again, since it may have come from the cache.
func (s *Source) Compose(images [][]byte) ([]byte, error) {
	plane := make([]uint8, Width*Height)
	for _, src := range images {
		img, err := pngcheck.Decode(src, Width, Height)
		if err != nil {
			return nil, err
		}
		burnt.Union(plane, opacities(img))
	}
	out := image.NewNRGBA(image.Rect(0, 0, Width, Height))
	for i, a := range plane {
		out.SetNRGBA(i%Width, i/Width, color.NRGBA{R: burnt.Red, G: burnt.Green, B: burnt.Blue, A: a})
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return nil, fmt.Errorf("encoding the burnt-area image: %w", err)
	}
	return buf.Bytes(), nil
}

// opacities answers each pixel's opacity, row by row.
func opacities(img image.Image) []uint8 {
	b := img.Bounds()
	out := make([]uint8, 0, b.Dx()*b.Dy())
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			out = append(out, uint8(a>>alphaShift))
		}
	}
	return out
}

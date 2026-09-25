// Package clouds reaches EUMETSAT's world cloud map (REQUIREMENTS.md 3.2.10):
// it reads the newest valid time the layer lists (FR-CLD-004), fetches that
// time's image (FR-CLD-016), refuses anything that is not the image asked for
// (FR-CLD-012) and draws it by the domain's rules (FR-CLD-006, FR-CLD-007), so
// the page is handed finished pixels.
package clouds

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/url"
	"time"

	"github.com/oernster/EarthNow/internal/domain/cloud"
	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch"
	"github.com/oernster/EarthNow/internal/infrastructure/pngcheck"
)

// Host is the one host the cloud layer reaches (NFR-PRIV-001).
const Host = "view.eumetsat.int"

// BaseURL is the service's GeoServer root.
const BaseURL = "https://" + Host + "/geoserver"

// The layer and the image FR-CLD-016 asks for: the whole world in CRS:84, at
// 2048 x 1024, as measured by the cloud spike (ASM-008). A replay asks for its
// images at half that (FR-RPL-015): measured 0.45 MB in 0.6 s each against
// 1.6 MB in 1.1 to 1.5 s.
const (
	layerWorkspace = "mumi"
	layerName      = "worldcloudmap_ir108"
	crs            = "CRS:84"
	bbox           = "-180,-90,180,90"
	Width          = 2048
	Height         = 1024
	ReplayWidth    = Width / 2
	ReplayHeight   = Height / 2
	wmsVersion     = "1.3.0"
	pngType        = "image/png"
)

// timeDimension names the WMS dimension listing the valid times.
const timeDimension = "time"

// Sentinel failures, each worded where it is raised.
var ErrNoTime = errors.New("the layer lists no valid time")

// Source implements ports.CloudSource over the service at base.
type Source struct {
	get  httpfetch.Getter
	base string
}

// New builds the adapter; base is BaseURL outside tests.
func New(get httpfetch.Getter, base string) *Source {
	return &Source{get: get, base: base}
}

// Latest implements ports.CloudSource: the time dimension's default, which is
// the newest valid time (measured 2026-09-24). The layer's own capabilities
// document is read (6.4 KB) rather than the whole service's (282 KB).
func (s *Source) Latest(ctx context.Context) (time.Time, error) {
	q := url.Values{"service": {"WMS"}, "version": {wmsVersion}, "request": {"GetCapabilities"}}
	raw := s.base + "/" + layerWorkspace + "/" + layerName + "/ows?" + q.Encode()
	body, err := httpfetch.Fresh(ctx, s.get, raw, httpfetch.AcceptXML)
	if err != nil {
		return time.Time{}, err
	}
	return newestTime(body)
}

// Image implements ports.CloudSource: validTime's image, drawn.
func (s *Source) Image(ctx context.Context, validTime time.Time) ([]byte, error) {
	return s.image(ctx, validTime, Width, Height)
}

// ReplayImage implements ports.CloudSource: validTime's image at the replay's
// size, drawn (FR-RPL-015).
func (s *Source) ReplayImage(ctx context.Context, validTime time.Time) ([]byte, error) {
	return s.image(ctx, validTime, ReplayWidth, ReplayHeight)
}

func (s *Source) image(ctx context.Context, validTime time.Time, width, height int) ([]byte, error) {
	q := url.Values{
		"service": {"WMS"}, "version": {wmsVersion}, "request": {"GetMap"},
		"layers": {layerWorkspace + ":" + layerName}, "styles": {""},
		"crs": {crs}, "bbox": {bbox},
		"width": {fmt.Sprint(width)}, "height": {fmt.Sprint(height)},
		"format": {pngType}, "transparent": {"true"},
		"time": {validTime.UTC().Format(time.RFC3339)},
	}
	body, err := httpfetch.Fresh(ctx, s.get, s.base+"/wms?"+q.Encode(), pngType)
	if err != nil {
		return nil, err
	}
	return Draw(body, width, height)
}

// newestTime reads the default of the layer's time dimension.
func newestTime(doc []byte) (time.Time, error) {
	dec := xml.NewDecoder(bytes.NewReader(doc))
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			return time.Time{}, ErrNoTime
		}
		if err != nil {
			return time.Time{}, fmt.Errorf("reading the cloud layer's capabilities: %w", err)
		}
		el, ok := tok.(xml.StartElement)
		if !ok || el.Name.Local != "Dimension" || attr(el, "name") != timeDimension {
			continue
		}
		at, err := time.Parse(time.RFC3339, attr(el, "default"))
		if err != nil {
			return time.Time{}, fmt.Errorf("%w: %v", ErrNoTime, err)
		}
		return at, nil
	}
}

func attr(el xml.StartElement, name string) string {
	for _, a := range el.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

// Draw turns the service's image into the one the globe shows: each pixel
// shaded by cloud.Shade, whose brightness is the grey source's red channel and
// whose data is absent where the source is transparent (measured 2026-09-24:
// every pixel grey, alpha only 0 or 255). An answer that is not a PNG of the
// requested size is refused (FR-CLD-012); the service's errors arrive as XML
// with status 200, which fails here.
func Draw(src []byte, width, height int) ([]byte, error) {
	img, err := pngcheck.Decode(src, width, height)
	if err != nil {
		return nil, err
	}
	out := image.NewNRGBA(image.Rect(0, 0, Width, Height))
	b := img.Bounds()
	for y := range Height {
		for x := range Width {
			c := color.NRGBAModel.Convert(img.At(b.Min.X+x, b.Min.Y+y)).(color.NRGBA)
			p := cloud.Shade(c.R, c.A != 0)
			out.SetNRGBA(x, y, color.NRGBA{R: p.R, G: p.G, B: p.B, A: p.A})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return nil, fmt.Errorf("encoding the cloud image: %w", err)
	}
	return buf.Bytes(), nil
}

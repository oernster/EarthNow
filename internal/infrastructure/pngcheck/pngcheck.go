// Package pngcheck decodes an image a map service answered, refusing anything
// that is not a PNG of the size asked for (REQUIREMENTS.md FR-CLD-012,
// FR-BA-013). Such services answer their errors as XML with status 200.
package pngcheck

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
)

// Sentinel failures, each worded where it is raised.
var (
	ErrNotAnImage = errors.New("the answer is not a PNG image")
	ErrWrongSize  = errors.New("the image is not the size asked for")
)

// Decode answers src decoded when it is a PNG of width by height. The size is
// checked before the pixels are decoded, so a foreign header never sets how
// much is set aside.
func Decode(src []byte, width, height int) (image.Image, error) {
	cfg, err := png.DecodeConfig(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotAnImage, err)
	}
	if cfg.Width != width || cfg.Height != height {
		return nil, fmt.Errorf("%w: %d x %d", ErrWrongSize, cfg.Width, cfg.Height)
	}
	img, err := png.Decode(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotAnImage, err)
	}
	return img, nil
}

package pngcheck

import (
	"bytes"
	"errors"
	"image"
	"image/png"
	"testing"
)

func encoded(t *testing.T, w, h int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewNRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestFRCLD012_FRBA013_OnlyAPNGOfTheSizeAskedForIsDecoded(t *testing.T) {
	t.Parallel()
	good := encoded(t, 4, 2)
	if img, err := Decode(good, 4, 2); err != nil || img.Bounds().Dx() != 4 {
		t.Fatalf("a good image: %v", err)
	}
	cases := []struct {
		name string
		src  []byte
		want error
	}{
		{"an empty body served with 200", nil, ErrNotAnImage},
		{"an XML exception served with 200", []byte(`<?xml version="1.0"?><ServiceExceptionReport/>`), ErrNotAnImage},
		{"a PNG of another size", encoded(t, 2, 1), ErrWrongSize},
		{"a PNG cut short", good[:len(good)-12], ErrNotAnImage},
	}
	for _, c := range cases {
		if _, err := Decode(c.src, 4, 2); !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
	}
}

package services

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/window"
)

// The burnt-area layer's image: the drawn days of a span composed as one
// (FR-BA-006), for the window ending now or a replay's days so far (FR-RPL-013).

// drawn answers the days that drew anything.
func drawn(days []ports.BurntDay) []ports.BurntDay {
	var out []ports.BurntDay
	for _, d := range days {
		if d.Drawn {
			out = append(out, d)
		}
	}
	return out
}

// imageKey names the days drawn and when each arrived, so a new key means a
// new image.
func imageKey(days []ports.BurntDay) string {
	parts := make([]string, len(days))
	for i, d := range days {
		parts[i] = d.Day.UTC().Format(dayKeyLayout) + "@" + d.RetrievedAt.UTC().Format(time.RFC3339Nano)
	}
	return strings.Join(parts, ",")
}

// Image answers the window's drawn days as one image, a data URL the page can
// draw; empty when none draws.
func (b *BurntAreas) Image() string {
	b.mu.Lock()
	w, now := b.win, b.clock.Now()
	b.mu.Unlock()
	return b.ImageAt(w, now, now)
}

// KeyAt names the image ImageAt answers, so the page asks for it only when it
// changes (FR-RPL-013).
func (b *BurntAreas) KeyAt(w window.Window, end, until time.Time) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	held, _ := b.heldFor(w, end, until)
	return imageKey(drawn(held))
}

// ImageAt answers the drawn days of w's span ending at end, up to until's day,
// as one image (FR-BA-006, FR-RPL-013). Each composed image is kept by its key
// until the held days change, since composing decodes every day.
func (b *BurntAreas) ImageAt(w window.Window, end, until time.Time) string {
	b.mu.Lock()
	held, _ := b.heldFor(w, end, until)
	shown := drawn(held)
	key := imageKey(shown)
	if url, ok := b.composed[key]; ok {
		defer b.mu.Unlock()
		return url
	}
	b.mu.Unlock()
	images := make([][]byte, len(shown))
	for i, d := range shown {
		images[i] = d.PNG
	}
	url := ""
	if len(images) > 0 {
		img, err := b.source.Compose(images)
		if err != nil {
			b.mu.Lock()
			defer b.mu.Unlock()
			b.notice = "The burnt-area maps could not be drawn: " + err.Error()
			return ""
		}
		url = pngDataURL + base64.StdEncoding.EncodeToString(img)
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.composed[key] = url
	return url
}

package httpfetch_test

import (
	"context"
	"errors"
	"testing"

	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch"
	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch/fetchtest"
)

func TestFresh_AnswersTheBodyAskedFor(t *testing.T) {
	t.Parallel()
	g := fetchtest.Answering([]byte("pixels"))
	body, err := httpfetch.Fresh(context.Background(), g, "https://example.org/map", "image/png")
	if err != nil || string(body) != "pixels" {
		t.Fatalf("got %q, %v", body, err)
	}
	if g.Asked[0] != "https://example.org/map" || g.Accepts[0] != "image/png" {
		t.Fatalf("asked %v accepting %v", g.Asked, g.Accepts)
	}
}

func TestFresh_RefusesNotModifiedAndPassesFailuresOn(t *testing.T) {
	t.Parallel()
	g := &fetchtest.Getter{Resp: httpfetch.Response{NotModified: true}}
	if _, err := httpfetch.Fresh(context.Background(), g, "https://example.org", "image/png"); !errors.Is(err, httpfetch.ErrNotChanged) {
		t.Fatalf("err = %v, want ErrNotChanged", err)
	}
	down := errors.New("offline")
	g = &fetchtest.Getter{Err: down}
	if _, err := httpfetch.Fresh(context.Background(), g, "https://example.org", "image/png"); !errors.Is(err, down) {
		t.Fatalf("err = %v, want the getter's", err)
	}
}

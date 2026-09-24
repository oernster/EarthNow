package main

import (
	"log"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/oernster/EarthNow/internal/application/dto"
)

// The image layers drawn over the globe: the clouds (3.2.10) and the burnt
// areas (3.2.13). Each is fetched on the driver's loop only while shown;
// the page is told when to ask for its state again.

// cloudsEvent tells the page to ask for the cloud layer's state again.
const cloudsEvent = "clouds-changed"

// burntEvent tells the page to ask for the burnt-area layer's state again.
const burntEvent = "burnt-changed"

// fetchClouds runs one cloud check and logs what happened (NFR-OBS-001). A
// panic is recorded as a failure first, as fetch does, so the check is retried.
func (a *App) fetchClouds() {
	defer func() {
		if r := recover(); r != nil {
			_, _ = a.clouds.Refresh(canceled())
			a.nudge()
			panic(r)
		}
	}()
	log.Printf("clouds: check start")
	t0 := time.Now()
	fresh, err := a.clouds.Refresh(a.ctx)
	switch {
	case err != nil:
		log.Printf("clouds: check failed after %v: %v; next attempt %v", time.Since(t0), err, a.clouds.NextWake())
	case fresh:
		log.Printf("clouds: new image in %v", time.Since(t0))
	default:
		log.Printf("clouds: no newer image (%v)", time.Since(t0))
	}
	wruntime.EventsEmit(a.ctx, cloudsEvent)
	a.nudge()
}

// fetchBurnt runs one burnt-area round and logs what happened (NFR-OBS-001),
// recording a panic as a failure first so the round is retried.
func (a *App) fetchBurnt() {
	defer func() {
		if r := recover(); r != nil {
			_, _ = a.burnt.Refresh(canceled())
			a.nudge()
			panic(r)
		}
	}()
	log.Printf("burnt areas: round start")
	t0 := time.Now()
	got, err := a.burnt.Refresh(a.ctx)
	if err != nil {
		log.Printf("burnt areas: %d days in %v, then failed: %v; next attempt %v", got, time.Since(t0), err, a.burnt.NextWake())
	} else {
		log.Printf("burnt areas: %d days in %v", got, time.Since(t0))
	}
	wruntime.EventsEmit(a.ctx, burntEvent)
	a.nudge()
}

// Clouds answers the cloud layer's state (FR-CLD-009, FR-CLD-013).
func (a *App) Clouds() dto.Clouds { return a.clouds.Status() }

// CloudImage answers the held cloud image as a data URL; empty when none.
func (a *App) CloudImage() string { return a.clouds.Image() }

// BurntAreas answers the burnt-area layer's state (FR-BA-008, FR-BA-009).
func (a *App) BurntAreas() dto.BurntAreas { return a.burnt.Status() }

// BurntImage answers the window's burnt areas as one data URL; empty when none
// draws (FR-BA-006).
func (a *App) BurntImage() string { return a.burnt.Image() }

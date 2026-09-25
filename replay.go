package main

import (
	"log"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/oernster/EarthNow/internal/application/dto"
)

// Replay (3.2.14): the page holds the position along the span and the span's
// end; these answer what the replay shows and fetch its cloud images on the
// driver's loop.

// replayEvent tells the page a replay cloud image has arrived.
const replayEvent = "replay-changed"

// ReplayFrame answers the replay at a position along the chosen window's span
// ending at endMs, Unix milliseconds (FR-RPL-009 to 019).
func (a *App) ReplayFrame(windowKey string, hiddenCategories, hiddenProviders []string, endMs int64, position float64) dto.ReplayFrame {
	frame, due := a.replay.Frame(windowKey, filterOf(hiddenCategories, hiddenProviders), time.UnixMilli(endMs), position)
	if due {
		a.nudge()
	}
	return frame
}

// ReplayCloudImage answers a replay cloud image by its valid time; empty when
// not held (FR-RPL-014).
func (a *App) ReplayCloudImage(validTime string) string { return a.replayClouds.Image(validTime) }

// ReplayBurntImage answers the burnt areas of the span's days so far
// (FR-RPL-013).
func (a *App) ReplayBurntImage(windowKey string, endMs int64, position float64) string {
	return a.replay.BurntImage(windowKey, time.UnixMilli(endMs), position)
}

// EndReplay releases the replay's cloud images as the scrubber returns to its
// end (FR-RPL-018).
func (a *App) EndReplay() { a.replay.End() }

// fetchReplayClouds runs one replay cloud round and logs it (NFR-OBS-001),
// recording a panic as a failure first so the round is not left running.
func (a *App) fetchReplayClouds() {
	defer func() {
		if r := recover(); r != nil {
			_ = a.replayClouds.Refresh(canceled())
			a.nudge()
			panic(r)
		}
	}()
	t0 := time.Now()
	if err := a.replayClouds.Refresh(a.ctx); err != nil {
		log.Printf("replay clouds: round failed after %v: %v", time.Since(t0), err)
	}
	wruntime.EventsEmit(a.ctx, replayEvent)
	a.nudge()
}

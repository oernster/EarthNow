// Storm trails (FR-TRL-001 to 003): which events draw one and how it looks. The
// positions arrive from the Go side already clipped to the window, oldest
// first and ending at the marker, so the page only draws them.
import {MARKER_ALTITUDE} from './markers'
import type {EventDTO} from './types'

// The fewest positions that make a line.
const MIN_POINTS = 2

// FR-TRL-002: faintest at the oldest fix, strongest at the marker, in the
// markers' neutral white so a track reads as the storm's and adds no colour of
// its own (a look the real window confirms by eye).
export const TRAIL_COLOURS = ['rgba(255, 255, 255, 0.08)', 'rgba(255, 255, 255, 0.85)']

// The line runs at the markers' altitude, so it ends in the storm's own marker.
export const TRAIL_ALTITUDE = MARKER_ALTITUDE

/** trailed answers the events that draw a trail: none while switched off (FR-TRL-003). */
export function trailed(events: readonly EventDTO[], shown: boolean): EventDTO[] {
    return shown ? events.filter(e => e.trail.length >= MIN_POINTS) : []
}

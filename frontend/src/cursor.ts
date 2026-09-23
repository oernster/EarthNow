// The globe's keyboard cursor (NFR-KBD-004): Up and Down walk the displayed
// events in the order the view gives them (newest first), wrapping at both ends.
// The cursor is held by event id, so a refresh that reorders or replaces the
// list carries it with the event rather than with a position.

/**
 * stepCursor answers the id the cursor moves to; null when there is nothing to
 * walk. From no cursor (or from an event no longer shown) Down enters at the
 * newest and Up at the oldest.
 */
export function stepCursor(ids: readonly string[], current: string | null, delta: 1 | -1): string | null {
    if (ids.length === 0) return null
    const at = current === null ? -1 : ids.indexOf(current)
    if (at < 0) return delta > 0 ? ids[0] : ids[ids.length - 1]
    return ids[(at + delta + ids.length) % ids.length]
}

// FR-GLB-006's altitude limits, in globe radii above the surface.
export const MIN_ALTITUDE = 0.15
export const MAX_ALTITUDE = 4.0

// ZOOM_FACTOR is one plus or minus press: the altitude divides or multiplies by it.
export const ZOOM_FACTOR = 1.25

/** zoomed answers the altitude after one zoom press, held inside the limits. */
export function zoomed(altitude: number, zoomIn: boolean): number {
    const next = zoomIn ? altitude / ZOOM_FACTOR : altitude * ZOOM_FACTOR
    return Math.min(MAX_ALTITUDE, Math.max(MIN_ALTITUDE, next))
}

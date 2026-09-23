import {describe, expect, it} from 'vitest'
import {MAX_ALTITUDE, MIN_ALTITUDE, stepCursor, zoomed} from './cursor'

// The view lists events newest first, so this is the order the cursor walks.
const ids = ['newest', 'middle', 'oldest']

describe('the globe cursor', () => {
    it('NFR-KBD-004 enters at the newest on Down and the oldest on Up', () => {
        expect(stepCursor(ids, null, 1)).toBe('newest')
        expect(stepCursor(ids, null, -1)).toBe('oldest')
    })

    it('NFR-KBD-004 walks newest first and wraps at both ends', () => {
        expect(stepCursor(ids, 'newest', 1)).toBe('middle')
        expect(stepCursor(ids, 'oldest', 1)).toBe('newest')
        expect(stepCursor(ids, 'newest', -1)).toBe('oldest')
    })

    it('re-enters when its event has left the view; quiet over none', () => {
        expect(stepCursor(ids, 'gone', 1)).toBe('newest')
        expect(stepCursor([], 'newest', 1)).toBeNull()
    })
})

describe('keyboard zoom', () => {
    it('FR-GLB-006 holds the altitude between its limits', () => {
        expect(zoomed(1, true)).toBeLessThan(1)
        expect(zoomed(1, false)).toBeGreaterThan(1)
        expect(zoomed(MIN_ALTITUDE, true)).toBe(MIN_ALTITUDE)
        expect(zoomed(MAX_ALTITUDE, false)).toBe(MAX_ALTITUDE)
    })
})

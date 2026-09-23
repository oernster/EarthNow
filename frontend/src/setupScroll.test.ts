// The setup program's auto-scroll (installer/frontend/dist/setup-scroll.js), ported
// with PigeonPost's autoScroll.test.ts. The setup page has no build step, so the
// script is plain JS; it exports itself to CommonJS outside the page, which is how
// it is loaded here. createRequire takes a path rather than import.meta.url, which
// is not a file URL under jsdom (measured).
import {createRequire} from 'node:module'
import {join} from 'node:path'
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'

interface State { phase: string; waitMs: number; ticksToStep: number }
interface View { scrollTop: number; maxScrollTop: number }
interface Machine {
    TICK_MS: number; START_HOLD_MS: number; DESCENT_PX: number; DESCENT_TICKS_PER_STEP: number
    BOTTOM_HOLD_MS: number; REWIND_PX: number; TOP_HOLD_MS: number; MANUAL_HOLD_MS: number
    initialAutoScrollState: () => State
    suspended: (s: State) => State
    autoScrollTick: (s: State, v: View) => {state: State; delta: number}
    startAutoScroll: (el: HTMLElement) => () => void
}

const load = createRequire(join(process.cwd(), 'package.json'))
const m = load('../installer/frontend/dist/setup-scroll.js') as Machine

const view = (scrollTop: number, maxScrollTop = 1000): View => ({scrollTop, maxScrollTop})

function run(state: State, ticks: number, start = 0, max = 1000) {
    let scrollTop = start
    let current = state
    for (let i = 0; i < ticks; i++) {
        const result = m.autoScrollTick(current, view(scrollTop, max))
        current = result.state
        scrollTop += result.delta
    }
    return {state: current, scrollTop}
}

const ticksFor = (ms: number) => Math.ceil(ms / m.TICK_MS)

describe('setup licence auto-scroll: opening', () => {
    it('holds still on open, before the first descent', () => {
        const {state, scrollTop} = run(m.initialAutoScrollState(), ticksFor(m.START_HOLD_MS) - 1)
        expect(state.phase).toBe('pauseTop')
        expect(scrollTop).toBe(0)
    })

    it('starts reading down once the start hold has run', () => {
        expect(run(m.initialAutoScrollState(), ticksFor(m.START_HOLD_MS)).state.phase).toBe('down')
    })
})

describe('setup licence auto-scroll: the reading pass', () => {
    const reading = (): State => ({phase: 'down', waitMs: 0, ticksToStep: m.DESCENT_TICKS_PER_STEP})

    it('advances a pixel every second tick, not every tick', () => {
        expect(run(reading(), 1).scrollTop).toBe(0)
        expect(run(reading(), 2).scrollTop).toBe(m.DESCENT_PX)
        expect(run(reading(), 20).scrollTop).toBe(10 * m.DESCENT_PX)
    })

    it('stops exactly at the end and holds there', () => {
        const {state, scrollTop} = run(reading(), 4, 998, 1000)
        expect(scrollTop).toBe(1000)
        expect(state.phase).toBe('pauseBottom')
        expect(state.waitMs).toBe(m.BOTTOM_HOLD_MS)
    })

    it('lands on the end without overshooting a part-pixel remainder', () => {
        expect(run(reading(), 2, 999.5, 1000).scrollTop).toBe(1000)
    })
})

describe('setup licence auto-scroll: the rewind', () => {
    const rewinding = (): State => ({phase: 'up', waitMs: 0, ticksToStep: m.DESCENT_TICKS_PER_STEP})

    it('travels far faster than the reading pass', () => {
        expect(run(rewinding(), 1, 500).scrollTop).toBe(500 - m.REWIND_PX)
        expect(m.REWIND_PX).toBeGreaterThan(m.DESCENT_PX * m.DESCENT_TICKS_PER_STEP)
    })

    it('settles at the top and holds before the next pass', () => {
        const {state, scrollTop} = run(rewinding(), 1, m.REWIND_PX - 1)
        expect(scrollTop).toBe(0)
        expect(state.phase).toBe('pauseTop')
        expect(state.waitMs).toBe(m.TOP_HOLD_MS)
    })

    it('goes back to reading after the top hold', () => {
        const held = {phase: 'pauseTop', waitMs: m.TOP_HOLD_MS, ticksToStep: 1}
        expect(run(held, ticksFor(m.TOP_HOLD_MS)).state.phase).toBe('down')
    })

    it('waits at the bottom, then rewinds rather than reading on', () => {
        const held = {phase: 'pauseBottom', waitMs: m.BOTTOM_HOLD_MS, ticksToStep: 1}
        expect(run(held, ticksFor(m.BOTTOM_HOLD_MS) - 1, 1000).state.phase).toBe('pauseBottom')
        expect(run(held, ticksFor(m.BOTTOM_HOLD_MS), 1000).state.phase).toBe('up')
    })
})

describe('setup licence auto-scroll: manual reading', () => {
    it('holds still for the whole suspension, then resumes from where the reader left it', () => {
        const held = m.suspended(m.initialAutoScrollState())
        expect(held.waitMs).toBe(m.MANUAL_HOLD_MS)
        expect(run(held, ticksFor(m.MANUAL_HOLD_MS) - 1, 300)).toEqual({state: expect.objectContaining({phase: 'manual'}), scrollTop: 300})
        expect(run(held, ticksFor(m.MANUAL_HOLD_MS), 300).state.phase).toBe('down')
    })

    it('rewinds instead when the reader has already scrolled to the very end', () => {
        expect(run(m.suspended(m.initialAutoScrollState()), ticksFor(m.MANUAL_HOLD_MS), 1000).state.phase).toBe('up')
    })

    it('consumes nothing while the content fits', () => {
        const opening = m.initialAutoScrollState()
        const {state, delta} = m.autoScrollTick(opening, view(0, 0))
        expect(state).toBe(opening)
        expect(delta).toBe(0)
    })
})

describe('setup licence auto-scroll: the driver', () => {
    beforeEach(() => vi.useFakeTimers())
    afterEach(() => vi.useRealTimers())

    function surface(): HTMLElement {
        const el = document.createElement('pre')
        Object.defineProperty(el, 'clientHeight', {configurable: true, value: 100})
        Object.defineProperty(el, 'scrollHeight', {configurable: true, value: 1000})
        return el
    }

    it('descends once the start hold has passed and stops when told', () => {
        const el = surface()
        const stop = m.startAutoScroll(el)
        vi.advanceTimersByTime(m.START_HOLD_MS)
        expect(el.scrollTop).toBe(0)
        vi.advanceTimersByTime(m.TICK_MS * m.DESCENT_TICKS_PER_STEP * 10)
        const reached = el.scrollTop
        expect(reached).toBeGreaterThan(0)
        stop()
        vi.advanceTimersByTime(m.START_HOLD_MS)
        expect(el.scrollTop).toBe(reached)
    })

    it('suspends when the reader touches it, then resumes', () => {
        const el = surface()
        const stop = m.startAutoScroll(el)
        vi.advanceTimersByTime(m.START_HOLD_MS + m.TICK_MS * m.DESCENT_TICKS_PER_STEP * 10)
        el.dispatchEvent(new Event('wheel'))
        const reached = el.scrollTop
        vi.advanceTimersByTime(m.MANUAL_HOLD_MS - m.TICK_MS)
        expect(el.scrollTop).toBe(reached)
        vi.advanceTimersByTime(m.TICK_MS * m.DESCENT_TICKS_PER_STEP * 10)
        expect(el.scrollTop).toBeGreaterThan(reached)
        stop()
    })
})

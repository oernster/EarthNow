// The house auto-scroll, in its one home for both pages (FR-HLP-005, CON-007): a
// long reading surface holds still on open, reads itself down slowly, holds at
// the tail, rewinds fast and repeats; any manual reading input suspends it and it
// resumes from wherever the reader left it.
//
// Ported from PigeonPost: the state machine is frontend/src/autoScroll.ts, the
// driver is the DOM half of hooks/useAutoScroll.ts. The pace is the app's, never a
// surface's own.
//
// It lives here because the setup program can embed only what sits under
// installer/; that page has no build step, so it is plain JavaScript typed
// by JSDoc. The application imports this same file (frontend/src/useAutoScroll.ts)
// rather than keeping a copy of its own.

// TICK_MS is the clock. Every hold is counted down in whole ticks.
export const TICK_MS = 40
// START_HOLD_MS is the stillness before the first descent.
export const START_HOLD_MS = 5000
// The descent is one pixel every second tick.
export const DESCENT_PX = 1
export const DESCENT_TICKS_PER_STEP = 2
// BOTTOM_HOLD_MS lets the tail be read before the rewind takes it away.
export const BOTTOM_HOLD_MS = 5000
// REWIND_PX is a reposition, not a reading pass, so it travels fast.
export const REWIND_PX = 15
// TOP_HOLD_MS is the breath before the next pass.
export const TOP_HOLD_MS = 2000
// MANUAL_HOLD_MS is the stillness after manual input before the cycle resumes.
export const MANUAL_HOLD_MS = 2500

/**
 * @typedef {{phase: 'down' | 'pauseBottom' | 'up' | 'pauseTop' | 'manual', waitMs: number, ticksToStep: number}} AutoScrollState
 * @typedef {{scrollTop: number, maxScrollTop: number}} AutoScrollView
 * @typedef {{state: AutoScrollState, delta: number}} AutoScrollStep
 */

/**
 * initialAutoScrollState opens in the top hold seeded with the start hold.
 * @returns {AutoScrollState}
 */
export function initialAutoScrollState() {
    return {phase: 'pauseTop', waitMs: START_HOLD_MS, ticksToStep: DESCENT_TICKS_PER_STEP}
}

/**
 * suspended is the state manual reading input puts the cycle into: a hold that
 * resumes at the reader's own position rather than restarting.
 * @param {AutoScrollState} state
 * @returns {AutoScrollState}
 */
export function suspended(state) {
    return {...state, phase: 'manual', waitMs: MANUAL_HOLD_MS}
}

/**
 * autoScrollTick advances the cycle one tick and reports how far to move,
 * clamped to the bounds. Content that does not overflow consumes nothing.
 * @param {AutoScrollState} state
 * @param {AutoScrollView} view
 * @returns {AutoScrollStep}
 */
export function autoScrollTick(state, view) {
    if (view.maxScrollTop <= 0) return {state, delta: 0}
    if (state.phase === 'down') return descend(state, view)
    if (state.phase === 'up') return rewind(state, view)
    return hold(state, view)
}

/**
 * @param {AutoScrollState} state
 * @param {AutoScrollView} view
 * @returns {AutoScrollStep}
 */
function hold(state, view) {
    const waitMs = state.waitMs - TICK_MS
    if (waitMs > 0) return {state: {...state, waitMs}, delta: 0}
    if (state.phase === 'pauseBottom') return {state: {...state, phase: 'up', waitMs: 0}, delta: 0}
    if (state.phase === 'manual' && view.scrollTop >= view.maxScrollTop) {
        return {state: {...state, phase: 'up', waitMs: 0}, delta: 0}
    }
    return {state: {phase: 'down', waitMs: 0, ticksToStep: DESCENT_TICKS_PER_STEP}, delta: 0}
}

/**
 * @param {AutoScrollState} state
 * @param {AutoScrollView} view
 * @returns {AutoScrollStep}
 */
function descend(state, view) {
    const ticksToStep = state.ticksToStep - 1
    if (ticksToStep > 0) return {state: {...state, ticksToStep}, delta: 0}
    const remaining = view.maxScrollTop - view.scrollTop
    if (remaining <= DESCENT_PX) {
        return {
            state: {phase: 'pauseBottom', waitMs: BOTTOM_HOLD_MS, ticksToStep: DESCENT_TICKS_PER_STEP},
            delta: Math.max(0, remaining),
        }
    }
    return {state: {...state, ticksToStep: DESCENT_TICKS_PER_STEP}, delta: DESCENT_PX}
}

/**
 * @param {AutoScrollState} state
 * @param {AutoScrollView} view
 * @returns {AutoScrollStep}
 */
function rewind(state, view) {
    if (view.scrollTop <= REWIND_PX) {
        return {
            state: {phase: 'pauseTop', waitMs: TOP_HOLD_MS, ticksToStep: DESCENT_TICKS_PER_STEP},
            delta: -view.scrollTop,
        }
    }
    return {state, delta: -REWIND_PX}
}

// MANUAL_EVENTS count as reading by hand. mousedown covers a press on the native
// scrollbar; focusin covers keyboard arrival.
const MANUAL_EVENTS = ['wheel', 'mousedown', 'touchstart', 'keydown', 'focusin']

/**
 * startAutoScroll makes an element read itself from a fresh cycle and answers
 * the function that stops it. Each opening of a surface starts over, as a modal
 * mounted again does in the reference.
 *
 * active says whether the surface is the one being looked at. While it answers
 * false the surface is FROZEN rather than suspended: the tick is skipped whole,
 * so phase, position and the remaining hold are all still there when it answers
 * true again (a surface beneath another modal, in the reference). Input is
 * ignored while frozen too.
 *
 * @param {HTMLElement} element
 * @param {() => boolean} [active]
 * @returns {() => void}
 */
export function startAutoScroll(element, active = () => true) {
    let state = initialAutoScrollState()
    // A frozen surface has no reader, so nothing reaching it can be one (scroll
    // skill: a frozen surface takes no input at all).
    const onManualInput = () => { if (active()) state = suspended(state) }
    for (const type of MANUAL_EVENTS) element.addEventListener(type, onManualInput, {passive: true})
    const timer = setInterval(() => {
        if (!active()) return
        const view = {scrollTop: element.scrollTop, maxScrollTop: element.scrollHeight - element.clientHeight}
        const next = autoScrollTick(state, view)
        state = next.state
        if (next.delta !== 0) element.scrollTop = view.scrollTop + next.delta
    }, TICK_MS)
    return () => {
        clearInterval(timer)
        for (const type of MANUAL_EVENTS) element.removeEventListener(type, onManualInput)
    }
}

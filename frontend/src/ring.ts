// The keyboard ring (NFR-KBD-001, 002, 006), ported from ED Voyage Companion's
// frontend/src/hooks.ts. Stops are marked with data-stop; the ring is recomputed
// on every move, so a rebuilt list is handled and a control that has just become
// disabled is skipped rather than stalling the ring.
import {useCallback, useEffect, useRef} from 'react'

// TEXT_TYPES are the inputs that own their horizontal arrows for the caret, so
// the ring is left with Tab. The reference exempts every INPUT but a slider;
// EarthNow's inputs are checkboxes and radios, whose Left and Right would
// otherwise stay native and never step the ring (keeb: the horizontal arrows
// step the ring everywhere but a text field).
const TEXT_TYPES = new Set(['text', 'search', 'email', 'url', 'number', 'password'])

function ownsArrows(target: HTMLElement | null): boolean {
    if (target instanceof HTMLTextAreaElement) return true
    return target instanceof HTMLInputElement && TEXT_TYPES.has(target.type)
}

// Enter equals Space (keeb). A browser gives Enter to a button but not to a
// checkbox or radio, so the ring hands those a click.
function isToggle(target: HTMLElement | null): target is HTMLInputElement {
    return target instanceof HTMLInputElement && (target.type === 'checkbox' || target.type === 'radio')
}

/**
 * noClickFocus is the mousedown handler of a region that is a stop on the ring
 * but must never take focus from a click (NFR-KBD-007): the default action of
 * mousedown is what moves focus, so it is refused. Dragging is unaffected,
 * since the globe's controls listen for pointer events rather than mousedown.
 */
export function noClickFocus(ev: {preventDefault: () => void}) {
    ev.preventDefault()
}

/** liveStops answers the usable stops inside root, in document order. */
export function liveStops(root: HTMLElement | null): HTMLElement[] {
    if (!root) return []
    return Array.from(root.querySelectorAll<HTMLElement>('[data-stop]')).filter(
        element =>
            !element.hasAttribute('disabled') &&
            element.getAttribute('aria-hidden') !== 'true' &&
            element.offsetParent !== null,
    )
}

/**
 * walkGroup is the keydown handler of a strip of peer options (the time window,
 * a set of radios): each option is its own stop, so Tab walks them bounded as
 * the ring carries on past the ends; Up and Down walk the same stops wrapping
 * (NFR-KBD-005). It moves focus only. Choosing stays with Enter or Space, so
 * walking never changes a setting under the reader.
 */
export function walkGroup(ev: React.KeyboardEvent<HTMLElement>) {
    if (ev.key !== 'ArrowDown' && ev.key !== 'ArrowUp') return
    ev.preventDefault()
    const options = liveStops(ev.currentTarget)
    if (options.length === 0) return
    const at = options.indexOf(document.activeElement as HTMLElement)
    const delta = ev.key === 'ArrowDown' ? 1 : -1
    const next = at < 0 ? 0 : (at + delta + options.length) % options.length
    options[next].focus()
}

/**
 * useRing wires one explicit focus ring over a container.
 *
 * Tab and Right move forward, Shift+Tab and Left move back; both wrap. The
 * horizontal arrows are tested first so they step the ring everywhere and focus
 * is never trapped. Nothing is focused until the first press: that is the main
 * window's neutral start (NFR-KBD-002).
 */
export function useRing(container: React.RefObject<HTMLElement | null>, enabled = true) {
    // Where the ring was when it last moved; null while it has never moved. Focus
    // can leave the ring (a click on empty space), so stepping from there
    // continues from the last stop rather than jumping back to the first.
    const mark = useRef<number | null>(null)

    const step = useCallback((delta: number) => {
        const list = liveStops(container.current)
        if (list.length === 0) return
        const active = document.activeElement as HTMLElement | null
        const found = active ? list.indexOf(active) : -1
        const neutral = delta > 0 ? 0 : list.length - 1
        const raw = found >= 0 ? found + delta : mark.current === null ? neutral : mark.current + delta
        const next = (raw + list.length) % list.length
        mark.current = next
        list[next].focus()
    }, [container])

    useEffect(() => {
        if (!enabled) return
        const onKey = (event: KeyboardEvent) => {
            const target = event.target as HTMLElement | null
            const forward = !event.shiftKey && (event.key === 'Tab' || event.key === 'ArrowRight')
            const back = event.key === 'ArrowLeft' || (event.key === 'Tab' && event.shiftKey)

            if (event.key === 'Enter' && isToggle(target)) {
                event.preventDefault()
                target.click()
                return
            }
            // A text field keeps its own arrows; Tab still leaves it.
            if (ownsArrows(target) && event.key !== 'Tab') return

            if (forward) {
                event.preventDefault()
                step(1)
            } else if (back) {
                event.preventDefault()
                step(-1)
            }
        }
        // Capture on the document: a key something else consumes never reaches a
        // bubble listener, while the ring has to see every press to be the ring.
        document.addEventListener('keydown', onKey, true)
        return () => document.removeEventListener('keydown', onKey, true)
    }, [step, enabled])

    return {step}
}

/**
 * useFirstStop focuses a surface's first usable stop when it opens and hands
 * focus back to whatever held it before once it closes (NFR-KBD-006). Dialogs
 * are the deliberate opposite of the main window: they start on their first
 * stop, because the user opened them to do the one thing they are for.
 *
 * key re-runs the entry for a surface that stays mounted while its content
 * changes, as the detail panel does from one event to the next.
 */
export function useFirstStop(container: React.RefObject<HTMLElement | null>, key: unknown = null) {
    const opener = useRef<HTMLElement | null>(null)

    useEffect(() => {
        if (opener.current === null) {
            const held = document.activeElement
            opener.current = held instanceof HTMLElement && held !== document.body ? held : null
        }
        liveStops(container.current)[0]?.focus()
    }, [container, key])

    useEffect(() => () => {
        const back = opener.current
        if (back && back.isConnected) back.focus()
    }, [])
}

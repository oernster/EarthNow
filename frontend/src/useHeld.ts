// FR-STS-007: the Refresh button turns while any provider refreshes, for at least
// one turn, so a fetch faster than a turn is still seen. The turn's length lives
// here once; the rail hands it to the stylesheet as a custom property.
import {useEffect, useRef, useState} from 'react'

/** REFRESH_TURN_MS is one turn of the Refresh button's icon (FR-STS-007). */
export const REFRESH_TURN_MS = 1000

/** useHeld answers `on`, held true for at least `minimumMs` once it turns true. */
export function useHeld(on: boolean, minimumMs: number): boolean {
    const [held, setHeld] = useState(on)
    const since = useRef(on ? Date.now() : 0)
    useEffect(() => {
        if (on) {
            if (!held) {
                since.current = Date.now()
                setHeld(true)
            }
            return
        }
        if (!held) return
        const timer = window.setTimeout(() => setHeld(false), Math.max(0, since.current + minimumMs - Date.now()))
        return () => window.clearTimeout(timer)
    }, [on, held, minimumMs])
    return held
}

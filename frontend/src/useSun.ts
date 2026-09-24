// Where the sun stands, for the day and night layer (FR-DAY-004). The page asks
// the Go side, which owns the astronomy, when the layer is shown and then every
// minute while it stays shown; hidden, it asks nothing.
import {useEffect, useState} from 'react'
import {api} from './api'
import type {SunDTO} from './types'

// FR-DAY-004's longest wait between two readings: the sun moves a quarter of a
// degree in it, under a pixel at the globe's fitted size.
export const SUN_POLL_MS = 60_000

export function useSun(shown: boolean, onProblem: (reason: string) => void): SunDTO | null {
    const [sun, setSun] = useState<SunDTO | null>(null)
    useEffect(() => {
        if (!shown) return
        const ask = () => { void api.sun(onProblem).then(s => { if (s) setSun(s) }) }
        ask()
        const timer = window.setInterval(ask, SUN_POLL_MS)
        return () => window.clearInterval(timer)
    }, [shown, onProblem])
    return sun
}

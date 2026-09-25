// The replay's state on the page (3.2.14): where the scrubber rests, whether it
// plays and the span's end, fixed when the replay starts (FR-RPL-003). Time
// travel ends only by Now (FR-RPL-024) or another window (FR-RPL-021). While
// replaying it asks the Go side for a frame at most every FRAME_ASK_MS; for an
// image only when the frame's key for it changes.
import {useCallback, useEffect, useRef, useState} from 'react'
import {api, on} from './api'
import type {ReplayFrameDTO} from './types'

// How often a playing replay asks for a frame.
export const FRAME_ASK_MS = 100
const MS_PER_SECOND = 1000
// The scrubber's right end: the span's end while replaying (FR-RPL-002).
const END = 1

type Refused = (reason: string) => void

export interface Replay {
    replaying: boolean
    position: number
    playing: boolean
    frame: ReplayFrameDTO | null
    cloudImage: string
    burntImage: string
    play: () => void
    pause: () => void
    seek: (position: number) => void
    toNow: () => void
}

// passSeconds is the chosen speed's pass (FR-RPL-025); null until the speeds are
// known, when play does not advance but a seek still shows its frame.
export function useReplay(windowKey: string, hiddenCategories: string[], hiddenProviders: string[],
    passSeconds: number | null, onProblem: Refused, now: () => number = Date.now): Replay {
    const [position, setPosition] = useState(END)
    const [playing, setPlaying] = useState(false)
    const [end, setEnd] = useState<number | null>(null)
    const [frame, setFrame] = useState<ReplayFrameDTO | null>(null)
    const [cloudImage, setCloudImage] = useState('')
    const [burntImage, setBurntImage] = useState('')
    const [asked, setAsked] = useState(0)
    const passMs = passSeconds === null ? null : passSeconds * MS_PER_SECOND

    // FR-RPL-024: Now stops play and shows the ordinary view again.
    const toNow = useCallback(() => {
        setPlaying(false)
        setPosition(END)
        setEnd(null)
        setFrame(null)
        void api.endReplay(onProblem)
    }, [onProblem])

    // FR-RPL-004: from now, play fixes the span and starts it; at the span's end
    // it starts the same span again; else it resumes.
    const play = useCallback(() => {
        if (end === null) {
            setEnd(now())
            setPosition(0)
        } else if (position >= END) {
            setPosition(0)
        }
        setPlaying(true)
    }, [end, position, now])
    const pause = useCallback(() => setPlaying(false), [])
    // FR-RPL-007: moving the scrubber pauses; the right end is the span's end, so
    // a seek there stays in the replay (FR-RPL-002). At now it is already there.
    const seek = useCallback((to: number) => {
        setPlaying(false)
        if (end === null && to >= END) return
        setEnd(e => e ?? now())
        setPosition(to)
    }, [end, now])

    // FR-RPL-021: another window returns to now.
    const shownWindow = useRef(windowKey)
    useEffect(() => {
        if (shownWindow.current === windowKey) return
        shownWindow.current = windowKey
        toNow()
    }, [windowKey, toNow])

    // The play loop: one pass of the span in the chosen speed's pass (FR-RPL-004,
    // FR-RPL-025); reaching the end holds there, paused (FR-RPL-005).
    useEffect(() => {
        if (!playing || passMs === null) return
        // The first frame's own timestamp starts the clock, so no time is counted
        // before the loop runs.
        let last: number | null = null
        let handle = requestAnimationFrame(function tick(t) {
            const step = last === null ? 0 : (t - last) / passMs
            last = t
            setPosition(p => Math.min(p + step, END))
            handle = requestAnimationFrame(tick)
        })
        return () => cancelAnimationFrame(handle)
    }, [playing, passMs])
    useEffect(() => { if (playing && position >= END) setPlaying(false) }, [playing, position])

    // A frame each FRAME_ASK_MS of the pass; again when a cloud image arrives.
    useEffect(() => on('replay-changed', () => setAsked(a => a + 1)), [])
    const at = passMs === null ? position : Math.floor(position * passMs / FRAME_ASK_MS) * FRAME_ASK_MS / passMs
    const latest = useRef(0)
    useEffect(() => {
        if (end === null) return
        const request = ++latest.current
        void api.replayFrame(windowKey, hiddenCategories, hiddenProviders, end, at, onProblem)
            .then(f => { if (f && request === latest.current) setFrame(f) })
    }, [end, at, asked, windowKey, hiddenCategories, hiddenProviders, onProblem])

    const cloudTime = frame?.cloudTime ?? ''
    useEffect(() => {
        if (cloudTime === '') { setCloudImage(''); return }
        void api.replayCloudImage(cloudTime, onProblem).then(url => { if (url !== null) setCloudImage(url) })
    }, [cloudTime, onProblem])
    const burntKey = frame?.burntKey ?? ''
    const place = useRef({windowKey, end, position})
    place.current = {windowKey, end, position}
    useEffect(() => {
        const {windowKey: w, end: e, position: p} = place.current
        if (burntKey === '' || e === null) { setBurntImage(''); return }
        void api.replayBurntImage(w, e, p, onProblem).then(url => { if (url !== null) setBurntImage(url) })
    }, [burntKey, onProblem])

    return {replaying: end !== null, position, playing, frame, cloudImage, burntImage, play, pause, seek, toNow}
}

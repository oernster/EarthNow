// The replay's state on the page (3.2.14): where the scrubber rests, whether it
// plays and the span's end, fixed when the scrubber leaves its right end
// (FR-RPL-003). While replaying it asks the Go side for a frame at most every
// FRAME_ASK_MS; for an image only when the frame's key for it changes.
import {useCallback, useEffect, useRef, useState} from 'react'
import {api, on} from './api'
import type {ReplayFrameDTO} from './types'

// FR-RPL-004: one pass of the whole span.
export const REPLAY_PASS_MS = 30_000
// How often a playing replay asks for a frame: three hundred a pass.
export const FRAME_ASK_MS = 100
// The scrubber's right end, which is now (FR-RPL-002).
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
}

export function useReplay(windowKey: string, hiddenCategories: string[], hiddenProviders: string[],
    onProblem: Refused, now: () => number = Date.now): Replay {
    const [position, setPosition] = useState(END)
    const [playing, setPlaying] = useState(false)
    const [end, setEnd] = useState<number | null>(null)
    const [frame, setFrame] = useState<ReplayFrameDTO | null>(null)
    const [cloudImage, setCloudImage] = useState('')
    const [burntImage, setBurntImage] = useState('')
    const [asked, setAsked] = useState(0)

    const toEnd = useCallback(() => {
        setPlaying(false)
        setPosition(END)
        setEnd(null)
        setFrame(null)
        void api.endReplay(onProblem)
    }, [onProblem])

    // FR-RPL-004: from the end, play starts the span afresh; else it resumes.
    const play = useCallback(() => {
        if (position >= END) {
            setEnd(now())
            setPosition(0)
        }
        setPlaying(true)
    }, [position, now])
    const pause = useCallback(() => setPlaying(false), [])
    // FR-RPL-007: moving the scrubber pauses; reaching the end is now again.
    const seek = useCallback((to: number) => {
        setPlaying(false)
        if (to >= END) {
            toEnd()
            return
        }
        setEnd(e => e ?? now())
        setPosition(to)
    }, [toEnd, now])

    // FR-RPL-021: another window returns the scrubber to its end.
    const shownWindow = useRef(windowKey)
    useEffect(() => {
        if (shownWindow.current === windowKey) return
        shownWindow.current = windowKey
        toEnd()
    }, [windowKey, toEnd])

    // The play loop: one pass of the span in REPLAY_PASS_MS (FR-RPL-004, 005).
    useEffect(() => {
        if (!playing) return
        // The first frame's own timestamp starts the clock, so no time is counted
        // before the loop runs.
        let last: number | null = null
        let handle = requestAnimationFrame(function tick(t) {
            const step = last === null ? 0 : (t - last) / REPLAY_PASS_MS
            last = t
            setPosition(p => Math.min(p + step, END))
            handle = requestAnimationFrame(tick)
        })
        return () => cancelAnimationFrame(handle)
    }, [playing])
    useEffect(() => { if (playing && position >= END) toEnd() }, [playing, position, toEnd])

    // A frame each FRAME_ASK_MS of the pass; again when a cloud image arrives.
    useEffect(() => on('replay-changed', () => setAsked(a => a + 1)), [])
    const tick = Math.floor(position * REPLAY_PASS_MS / FRAME_ASK_MS)
    const latest = useRef(0)
    useEffect(() => {
        if (end === null) return
        const request = ++latest.current
        void api.replayFrame(windowKey, hiddenCategories, hiddenProviders, end, tick * FRAME_ASK_MS / REPLAY_PASS_MS, onProblem)
            .then(f => { if (f && request === latest.current) setFrame(f) })
    }, [end, tick, asked, windowKey, hiddenCategories, hiddenProviders, onProblem])

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

    return {replaying: end !== null, position, playing, frame, cloudImage, burntImage, play, pause, seek}
}

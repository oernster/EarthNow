// Idle rotation (FR-GLB-002 to 004, FR-GLB-012, FR-GLB-018): input on the globe
// stops it; after the idle delay the camera returns to the fit altitude, keeping
// its facing, then turns again. The setting is the only switch. GlobeView owns
// the globe; this owns when it turns.
import {useEffect, useRef} from 'react'
import type {GlobeInstance} from 'globe.gl'
import {FOCUS_MS, nearAltitude} from './cursor'
import {fitOf} from './markers'

// OrbitControls documents speed 2.0 as one orbit in 30 s, so an orbit takes
// ORBIT_SECONDS_AT_UNIT_SPEED / speed; the period comes from the settings.
const ORBIT_SECONDS_AT_UNIT_SPEED = 60
// FR-GLB-002: rotation resumes after this long without input on the globe.
const IDLE_DELAY_MS = 10_000
// FR-GLB-003: the input that stops idle rotation.
const STOPPING_INPUT = ['pointerdown', 'wheel', 'keydown'] as const

export interface IdleRotation {
    // pause stops rotation for input and restarts the idle delay.
    pause: () => void
    // attach starts the globe as the setting wants and listens for input on el;
    // it answers the detach.
    attach: (g: GlobeInstance, el: HTMLElement) => () => void
}

export function useIdleRotation(globe: {current: GlobeInstance | null}, autoRotate: boolean,
    secondsPerRevolution: number): IdleRotation {
    // The settings as last rendered, read when the idle delay runs out and when a
    // globe is attached (which may come after this hook's effect has run).
    const wanted = useRef(autoRotate)
    const speed = useRef(ORBIT_SECONDS_AT_UNIT_SPEED / secondsPerRevolution)
    const timer = useRef<number | null>(null)
    // returning is true while the camera makes its way back to the fit altitude.
    const returning = useRef(false)

    const idle = useRef<IdleRotation | null>(null)
    // returnThenTurn is resume below, kept for the setting being switched on.
    const returnThenTurn = useRef<(() => void) | null>(null)
    if (idle.current === null) {
        const rotate = () => {
            timer.current = null
            returning.current = false
            if (globe.current) globe.current.controls().autoRotate = wanted.current
        }
        // resume runs when the idle delay runs out or the setting is switched on:
        // away from the fit altitude the camera returns to it first, then turns
        // (FR-GLB-018).
        const resume = () => {
            const g = globe.current
            if (!g || !wanted.current || nearAltitude(g.pointOfView().altitude, fitOf(g))) {
                rotate()
                return
            }
            returning.current = true
            g.pointOfView({altitude: fitOf(g)}, FOCUS_MS)
            timer.current = window.setTimeout(rotate, FOCUS_MS)
        }
        // Input during a return stops the camera where it is: setting a point of
        // view ends globe.gl's running tween.
        const pause = () => {
            const g = globe.current
            if (!g) return
            g.controls().autoRotate = false
            if (returning.current) {
                returning.current = false
                g.pointOfView({...g.pointOfView()})
            }
            if (timer.current !== null) window.clearTimeout(timer.current)
            timer.current = window.setTimeout(resume, IDLE_DELAY_MS)
        }
        const attach = (g: GlobeInstance, el: HTMLElement) => {
            g.controls().autoRotateSpeed = speed.current
            g.controls().autoRotate = wanted.current
            STOPPING_INPUT.forEach(name => el.addEventListener(name, pause))
            return () => {
                STOPPING_INPUT.forEach(name => el.removeEventListener(name, pause))
                if (timer.current !== null) window.clearTimeout(timer.current)
            }
        }
        idle.current = {pause, attach}
        returnThenTurn.current = resume
    }

    // Switching the setting on returns to the fit altitude and turns, unless input
    // is still inside its idle delay (FR-GLB-004, FR-GLB-012, FR-GLB-018). A speed
    // change while on is not a switch-on and leaves the camera where it is.
    useEffect(() => {
        const switchedOn = autoRotate && !wanted.current
        wanted.current = autoRotate
        speed.current = ORBIT_SECONDS_AT_UNIT_SPEED / secondsPerRevolution
        const g = globe.current
        if (!g) return
        const controls = g.controls()
        controls.autoRotateSpeed = speed.current
        if (!autoRotate) controls.autoRotate = false
        else if (timer.current !== null) return
        else if (switchedOn) returnThenTurn.current?.()
        else controls.autoRotate = true
    }, [globe, autoRotate, secondsPerRevolution])

    return idle.current
}

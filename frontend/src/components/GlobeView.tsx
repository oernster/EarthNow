// The globe (FR-GLB, FR-MRK): texture, fitted camera, idle rotation, emoji
// markers, hover place lines and selection with a camera focus.
import {forwardRef, useEffect, useImperativeHandle, useRef, useState} from 'react'
import Globe, {type GlobeInstance} from 'globe.gl'
import * as THREE from 'three'
import earthTexture from '../assets/earth.jpg'
import {api} from '../api'
import {categoryOf} from '../categories'
import {MAX_ALTITUDE, MIN_ALTITUDE, stepCursor, zoomed} from '../cursor'
import {fitAltitude, sprite} from '../markers'
import {noClickFocus} from '../ring'
import type {EventDTO} from '../types'

const MARKER_ALTITUDE = 0.01
// NFR-UX-003: the camera focus animation, used by Reset view as well.
const FOCUS_MS = 1000
// One plus or minus press animates over this long.
const ZOOM_MS = 250
// OrbitControls documents speed 2.0 as one orbit in 30 s, so an orbit takes
// ORBIT_SECONDS_AT_UNIT_SPEED / speed; the period comes from the settings.
const ORBIT_SECONDS_AT_UNIT_SPEED = 60
// FR-GLB-002: rotation resumes after this long without input on the globe.
const IDLE_DELAY_MS = 10_000
// FR-GLB-003: the input that stops idle rotation.
const STOPPING_INPUT = ['pointerdown', 'wheel', 'keydown'] as const
const TIP_OFFSET_PX = 14

interface Props {
    events: EventDTO[]
    selectedId: string | null
    autoRotate: boolean
    secondsPerRevolution: number
    onSelect: (e: EventDTO) => void
    onProblem: (reason: string) => void
}

// GlobeHandle is what the rail drives directly.
export interface GlobeHandle {
    resetView: () => void
    // zoom is one plus or minus press from the rail (FR-GLB-006's limits hold).
    zoom: (zoomIn: boolean) => void
}

interface Tip { x: number; y: number; title: string; place: string }

// zoomOnce is one zoom step, the same from a key and from the rail.
function zoomOnce(g: GlobeInstance, zoomIn: boolean) {
    g.pointOfView({altitude: zoomed(g.pointOfView().altitude, zoomIn)}, ZOOM_MS)
}

export const GlobeView = forwardRef<GlobeHandle, Props>(function GlobeView(
    {events, selectedId, autoRotate, secondsPerRevolution, onSelect, onProblem}, ref) {
    const host = useRef<HTMLDivElement>(null)
    const globe = useRef<GlobeInstance | null>(null)
    const hovered = useRef<EventDTO | null>(null)
    const pointer = useRef({x: 0, y: 0})
    const handlers = useRef({onSelect, onProblem})
    handlers.current = {onSelect, onProblem}
    const [tip, setTip] = useState<Tip | null>(null)
    // The setting as last rendered, read when the idle delay runs out.
    const rotationWanted = useRef(autoRotate)
    const idleTimer = useRef<number | null>(null)

    // pause stops rotation for input and restarts it after the idle delay.
    const pause = useRef(() => {
        const g = globe.current
        if (!g) return
        g.controls().autoRotate = false
        if (idleTimer.current !== null) window.clearTimeout(idleTimer.current)
        idleTimer.current = window.setTimeout(() => {
            idleTimer.current = null
            if (globe.current) globe.current.controls().autoRotate = rotationWanted.current
        }, IDLE_DELAY_MS)
    })

    // showTip words an event's tooltip at a point and adds its place line once the
    // lookup answers, unless the tooltip has moved on to another event by then.
    const showing = useRef<EventDTO | null>(null)
    const showTip = useRef((e: EventDTO, at: () => {x: number; y: number}) => {
        showing.current = e
        const title = `${categoryOf(e.category).emoji} ${e.title}`
        setTip({...at(), title, place: ''})
        void api.place(e.lat, e.lng, handlers.current.onProblem).then(place => {
            if (place !== null && showing.current === e) setTip({...at(), title, place})
        })
    })

    // The keyboard cursor (NFR-KBD-004), held by id so a refresh carries it.
    const cursor = useRef<string | null>(null)
    const shown = useRef(events)
    shown.current = events

    const centre = () => {
        const r = host.current?.getBoundingClientRect()
        return r ? {x: r.left + r.width / 2, y: r.top + r.height / 2} : {x: 0, y: 0}
    }

    const onKeyDown = (ev: React.KeyboardEvent) => {
        const g = globe.current
        if (!g) return
        if (ev.key === 'ArrowDown' || ev.key === 'ArrowUp') {
            ev.preventDefault()
            const id = stepCursor(shown.current.map(e => e.id), cursor.current, ev.key === 'ArrowDown' ? 1 : -1)
            const e = shown.current.find(x => x.id === id)
            cursor.current = id
            if (!e) return
            g.pointOfView({lat: e.lat, lng: e.lng}, FOCUS_MS)
            // FR-GEO-006: the cursor shows the tooltip hover would, where the
            // camera brings the event: the middle of the globe area.
            showTip.current(e, centre)
        } else if (ev.key === 'Enter' || ev.key === ' ') {
            ev.preventDefault()
            const e = shown.current.find(x => x.id === cursor.current)
            if (e) handlers.current.onSelect(e)
        } else if (ev.key === '+' || ev.key === '=' || ev.key === '-') {
            ev.preventDefault()
            zoomOnce(g, ev.key !== '-')
        }
    }

    const onBlur = () => {
        if (!hovered.current) setTip(null)
        showing.current = null
    }

    useImperativeHandle(ref, () => ({
        resetView: () => {
            const g = globe.current
            if (g) g.pointOfView({altitude: fitAltitude(g.camera() as THREE.PerspectiveCamera)}, FOCUS_MS)
        },
        zoom: (zoomIn: boolean) => {
            const g = globe.current
            if (g) zoomOnce(g, zoomIn)
        },
    }), [])

    useEffect(() => {
        const el = host.current
        if (!el) return
        const g = new Globe(el)
            .globeImageUrl(earthTexture)
            .backgroundColor('#000000')
            .objectLat('lat').objectLng('lng').objectAltitude(MARKER_ALTITUDE)
            .onObjectClick((d: object) => handlers.current.onSelect(d as EventDTO))
            .onObjectHover((d: object | null) => {
                const e = d as EventDTO | null
                hovered.current = e
                if (!e) { setTip(null); return }
                showTip.current(e, () => pointer.current)
            })
        globe.current = g
        const controls = g.controls()
        controls.autoRotate = rotationWanted.current
        // FR-GLB-006: the wheel zooms between the altitude limits. OrbitControls
        // measures distance from the centre, so a limit is the radius plus it.
        const radius = g.getGlobeRadius()
        controls.minDistance = radius * (1 + MIN_ALTITUDE)
        controls.maxDistance = radius * (1 + MAX_ALTITUDE)
        const camera = g.camera() as THREE.PerspectiveCamera
        const fit = () => {
            g.width(el.clientWidth).height(el.clientHeight)
            g.pointOfView({altitude: fitAltitude(camera)})
        }
        fit()
        const observer = new ResizeObserver(fit)
        observer.observe(el)
        const onMove = (ev: MouseEvent) => {
            pointer.current = {x: ev.clientX, y: ev.clientY}
            if (hovered.current) setTip(t => t && {...t, ...pointer.current})
        }
        el.addEventListener('mousemove', onMove)
        const onInput = () => pause.current()
        STOPPING_INPUT.forEach(name => el.addEventListener(name, onInput))
        return () => {
            observer.disconnect()
            el.removeEventListener('mousemove', onMove)
            STOPPING_INPUT.forEach(name => el.removeEventListener(name, onInput))
            if (idleTimer.current !== null) window.clearTimeout(idleTimer.current)
            g._destructor()
            globe.current = null
        }
    }, [])

    // The setting is the only switch (FR-GLB-004, FR-GLB-012). Switching it on
    // starts rotation at once unless input is still inside its idle delay.
    useEffect(() => {
        rotationWanted.current = autoRotate
        const g = globe.current
        if (!g) return
        const controls = g.controls()
        controls.autoRotateSpeed = ORBIT_SECONDS_AT_UNIT_SPEED / secondsPerRevolution
        if (!autoRotate) controls.autoRotate = false
        else if (idleTimer.current === null) controls.autoRotate = true
    }, [autoRotate, secondsPerRevolution])

    useEffect(() => {
        const g = globe.current
        if (!g) return
        g.objectThreeObject((d: object) => {
            const e = d as EventDTO
            return sprite(e, e.id === selectedId)
        })
        g.objectsData(events)
    }, [events, selectedId])

    useEffect(() => {
        const g = globe.current
        const chosen = events.find(e => e.id === selectedId)
        if (!g || !chosen) return
        pause.current()
        g.pointOfView({lat: chosen.lat, lng: chosen.lng}, FOCUS_MS)
    // Focus once per selection, not on every refresh of the same event.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [selectedId])

    // globe.gl owns the host's children, so the tooltip sits beside it.
    return <>
        <div ref={host} className="globe" data-stop tabIndex={-1} onKeyDown={onKeyDown} onBlur={onBlur}
            onMouseDown={noClickFocus}
            aria-label="Globe. Up and Down walk the events, Enter opens one, plus and minus zoom."/>
        {tip && <div className="tip" style={{left: tip.x + TIP_OFFSET_PX, top: tip.y + TIP_OFFSET_PX}}>
            <div className="tip-title">{tip.title}</div>
            {tip.place && <div className="tip-place">{tip.place}</div>}
        </div>}
    </>
})

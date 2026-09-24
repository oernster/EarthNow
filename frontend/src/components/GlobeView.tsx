// The globe (FR-GLB, FR-MRK, FR-CLD-008): texture, cloud layer, fitted camera,
// idle rotation, emoji markers, hover place lines and selection with a camera focus.
import {forwardRef, useEffect, useImperativeHandle, useRef, useState} from 'react'
import Globe, {type GlobeInstance} from 'globe.gl'
import * as THREE from 'three'
import earthTexture from '../assets/earth.jpg'
import {api} from '../api'
import {categoryCounts, categoryOf} from '../categories'
import {clusterEvents, layoutKey, type MarkerItem, separatingAltitude} from '../clusters'
import {MAX_ALTITUDE, MIN_ALTITUDE, stepCursor, zoomed} from '../cursor'
import {clusterSprite, fitAltitude, MARKER_ALTITUDE, rescale, sprite, viewHalfAngle} from '../markers'
import {makeCloudSphere, showCloudImage} from '../cloudLayer'
import {noClickFocus} from '../ring'
import type {EventDTO} from '../types'

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
// GLOBE_KEYS names the globe's keys (NFR-KBD-004), for its accessible name and
// for the tooltip while it holds focus.
const GLOBE_KEYS = 'Up and Down walk the events, Enter opens one, plus and minus zoom.'
// NO_EVENTS is what the cursor says when the window holds nothing to walk.
const NO_EVENTS = 'No events in this time window'
// FR-GLB-009: said in place of the globe where WebGL2 is missing.
export const NO_WEBGL2 = 'The globe needs WebGL2, which this computer does not offer. Updating the graphics driver usually adds it; the rest of the window still works.'

// hasWebGL2 reports whether the page can draw the globe at all.
export function hasWebGL2(): boolean {
    return document.createElement('canvas').getContext('webgl2') !== null
}

interface Props {
    events: EventDTO[]
    selectedId: string | null
    autoRotate: boolean
    secondsPerRevolution: number
    // cloudImage is the drawn cloud image as a data URL; empty draws none.
    cloudImage: string
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

// markerRadius is the sphere the markers sit on, in globe units.
function markerRadius(g: GlobeInstance): number {
    return g.getGlobeRadius() * (1 + MARKER_ALTITUDE)
}

function eventTitle(e: EventDTO): string {
    return `${categoryOf(e.category).emoji} ${e.title}`
}

// clusterTitle counts a cluster's members by category, largest first.
function clusterTitle(members: readonly EventDTO[]): string {
    const parts = categoryCounts(members).map(({category, count}) => `${category.emoji} ${count}`)
    return `${members.length} events: ${parts.join('  ')}`
}

export const GlobeView = forwardRef<GlobeHandle, Props>(function GlobeView(
    {events, selectedId, autoRotate, secondsPerRevolution, cloudImage, onSelect, onProblem}, ref) {
    const host = useRef<HTMLDivElement>(null)
    const [webgl2] = useState(hasWebGL2)
    const globe = useRef<GlobeInstance | null>(null)
    const hovered = useRef<MarkerItem | null>(null)
    const pointer = useRef({x: 0, y: 0})
    const handlers = useRef({onSelect, onProblem})
    handlers.current = {onSelect, onProblem}
    const [tip, setTip] = useState<Tip | null>(null)
    // keyboard is true while the globe holds focus, so the tooltip names the keys.
    const [keyboard, setKeyboard] = useState(false)
    // The setting as last rendered, read when the idle delay runs out.
    const rotationWanted = useRef(autoRotate)
    const idleTimer = useRef<number | null>(null)
    const clouds = useRef<THREE.Mesh | null>(null)

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

    // showTip words a marker's tooltip at a point and adds its place line once the
    // lookup answers, unless the tooltip has moved on to another marker by then.
    const showing = useRef<object | null>(null)
    const showTip = useRef((title: string, where: {lat: number; lng: number}, at: () => {x: number; y: number}) => {
        showing.current = where
        setTip({...at(), title, place: ''})
        void api.place(where.lat, where.lng, handlers.current.onProblem).then(place => {
            if (place !== null && showing.current === where) setTip({...at(), title, place})
        })
    })

    // The marker layout (FR-MRK-007): what is drawn at the current scale, the
    // altitude over the fit altitude, with the sprites made for it so a zoom can
    // resize them in place rather than drawing them again.
    const shown = useRef(events)
    shown.current = events
    const selected = useRef(selectedId)
    selected.current = selectedId
    const fitted = useRef(1)
    const scale = useRef(1)
    const sprites = useRef<THREE.Object3D[]>([])
    const laidOut = useRef('')
    const layout = (force: boolean) => {
        const g = globe.current
        if (!g) return
        const items = clusterEvents(shown.current, scale.current, markerRadius(g), selected.current)
        const key = layoutKey(items)
        if (!force && key === laidOut.current) return
        laidOut.current = key
        sprites.current = []
        g.objectsData(items)
    }
    const relayout = useRef(layout)
    relayout.current = layout

    // The keyboard cursor (NFR-KBD-004), held by id so a refresh carries it.
    const cursor = useRef<string | null>(null)

    const centre = () => {
        const r = host.current?.getBoundingClientRect()
        return r ? {x: r.left + r.width / 2, y: r.top + r.height / 2} : {x: 0, y: 0}
    }

    // walk moves the keyboard cursor one event and brings the camera to it.
    const walk = (delta: 1 | -1) => {
        const g = globe.current
        if (!g) return
        pause.current()
        const id = stepCursor(shown.current.map(e => e.id), cursor.current, delta)
        const e = shown.current.find(x => x.id === id)
        cursor.current = id
        if (!e) {
            setTip({...centre(), title: NO_EVENTS, place: ''})
            return
        }
        g.pointOfView({lat: e.lat, lng: e.lng}, FOCUS_MS)
        // FR-GEO-006: the cursor shows the tooltip hover would, where the
        // camera brings the event: the middle of the globe area.
        showTip.current(eventTitle(e), e, centre)
    }

    // Arriving by Tab shows where the ring landed (owner): the globe paints no
    // ring (NFR-KBD-007), so the cursor goes to an event at once, as Down would,
    // and the tooltip carries the keys. Returning resumes the event it left on.
    const onFocus = () => {
        setKeyboard(true)
        const current = shown.current.find(x => x.id === cursor.current)
        if (current) {
            showTip.current(eventTitle(current), current, centre)
        } else {
            walk(1)
        }
    }

    const onKeyDown = (ev: React.KeyboardEvent) => {
        const g = globe.current
        if (!g) return
        if (ev.key === 'ArrowDown' || ev.key === 'ArrowUp') {
            ev.preventDefault()
            walk(ev.key === 'ArrowDown' ? 1 : -1)
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
        setKeyboard(false)
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
        if (!el || !webgl2) return
        const g = new Globe(el)
            .globeImageUrl(earthTexture)
            .backgroundColor('#000000')
            .objectLat('lat').objectLng('lng').objectAltitude(MARKER_ALTITUDE)
            .objectThreeObject((d: object) => {
                const item = d as MarkerItem
                const made = item.kind === 'cluster'
                    ? clusterSprite(item.members, item.size, scale.current)
                    : sprite(item.event, item.event.id === selected.current, scale.current)
                sprites.current.push(made)
                return made
            })
            .onObjectClick((d: object) => {
                const item = d as MarkerItem
                if (item.kind === 'event') {
                    handlers.current.onSelect(item.event)
                    return
                }
                // FR-MRK-008: towards the cluster until its members separate.
                pause.current()
                const altitude = separatingAltitude(item.members, g.pointOfView().altitude,
                    fitted.current, markerRadius(g), viewHalfAngle(g.camera() as THREE.PerspectiveCamera))
                g.pointOfView({lat: item.lat, lng: item.lng, altitude}, FOCUS_MS)
            })
            .onObjectHover((d: object | null) => {
                const item = d as MarkerItem | null
                hovered.current = item
                if (!item) { setTip(null); return }
                const title = item.kind === 'cluster' ? clusterTitle(item.members) : eventTitle(item.event)
                showTip.current(title, item, () => pointer.current)
            })
            // A zoom resizes every sprite in place and regroups only when the
            // grouping changes; rotation leaves the altitude, so it costs nothing.
            .onZoom(({altitude}) => {
                const next = altitude / fitted.current
                if (next === scale.current) return
                scale.current = next
                sprites.current.forEach(s => rescale(s, next))
                relayout.current(false)
            })
        globe.current = g
        clouds.current = makeCloudSphere(g.getGlobeRadius())
        g.scene().add(clouds.current)
        const controls = g.controls()
        controls.autoRotate = rotationWanted.current
        // FR-GLB-006: the wheel zooms between the altitude limits. OrbitControls
        // measures distance from the centre, so a limit is the radius plus it.
        const radius = g.getGlobeRadius()
        controls.minDistance = radius * (1 + MIN_ALTITUDE)
        controls.maxDistance = radius * (1 + MAX_ALTITUDE)
        const camera = g.camera() as THREE.PerspectiveCamera
        const fit = () => {
            // A host with no size yet has no aspect: a fit then sets the camera to
            // NaN and every later fit keeps its NaN place, so the globe never draws
            // (measured with a host first attached at 0 x 0). The observer fits
            // again once there is a size.
            if (el.clientWidth === 0 || el.clientHeight === 0) return
            g.width(el.clientWidth).height(el.clientHeight)
            fitted.current = fitAltitude(camera)
            g.pointOfView({altitude: fitted.current})
            // At the fit altitude every marker is its own fit size.
            scale.current = 1
            sprites.current.forEach(s => rescale(s, scale.current))
            relayout.current(false)
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
            clouds.current = null
        }
    }, [webgl2])

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
        if (clouds.current) showCloudImage(clouds.current, cloudImage)
    }, [cloudImage])

    // New events or a new selection draw afresh; the sprites carry both.
    useEffect(() => relayout.current(true), [events, selectedId])

    useEffect(() => {
        const g = globe.current
        const chosen = events.find(e => e.id === selectedId)
        if (!g || !chosen) return
        pause.current()
        g.pointOfView({lat: chosen.lat, lng: chosen.lng}, FOCUS_MS)
    // Focus once per selection, not on every refresh of the same event.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [selectedId])

    if (!webgl2) return <div className="globe globe-missing" role="alert">{NO_WEBGL2}</div>

    // globe.gl owns the host's children, so the tooltip sits beside it.
    return <>
        <div ref={host} className="globe" data-stop tabIndex={-1} onKeyDown={onKeyDown} onFocus={onFocus} onBlur={onBlur}
            onMouseDown={noClickFocus} aria-label={`Globe. ${GLOBE_KEYS}`}/>
        {tip && <div className="tip" style={{left: tip.x + TIP_OFFSET_PX, top: tip.y + TIP_OFFSET_PX}}>
            <div className="tip-title">{tip.title}</div>
            {tip.place && <div className="tip-place">{tip.place}</div>}
            {keyboard && <div className="tip-keys">{GLOBE_KEYS}</div>}
        </div>}
    </>
})

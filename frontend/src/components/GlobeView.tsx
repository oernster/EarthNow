// The globe (FR-GLB, FR-MRK): texture, fitted camera, start view, idle rotation,
// emoji markers, hover place lines and selection with a camera focus. The clouds
// and day and night are the layers useGlobeLayers attaches (FR-CLD-008, FR-DAY-003).
import {forwardRef, useEffect, useImperativeHandle, useMemo, useRef, useState} from 'react'
import Globe, {type GlobeInstance} from 'globe.gl'
import * as THREE from 'three'
import earthTexture from '../assets/earth.jpg'
import {api} from '../api'
import {clusterTitle, eventTitle} from '../categories'
import {atClosest, clusterEvents, layoutKey, type MarkerItem, separatingAltitude} from '../clusters'
import {FOCUS_MS, MAX_ALTITUDE, MIN_ALTITUDE, stepCursor, zoomed} from '../cursor'
import {clusterSprite, fitAltitude, fitOf, MARKER_ALTITUDE, rescale, sprite, viewHalfAngle} from '../markers'
import {noClickFocus} from '../ring'
import {trailed} from '../trails'
import {useGlobeLayers} from '../useGlobeLayers'
import {useIdleRotation} from '../useIdleRotation'
import type {EventDTO, StartViewDTO, SunDTO} from '../types'

// One plus or minus press animates over this long.
const ZOOM_MS = 250
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
    // burntImage is the window's drawn burnt areas (FR-BA-006); empty draws none.
    burntImage: string
    // dayNightShown switches the day and night layer (FR-DAY-006); sun is
    // where the sun stands overhead, null until first asked.
    dayNightShown: boolean
    sun: SunDTO | null
    // start is where the globe first faces (FR-GLB-015); not found keeps 1.0.0's view.
    start: StartViewDTO
    // trailsShown switches the storm trails (FR-TRL-002, FR-TRL-003).
    trailsShown: boolean
    onSelect: (e: EventDTO) => void
    // onCluster lists a cluster the closest zoom cannot separate (FR-MRK-011).
    onCluster: (members: EventDTO[]) => void
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

export const GlobeView = forwardRef<GlobeHandle, Props>(function GlobeView(
    {events, selectedId, autoRotate, secondsPerRevolution, cloudImage, burntImage, dayNightShown, sun, start, trailsShown, onSelect, onCluster, onProblem}, ref) {
    const host = useRef<HTMLDivElement>(null)
    const [webgl2] = useState(hasWebGL2)
    const globe = useRef<GlobeInstance | null>(null)
    const hovered = useRef<MarkerItem | null>(null)
    const pointer = useRef({x: 0, y: 0})
    const handlers = useRef({onSelect, onCluster, onProblem})
    handlers.current = {onSelect, onCluster, onProblem}
    const [tip, setTip] = useState<Tip | null>(null)
    // keyboard is true while the globe holds focus, so the tooltip names the keys.
    const [keyboard, setKeyboard] = useState(false)
    const trails = useMemo(() => trailed(events, trailsShown), [events, trailsShown])
    const layers = useGlobeLayers(cloudImage, burntImage, dayNightShown, sun, trails)
    const opening = useRef(start)
    const idle = useIdleRotation(globe, autoRotate, secondsPerRevolution)

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
        idle.pause()
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
            if (g) g.pointOfView({altitude: fitOf(g)}, FOCUS_MS)
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
                // FR-MRK-011: at the closest zoom nothing more can part, so list them.
                if (atClosest(g.pointOfView().altitude)) {
                    handlers.current.onCluster(item.members)
                    return
                }
                // FR-MRK-008: towards the cluster until its members separate.
                idle.pause()
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
        layers.attach(g)
        const controls = g.controls()
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
        // FR-GLB-015: face the reader's country before the first frame turns it;
        // a later fit keeps the facing and sets only the altitude.
        if (opening.current.found) g.pointOfView({lat: opening.current.lat, lng: opening.current.lng})
        const observer = new ResizeObserver(fit)
        observer.observe(el)
        const onMove = (ev: MouseEvent) => {
            pointer.current = {x: ev.clientX, y: ev.clientY}
            if (hovered.current) setTip(t => t && {...t, ...pointer.current})
        }
        el.addEventListener('mousemove', onMove)
        const detachIdle = idle.attach(g, el)
        return () => {
            observer.disconnect()
            el.removeEventListener('mousemove', onMove)
            detachIdle()
            g._destructor()
            globe.current = null
            layers.detach()
        }
    }, [webgl2, layers, idle])

    // New events or a new selection draw afresh; the sprites carry both.
    useEffect(() => relayout.current(true), [events, selectedId])

    useEffect(() => {
        const g = globe.current
        const chosen = events.find(e => e.id === selectedId)
        if (!g || !chosen) return
        idle.pause()
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

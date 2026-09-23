// The globe (FR-GLB, FR-MRK): texture, fitted camera, idle rotation, emoji
// markers, hover place lines and selection with a camera focus.
import {useEffect, useRef, useState} from 'react'
import Globe, {type GlobeInstance} from 'globe.gl'
import * as THREE from 'three'
import earthTexture from '../assets/earth.jpg'
import {api} from '../api'
import {categoryOf} from '../categories'
import {fitAltitude, sprite} from '../markers'
import type {EventDTO} from '../types'

const MARKER_ALTITUDE = 0.01
const FOCUS_MS = 1000
// FR-GLB-002: one revolution per 240 s. OrbitControls documents speed 2.0 as
// one orbit in 30 s, so an orbit takes ORBIT_SECONDS_AT_UNIT_SPEED / speed.
const SECONDS_PER_REVOLUTION = 240
const ORBIT_SECONDS_AT_UNIT_SPEED = 60
const ROTATE_SPEED = ORBIT_SECONDS_AT_UNIT_SPEED / SECONDS_PER_REVOLUTION
const TIP_OFFSET_PX = 14

interface Props {
    events: EventDTO[]
    selectedId: string | null
    onSelect: (e: EventDTO) => void
    onProblem: (reason: string) => void
}

interface Tip { x: number; y: number; title: string; place: string }

export function GlobeView({events, selectedId, onSelect, onProblem}: Props) {
    const host = useRef<HTMLDivElement>(null)
    const globe = useRef<GlobeInstance | null>(null)
    const hovered = useRef<EventDTO | null>(null)
    const pointer = useRef({x: 0, y: 0})
    const handlers = useRef({onSelect, onProblem})
    handlers.current = {onSelect, onProblem}
    const [tip, setTip] = useState<Tip | null>(null)

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
                const title = `${categoryOf(e.category).emoji} ${e.title}`
                setTip({...pointer.current, title, place: ''})
                void api.place(e.lat, e.lng, handlers.current.onProblem).then(place => {
                    if (place !== null && hovered.current === e) setTip({...pointer.current, title, place})
                })
            })
        globe.current = g
        const controls = g.controls()
        controls.autoRotate = true
        controls.autoRotateSpeed = ROTATE_SPEED
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
        return () => {
            observer.disconnect()
            el.removeEventListener('mousemove', onMove)
            g._destructor()
            globe.current = null
        }
    }, [])

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
        g.controls().autoRotate = false
        g.pointOfView({lat: chosen.lat, lng: chosen.lng}, FOCUS_MS)
    // Focus once per selection, not on every refresh of the same event.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [selectedId])

    // globe.gl owns the host's children, so the tooltip sits beside it.
    return <>
        <div ref={host} className="globe"/>
        {tip && <div className="tip" style={{left: tip.x + TIP_OFFSET_PX, top: tip.y + TIP_OFFSET_PX}}>
            <div className="tip-title">{tip.title}</div>
            {tip.place && <div className="tip-place">{tip.place}</div>}
        </div>}
    </>
}

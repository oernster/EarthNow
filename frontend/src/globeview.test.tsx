// GlobeView's rules over a stand-in for globe.gl: jsdom has no WebGL, so the
// library is replaced by a recorder holding the same controls and camera. What
// is tested is the component's own logic, never globe.gl's drawing.
import {act, fireEvent, render, screen} from '@testing-library/react'
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import * as THREE from 'three'
import type {EventDTO} from './types'

const PLACE = '12 km N of Here, Land'
vi.mock('./api', () => ({api: {place: vi.fn(() => Promise.resolve(PLACE))}}))

interface Fake {
    controls: {autoRotate: boolean; autoRotateSpeed: number; minDistance: number; maxDistance: number}
    calls: Array<[string, unknown[]]>
    // altitude is what pointOfView() reports, as the camera would.
    altitude: number
}
const made: Fake[] = []

// makeGlobe answers a chainable recorder: every setter records its arguments and
// answers the recorder, as globe.gl's builder does.
function makeGlobe(): unknown {
    const fake: Fake = {controls: {autoRotate: false, autoRotateSpeed: 0, minDistance: 0, maxDistance: 0}, calls: [], altitude: 1.6}
    made.push(fake)
    const camera = new THREE.PerspectiveCamera(50, 1.5)
    const known: Record<string, (...a: unknown[]) => unknown> = {
        controls: () => fake.controls,
        camera: () => camera,
        getGlobeRadius: () => 100,
        _destructor: () => undefined,
    }
    const proxy: unknown = new Proxy({}, {
        get: (_, name: string) => known[name] ?? ((...args: unknown[]) => {
            fake.calls.push([name, args])
            return name === 'pointOfView' && args.length === 0 ? {lat: 0, lng: 0, altitude: fake.altitude} : proxy
        }),
    })
    return proxy
}
vi.mock('globe.gl', () => ({default: function Globe() { return makeGlobe() }}))

const {GlobeView} = await import('./components/GlobeView')
const {fitAltitude} = await import('./markers')

const quake: EventDTO = {
    id: 'USGS/ak1', provider: 'USGS', category: 'EARTHQUAKE', title: 'M 3.1 near Montana', description: '',
    lat: 61.899, lng: -150.919, at: '', dayOnly: false, reported: '', retrievedAt: '', retrieved: '',
    measurement: '', depth: '', trail: [], band: 2, sourceUrl: '', sourceText: '', ended: false,
}
const noop = () => undefined
const IDLE_MS = 10_000
// NFR-UX-003's focus duration, over which the return to the fit altitude runs.
const FOCUS = 1000

function draw(autoRotate: boolean, events: EventDTO[] = []) {
    const view = render(<GlobeView events={events} selectedId={null} autoRotate={autoRotate}
        secondsPerRevolution={60} cloudImage="" burntImage="" dayNightShown={false} sun={null} trailsShown={false} start={{found: false, lat: 0, lng: 0}} onSelect={noop} onCluster={noop} onProblem={noop}/>)
    return {view, globe: made[made.length - 1], host: document.querySelector<HTMLElement>('.globe')!}
}

// webgl2Only reports WebGL2 present and the 2D context the sprites draw with absent.
function webgl2Only() {
    vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockImplementation(
        ((kind: string) => (kind === 'webgl2' ? {} : null)) as never)
}

describe('the globe', () => {
    beforeEach(() => {
        made.length = 0
        webgl2Only()
        vi.useFakeTimers()
    })
    afterEach(() => {
        vi.useRealTimers()
        vi.restoreAllMocks()
    })

    it.each(['pointerDown', 'wheel', 'keyDown'] as const)('FR-GLB-003 stops idle rotation on %s', input => {
        const {globe, host} = draw(true)
        expect(globe.controls.autoRotate).toBe(true)
        fireEvent[input](host)
        expect(globe.controls.autoRotate).toBe(false)
    })

    it('FR-GLB-004 never rotates while auto-rotate is off, input or none', () => {
        const {globe, host} = draw(false)
        expect(globe.controls.autoRotate).toBe(false)
        fireEvent.pointerDown(host)
        act(() => { vi.advanceTimersByTime(3 * IDLE_MS) })
        expect(globe.controls.autoRotate).toBe(false)
    })

    it.each([['in', 0.5], ['out', 2]])('FR-GLB-018 zoomed %s, returns to the fit altitude before rotating', (_, times) => {
        const {globe, host} = draw(true)
        const fit = fitAltitude(new THREE.PerspectiveCamera(50, 1.5))
        globe.altitude = fit * times
        fireEvent.wheel(host)
        act(() => { vi.advanceTimersByTime(IDLE_MS) })
        expect(globe.calls).toContainEqual(['pointOfView', [{altitude: fit}, FOCUS]])
        expect(globe.controls.autoRotate).toBe(false)
        act(() => { vi.advanceTimersByTime(FOCUS) })
        expect(globe.controls.autoRotate).toBe(true)
    })

    it('FR-GLB-018 rotates at once when already at the fit altitude', () => {
        const {globe, host} = draw(true)
        globe.altitude = fitAltitude(new THREE.PerspectiveCamera(50, 1.5))
        fireEvent.pointerDown(host)
        const before = globe.calls.length
        act(() => { vi.advanceTimersByTime(IDLE_MS) })
        expect(globe.controls.autoRotate).toBe(true)
        expect(globe.calls.slice(before).some(([name, args]) => name === 'pointOfView' && args.length === 2)).toBe(false)
    })

    it('FR-GLB-018 input during the return stops it where it is and waits again', () => {
        const {globe, host} = draw(true)
        globe.altitude = 0.3
        fireEvent.pointerDown(host)
        act(() => { vi.advanceTimersByTime(IDLE_MS) })
        fireEvent.pointerDown(host)
        expect(globe.calls.at(-1)).toEqual(['pointOfView', [{lat: 0, lng: 0, altitude: 0.3}]])
        act(() => { vi.advanceTimersByTime(FOCUS) })
        expect(globe.controls.autoRotate).toBe(false)
        act(() => { vi.advanceTimersByTime(IDLE_MS) })
        expect(globe.calls.filter(([name, args]) => name === 'pointOfView' && args.length === 2)).toHaveLength(2)
    })

    it('FR-GLB-018 switching rotation back on while zoomed returns to the fit altitude first', () => {
        const {view, globe} = draw(true)
        const fit = fitAltitude(new THREE.PerspectiveCamera(50, 1.5))
        const set = (autoRotate: boolean, secondsPerRevolution = 60) => view.rerender(<GlobeView events={[]}
            selectedId={null} autoRotate={autoRotate} secondsPerRevolution={secondsPerRevolution} cloudImage="" burntImage=""
            dayNightShown={false} sun={null} trailsShown={false} start={{found: false, lat: 0, lng: 0}} onSelect={noop} onCluster={noop} onProblem={noop}/>)
        globe.altitude = 0.3
        set(false)
        set(true)
        expect(globe.calls).toContainEqual(['pointOfView', [{altitude: fit}, FOCUS]])
        expect(globe.controls.autoRotate).toBe(false)
        act(() => { vi.advanceTimersByTime(FOCUS) })
        expect(globe.controls.autoRotate).toBe(true)
        // A speed change while turning is not a switch-on: the camera stays put.
        globe.altitude = 0.3
        const before = globe.calls.length
        set(true, 120)
        expect(globe.calls.slice(before).some(([name, args]) => name === 'pointOfView' && args.length === 2)).toBe(false)
    })

    it('FR-GLB-018 switching rotation on inside the idle delay waits for the delay to end', () => {
        const {view, globe, host} = draw(true)
        globe.altitude = fitAltitude(new THREE.PerspectiveCamera(50, 1.5))
        const set = (autoRotate: boolean) => view.rerender(<GlobeView events={[]} selectedId={null} autoRotate={autoRotate}
            secondsPerRevolution={60} cloudImage="" burntImage="" dayNightShown={false} sun={null} trailsShown={false}
            start={{found: false, lat: 0, lng: 0}} onSelect={noop} onCluster={noop} onProblem={noop}/>)
        fireEvent.pointerDown(host)
        set(false)
        set(true)
        expect(globe.controls.autoRotate).toBe(false)
        act(() => { vi.advanceTimersByTime(IDLE_MS) })
        expect(globe.controls.autoRotate).toBe(true)
    })

    it('FR-GLB-004 and FR-GLB-011 apply the setting and its speed at launch and at once on a change', () => {
        const {view, globe} = draw(true)
        expect(globe.controls.autoRotateSpeed).toBe(1)
        const set = (autoRotate: boolean, secondsPerRevolution: number) => view.rerender(<GlobeView events={[]}
            selectedId={null} autoRotate={autoRotate} secondsPerRevolution={secondsPerRevolution} cloudImage="" burntImage=""
            dayNightShown={false} sun={null} trailsShown={false} start={{found: false, lat: 0, lng: 0}} onSelect={noop} onCluster={noop} onProblem={noop}/>)
        set(false, 60)
        expect(globe.controls.autoRotate).toBe(false)
        set(true, 120)
        expect(globe.controls.autoRotate).toBe(true)
        expect(globe.controls.autoRotateSpeed).toBe(0.5)
    })

    it('FR-GLB-012 rotates on idle whatever the reduced-motion setting says', () => {
        vi.stubGlobal('matchMedia', (query: string) => ({matches: query.includes('reduce'), media: query}))
        const {globe, host} = draw(true)
        globe.altitude = fitAltitude(new THREE.PerspectiveCamera(50, 1.5))
        expect(globe.controls.autoRotate).toBe(true)
        fireEvent.pointerDown(host)
        act(() => { vi.advanceTimersByTime(IDLE_MS) })
        expect(globe.controls.autoRotate).toBe(true)
        vi.unstubAllGlobals()
    })

    it('FR-MRK-001 draws one marker per event at its position', () => {
        const {globe} = draw(false, [quake])
        const placed = globe.calls.filter(([name]) => name === 'objectsData').at(-1)![1][0] as Array<Record<string, unknown>>
        expect(placed).toEqual([{kind: 'event', lat: 61.899, lng: -150.919, event: quake}])
        expect(globe.calls).toContainEqual(['objectLat', ['lat']])
        expect(globe.calls).toContainEqual(['objectLng', ['lng']])
    })

    it('NFR-UX-003 animates the camera to a selected event over 1,000 ms', () => {
        const {view, globe} = draw(false, [quake])
        view.rerender(<GlobeView events={[quake]} selectedId={quake.id} autoRotate={false}
            secondsPerRevolution={60} cloudImage="" burntImage="" dayNightShown={false} sun={null} trailsShown={false} start={{found: false, lat: 0, lng: 0}} onSelect={noop} onCluster={noop} onProblem={noop}/>)
        expect(globe.calls).toContainEqual(['pointOfView', [{lat: quake.lat, lng: quake.lng}, 1000]])
    })

    it('FR-GEO-006 shows the cursor event its tooltip with the place line, as hover does', async () => {
        const {host} = draw(false, [quake])
        fireEvent.focus(host)
        await act(async () => { await Promise.resolve() })
        expect(screen.getByText(PLACE)).toBeTruthy()
        expect(document.querySelector('.tip-title')!.textContent).toContain(quake.title)
    })

    it('FR-GLB-015 opens facing the start view, with rotation to follow from there', () => {
        render(<GlobeView events={[]} selectedId={null} autoRotate secondsPerRevolution={60} cloudImage="" burntImage=""
            dayNightShown={false} sun={null} trailsShown={false} start={{found: true, lat: 54.4027, lng: -2.1163}} onSelect={noop} onCluster={noop} onProblem={noop}/>)
        const globe = made[made.length - 1]
        expect(globe.calls).toContainEqual(['pointOfView', [{lat: 54.4027, lng: -2.1163}]])
        expect(globe.controls.autoRotate).toBe(true)
    })

    it('FR-GLB-016 opens as before when there is no start view', () => {
        const {globe} = draw(true)
        const facings = globe.calls.filter(([name, args]) => name === 'pointOfView' && 'lat' in (args[0] as object))
        expect(facings).toEqual([])
    })

    it('FR-DAY-003 draws the globe with the day and night material', () => {
        const {globe} = draw(false)
        const set = globe.calls.find(([name]) => name === 'globeMaterial')
        expect(set?.[1][0]).toBeInstanceOf(THREE.MeshPhongMaterial)
        expect(globe.calls).toContainEqual(['globeImageUrl', [expect.any(String)]])
    })

    it('FR-DAY-006 loads the night lights on the first show only, never while hidden', () => {
        const load = vi.spyOn(THREE.TextureLoader.prototype, 'load').mockImplementation(() => new THREE.Texture())
        const {view} = draw(false)
        expect(load).not.toHaveBeenCalled()
        const shown = (dayNightShown: boolean) => view.rerender(<GlobeView events={[]} selectedId={null} autoRotate={false}
            secondsPerRevolution={60} cloudImage="" burntImage="" dayNightShown={dayNightShown} sun={null} trailsShown={false} start={{found: false, lat: 0, lng: 0}} onSelect={noop} onCluster={noop} onProblem={noop}/>)
        shown(true)
        shown(false)
        shown(true)
        expect(load).toHaveBeenCalledTimes(1)
    })
})

describe('FR-MRK-011 activating a cluster', () => {
    beforeEach(() => {
        made.length = 0
        webgl2Only()
    })
    afterEach(() => vi.restoreAllMocks())

    it('zooms towards it from afar and lists it at the closest zoom', async () => {
        const {MIN_ALTITUDE} = await import('./cursor')
        const onCluster = vi.fn()
        const onSelect = vi.fn()
        render(<GlobeView events={[quake]} selectedId={null} autoRotate={false} secondsPerRevolution={60} cloudImage="" burntImage=""
            dayNightShown={false} sun={null} trailsShown={false} start={{found: false, lat: 0, lng: 0}}
            onSelect={onSelect} onCluster={onCluster} onProblem={noop}/>)
        const globe = made[made.length - 1]
        const click = globe.calls.find(([name]) => name === 'onObjectClick')![1][0] as (d: object) => void
        const cluster = {kind: 'cluster', lat: quake.lat, lng: quake.lng, key: 'k', members: [quake, {...quake, id: 'USGS/ak2'}], size: 1}
        click(cluster)
        expect(onCluster).not.toHaveBeenCalled()
        expect(globe.calls.some(([name, args]) => name === 'pointOfView' && args.length === 2)).toBe(true)
        globe.altitude = MIN_ALTITUDE
        click(cluster)
        expect(onCluster).toHaveBeenCalledWith(cluster.members)
        click({kind: 'event', lat: quake.lat, lng: quake.lng, event: quake})
        expect(onSelect).toHaveBeenCalledWith(quake)
    })
})

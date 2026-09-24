// The burnt-area layer on the page (REQUIREMENTS.md 3.2.13): the Settings box,
// the status line, the guide, the sphere the image is drawn on and the hook
// that asks for the image only when it changes. Drawing itself is WebGL's and
// is checked by a person (FR-BA-007 in TESTING.md).
import {act, fireEvent, render, renderHook, screen} from '@testing-library/react'
import {afterEach, describe, expect, it, vi} from 'vitest'
import * as THREE from 'three'
import {api} from './api'
import {SettingsDialog} from './components/SettingsDialog'
import {StatusLine} from './components/StatusLine'
import {guideSections} from './guide'
import {BURNT_ALTITUDE, BURNT_ORDER, CLOUD_ALTITUDE, CLOUD_ORDER, makeBurntSphere} from './imageLayers'
import type {BurntAreasDTO, ProviderDTO, SettingsDTO} from './types'
import {useGlobeLayers} from './useGlobeLayers'
import {type LayerSource, useLayer} from './useLayer'

const noop = () => undefined

const gwis: ProviderDTO = {name: 'GWIS', loading: false, refreshing: false, stale: false, retrieved: 'Retrieved 12 min ago', problem: '', nextAttempt: ''}

function burnt(patch: Partial<BurntAreasDTO> = {}): BurntAreasDTO {
    return {shown: true, line: 'Burnt areas: 17 to 23 Sep (UTC), retrieved 12 min ago', key: '2026-09-23@1', notice: '', provider: gwis, ...patch}
}

const SETTINGS: SettingsDTO = {autoRotate: true, magnitude: '2.5', speed: 'normal', windowKey: '7d', hiddenCategories: [],
    hiddenProviders: [], cloudsShown: false, dayNightShown: true, trailsShown: true, burntShown: false}

describe('FR-BA-010 the Settings box', () => {
    it('offers "Show burnt areas" and applies it at once', () => {
        const onChange = vi.fn()
        render(<SettingsDialog settings={SETTINGS} choices={{magnitudes: [], speeds: []}} onChange={onChange} onClose={noop}/>)
        const box = screen.getByLabelText('Show burnt areas') as HTMLInputElement
        expect(box.checked).toBe(false)
        fireEvent.click(box)
        expect(onChange).toHaveBeenCalledWith({burntShown: true})
    })
})

describe('FR-BA-008 the burnt-area line', () => {
    it('reads the days and their age while shown', () => {
        render(<StatusLine countLine="" providers={[]} problem="" burnt={burnt()}/>)
        expect(screen.getByText('Burnt areas: 17 to 23 Sep (UTC), retrieved 12 min ago').className).toBe('provider')
    })

    it('is marked while a day failed (FR-BA-012)', () => {
        render(<StatusLine countLine="" providers={[]} problem="" burnt={burnt({provider: {...gwis, problem: '23 Sep: timed out'}})}/>)
        const line = screen.getByText(/Burnt areas:/)
        expect(line.className).toContain('warn')
        expect(line.getAttribute('title')).toBe('23 Sep: timed out')
    })

    it('is absent while the layer is hidden (FR-BA-005)', () => {
        render(<StatusLine countLine="" providers={[]} problem="" burnt={burnt({shown: false})}/>)
        expect(screen.queryByText(/Burnt areas:/)).toBeNull()
    })
})

describe('FR-BA-016 the guide', () => {
    it('says how the maps are made and what they cannot show', () => {
        const text = guideSections('').find(s => s.heading === 'Burnt areas')?.paragraphs?.join(' ') ?? ''
        expect(text).toContain('one UTC day at a time')
        expect(text).toContain('the day it was mapped')
        expect(text).toContain('a small burn may not show')
        expect(text).toContain('may show none yet')
    })
})

describe('FR-BA-007 the burnt-area sphere', () => {
    afterEach(() => vi.restoreAllMocks())

    it('lies above the texture and beneath the clouds, drawn before them', () => {
        const sphere = makeBurntSphere(100)
        expect(BURNT_ALTITUDE).toBeGreaterThan(0)
        expect(BURNT_ALTITUDE).toBeLessThan(CLOUD_ALTITUDE)
        expect(BURNT_ORDER).toBeLessThan(CLOUD_ORDER)
        expect((sphere.geometry as THREE.SphereGeometry).parameters.radius).toBeCloseTo(100 * (1 + BURNT_ALTITUDE))
        expect(sphere.renderOrder).toBe(BURNT_ORDER)
        expect(sphere.visible).toBe(false)
    })

    it('is added to the globe and follows the image (FR-BA-006)', () => {
        vi.spyOn(THREE.TextureLoader.prototype, 'load').mockImplementation((_url, done) => {
            done?.(new THREE.Texture())
            return new THREE.Texture()
        })
        const added: THREE.Object3D[] = []
        const g: Record<string, unknown> = {globeMaterial: noop, getGlobeRadius: () => 100, scene: () => ({add: (o: THREE.Object3D) => added.push(o)})}
        for (const name of ['pathPoints', 'pathPointLat', 'pathPointLng', 'pathPointAlt', 'pathColor', 'pathTransitionDuration', 'pathsData']) g[name] = () => g
        const {result, rerender} = renderHook(({image}) => useGlobeLayers('', image, false, null, []), {initialProps: {image: ''}})
        act(() => result.current.attach(g as never))
        const sphere = added.find(o => o.renderOrder === BURNT_ORDER) as THREE.Mesh
        expect(sphere.visible).toBe(false)
        rerender({image: 'data:image/png;base64,AAAA'})
        expect(sphere.visible).toBe(true)
        rerender({image: ''})
        expect(sphere.visible).toBe(false)
    })
})

describe('the burnt-area calls', () => {
    afterEach(() => { delete (window as unknown as {go?: unknown}).go })

    it('reads the state and the image from the Go side', async () => {
        const state = burnt()
        const App = {BurntAreas: vi.fn(() => Promise.resolve(state)), BurntImage: vi.fn(() => Promise.resolve('data:image/png;base64,AAAA'))};
        (window as unknown as {go: unknown}).go = {main: {App}}
        expect(await api.burntAreas(noop)).toEqual(state)
        expect(await api.burntImage(noop)).toBe('data:image/png;base64,AAAA')
    })
})

describe('useLayer, the image layers\' reader (FR-CLD-016, FR-BA-006)', () => {
    afterEach(() => { delete (window as unknown as {runtime?: unknown}).runtime })

    it('asks for the image only when the key changes, again on each event', async () => {
        const handlers: Record<string, () => void> = {}
        const off = vi.fn();
        (window as unknown as {runtime: unknown}).runtime = {EventsOn: (name: string, cb: () => void) => { handlers[name] = cb; return off }}
        let state = burnt({key: 'a'})
        const source: LayerSource<BurntAreasDTO> = {
            state: vi.fn(() => Promise.resolve(state)),
            image: vi.fn(() => Promise.resolve(`image ${state.key}`)),
            key: b => b.key,
            events: ['burnt-changed'],
        }
        const {result, unmount} = renderHook(() => useLayer(source, true, noop))
        await act(async () => { await Promise.resolve(); await Promise.resolve() })
        expect(result.current[1]).toBe('image a')
        await act(async () => { handlers['burnt-changed'](); await Promise.resolve(); await Promise.resolve() })
        expect(source.image).toHaveBeenCalledTimes(1)
        state = burnt({key: 'b'})
        await act(async () => { handlers['burnt-changed'](); await Promise.resolve(); await Promise.resolve() })
        expect(result.current[1]).toBe('image b')
        expect(result.current[0]?.key).toBe('b')
        unmount()
        expect(off).toHaveBeenCalled()
    })

    it('reads nothing until ready; keeps what it had when refused', async () => {
        const source: LayerSource<BurntAreasDTO> = {state: vi.fn(() => Promise.resolve(null)), image: vi.fn(), key: b => b.key, events: []}
        const {result, rerender} = renderHook(({ready}) => useLayer(source, ready, noop), {initialProps: {ready: false}})
        expect(source.state).not.toHaveBeenCalled()
        rerender({ready: true})
        await act(async () => { await Promise.resolve() })
        expect(result.current).toEqual([null, ''])
        expect(source.image).not.toHaveBeenCalled()
    })
})

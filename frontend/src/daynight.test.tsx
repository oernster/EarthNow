// The day and night layer on the page (REQUIREMENTS.md 3.2.11): the rail
// button, the asking for the sun and the light both materials are drawn by.
// Drawing itself is WebGL's and is checked by a person (FR-DAY-003, FR-DAY-009
// in TESTING.md); here the shaders are only assembled, over three's own sources.
import {act, fireEvent, render, renderHook, screen} from '@testing-library/react'
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import * as THREE from 'three'
import {api} from './api'
import {Rail} from './components/Rail'
import {dimClouds, lightNights, makeGlobeMaterial, makeUniforms, placeSun, type ShaderParts, showDayNight, sunDirection} from './dayNight'
import {icons} from './icons'
import type {SunDTO} from './types'
import {useGlobeLayers} from './useGlobeLayers'
import {SUN_POLL_MS, useSun} from './useSun'

const noop = () => undefined

function rail(dayNightShown: boolean, onToggleDayNight: () => void = noop) {
    return render(<Rail autoRotate={false} attention={false} onToggleRotate={noop} cloudsShown={false}
        onToggleClouds={noop} dayNightShown={dayNightShown} onToggleDayNight={onToggleDayNight} onResetView={noop}
        onZoom={noop} onRefresh={noop} onStatus={noop} onSettings={noop} onHelp={noop} onDonate={noop}/>)
}

const SUN: SunDTO = {lat: 23.44, lng: -0.46, twilightDegrees: 6, cloudNightFloor: 0.25}

// compiled runs a material's injection over three's own shader for it.
function compiled(material: THREE.Material, lib: {vertexShader: string; fragmentShader: string}): ShaderParts {
    const shader: ShaderParts = {uniforms: {}, vertexShader: lib.vertexShader, fragmentShader: lib.fragmentShader}
    material.onBeforeCompile(shader as never, undefined as never)
    return shader
}

describe('FR-DAY-005 the day and night button', () => {
    it('shows the plain artwork and offers to show while hidden', () => {
        rail(false)
        const button = screen.getByLabelText('Show day and night')
        expect(button.querySelector('img')?.getAttribute('src')).toBe(icons.dayNight)
    })

    it('shows the negative and offers to hide while shown', () => {
        rail(true)
        const button = screen.getByLabelText('Hide day and night')
        expect(button.querySelector('img')?.getAttribute('src')).toBe(icons.dayNightHide)
    })
})

describe('FR-DAY-006 switching the layer', () => {
    it('asks for the switch on a press', () => {
        const onToggle = vi.fn()
        rail(true, onToggle)
        fireEvent.click(screen.getByLabelText('Hide day and night'))
        expect(onToggle).toHaveBeenCalledTimes(1)
    })
})

describe('FR-DAY-003 the sun on the globe', () => {
    // three-globe's polar2Cartesian: latitude 0, longitude 0 faces +Z, 90 E
    // faces +X and the north pole is +Y.
    it.each([
        [0, 0, [0, 0, 1]],
        [0, 90, [1, 0, 0]],
        [90, 0, [0, 1, 0]],
        [0, -90, [-1, 0, 0]],
    ])('places the sun over %s, %s in three-globe\'s frame', (lat, lng, want) => {
        const got = sunDirection(lat, lng)
        expect(got.x).toBeCloseTo(want[0])
        expect(got.y).toBeCloseTo(want[1])
        expect(got.z).toBeCloseTo(want[2])
    })

    it('takes the sun and the light figures that came with it', () => {
        const u = makeUniforms()
        placeSun(u, SUN)
        expect(u.sunDirection.value.equals(sunDirection(SUN.lat, SUN.lng))).toBe(true)
        expect(u.twilightDegrees.value).toBe(SUN.twilightDegrees)
        expect(u.cloudNightFloor.value).toBe(SUN.cloudNightFloor)
    })

    it('lights the globe by the day and the night by what is left, over three\'s Phong', () => {
        const u = makeUniforms()
        const shader = compiled(makeGlobeMaterial(u), THREE.ShaderLib.phong)
        expect(shader.uniforms.sunDirection).toBe(u.sunDirection)
        expect(shader.uniforms.dayOnly).toBe(u.dayOnly)
        expect(shader.vertexShader).toContain('vSunNormal = normalize(mat3(modelMatrix) * position);')
        expect(shader.fragmentShader).toContain('diffuseColor.rgb *= dayLight();')
        expect(shader.fragmentShader).toContain('totalEmissiveRadiance *= 1.0 - dayLight();')
    })

    it('keeps the night side dark until the lights arrive, then lets them through', () => {
        const material = makeGlobeMaterial(makeUniforms())
        expect(material.emissive.getHex()).toBe(0)
        const first = new THREE.Texture()
        lightNights(material, first)
        expect(material.emissiveMap).toBe(first)
        expect(first.colorSpace).toBe(THREE.SRGBColorSpace)
        expect(material.emissive.getHex()).toBe(new THREE.Color(1, 1, 1).getHex())
        const released = vi.spyOn(first, 'dispose')
        lightNights(material, new THREE.Texture())
        expect(released).toHaveBeenCalled()
    })
})

describe('FR-DAY-008 the layer hidden', () => {
    it('hands the globe a light of 1 everywhere', () => {
        const u = makeUniforms()
        expect(u.dayOnly.value).toBe(1)
        showDayNight(u, true)
        expect(u.dayOnly.value).toBe(0)
        showDayNight(u, false)
        expect(u.dayOnly.value).toBe(1)
    })
})

describe('FR-DAY-009 the clouds at night', () => {
    it('dims each cloud from the night floor by the same light as the globe', () => {
        const u = makeUniforms()
        const globe = compiled(makeGlobeMaterial(u), THREE.ShaderLib.phong)
        const cloud = new THREE.MeshBasicMaterial({transparent: true})
        dimClouds(cloud, u)
        const shader = compiled(cloud, THREE.ShaderLib.basic)
        expect(shader.uniforms.sunDirection).toBe(globe.uniforms.sunDirection)
        expect(shader.uniforms.cloudNightFloor).toBe(u.cloudNightFloor)
        expect(shader.fragmentShader).toContain('diffuseColor.a *= cloudNightFloor + (1.0 - cloudNightFloor) * dayLight();')
    })
})

describe('FR-DAY-004 asking for the sun', () => {
    let App: {Sun: ReturnType<typeof vi.fn>}
    beforeEach(() => {
        vi.useFakeTimers()
        App = {Sun: vi.fn(() => Promise.resolve(SUN))};
        (window as unknown as {go: unknown}).go = {main: {App}}
    })
    afterEach(() => {
        vi.useRealTimers()
        delete (window as unknown as {go?: unknown}).go
    })

    it('asks on showing and again within a minute while shown', async () => {
        const {result} = renderHook(() => useSun(true, noop))
        await act(async () => { await Promise.resolve() })
        expect(result.current).toEqual(SUN)
        expect(App.Sun).toHaveBeenCalledTimes(1)
        await act(async () => { vi.advanceTimersByTime(SUN_POLL_MS) })
        expect(App.Sun).toHaveBeenCalledTimes(2)
        expect(SUN_POLL_MS).toBeLessThanOrEqual(60_000)
    })

    it('asks nothing while hidden and stops when hidden', async () => {
        const {rerender} = renderHook(({shown}) => useSun(shown, noop), {initialProps: {shown: false}})
        await act(async () => { vi.advanceTimersByTime(2 * SUN_POLL_MS) })
        expect(App.Sun).not.toHaveBeenCalled()
        rerender({shown: true})
        rerender({shown: false})
        await act(async () => { vi.advanceTimersByTime(2 * SUN_POLL_MS) })
        expect(App.Sun).toHaveBeenCalledTimes(1)
    })

    it('reads the sun from the Go side', async () => {
        expect(await api.sun(noop)).toEqual(SUN)
    })

    it('keeps no sun when the Go side refuses', async () => {
        App.Sun.mockImplementation(() => Promise.reject(new Error('refused')))
        const onProblem = vi.fn()
        const {result} = renderHook(() => useSun(true, onProblem))
        await act(async () => { await Promise.resolve() })
        expect(result.current).toBeNull()
        expect(onProblem).toHaveBeenCalledWith('refused')
    })
})

describe('the layers on a globe (FR-DAY-003, FR-CLD-008)', () => {
    afterEach(() => vi.restoreAllMocks())

    it('attach, then follow the sun and the cloud image as they change', () => {
        vi.spyOn(THREE.TextureLoader.prototype, 'load').mockImplementation((_url, done) => {
            done?.(new THREE.Texture())
            return new THREE.Texture()
        })
        const setMaterial = vi.fn()
        const added: THREE.Object3D[] = []
        const g: Record<string, unknown> = {globeMaterial: setMaterial, getGlobeRadius: () => 100, scene: () => ({add: (o: THREE.Object3D) => added.push(o)})}
        for (const name of ['pathPoints', 'pathPointLat', 'pathPointLng', 'pathPointAlt', 'pathColor', 'pathTransitionDuration', 'pathsData']) g[name] = () => g
        const {result, rerender} = renderHook(({image, sun}) => useGlobeLayers(image, true, sun, []),
            {initialProps: {image: '', sun: null as SunDTO | null}})
        act(() => result.current.attach(g as never))
        const sphere = added[0] as THREE.Mesh
        expect(sphere.visible).toBe(false)
        rerender({image: 'data:image/png;base64,AAAA', sun: SUN})
        expect(sphere.visible).toBe(true)
        const shader = compiled(setMaterial.mock.calls[0][0] as THREE.Material, THREE.ShaderLib.phong)
        const placed = (shader.uniforms.sunDirection as {value: THREE.Vector3}).value
        expect(placed.equals(sunDirection(SUN.lat, SUN.lng))).toBe(true)
        act(() => result.current.detach())
        rerender({image: '', sun: SUN})
        expect(sphere.visible).toBe(true)
    })
})

describe('the start view call (FR-GLB-015)', () => {
    afterEach(() => { delete (window as unknown as {go?: unknown}).go })

    it('reads the start view from the Go side', async () => {
        const start = {found: true, lat: 54.4027, lng: -2.1163};
        (window as unknown as {go: unknown}).go = {main: {App: {StartView: vi.fn(() => Promise.resolve(start))}}}
        expect(await api.startView(noop)).toEqual(start)
    })
})

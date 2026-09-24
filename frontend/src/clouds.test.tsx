// The cloud layer on the page (REQUIREMENTS.md 3.2.10): the rail button, the
// status line and the sphere the image is drawn on. Drawing itself is WebGL's
// and is checked by a person (FR-CLD-008 in TESTING.md).
import {fireEvent, render, screen} from '@testing-library/react'
import {afterEach, describe, expect, it, vi} from 'vitest'
import * as THREE from 'three'
import {api} from './api'
import {CLOUD_ALTITUDE, makeCloudSphere, showCloudImage} from './cloudLayer'
import {Rail} from './components/Rail'
import {guideSections} from './guide'
import {StatusLine} from './components/StatusLine'
import {icons} from './icons'
import {MARKER_ALTITUDE} from './markers'
import type {CloudsDTO, ProviderDTO} from './types'

const noop = () => undefined

function rail(cloudsShown: boolean, onToggleClouds: () => void = noop) {
    return render(<Rail autoRotate={false} attention={false} onToggleRotate={noop} cloudsShown={cloudsShown}
        onToggleClouds={onToggleClouds} onResetView={noop} onZoom={noop} onRefresh={noop} onStatus={noop}
        onSettings={noop} onHelp={noop} onDonate={noop}/>)
}

const eumetsat: ProviderDTO = {name: 'EUMETSAT', loading: false, refreshing: false, stale: false, retrieved: 'Retrieved 2 min ago', problem: '', nextAttempt: ''}

function clouds(patch: Partial<CloudsDTO> = {}): CloudsDTO {
    return {shown: true, line: 'Clouds: image of 15:00 UTC, 3 h ago', validTime: '2026-09-24T15:00:00Z', notice: '', provider: eumetsat, ...patch}
}

describe('FR-CLD-001 the cloud button', () => {
    it('shows the plain artwork and offers to show while hidden', () => {
        rail(false)
        const button = screen.getByLabelText('Show clouds')
        expect(button.querySelector('img')?.getAttribute('src')).toBe(icons.cloudCover)
    })

    it('shows the negative and offers to hide while shown', () => {
        rail(true)
        const button = screen.getByLabelText('Hide clouds')
        expect(button.querySelector('img')?.getAttribute('src')).toBe(icons.cloudCoverHide)
    })
})

describe('FR-HLP-004 the guide', () => {
    it('names every rail button, the cloud button included', () => {
        rail(false)
        const named = new Set(guideSections('').flatMap(s => s.entries ?? []).map(e => e.name))
        const labels = Array.from(document.querySelectorAll('.rail-btn')).map(b => b.getAttribute('aria-label'))
        expect(labels).toContain('Show clouds')
        expect(labels.filter(l => !named.has(l ?? ''))).toEqual([])
    })
})

describe('FR-CLD-002 switching the layer', () => {
    it('asks for the switch on a press', () => {
        const onToggle = vi.fn()
        rail(false, onToggle)
        fireEvent.click(screen.getByLabelText('Show clouds'))
        expect(onToggle).toHaveBeenCalledTimes(1)
    })
})

describe('FR-CLD-009 the cloud line', () => {
    it('reads the image time and age while shown', () => {
        render(<StatusLine countLine="" providers={[]} problem="" clouds={clouds()}/>)
        expect(screen.getByText('Clouds: image of 15:00 UTC, 3 h ago').className).toBe('provider')
    })

    it('is marked while the image is stale or a fetch failed (FR-CLD-010, FR-CLD-011)', () => {
        render(<StatusLine countLine="" providers={[]} problem=""
            clouds={clouds({provider: {...eumetsat, stale: true}})}/>)
        expect(screen.getByText(/Clouds:/).className).toContain('warn')
    })

    it('is absent while the layer is hidden (FR-CLD-005)', () => {
        render(<StatusLine countLine="" providers={[]} problem="" clouds={clouds({shown: false})}/>)
        expect(screen.queryByText(/Clouds:/)).toBeNull()
    })
})

describe('FR-CLD-008 the cloud sphere', () => {
    afterEach(() => vi.restoreAllMocks())

    it('sits above the texture and beneath the markers, hidden until an image arrives', () => {
        const sphere = makeCloudSphere(100)
        expect(CLOUD_ALTITUDE).toBeGreaterThan(0)
        expect(CLOUD_ALTITUDE).toBeLessThan(MARKER_ALTITUDE)
        expect((sphere.geometry as THREE.SphereGeometry).parameters.radius).toBeCloseTo(100 * (1 + CLOUD_ALTITUDE))
        expect(sphere.visible).toBe(false)
    })

    it('draws an image when given one and hides again when given none (FR-CLD-013)', () => {
        const sphere = makeCloudSphere(100)
        const texture = new THREE.Texture()
        const load = vi.fn((_: string, done: (t: THREE.Texture) => void) => done(texture))
        showCloudImage(sphere, 'data:image/png;base64,AAAA', load)
        const material = sphere.material as THREE.MeshBasicMaterial
        expect(sphere.visible).toBe(true)
        expect(material.map).toBe(texture)
        expect(material.transparent).toBe(true)
        const replaced = vi.spyOn(texture, 'dispose')
        showCloudImage(sphere, 'data:image/png;base64,BBBB', (_, done) => done(new THREE.Texture()))
        expect(replaced).toHaveBeenCalled()
        showCloudImage(sphere, '', load)
        expect(sphere.visible).toBe(false)
        expect(load).toHaveBeenCalledTimes(1)
    })

    it('loads the image with three.js by default', () => {
        const load = vi.spyOn(THREE.TextureLoader.prototype, 'load').mockImplementation((_url, done) => {
            done?.(new THREE.Texture())
            return new THREE.Texture()
        })
        const sphere = makeCloudSphere(100)
        showCloudImage(sphere, 'data:image/png;base64,AAAA')
        expect(load.mock.calls[0][0]).toBe('data:image/png;base64,AAAA')
        expect(sphere.visible).toBe(true)
    })
})

describe('the cloud calls', () => {
    afterEach(() => { delete (window as unknown as {go?: unknown}).go })

    it('reads the state and the image from the Go side', async () => {
        const state = clouds()
        const App = {Clouds: vi.fn(() => Promise.resolve(state)), CloudImage: vi.fn(() => Promise.resolve('data:image/png;base64,AAAA'))};
        (window as unknown as {go: unknown}).go = {main: {App}}
        expect(await api.clouds(noop)).toEqual(state)
        expect(await api.cloudImage(noop)).toBe('data:image/png;base64,AAAA')
    })
})

import * as THREE from 'three'
import {beforeEach, describe, expect, it, vi} from 'vitest'
import {fitAltitude, rescale, sprite, viewHalfAngle} from './markers'
import type {EventDTO} from './types'

const wildfire = {category: 'WILDFIRE', band: 0} as EventDTO

describe('marker textures', () => {
    // jsdom has no 2D canvas and says so on stderr; answer none, as markers.ts allows.
    beforeEach(() => {
        vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue(null)
    })

    // Unmarked, three reads the canvas as linear and the emoji draws washed out.
    it.each([false, true])('are marked sRGB (selected %s)', selected => {
        const material = sprite(wildfire, selected, 1).material as THREE.SpriteMaterial
        expect(material.map?.colorSpace).toBe(THREE.SRGBColorSpace)
    })
})

describe('FR-GLB-013 the fit altitude', () => {
    // The globe's silhouette spans tan(its angular radius) against tan(the
    // narrower half-angle) of the shorter side. The camera stands at the altitude
    // plus one globe radius from the centre.
    const drawnShare = (camera: THREE.PerspectiveCamera) => {
        const angularRadius = Math.asin(1 / (1 + fitAltitude(camera)))
        return Math.tan(angularRadius) / Math.tan(viewHalfAngle(camera))
    }

    it.each([[50, 1264 / 761], [50, 0.6], [50, 1]])('draws the globe at 85 to 92%% of the shorter side (fov %s, aspect %s)', (fov, aspect) => {
        const share = drawnShare(new THREE.PerspectiveCamera(fov, aspect))
        expect(share).toBeGreaterThanOrEqual(0.85)
        expect(share).toBeLessThanOrEqual(0.92)
    })
})

describe('FR-MRK-010 marker size on screen', () => {
    // The scale is the altitude over the fit altitude. A marker's size in globe
    // units over its distance from the camera is what the screen shows, so size
    // growing in step with altitude holds it steady on screen.
    it.each([0.25, 1, 3])('draws a marker at its fit-altitude size times %s', scale => {
        const marker = new THREE.Object3D()
        marker.userData.size = 2
        rescale(marker, scale)
        expect(marker.scale.x).toBe(2 * scale)
        expect(marker.scale.y).toBe(2 * scale)
    })
})

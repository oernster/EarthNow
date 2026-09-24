import * as THREE from 'three'
import {beforeEach, describe, expect, it, vi} from 'vitest'
import {rescale, sprite} from './markers'
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

import * as THREE from 'three'
import {beforeEach, describe, expect, it, vi} from 'vitest'
import {sprite} from './markers'
import type {EventDTO} from './types'

const wildfire = {category: 'wildfires', band: 0} as EventDTO

describe('marker textures', () => {
    // jsdom has no 2D canvas and says so on stderr; answer none, as markers.ts allows.
    beforeEach(() => {
        vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue(null)
    })

    // Unmarked, three reads the canvas as linear and the emoji draws washed out.
    it.each([false, true])('are marked sRGB (selected %s)', selected => {
        const material = sprite(wildfire, selected).material as THREE.SpriteMaterial
        expect(material.map?.colorSpace).toBe(THREE.SRGBColorSpace)
    })
})

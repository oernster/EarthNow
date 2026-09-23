// Marker sprites: each category's emoji drawn once to a texture and shared
// (measured in the Phase 0 spike: 2,501 sprites held a 10 ms median frame).
import * as THREE from 'three'
import {categoryOf} from './categories'
import type {EventDTO} from './types'

const CANVAS_PX = 128
const EMOJI_PX = 84
const RING_WIDTH_PX = 8
const RING_RADIUS_PX = 58
const RING_COLOUR = '#f5f7fa'
const EMOJI_FONT = '"Segoe UI Emoji", "Apple Color Emoji", "Noto Color Emoji", sans-serif'

// Sprite sizes in globe units (the globe's radius is 100). Band 0 is every
// event without a magnitude to size by (FR-MRK-004); 1 to 4 are FR-MRK-003's.
const BAND_SIZES = [2.4, 1.6, 2.4, 3.2, 4.2]

const materials = new Map<string, THREE.SpriteMaterial>()

function material(emoji: string, selected: boolean): THREE.SpriteMaterial {
    const key = `${emoji}|${selected}`
    const cached = materials.get(key)
    if (cached) return cached
    const canvas = document.createElement('canvas')
    canvas.width = canvas.height = CANVAS_PX
    const ctx = canvas.getContext('2d')
    if (ctx) {
        ctx.font = `${EMOJI_PX}px ${EMOJI_FONT}`
        ctx.textAlign = 'center'
        ctx.textBaseline = 'middle'
        ctx.fillText(emoji, CANVAS_PX / 2, CANVAS_PX / 2)
        if (selected) {
            ctx.strokeStyle = RING_COLOUR
            ctx.lineWidth = RING_WIDTH_PX
            ctx.beginPath()
            ctx.arc(CANVAS_PX / 2, CANVAS_PX / 2, RING_RADIUS_PX, 0, 2 * Math.PI)
            ctx.stroke()
        }
    }
    const made = new THREE.SpriteMaterial({map: new THREE.CanvasTexture(canvas), depthWrite: false})
    materials.set(key, made)
    return made
}

export function sprite(e: EventDTO, selected: boolean): THREE.Sprite {
    const s = new THREE.Sprite(material(categoryOf(e.category).emoji, selected))
    const size = BAND_SIZES[e.band] ?? BAND_SIZES[0]
    s.scale.set(size, size, 1)
    return s
}

// The globe's radius as a share of the globe area's shorter side (FR-GLB-013;
// 0.9 measured at 88.2% drawn in the spike).
const FIT_FILL = 0.9

export function fitAltitude(camera: THREE.PerspectiveCamera): number {
    const vHalf = THREE.MathUtils.degToRad(camera.fov) / 2
    const hHalf = Math.atan(Math.tan(vHalf) * camera.aspect)
    return 1 / (FIT_FILL * Math.sin(Math.min(vHalf, hHalf))) - 1
}

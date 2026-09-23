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

// MARKER_SCALE enlarges every marker alike, keeping the bands in proportion: the
// first sizes read too small on the globe (owner, 2026-09-23).
const MARKER_SCALE = 2

// Sprite sizes in globe units at the fit altitude (the globe's radius is 100).
// Band 0 is every event without a magnitude to size by (FR-MRK-004); 1 to 4 are
// FR-MRK-003's.
const BAND_SIZES = [2.4, 1.6, 2.4, 3.2, 4.2].map(size => size * MARKER_SCALE)

/** markerSize answers a band's sprite size in globe units at the fit altitude. */
export function markerSize(band: number): number {
    return BAND_SIZES[band] ?? BAND_SIZES[0]
}

// DRAWN_SHARE is how much of a sprite's width the emoji itself covers, so two
// markers overlap on screen when their glyphs do, not their transparent corners.
export const DRAWN_SHARE = EMOJI_PX / CANVAS_PX

const materials = new Map<string, THREE.SpriteMaterial>()

// cached answers the material drawn under key, drawing it once on first use.
function cached(key: string, draw: (ctx: CanvasRenderingContext2D) => void): THREE.SpriteMaterial {
    const found = materials.get(key)
    if (found) return found
    const canvas = document.createElement('canvas')
    canvas.width = canvas.height = CANVAS_PX
    const ctx = canvas.getContext('2d')
    if (ctx) {
        ctx.textAlign = 'center'
        ctx.textBaseline = 'middle'
        draw(ctx)
    }
    // The canvas holds sRGB pixels; left unmarked, three reads them as linear and
    // encodes them again on output, which washed 🔥 from 243,134,60 to 250,190,133
    // (measured 2026-09-23 against three 0.186).
    const texture = new THREE.CanvasTexture(canvas)
    texture.colorSpace = THREE.SRGBColorSpace
    const made = new THREE.SpriteMaterial({map: texture, depthWrite: false})
    materials.set(key, made)
    return made
}

function ring(ctx: CanvasRenderingContext2D, colour: string) {
    ctx.strokeStyle = colour
    ctx.lineWidth = RING_WIDTH_PX
    ctx.beginPath()
    ctx.arc(CANVAS_PX / 2, CANVAS_PX / 2, RING_RADIUS_PX, 0, 2 * Math.PI)
    ctx.stroke()
}

// sized gives a sprite its fit-altitude size and draws it at the current scale.
function sized(s: THREE.Sprite, size: number, scale: number): THREE.Sprite {
    s.userData.size = size
    rescale(s, scale)
    return s
}

/**
 * rescale draws a marker at scale times its fit-altitude size. The scale is the
 * altitude over the fit altitude, so a marker keeps its launch size on screen
 * while the camera zooms; at a fixed size in globe units two overlapping markers
 * would grow with the gap between them and never separate (FR-MRK-008).
 */
export function rescale(s: THREE.Object3D, scale: number) {
    const size = (s.userData.size as number) * scale
    s.scale.set(size, size, 1)
}

export function sprite(e: EventDTO, selected: boolean, scale: number): THREE.Sprite {
    const emoji = categoryOf(e.category).emoji
    const made = cached(`${emoji}|${selected}`, ctx => {
        ctx.font = `${EMOJI_PX}px ${EMOJI_FONT}`
        ctx.fillText(emoji, CANVAS_PX / 2, CANVAS_PX / 2)
        if (selected) ring(ctx, RING_COLOUR)
    })
    return sized(new THREE.Sprite(made), markerSize(e.band), scale)
}

// The count sits inside the disc at COUNT_PX, shrunk until it spans no more than
// COUNT_FILL of the disc's width, so four digits fit as well as two.
const COUNT_PX = 60
const COUNT_FILL = 0.7
const countFont = (px: number) => `bold ${px}px "Segoe UI", sans-serif`

// token reads a colour from the page's palette (style.css), its one home.
function token(name: string): string {
    return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

/** clusterSprite draws FR-MRK-007's cluster marker: a disc reading its count. */
export function clusterSprite(count: number, size: number, scale: number): THREE.Sprite {
    const made = cached(`cluster|${count}`, ctx => {
        ctx.fillStyle = token('--surface-solid')
        ctx.beginPath()
        ctx.arc(CANVAS_PX / 2, CANVAS_PX / 2, RING_RADIUS_PX, 0, 2 * Math.PI)
        ctx.fill()
        ring(ctx, token('--text'))
        const text = String(count)
        ctx.font = countFont(COUNT_PX)
        const widest = 2 * RING_RADIUS_PX * COUNT_FILL
        ctx.font = countFont(Math.min(COUNT_PX, COUNT_PX * widest / ctx.measureText(text).width))
        ctx.fillStyle = token('--text')
        ctx.fillText(text, CANVAS_PX / 2, CANVAS_PX / 2)
    })
    return sized(new THREE.Sprite(made), size, scale)
}

// The globe's radius as a share of the globe area's shorter side (FR-GLB-013;
// 0.9 measured at 88.2% drawn in the spike).
const FIT_FILL = 0.9

/** viewHalfAngle answers the camera's narrower half-angle of view, in radians. */
export function viewHalfAngle(camera: THREE.PerspectiveCamera): number {
    const vHalf = THREE.MathUtils.degToRad(camera.fov) / 2
    const hHalf = Math.atan(Math.tan(vHalf) * camera.aspect)
    return Math.min(vHalf, hHalf)
}

export function fitAltitude(camera: THREE.PerspectiveCamera): number {
    return 1 / (FIT_FILL * Math.sin(viewHalfAngle(camera))) - 1
}

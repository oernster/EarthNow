// The image layers' spheres: the clouds (FR-CLD-008) and the burnt areas
// (FR-BA-007), each a sphere just above the globe texture and beneath every
// marker, turning with the globe because it sits in the same scene the camera
// orbits. The burnt areas lie on the ground, beneath the clouds. Each image
// arrives already drawn from Go as a data URL, so nothing here reaches the network.
import * as THREE from 'three'
import {MARKER_ALTITUDE} from './markers'

// Halfway between the surface and the markers, so the markers stay above it.
export const CLOUD_ALTITUDE = MARKER_ALTITUDE / 2

// Halfway between the surface and the clouds, so the clouds lie over it.
export const BURNT_ALTITUDE = CLOUD_ALTITUDE / 2

// three-globe builds its globe from 360 / globeCurvatureResolution segments
// with a default resolution of 4 degrees (read in three-globe.mjs); the layer
// spheres match it so they curve alike.
const CURVATURE_DEGREES = 4
const WIDTH_SEGMENTS = 360 / CURVATURE_DEGREES

// three-globe turns its globe mesh this way to face the prime meridian along
// Z (read in three-globe.mjs); an image's longitudes line up only when the
// sphere is turned alike.
const PRIME_MERIDIAN_TURN = -Math.PI / 2

// Drawn before the markers, which are transparent too, so they land on top;
// the burnt areas before the clouds, so the clouds land on them.
export const CLOUD_ORDER = -1
export const BURNT_ORDER = CLOUD_ORDER - 1

/** makeImageSphere answers a hidden layer sphere at altitude, drawn in order. */
function makeImageSphere(globeRadius: number, altitude: number, order: number): THREE.Mesh {
    const geometry = new THREE.SphereGeometry(globeRadius * (1 + altitude), WIDTH_SEGMENTS, WIDTH_SEGMENTS / 2)
    const material = new THREE.MeshBasicMaterial({transparent: true, depthWrite: false})
    const sphere = new THREE.Mesh(geometry, material)
    sphere.rotation.y = PRIME_MERIDIAN_TURN
    sphere.renderOrder = order
    sphere.visible = false
    return sphere
}

/** makeCloudSphere answers the hidden cloud sphere for a globe of radius. */
export function makeCloudSphere(globeRadius: number): THREE.Mesh {
    return makeImageSphere(globeRadius, CLOUD_ALTITUDE, CLOUD_ORDER)
}

/** makeBurntSphere answers the hidden burnt-area sphere for a globe of radius. */
export function makeBurntSphere(globeRadius: number): THREE.Mesh {
    return makeImageSphere(globeRadius, BURNT_ALTITUDE, BURNT_ORDER)
}

/**
 * showLayerImage draws url on a layer sphere; an empty url hides the sphere
 * (the layer hidden or nothing to draw: FR-CLD-013 draws no veil then and
 * FR-BA-009 draws no burnt area). The previous texture is released once the
 * new one has loaded.
 */
export function showLayerImage(sphere: THREE.Mesh, url: string, load: (url: string, done: (t: THREE.Texture) => void) => void = loadTexture): void {
    const material = sphere.material as THREE.MeshBasicMaterial
    if (url === '') {
        sphere.visible = false
        return
    }
    load(url, texture => {
        texture.colorSpace = THREE.SRGBColorSpace
        material.map?.dispose()
        material.map = texture
        material.needsUpdate = true
        sphere.visible = true
    })
}

function loadTexture(url: string, done: (t: THREE.Texture) => void): void {
    new THREE.TextureLoader().load(url, done)
}

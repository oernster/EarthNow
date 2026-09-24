// The cloud layer's sphere (FR-CLD-008): a second sphere just above the globe
// texture and beneath every marker, turning with the globe because it sits in
// the same scene the camera orbits. The image arrives already drawn from Go
// (FR-CLD-006, FR-CLD-007) as a data URL, so nothing here reaches the network.
import * as THREE from 'three'
import {MARKER_ALTITUDE} from './markers'

// Halfway between the surface and the markers, so the markers stay above it.
export const CLOUD_ALTITUDE = MARKER_ALTITUDE / 2

// three-globe builds its globe from 360 / globeCurvatureResolution segments
// with a default resolution of 4 degrees (read in three-globe.mjs); the cloud
// sphere matches it so the two curve alike.
const CURVATURE_DEGREES = 4
const WIDTH_SEGMENTS = 360 / CURVATURE_DEGREES

// three-globe turns its globe mesh this way to face the prime meridian along
// Z (read in three-globe.mjs); the cloud image's longitudes line up only when
// the sphere is turned alike.
const PRIME_MERIDIAN_TURN = -Math.PI / 2

// Drawn before the markers, which are transparent too, so they land on top.
const BENEATH_MARKERS = -1

/** makeCloudSphere answers the hidden cloud sphere for a globe of radius. */
export function makeCloudSphere(globeRadius: number): THREE.Mesh {
    const geometry = new THREE.SphereGeometry(globeRadius * (1 + CLOUD_ALTITUDE), WIDTH_SEGMENTS, WIDTH_SEGMENTS / 2)
    const material = new THREE.MeshBasicMaterial({transparent: true, depthWrite: false})
    const sphere = new THREE.Mesh(geometry, material)
    sphere.rotation.y = PRIME_MERIDIAN_TURN
    sphere.renderOrder = BENEATH_MARKERS
    sphere.visible = false
    return sphere
}

/**
 * showCloudImage draws url on the sphere; an empty url hides the sphere
 * (the layer hidden or no image held: FR-CLD-013 draws no veil then).
 * The previous texture is released once the new one has loaded.
 */
export function showCloudImage(sphere: THREE.Mesh, url: string, load: (url: string, done: (t: THREE.Texture) => void) => void = loadTexture): void {
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

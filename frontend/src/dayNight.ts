// The day and night layer (FR-DAY-003, FR-DAY-008, FR-DAY-009). The globe keeps
// three-globe's own lit material, the day texture its map as before, with the
// night lights added as an emissive map; one light factor per point scales the
// day by it and the lights by what is left. The cloud sphere dims by the same
// factor. Both read one set of uniforms, so the sun moves them together.
//
// The sun is placed in world space, where the globe mesh stays still and only
// the camera turns (OrbitControls), so rotating the view needs no uniform. The
// light rule is FR-DAY-002's, taught in internal/domain/sun: the twilight limit
// and the night floor arrive from there with the sun (SunDTO), so the shader
// below restates the ramp's shape and holds none of its figures.
import * as THREE from 'three'
import type {SunDTO} from './types'

// The uniforms both materials share. dayOnly is 1 while the layer is hidden,
// which makes the light 1 everywhere: the globe is then drawn exactly as its
// plain material draws it (FR-DAY-008) and no cloud is dimmed.
export interface DayNightUniforms {
    sunDirection: {value: THREE.Vector3}
    twilightDegrees: {value: number}
    cloudNightFloor: {value: number}
    dayOnly: {value: number}
}

export function makeUniforms(): DayNightUniforms {
    return {
        sunDirection: {value: new THREE.Vector3(0, 0, 1)},
        twilightDegrees: {value: 1},
        cloudNightFloor: {value: 1},
        dayOnly: {value: 1},
    }
}

// NO_LIGHT is three-globe's own globe colour, black (read in three-globe.mjs);
// FULL_LIGHT lets the night texture through as it is.
const NO_LIGHT = 0x000000
const FULL_LIGHT = 0xffffff

const DEGREES_TO_RADIANS = Math.PI / 180
const RIGHT_ANGLE = 90

/**
 * sunDirection is the unit vector from the globe's centre to lat, lng, in
 * three-globe's convention (polar2Cartesian, read in three-globe.mjs), which is
 * where the markers and the texture's own points sit.
 */
export function sunDirection(lat: number, lng: number): THREE.Vector3 {
    const phi = (RIGHT_ANGLE - lat) * DEGREES_TO_RADIANS
    const theta = (RIGHT_ANGLE - lng) * DEGREES_TO_RADIANS
    return new THREE.Vector3(Math.sin(phi) * Math.cos(theta), Math.cos(phi), Math.sin(phi) * Math.sin(theta))
}

/** placeSun moves the sun and takes the light figures that came with it. */
export function placeSun(u: DayNightUniforms, sun: SunDTO): void {
    u.sunDirection.value.copy(sunDirection(sun.lat, sun.lng))
    u.twilightDegrees.value = sun.twilightDegrees
    u.cloudNightFloor.value = sun.cloudNightFloor
}

/** showDayNight switches the layer: hidden draws the day everywhere (FR-DAY-008). */
export function showDayNight(u: DayNightUniforms, shown: boolean): void {
    u.dayOnly.value = shown ? 0 : 1
}

// The world-space normal of a sphere centred on its mesh's origin is its
// position turned by the mesh, so no geometry normal is needed.
const VERTEX_DECLARATIONS = 'varying vec3 vSunNormal;\n'
const VERTEX_NORMAL = '#include <begin_vertex>\nvSunNormal = normalize(mat3(modelMatrix) * position);'

// FR-DAY-002's ramp, its figures from the uniforms: none at the twilight limit
// below the horizon, all at the limit above, linear between.
const FRAGMENT_DECLARATIONS = `
varying vec3 vSunNormal;
uniform vec3 sunDirection;
uniform float twilightDegrees;
uniform float cloudNightFloor;
uniform float dayOnly;
float dayLight() {
    float elevation = degrees(asin(clamp(dot(normalize(vSunNormal), sunDirection), -1.0, 1.0)));
    return max(dayOnly, clamp((elevation + twilightDegrees) / (2.0 * twilightDegrees), 0.0, 1.0));
}
`

// On the globe: the day's colour by the light (FR-DAY-003), the night's lights
// by what is left of it.
const GLOBE_DAY = '#include <map_fragment>\ndiffuseColor.rgb *= dayLight();'
const GLOBE_NIGHT = '#include <emissivemap_fragment>\ntotalEmissiveRadiance *= 1.0 - dayLight();'

// On a cloud: FR-DAY-009's share of its opacity, from the night floor to all.
const CLOUD_DIM = '#include <map_fragment>\ndiffuseColor.a *= cloudNightFloor + (1.0 - cloudNightFloor) * dayLight();'

// The part of a compiled shader the injection touches (three's WebGLProgramParametersWithUniforms).
export interface ShaderParts {
    uniforms: Record<string, {value: unknown}>
    vertexShader: string
    fragmentShader: string
}

function inject(shader: ShaderParts, u: DayNightUniforms, fragment: [string, string][]): void {
    Object.assign(shader.uniforms, u)
    shader.vertexShader = VERTEX_DECLARATIONS + shader.vertexShader.replace('#include <begin_vertex>', VERTEX_NORMAL)
    let body = shader.fragmentShader
    for (const [chunk, replacement] of fragment) body = body.replace(chunk, replacement)
    shader.fragmentShader = FRAGMENT_DECLARATIONS + body
}

/**
 * makeGlobeMaterial answers three-globe's default globe material (a black
 * Phong, read in three-globe.mjs) with the night lights as its emissive map and
 * the light factor injected. three-globe still lays the day texture on as its
 * map, exactly as it does on its own material.
 */
export function makeGlobeMaterial(u: DayNightUniforms): THREE.MeshPhongMaterial {
    const material = new THREE.MeshPhongMaterial({color: NO_LIGHT, emissive: NO_LIGHT})
    material.onBeforeCompile = shader => inject(shader, u, [
        ['#include <map_fragment>', GLOBE_DAY],
        ['#include <emissivemap_fragment>', GLOBE_NIGHT],
    ])
    return material
}

/**
 * lightNights lays the night lights texture on the globe material. The emissive
 * colour stays black until then, so the night side never glows plain white
 * while the texture is on its way.
 */
export function lightNights(material: THREE.MeshPhongMaterial, texture: THREE.Texture): void {
    texture.colorSpace = THREE.SRGBColorSpace
    material.emissiveMap?.dispose()
    material.emissiveMap = texture
    material.emissive.set(FULL_LIGHT)
    material.needsUpdate = true
}

/** dimClouds makes the cloud sphere's material dim by the light (FR-DAY-009). */
export function dimClouds(material: THREE.MeshBasicMaterial, u: DayNightUniforms): void {
    material.onBeforeCompile = shader => inject(shader, u, [['#include <map_fragment>', CLOUD_DIM]])
}

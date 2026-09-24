// The globe's layers (FR-CLD-008, FR-BA-007, FR-DAY-003 to 009, FR-TRL-002): the
// cloud and burnt-area spheres, the day and night light and the storm trails, attached to a globe
// once one exists and driven by their props after that. GlobeView owns the camera and the markers; this owns
// what is drawn on and over the globe's surface.
import {useEffect, useRef, useState} from 'react'
import type {GlobeInstance} from 'globe.gl'
import * as THREE from 'three'
import earthNight from './assets/earth-night.jpg'
import {makeBurntSphere, makeCloudSphere, showLayerImage} from './imageLayers'
import {dimClouds, lightNights, makeGlobeMaterial, makeUniforms, placeSun, showDayNight} from './dayNight'
import {TRAIL_ALTITUDE, TRAIL_COLOURS} from './trails'
import type {EventDTO, SunDTO} from './types'

export interface GlobeLayers {
    attach: (g: GlobeInstance) => void
    detach: () => void
}

export function useGlobeLayers(cloudImage: string, burntImage: string, dayNightShown: boolean, sun: SunDTO | null,
    trails: EventDTO[]): GlobeLayers {
    // The light both the globe and the clouds are drawn by (FR-DAY-003, FR-DAY-009).
    const [light] = useState(makeUniforms)
    const material = useRef<THREE.MeshPhongMaterial | null>(null)
    const clouds = useRef<THREE.Mesh | null>(null)
    const burnt = useRef<THREE.Mesh | null>(null)
    const nightsAsked = useRef(false)
    const globe = useRef<GlobeInstance | null>(null)
    const latest = useRef({cloudImage, burntImage, dayNightShown, trails})
    latest.current = {cloudImage, burntImage, dayNightShown, trails}

    // The switch (FR-DAY-006, FR-DAY-008). The night lights are large, so they
    // load on the first show rather than for a reader who keeps the layer off.
    const switchDayNight = useRef((shown: boolean) => {
        showDayNight(light, shown)
        const m = material.current
        if (!shown || !m || nightsAsked.current) return
        nightsAsked.current = true
        new THREE.TextureLoader().load(earthNight, t => lightNights(m, t))
    })

    // attach applies the props as they stand, so it does not matter whether the
    // effects below ran before the globe existed.
    const layers = useRef<GlobeLayers>({
        attach: g => {
            globe.current = g
            g.pathPoints('trail').pathPointLat((p: object) => (p as number[])[0])
                .pathPointLng((p: object) => (p as number[])[1]).pathPointAlt(TRAIL_ALTITUDE)
                .pathColor(() => TRAIL_COLOURS).pathTransitionDuration(0)
                .pathsData(latest.current.trails)
            material.current = makeGlobeMaterial(light)
            g.globeMaterial(material.current)
            clouds.current = makeCloudSphere(g.getGlobeRadius())
            dimClouds(clouds.current.material as THREE.MeshBasicMaterial, light)
            g.scene().add(clouds.current)
            showLayerImage(clouds.current, latest.current.cloudImage)
            // Burnt areas are data, not scenery: the night side does not dim them.
            burnt.current = makeBurntSphere(g.getGlobeRadius())
            g.scene().add(burnt.current)
            showLayerImage(burnt.current, latest.current.burntImage)
            switchDayNight.current(latest.current.dayNightShown)
        },
        detach: () => {
            globe.current = null
            clouds.current = null
            burnt.current = null
            material.current = null
        },
    })

    useEffect(() => {
        if (clouds.current) showLayerImage(clouds.current, cloudImage)
    }, [cloudImage])
    useEffect(() => {
        if (burnt.current) showLayerImage(burnt.current, burntImage)
    }, [burntImage])
    useEffect(() => switchDayNight.current(dayNightShown), [dayNightShown])
    useEffect(() => { if (sun) placeSun(light, sun) }, [sun, light])
    useEffect(() => { globe.current?.pathsData(trails) }, [trails])

    return layers.current
}

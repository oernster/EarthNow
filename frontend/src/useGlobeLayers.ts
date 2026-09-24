// The globe's two layers (FR-CLD-008, FR-DAY-003 to 009): the cloud sphere and
// the day and night light, attached to a globe once one exists and driven by
// their props after that. GlobeView owns the camera and the markers; this owns
// what is drawn on and over the globe's surface.
import {useEffect, useRef, useState} from 'react'
import type {GlobeInstance} from 'globe.gl'
import * as THREE from 'three'
import earthNight from './assets/earth-night.jpg'
import {makeCloudSphere, showCloudImage} from './cloudLayer'
import {dimClouds, lightNights, makeGlobeMaterial, makeUniforms, placeSun, showDayNight} from './dayNight'
import type {SunDTO} from './types'

export interface GlobeLayers {
    attach: (g: GlobeInstance) => void
    detach: () => void
}

export function useGlobeLayers(cloudImage: string, dayNightShown: boolean, sun: SunDTO | null): GlobeLayers {
    // The light both the globe and the clouds are drawn by (FR-DAY-003, FR-DAY-009).
    const [light] = useState(makeUniforms)
    const material = useRef<THREE.MeshPhongMaterial | null>(null)
    const clouds = useRef<THREE.Mesh | null>(null)
    const nightsAsked = useRef(false)
    const latest = useRef({cloudImage, dayNightShown})
    latest.current = {cloudImage, dayNightShown}

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
            material.current = makeGlobeMaterial(light)
            g.globeMaterial(material.current)
            clouds.current = makeCloudSphere(g.getGlobeRadius())
            dimClouds(clouds.current.material as THREE.MeshBasicMaterial, light)
            g.scene().add(clouds.current)
            showCloudImage(clouds.current, latest.current.cloudImage)
            switchDayNight.current(latest.current.dayNightShown)
        },
        detach: () => {
            clouds.current = null
            material.current = null
        },
    })

    useEffect(() => {
        if (clouds.current) showCloudImage(clouds.current, cloudImage)
    }, [cloudImage])
    useEffect(() => switchDayNight.current(dayNightShown), [dayNightShown])
    useEffect(() => { if (sun) placeSun(light, sun) }, [sun, light])

    return layers.current
}

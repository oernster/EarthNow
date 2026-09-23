import {describe, expect, it} from 'vitest'
import {clusterEvents, layoutKey, separatingAltitude, spreadAltitude, type Cluster} from './clusters'
import {MIN_ALTITUDE, zoomed} from './cursor'
import {markerSize} from './markers'
import type {EventDTO} from './types'

// The sphere the markers sit on and a fit altitude, as the globe gives them.
const RADIUS = 101
const FIT = 1.6
const LAUNCH = 1
const EARTH_RADIUS_KM = 6371
const KM_PER_DEGREE = EARTH_RADIUS_KM * Math.PI / 180
// The narrower half-angle of a 50 degree field of view, the globe's camera.
const HALF_ANGLE = 25 * Math.PI / 180
// clusters.ts SPREAD_FILL, restated so the geometry is checked independently.
const SPREAD_FILL_MEASURED = 0.8

function at(id: string, lat: number, lng = 0, band = 0): EventDTO {
    return {id, lat, lng, band, category: 'WILDFIRE'} as EventDTO
}

describe('clusterEvents', () => {
    it('draws two events 1 km apart at launch altitude as one cluster reading 2 (FR-MRK-007)', () => {
        const items = clusterEvents([at('a', 0), at('b', 1 / KM_PER_DEGREE)], LAUNCH, RADIUS, null)
        expect(items).toHaveLength(1)
        expect(items[0].kind).toBe('cluster')
        expect((items[0] as Cluster).members).toHaveLength(2)
    })

    it('leaves events that do not overlap alone', () => {
        const items = clusterEvents([at('a', 0), at('b', 30)], LAUNCH, RADIUS, null)
        expect(items.map(i => i.kind)).toEqual(['event', 'event'])
    })

    it('follows overlap through a chain', () => {
        // Neighbours 1.5 degrees apart overlap at launch; the ends, 3 apart, do not.
        const items = clusterEvents([at('a', 0), at('b', 1.5), at('c', 3)], LAUNCH, RADIUS, null)
        expect(items).toHaveLength(1)
        expect((items[0] as Cluster).members.map(e => e.id)).toEqual(['a', 'b', 'c'])
        expect(clusterEvents([at('a', 0), at('c', 3)], LAUNCH, RADIUS, null)).toHaveLength(2)
    })

    it('never clusters the selected event', () => {
        const items = clusterEvents([at('a', 0), at('b', 0.001), at('c', 0.002)], LAUNCH, RADIUS, 'b')
        expect(items.map(i => i.kind)).toEqual(['cluster', 'event'])
        expect(items[1].kind === 'event' && items[1].event.id).toBe('b')
    })

    it('places a cluster at its members mean position and sizes it by its largest', () => {
        const items = clusterEvents([at('a', 0, 0, 1), at('b', 1, 0, 4)], LAUNCH, RADIUS, null)
        const c = items[0] as Cluster
        expect(c.lat).toBeCloseTo(0.5)
        expect(c.lng).toBeCloseTo(0)
        expect(c.size).toBe(markerSize(4))
    })

    it('separates overlapping markers as the scale falls', () => {
        const pair = [at('a', 0), at('b', 1)]
        expect(clusterEvents(pair, LAUNCH, RADIUS, null)).toHaveLength(1)
        expect(clusterEvents(pair, LAUNCH / 4, RADIUS, null)).toHaveLength(2)
    })
})

describe('layoutKey', () => {
    it('changes when events regroup and not otherwise', () => {
        const pair = [at('a', 0), at('b', 1)]
        const together = layoutKey(clusterEvents(pair, LAUNCH, RADIUS, null))
        expect(layoutKey(clusterEvents(pair, LAUNCH * 0.99, RADIUS, null))).toBe(together)
        expect(layoutKey(clusterEvents(pair, LAUNCH / 4, RADIUS, null))).not.toBe(together)
    })
})

describe('spreadAltitude', () => {
    it('puts the widest member at the spread share of the half-angle', () => {
        // Members 10 degrees apart sit 5 degrees either side of their mean.
        const altitude = spreadAltitude([at('a', 0), at('b', 10)], HALF_ANGLE)
        const theta = 5 * Math.PI / 180
        const seen = Math.atan(Math.sin(theta) / (1 + altitude - Math.cos(theta)))
        expect(seen).toBeCloseTo(HALF_ANGLE * SPREAD_FILL_MEASURED)
    })
})

describe('separatingAltitude (FR-MRK-008)', () => {
    it('stops at the first step at which every member stands alone', () => {
        const pair = [at('a', 0), at('b', 1)]
        const altitude = separatingAltitude(pair, FIT, FIT, RADIUS, HALF_ANGLE)
        expect(clusterEvents(pair, altitude / FIT, RADIUS, null)).toHaveLength(2)
        expect(clusterEvents(pair, zoomed(altitude, false) / FIT, RADIUS, null)).toHaveLength(1)
    })

    it('stops once the members fill the view in several groups', () => {
        // A 12 degree chain holding a pair 100 m apart that no zoom parts.
        const chain = [0, 1.5, 3, 4.5, 6, 7.5, 9, 10.5, 12].map((lat, i) => at(`c${i}`, lat))
        const members = [...chain, at('twin', 100 / 1000 / KM_PER_DEGREE)]
        const altitude = separatingAltitude(members, FIT, FIT, RADIUS, HALF_ANGLE)
        expect(altitude).toBeGreaterThan(MIN_ALTITUDE)
        expect(altitude).toBeLessThanOrEqual(spreadAltitude(members, HALF_ANGLE))
        expect(zoomed(altitude, false)).toBeGreaterThan(spreadAltitude(members, HALF_ANGLE))
        expect(clusterEvents(members, altitude / FIT, RADIUS, null)).toHaveLength(members.length - 1)
    })

    it('never moves the camera away from a cluster already wider than the view', () => {
        const chain = [0, 1.5, 3, 4.5, 6, 7.5, 9, 10.5, 12].map((lat, i) => at(`c${i}`, lat))
        const low = 0.3
        expect(separatingAltitude(chain, low, FIT, RADIUS, HALF_ANGLE)).toBeLessThan(low)
    })

    it('stops at the minimum altitude when the members never separate', () => {
        const pair = [at('a', 0), at('b', 1 / KM_PER_DEGREE)]
        expect(separatingAltitude(pair, FIT, FIT, RADIUS, HALF_ANGLE)).toBe(MIN_ALTITUDE)
        expect(separatingAltitude(pair, MIN_ALTITUDE, FIT, RADIUS, HALF_ANGLE)).toBe(MIN_ALTITUDE)
    })
})

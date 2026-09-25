// Marker clustering (FR-MRK-007, FR-MRK-008). globe.gl offers none, so it is
// written here, pure: no DOM, no camera, only positions, sizes and a scale.
//
// Two markers overlap on screen when their drawn glyphs do. Every marker keeps
// its launch size on screen (markers.ts rescale), so the glyph's radius in globe
// units is its fit-altitude radius times the scale (altitude over fit altitude)
// and the test reduces to the straight-line distance between the two positions
// against the sum of the two radii. That holds exactly for markers facing the
// camera; towards the limb the globe foreshortens the gap and markers there can
// touch on screen while held apart here.
//
// Overlap is followed through: where A overlaps B and B overlaps C, all three
// are one cluster, since drawing A and C apart would still leave B under both.
import {HALF_STEP, MIN_ALTITUDE, zoomed} from './cursor'
import {DRAWN_SHARE, markerSize} from './markers'
import type {EventDTO} from './types'

export interface Placed {
    kind: 'event'
    lat: number
    lng: number
    event: EventDTO
}

export interface Cluster {
    kind: 'cluster'
    lat: number
    lng: number
    // key names the members, so a layout can tell whether anything regrouped.
    key: string
    members: EventDTO[]
    // size is the largest member's fit-altitude size, so a cluster is never
    // drawn smaller than the biggest marker it stands for.
    size: number
}

export type MarkerItem = Placed | Cluster

type Vec = [number, number, number]

function unit(lat: number, lng: number): Vec {
    const phi = lat * Math.PI / 180
    const lambda = lng * Math.PI / 180
    return [Math.cos(phi) * Math.cos(lambda), Math.cos(phi) * Math.sin(lambda), Math.sin(phi)]
}

function root(parent: number[], i: number): number {
    while (parent[i] !== i) {
        parent[i] = parent[parent[i]]
        i = parent[i]
    }
    return i
}

// mean answers the direction of the points' mean position on the sphere.
function mean(points: Vec[]): Vec {
    const [x, y, z] = points.reduce<Vec>((sum, p) => [sum[0] + p[0], sum[1] + p[1], sum[2] + p[2]], [0, 0, 0])
    const length = Math.hypot(x, y, z)
    return [x / length, y / length, z / length]
}

function latLng([x, y, z]: Vec): {lat: number; lng: number} {
    return {
        lat: Math.atan2(z, Math.hypot(x, y)) * 180 / Math.PI,
        lng: Math.atan2(y, x) * 180 / Math.PI,
    }
}

/**
 * clusterEvents answers what the globe draws at a scale: each event alone where
 * nothing overlaps it, one cluster for each overlapping group. radius is the
 * sphere the markers sit on, in globe units. The pinned event (the selection)
 * is never clustered, so its selection ring is always visible (FR-MRK-006).
 * Items keep the order of their first member in events.
 */
export function clusterEvents(events: readonly EventDTO[], scale: number, radius: number,
    pinnedId: string | null): MarkerItem[] {
    const free = events.filter(e => e.id !== pinnedId)
    const points = free.map(e => unit(e.lat, e.lng))
    const reach = free.map(e => markerSize(e.band) * scale * DRAWN_SHARE / 2)
    const parent = free.map((_, i) => i)
    for (let i = 0; i < free.length; i++) {
        for (let j = i + 1; j < free.length; j++) {
            const [a, b] = [points[i], points[j]]
            const apart = radius * Math.hypot(a[0] - b[0], a[1] - b[1], a[2] - b[2])
            if (apart < reach[i] + reach[j]) parent[root(parent, i)] = root(parent, j)
        }
    }
    const groups = new Map<number, number[]>()
    free.forEach((_, i) => {
        const r = root(parent, i)
        groups.set(r, [...(groups.get(r) ?? []), i])
    })
    const items: MarkerItem[] = [...groups.values()].map(group => {
        if (group.length === 1) {
            const e = free[group[0]]
            return {kind: 'event', lat: e.lat, lng: e.lng, event: e}
        }
        const members = group.map(i => free[i])
        return {
            kind: 'cluster',
            ...latLng(mean(group.map(i => points[i]))),
            key: members.map(e => e.id).sort().join(' '),
            members,
            size: Math.max(...members.map(e => markerSize(e.band))),
        }
    })
    const pinned = events.find(e => e.id === pinnedId)
    if (pinned) items.push({kind: 'event', lat: pinned.lat, lng: pinned.lng, event: pinned})
    return items
}

/** layoutKey names a layout, equal for two layouts that group alike. */
export function layoutKey(items: readonly MarkerItem[]): string {
    return items.map(item => item.kind === 'cluster' ? `[${item.key}]` : item.event.id).join(' ')
}

// SPREAD_FILL is the share of the view's half-angle the members may span once
// the camera has come to them, so none lands at the very edge.
const SPREAD_FILL = 0.8

/**
 * spreadAltitude answers the altitude, in globe radii, at which a camera over
 * the members' mean position sees every member within SPREAD_FILL of its
 * narrower half-angle of view. A member at angle theta from the view's centre
 * appears at atan(sin theta / (1 + altitude - cos theta)) from the camera's
 * axis; solving that for the altitude gives the form below.
 */
export function spreadAltitude(members: readonly EventDTO[], halfAngle: number): number {
    const points = members.map(e => unit(e.lat, e.lng))
    const centre = mean(points)
    const widest = Math.max(...points.map(p =>
        Math.acos(Math.min(1, p[0] * centre[0] + p[1] * centre[1] + p[2] * centre[2]))))
    return Math.cos(widest) + Math.sin(widest) / Math.tan(halfAngle * SPREAD_FILL) - 1
}

/**
 * atClosest reports whether the camera is at its closest zoom, where activating a
 * cluster can separate nothing more (FR-MRK-011). Within half a zoom step of the
 * minimum counts, since the camera reads its own altitude back with float error.
 */
export function atClosest(altitude: number): boolean {
    return altitude <= MIN_ALTITUDE * HALF_STEP
}

/**
 * separatingAltitude answers where activating a cluster takes the camera
 * (FR-MRK-008), stepping down from altitude a zoom press at a time and stopping
 * at the first step where every member stands alone. It also stops where the
 * members have come apart into several groups and already fill the view
 * (spreadAltitude), since going lower would lose some off its edge. Failing
 * both it stops at the minimum altitude, where members at one place stay together.
 */
export function separatingAltitude(members: readonly EventDTO[], altitude: number,
    fit: number, radius: number, halfAngle: number): number {
    const spread = spreadAltitude(members, halfAngle)
    let at = altitude
    do {
        at = zoomed(at, true)
        const groups = clusterEvents(members, at / fit, radius, null).length
        if (groups === members.length || (groups > 1 && at <= spread)) return at
    } while (at > MIN_ALTITUDE)
    return at
}

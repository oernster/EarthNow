// A cluster the closest zoom cannot separate (FR-MRK-011, FR-MRK-012): the
// reproduction that found it, the closest-zoom test and the member list.
import {fireEvent, render, screen} from '@testing-library/react'
import {describe, expect, it, vi} from 'vitest'
import {atClosest, clusterEvents} from './clusters'
import {ClusterList} from './components/ClusterList'
import {MIN_ALTITUDE, ZOOM_FACTOR} from './cursor'
import type {EventDTO} from './types'

function quake(id: string, lat: number, lng: number): EventDTO {
    return {id, provider: 'USGS', category: 'EARTHQUAKE', title: `M 4.3 ${id}`, description: '', lat, lng, at: '', dayOnly: false,
        reported: 'Observed 3 h ago', retrievedAt: '', retrieved: '', measurement: '', depth: '', trail: [], band: 1,
        sourceUrl: '', sourceText: '', ended: false}
}

// Two M4.3 quakes 15 km NW of Mantoudi, Greece, 2.0 km apart (USGS, 2026-09-25).
const PAIR = [quake('m1', 38.9246, 23.3956), quake('m2', 38.915, 23.3755)]

describe('FR-MRK-011 a cluster the closest zoom cannot separate', () => {
    it('stays one cluster at the minimum altitude, which is why it needs a list', () => {
        const fit = 1.629
        expect(clusterEvents(PAIR, MIN_ALTITUDE / fit, 101, null)).toHaveLength(1)
    })

    it('counts the camera at its closest within half a zoom step of the minimum', () => {
        expect(atClosest(MIN_ALTITUDE)).toBe(true)
        expect(atClosest(MIN_ALTITUDE * 1.0001)).toBe(true)
        expect(atClosest(MIN_ALTITUDE * ZOOM_FACTOR)).toBe(false)
    })
})

describe('FR-MRK-012 the member list', () => {
    it('names each member with its time and opens the one chosen', () => {
        const onChoose = vi.fn()
        render(<ClusterList members={PAIR} onChoose={onChoose} onClose={() => undefined}/>)
        expect(screen.getByRole('dialog').textContent).toContain('2 events')
        fireEvent.click(screen.getByText(/M 4\.3 m2/))
        expect(onChoose).toHaveBeenCalledWith(PAIR[1])
        expect(screen.getAllByText('Observed 3 h ago')).toHaveLength(2)
    })
})

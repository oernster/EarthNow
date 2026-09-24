import {describe, expect, it} from 'vitest'
import {categoryCounts, clusterTitle, eventTitle} from './categories'

describe('the tooltip titles (FR-MRK-005, FR-MRK-007)', () => {
    it('words an event by its emoji and a cluster by its counts, largest first', () => {
        expect(eventTitle({category: 'VOLCANO', title: 'Etna'})).toBe('🌋 Etna')
        const members = ['FLOOD', 'EARTHQUAKE', 'EARTHQUAKE'].map(category => ({category}))
        expect(clusterTitle(members)).toBe('3 events: 〰️ 2  🌊 1')
    })
})

describe('categoryCounts', () => {
    it('counts by category, the most numerous first', () => {
        const events = ['VOLCANO', 'EARTHQUAKE', 'EARTHQUAKE', 'NOT_A_CATEGORY'].map(category => ({category}))
        expect(categoryCounts(events).map(c => [c.category.key, c.count]))
            .toEqual([['EARTHQUAKE', 2], ['VOLCANO', 1], ['OTHER', 1]])
    })
})

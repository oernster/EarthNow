import {describe, expect, it} from 'vitest'
import {categoryCounts} from './categories'

describe('categoryCounts', () => {
    it('counts by category, the most numerous first', () => {
        const events = ['VOLCANO', 'EARTHQUAKE', 'EARTHQUAKE', 'NOT_A_CATEGORY'].map(category => ({category}))
        expect(categoryCounts(events).map(c => [c.category.key, c.count]))
            .toEqual([['EARTHQUAKE', 2], ['VOLCANO', 1], ['OTHER', 1]])
    })
})

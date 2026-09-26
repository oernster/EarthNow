import {render, screen, within} from '@testing-library/react'
import {describe, expect, it} from 'vitest'
import {CATEGORIES, categoryOf} from './categories'
import {Key} from './components/Key'

const noop = () => undefined

describe('FR-FLT-001 the key offers every category', () => {
    // A store holding only quakes and storms still offers all seven toggles, so the
    // key stays steady as events come and go; the rest read 0 (FR-KEY-004).
    it('lists every category in DATA-002 order, a category with no events reading 0', () => {
        render(<Key counts={{EARTHQUAKE: 3, SEVERE_STORM: 1}} providers={[]}
            hiddenCategories={new Set()} hiddenProviders={new Set()}
            onToggleCategory={noop} onToggleProvider={noop} onShowAll={noop}/>)
        const rows = screen.getAllByRole('button', {pressed: true})
        expect(rows.map(r => within(r).getByText(/./, {selector: '.key-name'}).textContent))
            .toEqual(CATEGORIES.map(c => c.name))
        const count = (name: string) =>
            rows.find(r => r.textContent?.includes(name))?.querySelector('.key-count')?.textContent
        expect(count('Earthquake')).toBe('3')
        expect(count('Volcano')).toBe('0')
        expect(screen.getByRole('button', {name: 'All events'})).toBeTruthy()
    })

    // DATA-002 as amended (45): only the categories a source feeds have a row;
    // landslides, drought and dust haze are Other.
    it('DATA-002 lists only the seven fed categories, Other taking any other kind', () => {
        render(<Key counts={{}} providers={[]} hiddenCategories={new Set()} hiddenProviders={new Set()}
            onToggleCategory={noop} onToggleProvider={noop} onShowAll={noop}/>)
        expect(screen.getAllByRole('button', {pressed: true}).map(r => r.querySelector('.key-name')!.textContent))
            .toEqual(['Earthquake', 'Volcano', 'Wildfire', 'Severe storm', 'Flood', 'Ice', 'Other'])
        expect(categoryOf('LANDSLIDE').key).toBe('OTHER')
    })

    // FR-KEY-001: each row is the category's emoji, then its name.
    it('FR-KEY-001 draws each category as its emoji beside its name', () => {
        render(<Key counts={{}} providers={[]} hiddenCategories={new Set()} hiddenProviders={new Set()}
            onToggleCategory={noop} onToggleProvider={noop} onShowAll={noop}/>)
        const rows = screen.getAllByRole('button', {pressed: true})
        expect(rows.map(r => [r.querySelector('.key-emoji')!.textContent, r.querySelector('.key-name')!.textContent]))
            .toEqual(CATEGORIES.map(c => [c.emoji, c.name]))
        for (const r of rows) expect(r.firstElementChild!.className).toBe('key-emoji')
    })
})

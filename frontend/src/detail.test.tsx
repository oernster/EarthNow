// The detail panel's content and its closing (FR-SEL-002, 003, 007, 008,
// FR-GEO-007).
import {act, fireEvent, render, screen} from '@testing-library/react'
import {describe, expect, it, vi} from 'vitest'
import {useState} from 'react'
import {guideSections} from './guide'
import type {EventDTO} from './types'

const PLACE = '49 km SW of Montana, Alaska, United States of America'
vi.mock('./api', () => ({api: {place: vi.fn(() => Promise.resolve(PLACE)), openSource: vi.fn()}}))

const {DetailPanel} = await import('./components/DetailPanel')

const AT = '2026-09-20T10:00:00Z'
const RETRIEVED_AT = '2026-09-20T10:05:00Z'
const full: EventDTO = {
    id: 'USGS/ak1', provider: 'USGS', category: 'EARTHQUAKE', title: 'M 3.1 near Montana', description: '',
    lat: 61.899, lng: -150.919, at: AT, dayOnly: false, reported: 'Reported 2 h ago', retrievedAt: RETRIEVED_AT,
    retrieved: 'Retrieved 1 h ago', measurement: 'Magnitude 3.1', depth: '18.4 km, shallow', trail: [], band: 2, sourceUrl: 'https://earthquake.usgs.gov/e',
    sourceText: '', ended: false, ongoing: false,
}
const noop = () => undefined

async function show(event: EventDTO, inView = true) {
    render(<DetailPanel event={event} inView={inView} onClose={noop} onProblem={noop}/>)
    await act(async () => { await Promise.resolve() })
}

const row = (term: string) => screen.queryByText(term, {selector: 'dt'})?.nextElementSibling?.textContent ?? null

describe('the detail panel', () => {
    it('FR-SEL-002 shows every field of an event; a missing measurement omits its row', async () => {
        await show(full)
        expect(screen.getByRole('heading').textContent).toContain(full.title)
        expect(row('Category')).toBe('Earthquake')
        expect(row('Provider')).toBe('USGS')
        expect(row('Position')).toBe('61.8990, -150.9190')
        expect(row('Event time')).toContain(full.reported)
        expect(row('Retrieved')).toContain(full.retrieved)
        expect(row('Measurement')).toBe('Magnitude 3.1')
        expect(screen.getByRole('button', {name: 'Open the source page'})).toBeTruthy()
    })

    it('FR-SEL-002 leaves out the measurement row where there is none', async () => {
        await show({...full, measurement: ''})
        expect(row('Measurement')).toBeNull()
    })

    it('FR-SEL-010 shows the depth as the Go side words it', async () => {
        await show(full)
        expect(row('Depth')).toBe('18.4 km, shallow')
    })

    it('FR-SEL-014 has the guide explain the fixed depth of 10 km', () => {
        const words = guideSections('').flatMap(s => s.paragraphs ?? []).join(' ')
        expect(words).toContain('assigns 10 km, a fixed depth')
    })

    it('FR-SEL-013 leaves out the depth row where there is none', async () => {
        await show({...full, depth: ''})
        expect(row('Depth')).toBeNull()
    })

    it('FR-SEL-003 gives each time as freshness wording, exact UTC and local time', async () => {
        await show(full)
        for (const [term, iso, fresh] of [['Event time', AT, full.reported], ['Retrieved', RETRIEVED_AT, full.retrieved]]) {
            const text = row(term)!
            expect(text).toContain(fresh)
            expect(text).toContain(`${iso.slice(0, 10)} ${iso.slice(11, 19)} UTC`)
            expect(text).toContain(`${new Date(iso).toLocaleString()} local`)
        }
    })

    it('FR-SEL-004 gives an ongoing event its report week and no exact time', async () => {
        const reported = 'Continuing: report for 10 to 16 Sep 2026, issued 17 Sep 2026'
        await show({...full, provider: 'GVP', category: 'VOLCANO', at: '2026-09-17T00:00:00Z', dayOnly: true, reported, ongoing: true})
        expect(row('Event time')).toBe(reported)
    })

    it('FR-GEO-007 repeats the place line', async () => {
        await show(full)
        expect(row('Where')).toBe(PLACE)
    })

    it('FR-SEL-008 stays open saying an event has left the current view', async () => {
        await show(full, false)
        expect(screen.getByText('This event is no longer in the current view.')).toBeTruthy()
    })

    it('FR-SEL-007 closing clears the selection and hands focus back to the opener', async () => {
        // The selection's owner, as App.tsx holds it: Close sets it to none.
        function Owner() {
            const [selected, setSelected] = useState<EventDTO | null>(null)
            return <>
                <button onClick={() => setSelected(full)}>opener</button>
                {selected && <DetailPanel event={selected} inView onClose={() => setSelected(null)} onProblem={noop}/>}
            </>
        }
        render(<Owner/>)
        const opener = screen.getByRole('button', {name: 'opener'})
        opener.focus()
        fireEvent.click(opener)
        await act(async () => { await Promise.resolve() })
        fireEvent.click(screen.getByRole('button', {name: 'Close'}))
        expect(screen.queryByLabelText('Event details')).toBeNull()
        expect(document.activeElement).toBe(opener)
    })
})

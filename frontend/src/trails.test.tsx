// Storm trails on the page (REQUIREMENTS.md 3.2.12): which events draw one, what
// the globe is handed and the Settings box. The line itself is WebGL's and is
// checked by a person (FR-TRL-002 in TESTING.md).
import {act, fireEvent, render, renderHook, screen} from '@testing-library/react'
import {describe, expect, it, vi} from 'vitest'
import {SettingsDialog} from './components/SettingsDialog'
import {TRAIL_ALTITUDE, TRAIL_COLOURS, trailed} from './trails'
import type {EventDTO, SettingsDTO} from './types'
import {useGlobeLayers} from './useGlobeLayers'

function event(id: string, trail: [number, number][]): EventDTO {
    return {id, provider: 'EONET', category: 'SEVERE_STORM', title: id, description: '', lat: 0, lng: 0, at: '',
        dayOnly: false, reported: '', retrievedAt: '', retrieved: '', measurement: '', depth: '', trail, band: 0,
        sourceUrl: '', sourceText: '', ended: false}
}
const storm = event('Polo', [[10, -60], [12, -62], [14, -63]])
const quake = event('quake', [])

// fakeGlobe records every call and answers itself, as globe.gl's builder does.
function fakeGlobe() {
    const calls: Array<[string, unknown[]]> = []
    const g: Record<string, unknown> = {getGlobeRadius: () => 100, scene: () => ({add: () => undefined})}
    return {calls, g: new Proxy(g, {get: (t, name: string) => t[name] ?? ((...args: unknown[]) => { calls.push([name, args]); return proxyOf() })})}
    function proxyOf(): unknown { return fake.g }
}
let fake: ReturnType<typeof fakeGlobe>

describe('FR-TRL-001 which events draw a trail', () => {
    it('keeps storms with two points or more; none while switched off (FR-TRL-003)', () => {
        expect(trailed([storm, quake, event('one fix', [[1, 1]])], true)).toEqual([storm])
        expect(trailed([storm], false)).toEqual([])
    })
})

describe('FR-TRL-002 the trail on the globe', () => {
    it('hands the globe each storm faded from faint to strong at the marker altitude', () => {
        fake = fakeGlobe()
        const {result, rerender} = renderHook(({trails}) => useGlobeLayers('', false, null, trails),
            {initialProps: {trails: [storm]}})
        act(() => result.current.attach(fake.g as never))
        expect(fake.calls).toContainEqual(['pathPoints', ['trail']])
        expect(fake.calls).toContainEqual(['pathPointAlt', [TRAIL_ALTITUDE]])
        const colour = fake.calls.find(([name]) => name === 'pathColor')![1][0] as () => string[]
        expect(colour()).toEqual(TRAIL_COLOURS)
        expect(fake.calls.filter(([name]) => name === 'pathsData').at(-1)).toEqual(['pathsData', [[storm]]])
        const lat = fake.calls.find(([name]) => name === 'pathPointLat')![1][0] as (p: object) => number
        const lng = fake.calls.find(([name]) => name === 'pathPointLng')![1][0] as (p: object) => number
        expect([lat([10, -60]), lng([10, -60])]).toEqual([10, -60])
        rerender({trails: []})
        expect(fake.calls.filter(([name]) => name === 'pathsData').at(-1)).toEqual(['pathsData', [[]]])
    })
})

describe('FR-TRL-004 the Settings box', () => {
    it('offers "Show storm tracks" and applies a change at once', () => {
        const onChange = vi.fn()
        const settings = {autoRotate: true, magnitude: '2.5', speed: 'normal', windowKey: '24h', hiddenCategories: [],
            hiddenProviders: [], cloudsShown: false, dayNightShown: true, trailsShown: true} satisfies SettingsDTO
        render(<SettingsDialog settings={settings} choices={{magnitudes: [], speeds: []}} onChange={onChange} onClose={() => undefined}/>)
        const box = screen.getByLabelText('Show storm tracks') as HTMLInputElement
        expect(box.checked).toBe(true)
        fireEvent.click(box)
        expect(onChange).toHaveBeenCalledWith({trailsShown: false})
    })
})

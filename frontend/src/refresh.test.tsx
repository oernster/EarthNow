// FR-STS-007: a refresh that is under way says so in the status line and on
// the Refresh button; a fast one is still seen for one turn.
import {act, render, renderHook} from '@testing-library/react'
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {Rail} from './components/Rail'
import {StatusLine} from './components/StatusLine'
import {RAIL_LABELS} from './railLabels'
import type {ProviderDTO} from './types'
import {REFRESH_TURN_MS, useHeld} from './useHeld'

const noop = () => undefined

const usgs = (over: Partial<ProviderDTO>): ProviderDTO =>
    ({name: 'USGS', loading: false, refreshing: false, stale: false, retrieved: 'Retrieved under a minute ago', problem: '', nextAttempt: '', ...over})

function refreshButton(refreshing: boolean): HTMLElement {
    render(<Rail autoRotate={false} attention={false} refreshing={refreshing} onToggleRotate={noop} onResetView={noop}
        onZoom={noop} onRefresh={noop} onStatus={noop} onSettings={noop} onHelp={noop} onDonate={noop}/>)
    return document.querySelector<HTMLElement>(`[aria-label="${RAIL_LABELS.refresh}"]`)!
}

describe('FR-STS-007 a refresh under way says so', () => {
    it('reads refreshing in the status line, loading still reading loading', () => {
        render(<StatusLine countLine="" problem="" providers={[usgs({refreshing: true}), usgs({name: 'EONET', loading: true})]}/>)
        expect(document.querySelector('.status')!.textContent).toContain('USGS: refreshing')
        expect(document.querySelector('.status')!.textContent).toContain('EONET: loading')
    })

    it('turns the Refresh button while refreshing and marks it busy', () => {
        const button = refreshButton(true)
        expect(button.classList.contains('busy')).toBe(true)
        expect(button.getAttribute('aria-busy')).toBe('true')
        expect((document.querySelector('.rail') as HTMLElement).style.getPropertyValue('--refresh-turn')).toBe(`${REFRESH_TURN_MS}ms`)
    })

    it('leaves the Refresh button still otherwise', () => {
        expect(refreshButton(false).classList.contains('busy')).toBe(false)
    })
})

describe('FR-STS-007 a fast refresh is held for one turn', () => {
    beforeEach(() => { vi.useFakeTimers() })
    afterEach(() => { vi.useRealTimers() })

    it('holds a short burst for the full turn, then lets go', () => {
        const {result, rerender} = renderHook(({on}) => useHeld(on, REFRESH_TURN_MS), {initialProps: {on: false}})
        expect(result.current).toBe(false)
        rerender({on: true})
        expect(result.current).toBe(true)
        act(() => { vi.advanceTimersByTime(REFRESH_TURN_MS / 4) })
        rerender({on: false})
        expect(result.current).toBe(true)
        act(() => { vi.advanceTimersByTime(REFRESH_TURN_MS / 2) })
        expect(result.current).toBe(true)
        act(() => { vi.advanceTimersByTime(REFRESH_TURN_MS / 4) })
        expect(result.current).toBe(false)
    })

    it('lets a long refresh go as soon as it ends', () => {
        const {result, rerender} = renderHook(({on}) => useHeld(on, REFRESH_TURN_MS), {initialProps: {on: true}})
        act(() => { vi.advanceTimersByTime(REFRESH_TURN_MS * 2) })
        rerender({on: false})
        act(() => { vi.advanceTimersByTime(0) })
        expect(result.current).toBe(false)
    })
})

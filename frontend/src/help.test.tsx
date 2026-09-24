// The rail's newer surfaces: the reading dialogs, the Help menu, the provider
// status popover, the guide's words and the rail's order (FR-RAIL-002, FR-HLP,
// FR-STS-003).
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {act, fireEvent, render, screen} from '@testing-library/react'
import {useRef} from 'react'
import {DetailPanel} from './components/DetailPanel'
import {Dialog} from './components/Dialog'
import {HelpDialog} from './components/HelpDialogs'
import {HelpMenu} from './components/HelpMenu'
import {Rail} from './components/Rail'
import {ProductName} from './product'
import {donateLabel} from './railLabels'
import {donate} from './donate'
import {StatusPanel, needsAttention} from './components/StatusPanel'
import {CATEGORIES} from './categories'
import {COVERS} from './guide'
import {indexOnRing, liveStops, useRing} from './ring'
import {isTopmostSurface} from './useAutoScroll'
import {layOut, named, overflowing, unlayOut} from './testLayout'
import type {ProviderDTO} from './types'

const key = (k: string) => fireEvent.keyDown(document.activeElement ?? document.body, {key: k})

const provider = (over: Partial<ProviderDTO> = {}): ProviderDTO =>
    ({name: 'USGS', loading: false, refreshing: false, stale: false, retrieved: 'Retrieved 3 min ago', problem: '', nextAttempt: '', ...over})

describe('a reading body', () => {
    beforeEach(layOut)
    afterEach(unlayOut)

    it('is a stop only while it overflows (NFR-KBD-007)', () => {
        render(<Dialog title="Licence" onClose={() => undefined} reading><p>text</p></Dialog>)
        const body = document.querySelector<HTMLElement>('[data-reading]')!
        const frame = screen.getByRole('dialog')
        overflowing(body, false)
        expect(liveStops(frame)).not.toContain(body)
        overflowing(body, true)
        expect(liveStops(frame)).toContain(body)
    })

    it('is never the stop a dialog opens on, even while it overflows', () => {
        // Every element overflows, so the body is a live stop ahead of Close.
        vi.spyOn(HTMLElement.prototype, 'scrollHeight', 'get').mockReturnValue(1000)
        render(<Dialog title="Licence" onClose={() => undefined} reading><p>text</p></Dialog>)
        expect(named()).toBe('Close')
        vi.restoreAllMocks()
    })

    it('refuses focus from a click', () => {
        render(<Dialog title="Licence" onClose={() => undefined} reading><p>text</p></Dialog>)
        const body = document.querySelector<HTMLElement>('[data-reading]')!
        expect(fireEvent.mouseDown(body)).toBe(false)
    })

    it('closes on Escape', () => {
        const onClose = vi.fn()
        render(<Dialog title="Settings" onClose={onClose}><p>text</p></Dialog>)
        fireEvent.keyDown(window, {key: 'Escape'})
        expect(onClose).toHaveBeenCalledOnce()
    })
})

describe('the Help menu', () => {
    beforeEach(layOut)
    afterEach(unlayOut)

    function Shell({onChoose}: {onChoose: (k: string) => void}) {
        const shell = useRef<HTMLDivElement>(null)
        useRing(shell)
        return <div ref={shell}>
            <button data-stop type="button">before</button>
            <HelpMenu icon="" onChoose={onChoose}/>
            <button data-stop type="button">after</button>
        </div>
    }

    it('drops open on Down onto its first item and walks the items wrapping', () => {
        render(<Shell onChoose={() => undefined}/>)
        screen.getByLabelText('Help').focus()
        key('ArrowDown')
        expect(named()).toBe('Guide')
        key('ArrowUp')
        expect(named()).toBe('Third-party notices')
        key('ArrowDown')
        expect(named()).toBe('Guide')
    })

    it('closes on Escape and hands focus back to the Help button', () => {
        render(<Shell onChoose={() => undefined}/>)
        screen.getByLabelText('Help').focus()
        key('ArrowDown')
        key('Escape')
        expect(screen.queryByRole('menu')).toBeNull()
        expect(named()).toBe('Help')
    })

    it('steps the ring on from the Help button when Tab leaves the menu', () => {
        render(<Shell onChoose={() => undefined}/>)
        screen.getByLabelText('Help').focus()
        key('ArrowDown')
        key('Tab')
        expect(named()).toBe('after')
        expect(screen.queryByRole('menu')).toBeNull()
    })

    it('chooses on activation and leaves focus on the button', () => {
        const onChoose = vi.fn()
        render(<Shell onChoose={onChoose}/>)
        screen.getByLabelText('Help').focus()
        key('ArrowDown')
        fireEvent.click(screen.getByText('Licence'))
        expect(onChoose).toHaveBeenCalledWith('licence')
        expect(named()).toBe('Help')
    })

    it('closes on a press outside it', () => {
        render(<Shell onChoose={() => undefined}/>)
        fireEvent.click(screen.getByLabelText('Help'))
        expect(screen.getByRole('menu')).toBeTruthy()
        fireEvent.mouseDown(document.body)
        expect(screen.queryByRole('menu')).toBeNull()
    })

    it('counts focus in a popup as resting on the stop that controls it', () => {
        render(<Shell onChoose={() => undefined}/>)
        screen.getByLabelText('Help').focus()
        key('ArrowDown')
        const list = liveStops(document.body)
        const button = screen.getByRole('button', {name: 'Help'})
        expect(indexOnRing(list, document.activeElement as HTMLElement)).toBe(list.indexOf(button))
        expect(indexOnRing(list, document.createElement('span'))).toBe(-1)
    })
})

describe('FR-RAIL-002 the rail', () => {
    // FR-DON-008: donate is last, straight after Help. NFR-A11Y-003: every rail
    // button is icon-only, so each carries its name as tooltip and accessible name.
    it('holds its buttons in the stated order, FR-DON-008 donate last, NFR-A11Y-003 each named', () => {
        const noop = () => undefined
        render(<Rail autoRotate={false} attention={false} onToggleRotate={noop} cloudsShown={false} onToggleClouds={noop} onResetView={noop} onZoom={noop}
            onRefresh={noop} onStatus={noop} onSettings={noop} onHelp={noop} onDonate={noop}/>)
        const buttons = Array.from(document.querySelectorAll('.rail-btn'))
        const labels = buttons.map(b => b.getAttribute('aria-label'))
        expect(labels).toEqual(['Start rotating', 'Show clouds', 'Reset view', 'Zoom in', 'Zoom out', 'Refresh now',
            'Provider status', 'Settings', 'Help', donateLabel('')])
        for (const b of buttons) {
            expect(b.getAttribute('aria-label')).toBeTruthy()
            expect(b.getAttribute('data-tip')).toBe(b.getAttribute('aria-label'))
        }
    })

    it('zooms in and out from its buttons', () => {
        const onZoom = vi.fn()
        const noop = () => undefined
        render(<Rail autoRotate attention onToggleRotate={noop} cloudsShown={false} onToggleClouds={noop} onResetView={noop} onZoom={onZoom}
            onRefresh={noop} onStatus={noop} onSettings={noop} onHelp={noop} onDonate={noop}/>)
        fireEvent.click(screen.getByLabelText('Zoom in'))
        fireEvent.click(screen.getByLabelText('Zoom out'))
        expect(onZoom.mock.calls).toEqual([[true], [false]])
        expect(screen.getByLabelText('Provider status: something to report').className).toContain('attention')
    })
})

describe('FR-STS-003 the provider status popover', () => {
    it('gives a failed fetch its reason in words and its next attempt', () => {
        render(<StatusPanel providers={[provider({problem: 'the server answered 503', nextAttempt: 'next attempt in 4 min'})]}
            notice="" onClose={() => undefined}/>)
        expect(screen.getByText('The last attempt failed: the server answered 503')).toBeTruthy()
        expect(screen.getByText('Next attempt in 4 min.')).toBeTruthy()
    })

    it('states loading, stale and the standing notices (FR-STS-005, FR-SET-004)', () => {
        render(<StatusPanel providers={[provider({name: 'EONET', loading: true}), provider({stale: true})]}
            notice="No settings file was read" onClose={() => undefined}/>)
        expect(screen.getByText('Loading for the first time.')).toBeTruthy()
        expect(screen.getByText('Retrieved 3 min ago (stale).')).toBeTruthy()
        expect(screen.getByText('No settings file was read')).toBeTruthy()
    })

    it('asks for attention only when there is something to report', () => {
        expect(needsAttention([provider()], '')).toBe(false)
        expect(needsAttention([provider({stale: true})], '')).toBe(true)
        expect(needsAttention([provider({problem: 'x'})], '')).toBe(true)
        expect(needsAttention([provider()], 'notice')).toBe(true)
    })
})

describe('the help dialogs', () => {
    const bound = {
        About: () => Promise.resolve({name: 'Product', version: '1.2.3', copyright: '© Someone 2026', licence: 'a licence', attributions: ['credit one']}),
        Licence: () => Promise.resolve('the licence text'),
        Notices: () => Promise.resolve('the notices text'),
    }
    beforeEach(() => { (window as unknown as {go: unknown}).go = {main: {App: bound}} })
    afterEach(() => { delete (window as unknown as {go?: unknown}).go })

    it('FR-HLP-001 shows the name, version, licence and credits', async () => {
        render(<HelpDialog kind="about" onClose={() => undefined} onProblem={() => undefined}/>)
        expect(await screen.findByText('Version 1.2.3')).toBeTruthy()
        expect(screen.getByText('Product')).toBeTruthy()
        expect(screen.getByText('Licensed under the a licence.')).toBeTruthy()
        expect(screen.getByText('credit one')).toBeTruthy()
        expect(screen.getByText('© Someone 2026')).toBeTruthy()
    })

    it('FR-HLP-002 and 003 show the licence and the notices in full', async () => {
        const {unmount} = render(<HelpDialog kind="licence" onClose={() => undefined} onProblem={() => undefined}/>)
        expect(await screen.findByText('the licence text')).toBeTruthy()
        unmount()
        render(<HelpDialog kind="notices" onClose={() => undefined} onProblem={() => undefined}/>)
        expect(await screen.findByText('the notices text')).toBeTruthy()
    })

    it('FR-HLP-004 the guide names every category with what it covers', () => {
        render(<HelpDialog kind="guide" onClose={() => undefined} onProblem={() => undefined}/>)
        for (const c of CATEGORIES) expect(screen.getByText(c.name)).toBeTruthy()
        expect(Object.keys(COVERS).sort()).toEqual(CATEGORIES.map(c => c.key).sort())
    })

    it('says so when the backend refuses', async () => {
        const onProblem = vi.fn()
        delete (window as unknown as {go?: unknown}).go
        render(<HelpDialog kind="licence" onClose={() => undefined} onProblem={onProblem}/>)
        await act(async () => undefined)
        expect(onProblem).toHaveBeenCalled()
    })
})

describe('FR-HLP-006 a surface under a dialog', () => {
    it('reads only while it is the topmost surface', () => {
        render(<div>
            <div className="detail"><div data-testid="panel"/></div>
        </div>)
        const panel = screen.getByTestId('panel')
        expect(isTopmostSurface(panel)).toBe(true)
        render(<Dialog title="Settings" onClose={() => undefined}><div data-testid="inside"/></Dialog>)
        expect(isTopmostSurface(panel)).toBe(false)
        expect(isTopmostSurface(screen.getByTestId('inside'))).toBe(true)
    })
})

describe('FR-DON the donate button', () => {
    afterEach(() => { delete (window as unknown as {go?: unknown}).go })

    it('FR-DON-001 and FR-DON-005 is a rail button at its foot, named as PigeonPost names it', () => {
        const onDonate = vi.fn()
        const noop = () => undefined
        // The name arrives from the Go side (internal/product) through the context.
        render(<ProductName.Provider value="Product"><Rail autoRotate={false} attention={false}
            onToggleRotate={noop} cloudsShown={false} onToggleClouds={noop} onResetView={noop} onZoom={noop} onRefresh={noop} onStatus={noop}
            onSettings={noop} onHelp={noop} onDonate={onDonate}/></ProductName.Provider>)
        const button = screen.getByLabelText('Donate to support Product')
        expect(button.className).toBe('rail-btn rail-foot')
        expect(button.getAttribute('data-tip')).toBe('Donate to support Product')
        expect(button.closest('.rail-actions')).toBeNull()
        fireEvent.click(button)
        expect(onDonate).toHaveBeenCalledOnce()
    })

    it('FR-DON-003 asks the Go side to open the page, holding no address itself', async () => {
        const Donate = vi.fn(() => Promise.resolve())
        ;(window as unknown as {go: unknown}).go = {main: {App: {Donate}}}
        const onProblem = vi.fn()
        donate(onProblem)
        await act(async () => undefined)
        expect(Donate).toHaveBeenCalledOnce()
        expect(onProblem).not.toHaveBeenCalled()
    })

    it('FR-DON-006 says so when the page could not be opened', async () => {
        const Donate = vi.fn(() => Promise.reject(new Error('not an https link')))
        ;(window as unknown as {go: unknown}).go = {main: {App: {Donate}}}
        const onProblem = vi.fn()
        donate(onProblem)
        await act(async () => undefined)
        expect(onProblem).toHaveBeenCalledWith('The donation page could not be opened: not an https link')
    })
})

describe('FR-PRV-001 an ended event', () => {
    afterEach(() => { delete (window as unknown as {go?: unknown}).go })

    it('says in its detail panel that the source has ended it', async () => {
        ;(window as unknown as {go: unknown}).go = {main: {App: {Place: () => Promise.resolve('')}}}
        const e = {id: 'EONET:1', provider: 'EONET', category: 'WILDFIRE', title: 'Fire', description: '', lat: 1, lng: 2,
            at: '2026-09-20T00:00:00Z', dayOnly: true, reported: 'r', retrievedAt: '2026-09-23T00:00:00Z', retrieved: 'x',
            measurement: '', band: 0, sourceUrl: '', sourceText: '', ended: true}
        const {unmount} = render(<DetailPanel event={e} inView onClose={() => undefined} onProblem={() => undefined}/>)
        await act(async () => undefined)
        expect(screen.getByText('EONET has marked this event as ended.')).toBeTruthy()
        unmount()
        render(<DetailPanel event={{...e, ended: false}} inView onClose={() => undefined} onProblem={() => undefined}/>)
        await act(async () => undefined)
        expect(screen.queryByText('EONET has marked this event as ended.')).toBeNull()
    })
})

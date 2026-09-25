// Replay on the page (REQUIREMENTS.md 3.2.14): the controls, the hook that
// plays the span and asks for its frames and images, the status lines and the
// guide. Drawing is WebGL's and is checked by a person (FR-RPL-020 in TESTING.md).
import {act, fireEvent, render, renderHook, screen} from '@testing-library/react'
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {REPLAY_LABELS, ReplayControls} from './components/ReplayControls'
import {StatusLine} from './components/StatusLine'
import {TimeWindow} from './components/TimeWindow'
import {guideSections} from './guide'
import {icons} from './icons'
import type {ReplayFrameDTO, ReplaySpeedDTO} from './types'
import {useReplay} from './useReplay'

const noop = () => undefined
const NOW = Date.UTC(2026, 8, 24, 12)
// Steady values, as App passes: a fresh list each render would ask for a frame each render.
const NONE: string[] = []
const now = () => NOW
// The speeds as the Go side offers them (FR-RPL-025).
const HALF: ReplaySpeedDTO = {key: 'half', label: '0.5x', passSeconds: 60}
const NORMAL: ReplaySpeedDTO = {key: 'normal', label: '1x', passSeconds: 30}
const DOUBLE: ReplaySpeedDTO = {key: 'double', label: '2x', passSeconds: 15}
const MS_PER_SECOND = 1000
const PASS_MS = NORMAL.passSeconds * MS_PER_SECOND

function frame(patch: Partial<ReplayFrameDTO> = {}): ReplayFrameDTO {
    return {
        view: {windowKey: '7d', countLine: '1 event up to 21 Sep 00:00 UTC in the last 7 days', events: [], counts: {}, providers: [], notice: ''},
        at: '2026-09-21T00:00:00Z', line: 'Replay: 21 Sep 00:00 UTC', sun: {lat: 0, lng: 0, twilightDegrees: 6, cloudNightFloor: 0.25},
        cloudTime: '', cloudsLine: '', burntKey: '',
        cloudsProvider: {name: 'EUMETSAT (replay)', loading: false, refreshing: false, stale: false, retrieved: '', problem: '', nextAttempt: ''},
        ...patch,
    }
}

interface Handlers { onPlay: () => void, onPause: () => void, onSeek: (p: number) => void, onNow: () => void, onSpeed: () => void }

function controls(playing: boolean, handlers: Partial<Handlers> = {}, replaying = true, speed: ReplaySpeedDTO | null = NORMAL) {
    return render(<ReplayControls position={0.5} playing={playing} replaying={replaying} speed={speed} next={speed && DOUBLE}
        onPlay={handlers.onPlay ?? noop} onPause={handlers.onPause ?? noop} onSeek={handlers.onSeek ?? noop}
        onNow={handlers.onNow ?? noop} onSpeed={handlers.onSpeed ?? noop}/>)
}

describe('FR-RPL-008 the Play/Pause button', () => {
    it('offers to play with the play artwork while stopped, to pause while playing', () => {
        const {unmount} = controls(false)
        expect(screen.getByLabelText(REPLAY_LABELS.play).querySelector('img')?.getAttribute('src')).toBe(icons.play)
        unmount()
        controls(true)
        expect(screen.getByLabelText(REPLAY_LABELS.pause).querySelector('img')?.getAttribute('src')).toBe(icons.pause)
    })

    it('asks to play or pause on a press', () => {
        const onPlay = vi.fn()
        controls(false, {onPlay})
        fireEvent.click(screen.getByLabelText(REPLAY_LABELS.play))
        expect(onPlay).toHaveBeenCalledTimes(1)
    })
})

describe('NFR-KBD-009 the replay on the ring and the keys', () => {
    it('is four stops after the time window while replaying; Space on the scrubber plays or pauses; a step seeks (FR-RPL-007)', () => {
        const onPause = vi.fn()
        const onSeek = vi.fn()
        render(<div><TimeWindow windows={[{key: '24h', label: '24 h'}, {key: '7d', label: '7 days'}]} selected="7d" onChoose={noop}/>
            <ReplayControls position={0.5} playing replaying speed={NORMAL} next={DOUBLE} onPlay={noop} onPause={onPause}
                onSeek={onSeek} onNow={noop} onSpeed={noop}/></div>)
        const stops = Array.from(document.querySelectorAll('[data-stop]')).map(e => e.getAttribute('aria-label') ?? e.textContent)
        expect(stops).toEqual(['24 h', REPLAY_LABELS.pause, REPLAY_LABELS.position, REPLAY_LABELS.speed('1x', '2x'),
            REPLAY_LABELS.nowName])
        const scrubber = screen.getByLabelText(REPLAY_LABELS.position) as HTMLInputElement
        expect(scrubber.step).toBe('10')
        fireEvent.keyDown(scrubber, {key: ' '})
        expect(onPause).toHaveBeenCalledTimes(1)
        fireEvent.keyDown(scrubber, {key: 'a'})
        fireEvent.change(scrubber, {target: {value: '250'}})
        expect(onSeek).toHaveBeenCalledWith(0.25)
    })
})

describe('FR-RPL-024 the Now button', () => {
    it('is hidden outside a replay; while replaying it reads Now and returns to the present', () => {
        const onNow = vi.fn()
        const {unmount} = controls(false, {onNow}, false)
        expect(screen.queryByLabelText(REPLAY_LABELS.nowName)).toBeNull()
        unmount()
        controls(true, {onNow})
        const now = screen.getByLabelText(REPLAY_LABELS.nowName)
        expect(now.textContent).toBe(REPLAY_LABELS.now)
        fireEvent.click(now)
        expect(onNow).toHaveBeenCalledTimes(1)
    })
})

describe('FR-RPL-025 the speed button', () => {
    it('reads the chosen speed, names the next and asks for it on a press', () => {
        const onSpeed = vi.fn()
        controls(false, {onSpeed})
        const button = screen.getByLabelText(REPLAY_LABELS.speed('1x', '2x'))
        expect(button.textContent).toBe('1x')
        fireEvent.click(button)
        expect(onSpeed).toHaveBeenCalledTimes(1)
    })

    it('is not shown until the speeds are known', () => {
        controls(false, {}, true, null)
        expect(screen.queryByText('1x')).toBeNull()
    })
})

describe('FR-RPL-019 the replay lines', () => {
    it('shows the instant and the cloud images counted in', () => {
        render(<StatusLine countLine="" providers={[]} problem="" replayLines={['Replay: 21 Sep 00:00 UTC', 'Clouds 12 of 56', '']}/>)
        expect(screen.getByText('Replay: 21 Sep 00:00 UTC')).toBeTruthy()
        expect(screen.getByText('Clouds 12 of 56')).toBeTruthy()
    })
})

describe('FR-RPL-023 the guide', () => {
    it('says Replay shows what the sources hold now, with softer clouds', () => {
        const text = guideSections('').find(s => s.heading === 'Replay')?.paragraphs?.join(' ') ?? ''
        expect(text).toContain('what the sources hold now')
        expect(text).toContain('revised or withdrawn')
        expect(text).toContain('softer images')
    })

    it('FR-RPL-024 FR-RPL-025 says the replay holds until Now and names the speeds', () => {
        const text = guideSections('').find(s => s.heading === 'Replay')?.paragraphs?.join(' ') ?? ''
        expect(text).toContain('holds at the end of the window until Now returns to the present')
        expect(text).toContain('half, normal or double speed')
    })
})

describe('useReplay, the replay on the page', () => {
    let frames: Map<number, (t: number) => void>
    let handle: number
    let clock: number
    let App: Record<string, ReturnType<typeof vi.fn>>
    let handlers: Record<string, () => void>

    beforeEach(() => {
        frames = new Map()
        handle = 0
        clock = 0
        handlers = {}
        vi.stubGlobal('requestAnimationFrame', (cb: (t: number) => void) => { frames.set(++handle, cb); return handle })
        vi.stubGlobal('cancelAnimationFrame', (h: number) => frames.delete(h))
        App = {
            ReplayFrame: vi.fn(() => Promise.resolve(frame())),
            ReplayCloudImage: vi.fn(() => Promise.resolve('data:image/png;base64,CLOUD')),
            ReplayBurntImage: vi.fn(() => Promise.resolve('data:image/png;base64,BURNT')),
            EndReplay: vi.fn(() => Promise.resolve()),
        };
        (window as unknown as {go: unknown}).go = {main: {App}};
        (window as unknown as {runtime: unknown}).runtime = {EventsOn: (name: string, cb: () => void) => { handlers[name] = cb; return noop }}
    })
    afterEach(() => {
        vi.unstubAllGlobals()
        vi.restoreAllMocks()
        delete (window as unknown as {go?: unknown}).go
        delete (window as unknown as {runtime?: unknown}).runtime
    })

    // advance runs the animation frames queued, as the browser would at t ms.
    function advance(ms: number) {
        clock += ms
        const queued = [...frames.values()]
        frames.clear()
        act(() => queued.forEach(cb => cb(clock)))
    }
    const settle = () => act(async () => { await Promise.resolve(); await Promise.resolve() })
    const hook = (windowKey = '7d', passSeconds: number | null = NORMAL.passSeconds) =>
        renderHook(({w}) => useReplay(w, NONE, NONE, passSeconds, noop, now), {initialProps: {w: windowKey}})

    it('FR-RPL-002 rests at now, asking for no frame', async () => {
        const {result} = hook()
        await settle()
        expect(result.current.replaying).toBe(false)
        expect(result.current.position).toBe(1)
        expect(App.ReplayFrame).not.toHaveBeenCalled()
    })

    it('FR-RPL-003 FR-RPL-004 FR-RPL-005 plays the span from its start in 30 s, then holds at its end', async () => {
        const {result} = hook()
        act(() => result.current.play())
        await settle()
        expect(result.current.replaying).toBe(true)
        expect(App.ReplayFrame).toHaveBeenLastCalledWith('7d', [], [], NOW, 0)
        advance(0)
        advance(PASS_MS / 2)
        expect(result.current.position).toBeCloseTo(0.5)
        advance(PASS_MS / 2)
        await settle()
        expect(result.current.playing).toBe(false)
        expect(result.current.position).toBe(1)
        expect(result.current.replaying).toBe(true)
        expect(App.ReplayFrame).toHaveBeenLastCalledWith('7d', [], [], NOW, 1)
        expect(App.EndReplay).not.toHaveBeenCalled()
        // Play at the span's end starts the same span again.
        act(() => result.current.play())
        expect(result.current.position).toBe(0)
    })

    it('FR-RPL-024 Now stops play and returns to the ordinary view', async () => {
        const {result} = hook()
        act(() => result.current.play())
        advance(0)
        advance(PASS_MS / 4)
        act(() => result.current.toNow())
        await settle()
        expect(result.current.playing).toBe(false)
        expect(result.current.replaying).toBe(false)
        expect(result.current.position).toBe(1)
        expect(App.EndReplay).toHaveBeenCalledTimes(1)
    })

    it.each([[HALF, 0.25], [DOUBLE, 1]])('FR-RPL-025 at %o a quarter-minute of play reaches %s', async (speed, reached) => {
        const {result} = hook('7d', speed.passSeconds)
        act(() => result.current.play())
        advance(0)
        advance(PASS_MS / 2)
        expect(result.current.position).toBeCloseTo(reached)
    })

    it('FR-RPL-025 before the speeds are known play waits while a seek still shows its frame', async () => {
        const {result} = hook('7d', null)
        act(() => result.current.play())
        advance(0)
        advance(PASS_MS)
        expect(result.current.position).toBe(0)
        act(() => result.current.seek(0.25))
        await settle()
        expect(App.ReplayFrame).toHaveBeenLastCalledWith('7d', [], [], NOW, 0.25)
    })

    it('FR-RPL-006 FR-RPL-007 pause holds; a seek pauses and moves; the right end is the span\'s end', async () => {
        const {result} = hook()
        act(() => result.current.play())
        advance(0)
        advance(PASS_MS / 3)
        act(() => result.current.pause())
        advance(PASS_MS)
        expect(result.current.position).toBeCloseTo(1 / 3)
        act(() => result.current.seek(0.25))
        expect(result.current.position).toBe(0.25)
        act(() => result.current.play())
        expect(result.current.position).toBe(0.25)
        act(() => result.current.seek(1))
        expect(result.current.replaying).toBe(true)
        expect(result.current.position).toBe(1)
    })

    it('FR-RPL-002 a seek to the right end outside a replay stays in the ordinary view', async () => {
        const {result} = hook()
        act(() => result.current.seek(1))
        await settle()
        expect(result.current.replaying).toBe(false)
        expect(App.ReplayFrame).not.toHaveBeenCalled()
    })

    it('FR-RPL-003 a seek from the end fixes the span at that instant', async () => {
        const {result} = hook()
        act(() => result.current.seek(0.5))
        await settle()
        expect(App.ReplayFrame).toHaveBeenLastCalledWith('7d', [], [], NOW, 0.5)
    })

    it('FR-RPL-021 another window returns to now', async () => {
        const {result, rerender} = hook()
        act(() => result.current.seek(0.5))
        rerender({w: '24h'})
        await settle()
        expect(result.current.replaying).toBe(false)
    })

    it('FR-RPL-013 FR-RPL-014 asks for each image once its key names one; again on news', async () => {
        App.ReplayFrame.mockImplementation(() => Promise.resolve(frame({cloudTime: '2026-09-20T21:00:00Z', burntKey: 'k'})))
        const {result} = hook()
        act(() => result.current.seek(0.5))
        await settle()
        await settle()
        expect(result.current.cloudImage).toBe('data:image/png;base64,CLOUD')
        expect(result.current.burntImage).toBe('data:image/png;base64,BURNT')
        expect(App.ReplayCloudImage).toHaveBeenCalledWith('2026-09-20T21:00:00Z')
        const asked = App.ReplayFrame.mock.calls.length
        act(() => handlers['replay-changed']())
        await settle()
        expect(App.ReplayFrame.mock.calls.length).toBe(asked + 1)
        App.ReplayFrame.mockImplementation(() => Promise.resolve(frame()))
        act(() => result.current.seek(0.2))
        await settle()
        await settle()
        expect(result.current.cloudImage).toBe('')
        expect(result.current.burntImage).toBe('')
    })
})

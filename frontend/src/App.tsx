// The window: the action rail down the left, the globe filling the middle, the
// key down the right (FR-RAIL-001, FR-KEY-001), the time window above the globe,
// the status line beneath it and the detail panel over its right edge.
import {useCallback, useEffect, useMemo, useRef, useState} from 'react'
import {api, on} from './api'
import {DetailPanel} from './components/DetailPanel'
import {HelpDialog, type HelpKind} from './components/HelpDialogs'
import {GlobeView, type GlobeHandle} from './components/GlobeView'
import {Key} from './components/Key'
import {Rail} from './components/Rail'
import {SettingsDialog} from './components/SettingsDialog'
import {lastRefreshed, StatusLine} from './components/StatusLine'
import {StatusPanel, needsAttention} from './components/StatusPanel'
import {TimeWindow} from './components/TimeWindow'
import {ReplayControls} from './components/ReplayControls'
import {settleKeyboard} from '../../installer/frontend/dist/settle-keyboard.js'
import {donate} from './donate'
import {icons} from './icons'
import {ProductName} from './product'
import {useRing} from './ring'
import type {BurntAreasDTO, ChoiceDTO, CloudsDTO, EventDTO, SettingChoicesDTO, SettingsDTO, StartViewDTO, ViewDTO} from './types'
import {REFRESH_TURN_MS, useHeld} from './useHeld'
import {type LayerSource, useLayer} from './useLayer'
import {useReplay} from './useReplay'
import {useSun} from './useSun'

// NONE is the empty filter list, one value so the replay's requests stay steady.
const NONE: string[] = []
const EMPTY_VIEW: ViewDTO = {windowKey: '', countLine: '', events: [], counts: {}, providers: [], notice: ''}
// A refused start view opens the globe as before (FR-GLB-016).
const NO_START: StartViewDTO = {found: false, lat: 0, lng: 0}

// The image layers (FR-CLD-016, FR-BA-006). Each line's age is re-read
// whenever the events are, so it keeps up.
const CLOUD_LAYER: LayerSource<CloudsDTO> = {
    state: api.clouds, image: api.cloudImage, key: c => c.validTime, events: ['events-changed', 'clouds-changed'],
}
const BURNT_LAYER: LayerSource<BurntAreasDTO> = {
    state: api.burntAreas, image: api.burntImage, key: b => b.key, events: ['events-changed', 'burnt-changed'],
}

function toggled(list: string[], key: string): string[] {
    return list.includes(key) ? list.filter(k => k !== key) : [...list, key]
}

export default function App() {
    const [windows, setWindows] = useState<ChoiceDTO[]>([])
    const [settings, setSettings] = useState<SettingsDTO | null>(null)
    const [choices, setChoices] = useState<SettingChoicesDTO | null>(null)
    const [settingsOpen, setSettingsOpen] = useState(false)
    const [statusOpen, setStatusOpen] = useState(false)
    const [help, setHelp] = useState<HelpKind | null>(null)
    const [view, setView] = useState<ViewDTO>(EMPTY_VIEW)
    const [selected, setSelected] = useState<EventDTO | null>(null)
    const [problem, setProblem] = useState('')
    const [product, setProduct] = useState('')
    // Where the globe first faces (FR-GLB-015); the globe waits for the answer.
    const [start, setStart] = useState<StartViewDTO | null>(null)
    const globe = useRef<GlobeHandle>(null)
    // NFR-KBD-001: one ring over the window, inert while a modal owns the keys.
    const shell = useRef<HTMLDivElement>(null)
    const modalOpen = settingsOpen || statusOpen || help !== null
    useRing(shell, !modalOpen)
    const onProblem = useCallback((reason: string) => setProblem(reason), [])

    const windowKey = settings?.windowKey ?? ''
    const hiddenCategories = useMemo(() => new Set(settings?.hiddenCategories ?? []), [settings?.hiddenCategories])
    const hiddenProviders = useMemo(() => new Set(settings?.hiddenProviders ?? []), [settings?.hiddenProviders])

    const load = useCallback(() => {
        void api.view(windowKey, [...hiddenCategories], [...hiddenProviders], onProblem).then(v => { if (v) setView(v) })
    }, [windowKey, hiddenCategories, hiddenProviders, onProblem])

    // change applies a setting at once and saves it (FR-FLT-005, FR-GLB-011);
    // the answer is the settings as held, which the view then follows.
    const latest = useRef<SettingsDTO | null>(null)
    latest.current = settings
    const change = useCallback((patch: Partial<SettingsDTO>) => {
        if (!latest.current) return
        const next = {...latest.current, ...patch}
        latest.current = next
        setSettings(next)
        void api.saveSettings(next, onProblem).then(held => { if (held) setSettings(held) })
    }, [onProblem])

    // FR-PRV-010: every press says when the last manual refresh was made, so a
    // press the cooldown refused still shows the refresh that stands.
    const [lastRefresh, setLastRefresh] = useState<number | null>(null)
    const refresh = () => {
        void api.refreshNow(onProblem).then(at => { if (at !== null) setLastRefresh(at) })
    }

    // NFR-KBD-002: the window starts neutral while still holding the keyboard, so the
    // first Tab reaches the ring. A launch that lost the race inside Wails asks for
    // it; nothing on the page is focused either way (the setup page's repair).
    useEffect(() => settleKeyboard(() => api.takeKeyboard(onProblem)), [onProblem])

    useEffect(() => {
        void api.windows(onProblem).then(w => { if (w) setWindows(w) })
        void api.settings(onProblem).then(s => { if (s) setSettings(s) })
        void api.settingChoices(onProblem).then(c => { if (c) setChoices(c) })
        void api.startView(onProblem).then(s => setStart(s ?? NO_START))
        // The product's name has its home on the Go side; About carries it.
        void api.about(onProblem).then(a => { if (a) setProduct(a.name) })
    }, [onProblem])
    useEffect(() => { document.title = product }, [product])
    const ready = settings !== null
    const [clouds, cloudImage] = useLayer(CLOUD_LAYER, ready, onProblem)
    const [burnt, burntImage] = useLayer(BURNT_LAYER, ready, onProblem)
    useEffect(() => {
        if (!ready) return
        load()
        const offChanged = on('events-changed', load)
        const offProblem = on('problem', (reason) => onProblem(String(reason)))
        return () => { offChanged(); offProblem() }
    }, [load, onProblem, ready])

    const dayNightShown = settings?.dayNightShown ?? false
    const liveSun = useSun(dayNightShown, onProblem)
    // Replay (3.2.14): while the scrubber is off its end, the frame stands in for
    // the live view, sun and images.
    const replay = useReplay(windowKey, settings?.hiddenCategories ?? NONE, settings?.hiddenProviders ?? NONE, onProblem)
    const frame = replay.replaying ? replay.frame : null
    const shown = frame?.view ?? view
    const sun = frame?.sun ?? liveSun
    // FR-SEL-008: a selected event that leaves the view keeps its panel, marked.
    const current = selected ? shown.events.find(e => e.id === selected.id) : undefined
    const shownDetail = current ?? selected
    const speed = choices?.speeds.find(s => s.key === settings?.speed)
    const refreshing = useHeld(view.providers.some(p => p.refreshing), REFRESH_TURN_MS)
    // FR-CLD-011, FR-BA-012: each image layer's service joins the popover while
    // the layer is shown.
    const replayClouds = frame && clouds?.shown ? [frame.cloudsProvider] : []
    const statusProviders = [...view.providers, ...[clouds, burnt].filter(l => l?.shown).map(l => l!.provider), ...replayClouds]

    return <ProductName.Provider value={product}><div ref={shell} className="app">
        <Rail autoRotate={settings?.autoRotate ?? false}
            onToggleRotate={() => change({autoRotate: !settings?.autoRotate})}
            cloudsShown={settings?.cloudsShown ?? false}
            onToggleClouds={() => change({cloudsShown: !settings?.cloudsShown})}
            dayNightShown={dayNightShown}
            onToggleDayNight={() => change({dayNightShown: !dayNightShown})}
            attention={needsAttention(statusProviders, view.notice)}
            onResetView={() => globe.current?.resetView()}
            onZoom={zoomIn => globe.current?.zoom(zoomIn)}
            onRefresh={refresh}
            refreshing={refreshing}
            onStatus={() => setStatusOpen(true)}
            onSettings={() => setSettingsOpen(true)}
            onHelp={setHelp}
            onDonate={() => donate(onProblem)}/>
        <main className="stage">
            {/* The time window comes before the globe in the document so the ring
                meets them in reading order, top to bottom (keeb invariant 1). */}
            <div className="top-bar">
                <TimeWindow windows={windows} selected={windowKey} onChoose={k => change({windowKey: k})}/>
                <ReplayControls position={replay.position} playing={replay.playing}
                    onPlay={replay.play} onPause={replay.pause} onSeek={replay.seek}/>
            </div>
            {speed && settings && start && <GlobeView ref={globe} start={start} events={shown.events} selectedId={selected?.id ?? null}
                autoRotate={settings.autoRotate} secondsPerRevolution={speed.secondsPerRevolution}
                cloudImage={settings.cloudsShown ? (frame ? replay.cloudImage : cloudImage) : ''}
                burntImage={settings.burntShown ? (frame ? replay.burntImage : burntImage) : ''}
                dayNightShown={dayNightShown} sun={sun} trailsShown={settings.trailsShown}
                onSelect={setSelected} onProblem={onProblem}/>}
            <StatusLine countLine={shown.countLine} providers={view.providers} problem={problem}
                note={lastRefresh === null ? '' : lastRefreshed(lastRefresh)} clouds={clouds} burnt={burnt}
                replayLines={frame ? [frame.line, clouds?.shown ? frame.cloudsLine : ''] : []}/>
            {shownDetail && <DetailPanel event={shownDetail} inView={current !== undefined} onClose={() => setSelected(null)} onProblem={onProblem}/>}
        </main>
        <aside className="side">
            {/* The mark IS the heading: its artwork carries the product's name
                (owner), so its alternative text is that name. */}
            <h1><img className="brand-mark" src={icons.appMark} alt={product} draggable={false}/></h1>
            <Key counts={shown.counts} providers={view.providers}
                hiddenCategories={hiddenCategories} hiddenProviders={hiddenProviders}
                onToggleCategory={k => change({hiddenCategories: toggled(settings?.hiddenCategories ?? [], k)})}
                onToggleProvider={n => change({hiddenProviders: toggled(settings?.hiddenProviders ?? [], n)})}
                onShowAll={() => change({hiddenCategories: [], hiddenProviders: []})}/>
        </aside>
        {settingsOpen && settings && choices && <SettingsDialog settings={settings} choices={choices}
            onChange={change} onClose={() => setSettingsOpen(false)}/>}
        {statusOpen && <StatusPanel providers={statusProviders} notice={view.notice} onClose={() => setStatusOpen(false)}/>}
        {help && <HelpDialog kind={help} onClose={() => setHelp(null)} onProblem={onProblem}/>}
    </div></ProductName.Provider>
}

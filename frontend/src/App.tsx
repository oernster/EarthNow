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
import {settleKeyboard} from '../../installer/frontend/dist/settle-keyboard.js'
import {donate} from './donate'
import {icons} from './icons'
import {ProductName} from './product'
import {useRing} from './ring'
import type {ChoiceDTO, CloudsDTO, EventDTO, SettingChoicesDTO, SettingsDTO, ViewDTO} from './types'
import {REFRESH_TURN_MS, useHeld} from './useHeld'
import {useSun} from './useSun'

const EMPTY_VIEW: ViewDTO = {windowKey: '', countLine: '', events: [], counts: {}, providers: [], notice: ''}

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
    const [clouds, setClouds] = useState<CloudsDTO | null>(null)
    const [cloudImage, setCloudImage] = useState('')
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

    // loadClouds reads the cloud layer's state and asks for the image only when
    // the held one has been replaced, since it is large (FR-CLD-016).
    const cloudTime = useRef('')
    const loadClouds = useCallback(() => {
        void api.clouds(onProblem).then(c => {
            if (!c) return
            setClouds(c)
            if (c.validTime === cloudTime.current) return
            cloudTime.current = c.validTime
            void api.cloudImage(onProblem).then(url => { if (url !== null) setCloudImage(url) })
        })
    }, [onProblem])

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
        // The product's name has its home on the Go side; About carries it.
        void api.about(onProblem).then(a => { if (a) setProduct(a.name) })
    }, [onProblem])
    useEffect(() => { document.title = product }, [product])
    const ready = settings !== null
    useEffect(() => {
        if (!ready) return
        load()
        loadClouds()
        const offChanged = on('events-changed', load)
        // The cloud line's age is re-read whenever the events are, so it keeps up.
        const offAged = on('events-changed', loadClouds)
        const offClouds = on('clouds-changed', loadClouds)
        const offProblem = on('problem', (reason) => onProblem(String(reason)))
        return () => { offChanged(); offAged(); offClouds(); offProblem() }
    }, [load, loadClouds, onProblem, ready])

    // FR-SEL-008: a selected event that leaves the view keeps its panel, marked.
    const current = selected ? view.events.find(e => e.id === selected.id) : undefined
    const shownDetail = current ?? selected
    const speed = choices?.speeds.find(s => s.key === settings?.speed)
    const refreshing = useHeld(view.providers.some(p => p.refreshing), REFRESH_TURN_MS)
    // FR-CLD-011: the cloud service joins the popover while the layer is shown.
    const statusProviders = clouds?.shown ? [...view.providers, clouds.provider] : view.providers
    const dayNightShown = settings?.dayNightShown ?? false
    const sun = useSun(dayNightShown, onProblem)

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
            </div>
            {speed && settings && <GlobeView ref={globe} events={view.events} selectedId={selected?.id ?? null}
                autoRotate={settings.autoRotate} secondsPerRevolution={speed.secondsPerRevolution}
                cloudImage={settings.cloudsShown ? cloudImage : ''}
                dayNightShown={dayNightShown} sun={sun}
                onSelect={setSelected} onProblem={onProblem}/>}
            <StatusLine countLine={view.countLine} providers={view.providers} problem={problem}
                note={lastRefresh === null ? '' : lastRefreshed(lastRefresh)} clouds={clouds}/>
            {shownDetail && <DetailPanel event={shownDetail} inView={current !== undefined} onClose={() => setSelected(null)} onProblem={onProblem}/>}
        </main>
        <aside className="side">
            {/* The mark IS the heading: its artwork carries the product's name
                (owner), so its alternative text is that name. */}
            <h1><img className="brand-mark" src={icons.appMark} alt={product} draggable={false}/></h1>
            <Key counts={view.counts} providers={view.providers}
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

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
import {StatusLine} from './components/StatusLine'
import {StatusPanel, needsAttention} from './components/StatusPanel'
import {TimeWindow} from './components/TimeWindow'
import {settleKeyboard} from '../../installer/frontend/dist/settle-keyboard.js'
import {donate} from './donate'
import {icons} from './icons'
import {useRing} from './ring'
import type {ChoiceDTO, EventDTO, SettingChoicesDTO, SettingsDTO, ViewDTO} from './types'

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

    const refresh = () => {
        void api.refreshNow(onProblem).then(wait => { if (wait !== null) setProblem(wait) })
    }

    // NFR-KBD-002: the window starts neutral while still holding the keyboard, so the
    // first Tab reaches the ring. A launch that lost the race inside Wails asks for
    // it; nothing on the page is focused either way (the setup page's repair).
    useEffect(() => settleKeyboard(() => api.takeKeyboard(onProblem)), [onProblem])

    useEffect(() => {
        void api.windows(onProblem).then(w => { if (w) setWindows(w) })
        void api.settings(onProblem).then(s => { if (s) setSettings(s) })
        void api.settingChoices(onProblem).then(c => { if (c) setChoices(c) })
    }, [onProblem])
    const ready = settings !== null
    useEffect(() => {
        if (!ready) return
        load()
        const offChanged = on('events-changed', load)
        const offProblem = on('problem', (reason) => onProblem(String(reason)))
        return () => { offChanged(); offProblem() }
    }, [load, onProblem, ready])

    // FR-SEL-008: a selected event that leaves the view keeps its panel, marked.
    const current = selected ? view.events.find(e => e.id === selected.id) : undefined
    const shownDetail = current ?? selected
    const speed = choices?.speeds.find(s => s.key === settings?.speed)

    return <div ref={shell} className="app">
        <Rail autoRotate={settings?.autoRotate ?? false}
            onToggleRotate={() => change({autoRotate: !settings?.autoRotate})}
            attention={needsAttention(view.providers, view.notice)}
            onResetView={() => globe.current?.resetView()}
            onZoom={zoomIn => globe.current?.zoom(zoomIn)}
            onRefresh={refresh}
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
                onSelect={setSelected} onProblem={onProblem}/>}
            <StatusLine countLine={view.countLine} providers={view.providers} problem={problem}/>
            {shownDetail && <DetailPanel event={shownDetail} inView={current !== undefined} onClose={() => setSelected(null)} onProblem={onProblem}/>}
        </main>
        <aside className="side">
            {/* The mark IS the heading: its artwork carries the product's name
                (owner), so its alternative text is that name. */}
            <h1><img className="brand-mark" src={icons.appMark} alt="EarthNow" draggable={false}/></h1>
            <Key counts={view.counts} providers={view.providers}
                hiddenCategories={hiddenCategories} hiddenProviders={hiddenProviders}
                onToggleCategory={k => change({hiddenCategories: toggled(settings?.hiddenCategories ?? [], k)})}
                onToggleProvider={n => change({hiddenProviders: toggled(settings?.hiddenProviders ?? [], n)})}
                onShowAll={() => change({hiddenCategories: [], hiddenProviders: []})}/>
        </aside>
        {settingsOpen && settings && choices && <SettingsDialog settings={settings} choices={choices}
            onChange={change} onClose={() => setSettingsOpen(false)}/>}
        {statusOpen && <StatusPanel providers={view.providers} notice={view.notice} onClose={() => setStatusOpen(false)}/>}
        {help && <HelpDialog kind={help} onClose={() => setHelp(null)} onProblem={onProblem}/>}
    </div>
}

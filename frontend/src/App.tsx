// The window: key down the left, the globe filling the rest, the time window
// above it, the status line beneath it and the detail panel over its right edge.
import {useCallback, useEffect, useState} from 'react'
import {api, on} from './api'
import {DetailPanel} from './components/DetailPanel'
import {GlobeView} from './components/GlobeView'
import {Key} from './components/Key'
import {StatusLine} from './components/StatusLine'
import {TimeWindow} from './components/TimeWindow'
import type {EventDTO, ViewDTO, WindowDTO} from './types'

const DEFAULT_WINDOW = '24h'
const EMPTY_VIEW: ViewDTO = {windowKey: DEFAULT_WINDOW, countLine: '', events: [], counts: {}, providers: [], notice: ''}

export default function App() {
    const [windows, setWindows] = useState<WindowDTO[]>([])
    const [windowKey, setWindowKey] = useState(DEFAULT_WINDOW)
    const [view, setView] = useState<ViewDTO>(EMPTY_VIEW)
    const [selected, setSelected] = useState<EventDTO | null>(null)
    const [problem, setProblem] = useState('')
    const [hiddenCategories, setHiddenCategories] = useState<ReadonlySet<string>>(new Set())
    const [hiddenProviders, setHiddenProviders] = useState<ReadonlySet<string>>(new Set())
    const onProblem = useCallback((reason: string) => setProblem(reason), [])

    const load = useCallback(() => {
        void api.view(windowKey, [...hiddenCategories], [...hiddenProviders], onProblem).then(v => { if (v) setView(v) })
    }, [windowKey, hiddenCategories, hiddenProviders, onProblem])

    const toggle = (set: ReadonlySet<string>, key: string) => {
        const next = new Set(set)
        if (!next.delete(key)) next.add(key)
        return next
    }
    const refresh = () => {
        void api.refreshNow(onProblem).then(wait => { if (wait !== null) setProblem(wait) })
    }

    useEffect(() => { void api.windows(onProblem).then(w => { if (w) setWindows(w) }) }, [onProblem])
    useEffect(() => {
        load()
        const offChanged = on('events-changed', load)
        const offProblem = on('problem', (reason) => onProblem(String(reason)))
        return () => { offChanged(); offProblem() }
    }, [load, onProblem])

    // FR-SEL-008: a selected event that leaves the view keeps its panel, marked.
    const current = selected ? view.events.find(e => e.id === selected.id) : undefined
    const shownDetail = current ?? selected

    return <div className="app">
        <aside className="side">
            <h1>EarthNow</h1>
            <Key counts={view.counts} providers={view.providers}
                hiddenCategories={hiddenCategories} hiddenProviders={hiddenProviders}
                onToggleCategory={k => setHiddenCategories(s => toggle(s, k))}
                onToggleProvider={n => setHiddenProviders(s => toggle(s, n))}
                onShowAll={() => { setHiddenCategories(new Set()); setHiddenProviders(new Set()) }}/>
        </aside>
        <main className="stage">
            <GlobeView events={view.events} selectedId={selected?.id ?? null} onSelect={setSelected} onProblem={onProblem}/>
            <div className="top-bar">
                <TimeWindow windows={windows} selected={windowKey} onChoose={setWindowKey}/>
                <button className="refresh" onClick={refresh} title="Fetch the latest data from every source now">Refresh</button>
            </div>
            <StatusLine countLine={view.countLine} providers={view.providers} problem={problem || view.notice}/>
            {shownDetail && <DetailPanel event={shownDetail} inView={current !== undefined} onClose={() => setSelected(null)} onProblem={onProblem}/>}
        </main>
    </div>
}

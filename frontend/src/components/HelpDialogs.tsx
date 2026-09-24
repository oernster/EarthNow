// The four dialogs behind the Help button (FR-HLP-001 to 005): About, the guide,
// the licence and the third-party notices. Each is the shared Dialog in its
// reading form, so each reads itself with Close pinned beneath (PigeonPost's
// AboutModal, GuideModal and LicenceModal). The guide's words ship with the page;
// the rest are asked of the Go side, which holds the one copy of each.
import {useEffect, useState} from 'react'
import {api} from '../api'
import {guideSections} from '../guide'
import {useProductName} from '../product'
import type {AboutDTO} from '../types'
import {Dialog} from './Dialog'

export type HelpKind = 'guide' | 'about' | 'licence' | 'notices'

// HELP_ENTRIES is the Help menu, in the order it is drawn.
export const HELP_ENTRIES: readonly {kind: HelpKind; label: string}[] = [
    {kind: 'guide', label: 'Guide'},
    {kind: 'about', label: 'About'},
    {kind: 'licence', label: 'Licence'},
    {kind: 'notices', label: 'Third-party notices'},
]

interface Props {
    kind: HelpKind
    onClose: () => void
    onProblem: (reason: string) => void
}

export function HelpDialog({kind, onClose, onProblem}: Props) {
    const title = HELP_ENTRIES.find(e => e.kind === kind)?.label ?? ''
    return <Dialog title={title} onClose={onClose} reading>
        {kind === 'guide' && <Guide/>}
        {kind === 'about' && <About onProblem={onProblem}/>}
        {kind === 'licence' && <Text load={api.licence} onProblem={onProblem}/>}
        {kind === 'notices' && <Text load={api.notices} onProblem={onProblem}/>}
    </Dialog>
}

function Guide() {
    const product = useProductName()
    return <>
        {guideSections(product).map(s => <section key={s.heading}>
            <h3>{s.heading}</h3>
            {s.intro && <p>{s.intro}</p>}
            {s.entries?.map(e => <p key={e.name} className="guide-entry">
                {e.icon
                    ? <img className="guide-icon" src={e.icon} alt="" draggable={false}/>
                    : <span className="guide-icon guide-emoji" aria-hidden="true">{e.emoji}</span>}
                <span><b>{e.name}</b>: {e.text}</span>
            </p>)}
            {s.paragraphs?.map(p => <p key={p}>{p}</p>)}
        </section>)}
    </>
}

function About({onProblem}: {onProblem: (reason: string) => void}) {
    const [about, setAbout] = useState<AboutDTO | null>(null)
    useEffect(() => { void api.about(onProblem).then(setAbout) }, [onProblem])
    if (!about) return null
    return <>
        <p className="about-name">{about.name}</p>
        <p>Version {about.version}</p>
        <p>{about.copyright}</p>
        <p>Licensed under the {about.licence}.</p>
        <h3>Data and imagery</h3>
        {about.attributions.map(a => <p key={a}>{a}</p>)}
    </>
}

function Text({load, onProblem}: {load: (onRefused: (reason: string) => void) => Promise<string | null>; onProblem: (reason: string) => void}) {
    const [text, setText] = useState<string | null>(null)
    useEffect(() => { void load(onProblem).then(setText) }, [load, onProblem])
    return text === null ? null : <pre className="reading-text">{text}</pre>
}

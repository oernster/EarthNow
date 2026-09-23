// The detail panel (FR-SEL-002 to 008, FR-GEO-007): concise source facts with
// every time given as freshness wording, exact UTC and local time.
import {useEffect, useState} from 'react'
import {api} from '../api'
import {categoryOf} from '../categories'
import type {EventDTO} from '../types'

interface Props {
    event: EventDTO
    inView: boolean
    onClose: () => void
    onProblem: (reason: string) => void
}

const COORD_DECIMALS = 4

function utc(iso: string, dayOnly: boolean): string {
    const [date, rest] = iso.split('T')
    return dayOnly ? `${date} (UTC date)` : `${date} ${rest.replace('Z', '')} UTC`
}

function local(iso: string, dayOnly: boolean): string {
    const d = new Date(iso)
    return dayOnly ? '' : d.toLocaleString()
}

export function DetailPanel({event, inView, onClose, onProblem}: Props) {
    const [place, setPlace] = useState('')
    const cat = categoryOf(event.category)

    useEffect(() => {
        setPlace('')
        void api.place(event.lat, event.lng, onProblem).then(p => { if (p !== null) setPlace(p) })
    }, [event.id, event.lat, event.lng, onProblem])

    useEffect(() => {
        const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
        window.addEventListener('keydown', onKey)
        return () => window.removeEventListener('keydown', onKey)
    }, [onClose])

    const eventLocal = local(event.at, event.dayOnly)
    return <aside className="detail" aria-label="Event details">
        <h2>{cat.emoji} {event.title}</h2>
        {!inView && <p className="detail-note">This event is no longer in the current view.</p>}
        <dl>
            <dt>Category</dt><dd>{cat.name}</dd>
            {place && <><dt>Where</dt><dd>{place}</dd></>}
            <dt>Position</dt><dd>{event.lat.toFixed(COORD_DECIMALS)}, {event.lng.toFixed(COORD_DECIMALS)}</dd>
            <dt>Event time</dt><dd>{event.reported}<br/>{utc(event.at, event.dayOnly)}{eventLocal && <><br/>{eventLocal} local</>}</dd>
            {event.measurement && <><dt>Measurement</dt><dd>{event.measurement}</dd></>}
            {event.description && <><dt>Source text</dt><dd>{event.description}</dd></>}
            <dt>Provider</dt><dd>{event.provider}</dd>
            <dt>Retrieved</dt><dd>{event.retrieved}<br/>{utc(event.retrievedAt, false)}</dd>
            {event.sourceUrl && <><dt>Source</dt><dd>
                <button className="link" onClick={() => void api.openSource(event.sourceUrl, onProblem)}>Open the source page</button>
            </dd></>}
            {!event.sourceUrl && event.sourceText && <><dt>Source</dt><dd className="plain">{event.sourceText}</dd></>}
        </dl>
        <div className="detail-actions"><button onClick={onClose}>Close</button></div>
    </aside>
}

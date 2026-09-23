// The detail panel (FR-SEL-002 to 008, FR-GEO-007): concise source facts with
// every time given as freshness wording, exact UTC and local time.
import {useEffect, useRef, useState} from 'react'
import {api} from '../api'
import {categoryOf} from '../categories'
import {noClickFocus, useFirstStop} from '../ring'
import {useAutoScroll} from '../useAutoScroll'
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
    const panel = useRef<HTMLElement>(null)
    // NFR-KBD-006: opens on its first stop for each event shown; closing hands
    // focus back to the opener (FR-SEL-007).
    useFirstStop(panel, event.id)
    // FR-HLP-006: a long panel reads itself, Close pinned beneath it. It holds
    // still while a dialog is open over it (useAutoScroll).
    const autoScroll = useAutoScroll()

    useEffect(() => {
        setPlace('')
        void api.place(event.lat, event.lng, onProblem).then(p => { if (p !== null) setPlace(p) })
    }, [event.id, event.lat, event.lng, onProblem])

    useEffect(() => {
        // A dialog over the panel takes Escape first; one press closes one level.
        const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape' && !document.querySelector('.scrim')) onClose() }
        window.addEventListener('keydown', onKey)
        return () => window.removeEventListener('keydown', onKey)
    }, [onClose])

    const eventLocal = local(event.at, event.dayOnly)
    return <aside ref={panel} className="detail" aria-label="Event details">
        <h2>{cat.emoji} {event.title}</h2>
        <div key={event.id} ref={autoScroll} className="reading-body" data-stop data-reading tabIndex={-1} onMouseDown={noClickFocus}>
            {!inView && <p className="detail-note">This event is no longer in the current view.</p>}
            {event.ended && <p className="detail-note">{event.provider} has marked this event as ended.</p>}
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
                    <button data-stop className="link" onClick={() => void api.openSource(event.sourceUrl, onProblem)}>Open the source page</button>
                </dd></>}
                {!event.sourceUrl && event.sourceText && <><dt>Source</dt><dd className="plain">{event.sourceText}</dd></>}
            </dl>
        </div>
        <div className="detail-actions"><button data-stop onClick={onClose}>Close</button></div>
    </aside>
}

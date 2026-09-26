// The provider status popover (FR-STS-003, FR-STS-005, FR-SET-004), opened from
// the rail's status button. Each provider's state in full, a failed fetch's
// reason in words with the time of the next attempt and the standing notices
// (a settings file or cache that could not be read). The status line beneath
// the globe keeps the short form.
import type {ProviderDTO} from '../types'
import {Dialog} from './Dialog'

interface Props {
    providers: ProviderDTO[]
    notice: string
    onClose: () => void
}

/** needsAttention says whether the status has anything to report beyond the ordinary. */
export function needsAttention(providers: ProviderDTO[], notice: string): boolean {
    return notice !== '' || providers.some(p => p.stale || p.problem !== '' || p.notice !== '')
}

function state(p: ProviderDTO): string {
    if (p.loading) return 'Loading for the first time.'
    if (!p.retrieved) return 'Not retrieved yet.'
    return p.stale ? `${p.retrieved} (stale).` : `${p.retrieved}.`
}

export function StatusPanel({providers, notice, onClose}: Props) {
    return <Dialog title="Provider status" onClose={onClose}>
        {providers.map(p => <section key={p.name} className="status-entry">
            <h3>{p.name}</h3>
            <p>{state(p)}</p>
            {p.notice && <p className="warn">{p.notice}.</p>}
            {p.problem && <p className="warn">The last attempt failed: {p.problem}</p>}
            {p.problem && p.nextAttempt && <p>{p.nextAttempt.charAt(0).toUpperCase() + p.nextAttempt.slice(1)}.</p>}
        </section>)}
        {notice && <section className="status-entry">
            <h3>Notices</h3>
            <p className="warn">{notice}</p>
        </section>}
    </Dialog>
}

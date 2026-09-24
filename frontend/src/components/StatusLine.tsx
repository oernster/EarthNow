// The status area (FR-CNT-001, FR-STS-001 to 003, FR-CLD-009): the count line,
// each provider's freshness and the cloud image's time, never the word "live".
import type {CloudsDTO, ProviderDTO} from '../types'

interface Props {
    countLine: string
    providers: ProviderDTO[]
    problem: string
    // note is ordinary news, not a fault: when the last manual refresh was made (FR-PRV-010).
    note?: string
    // clouds is the cloud layer's state; its line shows only while the layer does.
    clouds?: CloudsDTO | null
}

/** lastRefreshed words when the last manual refresh was made, in local time (FR-PRV-010). */
export function lastRefreshed(at: number): string {
    return `Last refreshed at ${new Date(at).toLocaleTimeString()}`
}

function providerLine(p: ProviderDTO): string {
    if (p.loading) return `${p.name}: loading`
    if (p.refreshing) return `${p.name}: refreshing`
    const state = p.retrieved ? `${p.name}: ${p.retrieved.toLowerCase()}${p.stale ? ' (stale)' : ''}` : p.name
    return p.problem ? `${state}; last attempt failed, ${p.nextAttempt}` : state
}

export function StatusLine({countLine, providers, problem, note = '', clouds = null}: Props) {
    const cloud = clouds?.shown ? clouds.provider : null
    return <div className="status" role="status">
        <span className="count">{countLine}</span>
        {providers.map(p => <span key={p.name} className={p.problem || p.stale ? 'provider warn' : 'provider'} title={p.problem}>
            {providerLine(p)}
        </span>)}
        {clouds && cloud && <span className={cloud.problem || cloud.stale ? 'provider warn' : 'provider'} title={cloud.problem}>
            {clouds.line}
        </span>}
        {note && <span className="provider">{note}</span>}
        {problem && <span className="provider warn">{problem}</span>}
    </div>
}

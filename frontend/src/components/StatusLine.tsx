// The status area (FR-CNT-001, FR-STS-001 to 003): the count line plus each
// provider's freshness, never the word "live".
import type {ProviderDTO} from '../types'

interface Props {
    countLine: string
    providers: ProviderDTO[]
    problem: string
}

function providerLine(p: ProviderDTO): string {
    if (p.loading) return `${p.name}: loading`
    const state = p.retrieved ? `${p.name}: ${p.retrieved.toLowerCase()}${p.stale ? ' (stale)' : ''}` : p.name
    return p.problem ? `${state}; last attempt failed, ${p.nextAttempt}` : state
}

export function StatusLine({countLine, providers, problem}: Props) {
    return <div className="status" role="status">
        <span className="count">{countLine}</span>
        {providers.map(p => <span key={p.name} className={p.problem || p.stale ? 'provider warn' : 'provider'} title={p.problem}>
            {providerLine(p)}
        </span>)}
        {problem && <span className="provider warn">{problem}</span>}
    </div>
}

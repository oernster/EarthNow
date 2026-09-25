// The status area (FR-CNT-001, FR-STS-001 to 003, FR-CLD-009, FR-BA-008): the
// count line, each provider's freshness, the cloud image's time and the burnt
// areas' days, never the word "live".
import type {ProviderDTO} from '../types'

// A layer's state as the status line reads it: the cloud and burnt-area DTOs
// both carry these.
interface LayerState {
    shown: boolean
    line: string
    provider: ProviderDTO
}

interface Props {
    countLine: string
    providers: ProviderDTO[]
    problem: string
    // note is ordinary news, not a fault: when the last manual refresh was made (FR-PRV-010).
    note?: string
    // clouds and burnt are the image layers' states; each line shows only while
    // its layer does.
    clouds?: LayerState | null
    burnt?: LayerState | null
    // replayLines are the replay's instant and its cloud images' progress
    // (FR-RPL-016, FR-RPL-019); empty while not replaying.
    replayLines?: string[]
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

function layerLine(layer: LayerState | null) {
    if (!layer?.shown) return null
    const p = layer.provider
    return <span className={p.problem || p.stale ? 'provider warn' : 'provider'} title={p.problem}>{layer.line}</span>
}

export function StatusLine({countLine, providers, problem, note = '', clouds = null, burnt = null, replayLines = []}: Props) {
    return <div className="status" role="status">
        <span className="count">{countLine}</span>
        {providers.map(p => <span key={p.name} className={p.problem || p.stale ? 'provider warn' : 'provider'} title={p.problem}>
            {providerLine(p)}
        </span>)}
        {replayLines.filter(l => l !== '').map(l => <span key={l} className="provider">{l}</span>)}
        {layerLine(clouds)}
        {layerLine(burnt)}
        {note && <span className="provider">{note}</span>}
        {problem && <span className="provider warn">{problem}</span>}
    </div>
}

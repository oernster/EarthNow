// The key down the left side (FR-KEY-001 to 004), which is also the filter
// (FR-FLT-001 to 004): each category row and each provider toggles what the
// globe shows; "All events" puts everything back.
import {CATEGORIES} from '../categories'
import type {ProviderDTO} from '../types'

interface Props {
    counts: Record<string, number>
    providers: ProviderDTO[]
    hiddenCategories: ReadonlySet<string>
    hiddenProviders: ReadonlySet<string>
    onToggleCategory: (key: string) => void
    onToggleProvider: (name: string) => void
    onShowAll: () => void
}

export function Key(p: Props) {
    const filtered = p.hiddenCategories.size > 0 || p.hiddenProviders.size > 0
    return <nav className="key" aria-label="Key and filters">
        <ul>
            {CATEGORIES.map(c => {
                const hidden = p.hiddenCategories.has(c.key)
                return <li key={c.key}>
                    <button className={hidden ? 'key-row hidden' : 'key-row'} aria-pressed={!hidden}
                        title={hidden ? `Show ${c.name.toLowerCase()} events` : `Hide ${c.name.toLowerCase()} events`}
                        onClick={() => p.onToggleCategory(c.key)}>
                        <span className="key-emoji" aria-hidden="true">{c.emoji}</span>
                        <span className="key-name">{c.name}</span>
                        <span className="key-count">{p.counts[c.key] ?? 0}</span>
                    </button>
                </li>
            })}
        </ul>
        <div className="key-sources">
            {p.providers.map(s => {
                const hidden = p.hiddenProviders.has(s.name)
                return <button key={s.name} className={hidden ? 'source hidden' : 'source'} aria-pressed={!hidden}
                    title={hidden ? `Show ${s.name} events` : `Hide ${s.name} events`}
                    onClick={() => p.onToggleProvider(s.name)}>{s.name}</button>
            })}
        </div>
        <button className="show-all" disabled={!filtered} onClick={p.onShowAll}>All events</button>
    </nav>
}

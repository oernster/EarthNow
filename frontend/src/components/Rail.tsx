// The action rail down the left side (FR-RAIL-001 to 003). The order is
// FR-RAIL-002's: the view first (rotation, Reset view, zoom), then the data
// (Refresh, provider status), then Settings and Help. The donate button sits at
// the rail's foot (FR-DON-001).
import {icons} from '../icons'
import type {HelpKind} from './HelpDialogs'
import {HelpMenu} from './HelpMenu'
import {RailButton} from './RailButton'

interface Props {
    autoRotate: boolean
    // attention is true while the provider status has something to report.
    attention: boolean
    onToggleRotate: () => void
    onResetView: () => void
    onZoom: (zoomIn: boolean) => void
    onRefresh: () => void
    onStatus: () => void
    onSettings: () => void
    onHelp: (kind: HelpKind) => void
    onDonate: () => void
}

// FR-DON-005: the tooltip and accessible name, as PigeonPost's.
export const DONATE_LABEL = 'Donate to support EarthNow'

export function Rail(p: Props) {
    // NFR-UX-004: the button shows the state it switches TO. While rotating it
    // shows the crossed icon and offers to stop.
    const rotation = p.autoRotate
        ? {label: 'Stop rotating', icon: icons.rotateStop}
        : {label: 'Start rotating', icon: icons.rotate}
    const statusLabel = p.attention ? 'Provider status: something to report' : 'Provider status'
    return <nav className="rail" aria-label="Actions">
        <div className="rail-actions">
            <RailButton {...rotation} onClick={p.onToggleRotate}/>
            <RailButton label="Reset view" icon={icons.resetView} onClick={p.onResetView}/>
            <RailButton label="Zoom in" icon={icons.zoomIn} onClick={() => p.onZoom(true)}/>
            <RailButton label="Zoom out" icon={icons.zoomOut} onClick={() => p.onZoom(false)}/>
            <RailButton label="Refresh now" icon={icons.refresh} onClick={p.onRefresh}/>
            <RailButton label={statusLabel} icon={icons.status} attention={p.attention} onClick={p.onStatus}/>
            <RailButton label="Settings" icon={icons.settings} onClick={p.onSettings}/>
            <HelpMenu icon={icons.help} onChoose={p.onHelp}/>
        </div>
        {/* FR-DON-001: the donate button is the rail's own button, pinned to its
            foot with no tray of its own. It belongs to nothing on screen, so it
            sits apart, where nothing else is reached by accident. */}
        <RailButton label={DONATE_LABEL} icon={icons.donate} className="rail-foot" onClick={p.onDonate}/>
    </nav>
}

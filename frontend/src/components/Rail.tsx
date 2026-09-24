// The action rail down the left side (FR-RAIL-001 to 003). The order is
// FR-RAIL-002's: the view first (rotation, clouds, Reset view, zoom), then the data
// (Refresh, provider status), then Settings and Help. The donate button sits at
// the rail's foot (FR-DON-001). Every name is read from railLabels.
import type {CSSProperties} from 'react'
import {icons} from '../icons'
import {useProductName} from '../product'
import {donateLabel, RAIL_LABELS} from '../railLabels'
import {REFRESH_TURN_MS} from '../useHeld'
import type {HelpKind} from './HelpDialogs'
import {HelpMenu} from './HelpMenu'
import {RailButton} from './RailButton'

interface Props {
    autoRotate: boolean
    // attention is true while the provider status has something to report.
    attention: boolean
    onToggleRotate: () => void
    cloudsShown: boolean
    onToggleClouds: () => void
    onResetView: () => void
    onZoom: (zoomIn: boolean) => void
    onRefresh: () => void
    // refreshing turns the Refresh button's icon while a fetch runs (FR-STS-007).
    refreshing?: boolean
    onStatus: () => void
    onSettings: () => void
    onHelp: (kind: HelpKind) => void
    onDonate: () => void
}

export function Rail(p: Props) {
    const product = useProductName()
    // NFR-UX-004: the button shows the state it switches TO. While rotating it
    // shows the crossed icon and offers to stop.
    const rotation = p.autoRotate
        ? {label: RAIL_LABELS.stopRotating, icon: icons.rotateStop}
        : {label: RAIL_LABELS.startRotating, icon: icons.rotate}
    // FR-CLD-001: hidden shows the plain artwork; shown adds the negative.
    const clouds = p.cloudsShown
        ? {label: RAIL_LABELS.hideClouds, icon: icons.cloudCoverHide}
        : {label: RAIL_LABELS.showClouds, icon: icons.cloudCover}
    // The turn's length has its one home in useHeld; the stylesheet reads it here.
    const turn = {'--refresh-turn': `${REFRESH_TURN_MS}ms`} as CSSProperties
    return <nav className="rail" aria-label="Actions" style={turn}>
        <div className="rail-actions">
            <RailButton {...rotation} onClick={p.onToggleRotate}/>
            <RailButton {...clouds} onClick={p.onToggleClouds}/>
            <RailButton label={RAIL_LABELS.resetView} icon={icons.resetView} onClick={p.onResetView}/>
            <RailButton label={RAIL_LABELS.zoomIn} icon={icons.zoomIn} onClick={() => p.onZoom(true)}/>
            <RailButton label={RAIL_LABELS.zoomOut} icon={icons.zoomOut} onClick={() => p.onZoom(false)}/>
            <RailButton label={RAIL_LABELS.refresh} icon={icons.refresh} onClick={p.onRefresh}
                className={p.refreshing ? 'busy' : undefined} aria-busy={p.refreshing}/>
            <RailButton label={p.attention ? RAIL_LABELS.statusAttention : RAIL_LABELS.status} icon={icons.status}
                attention={p.attention} onClick={p.onStatus}/>
            <RailButton label={RAIL_LABELS.settings} icon={icons.settings} onClick={p.onSettings}/>
            <HelpMenu icon={icons.help} onChoose={p.onHelp}/>
        </div>
        {/* FR-DON-001: the donate button is the rail's own button, pinned to its
            foot with no tray of its own. It belongs to nothing on screen, so it
            sits apart, where nothing else is reached by accident. */}
        <RailButton label={donateLabel(product)} icon={icons.donate} className="rail-foot" onClick={p.onDonate}/>
    </nav>
}

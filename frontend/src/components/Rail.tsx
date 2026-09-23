// The action rail down the left side (FR-RAIL-001 to 003), in the house
// nav-band style ported from PigeonPost's TitleBar: artwork buttons, each named
// by an immediate tooltip and an accessible name. The foot of the rail is kept
// for the donate tray (FR-DON-001).
import {icons} from '../icons'

interface Props {
    autoRotate: boolean
    onToggleRotate: () => void
    onResetView: () => void
    onRefresh: () => void
    onSettings: () => void
}

interface Action {
    label: string
    icon: string
    onClick: () => void
}

function RailButton({label, icon, onClick}: Action) {
    return <button data-stop className="rail-btn" data-tip={label} aria-label={label} onClick={onClick}>
        <img src={icon} alt="" draggable={false}/>
    </button>
}

export function Rail(p: Props) {
    // NFR-UX-004: the button shows the state it switches TO. While rotating it
    // shows the crossed icon and offers to stop.
    const rotation: Action = p.autoRotate
        ? {label: 'Stop rotating', icon: icons.rotateStop, onClick: p.onToggleRotate}
        : {label: 'Start rotating', icon: icons.rotate, onClick: p.onToggleRotate}
    return <nav className="rail" aria-label="Actions">
        <div className="rail-actions">
            <RailButton {...rotation}/>
            <RailButton label="Reset view" icon={icons.resetView} onClick={p.onResetView}/>
            <RailButton label="Refresh now" icon={icons.refresh} onClick={p.onRefresh}/>
            <RailButton label="Settings" icon={icons.settings} onClick={p.onSettings}/>
        </div>
    </nav>
}

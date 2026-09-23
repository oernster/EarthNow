// The time window control (FR-TW-001): one button per window.
import type {WindowDTO} from '../types'

interface Props {
    windows: WindowDTO[]
    selected: string
    onChoose: (key: string) => void
}

export function TimeWindow({windows, selected, onChoose}: Props) {
    return <div className="time-window" role="group" aria-label="Time window">
        {windows.map(w => <button
            key={w.key}
            className={w.key === selected ? 'segment current' : 'segment'}
            aria-pressed={w.key === selected}
            onClick={() => onChoose(w.key)}
        >{w.label}</button>)}
    </div>
}

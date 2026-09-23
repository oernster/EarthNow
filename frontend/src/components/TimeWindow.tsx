// The time window control (FR-TW-001): one button per window. On the ring it is
// one stop per option (NFR-KBD-005): Tab walks the options in drawn order and
// leaves at the ends, because the ring carries on past them; Up and Down walk
// the same options wrapping. The current option does nothing when pressed, so
// it is not a stop (keeb invariant 4).
import {walkGroup} from '../ring'
import type {ChoiceDTO} from '../types'

interface Props {
    windows: ChoiceDTO[]
    selected: string
    onChoose: (key: string) => void
}

export function TimeWindow({windows, selected, onChoose}: Props) {
    return <div className="time-window" role="group" aria-label="Time window" onKeyDown={walkGroup}>
        {windows.map(w => {
            const current = w.key === selected
            return <button
                key={w.key}
                className={current ? 'segment current' : 'segment'}
                aria-pressed={current}
                data-stop={current ? undefined : true}
                onClick={() => onChoose(w.key)}
            >{w.label}</button>
        })}
    </div>
}

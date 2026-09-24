// The settings dialog (FR-SET-001 to 003). Every control applies and saves at
// once; the dialog and the rail's rotation button read the one settings value,
// so they cannot disagree (FR-GLB-011).
//
// The frame, its ring, Escape and Close are the shared Dialog's. Each radio set
// is a strip of peers like the time window: the chosen option is not a stop, Up
// and Down walk the rest and Enter or Space chooses.
import {walkGroup} from '../ring'
import type {SettingChoicesDTO, SettingsDTO} from '../types'
import {Dialog} from './Dialog'

const SECONDS_PER_MINUTE = 60

interface Props {
    settings: SettingsDTO
    choices: SettingChoicesDTO
    onChange: (patch: Partial<SettingsDTO>) => void
    onClose: () => void
}

export function SettingsDialog({settings, choices, onChange, onClose}: Props) {
    return <Dialog title="Settings" onClose={onClose}>
        <label className="field-row">
            <input data-stop type="checkbox" checked={settings.autoRotate}
                onChange={e => onChange({autoRotate: e.target.checked})}/>
            Rotate the globe when idle
        </label>
        <label className="field-row">
            <input data-stop type="checkbox" checked={settings.trailsShown}
                onChange={e => onChange({trailsShown: e.target.checked})}/>
            Show storm tracks
        </label>
        <fieldset onKeyDown={walkGroup}>
            <legend>Idle rotation speed</legend>
            {choices.speeds.map(s => <label key={s.key} className="field-row">
                <input data-stop={settings.speed === s.key ? undefined : true} type="radio" name="speed"
                    checked={settings.speed === s.key} onChange={() => onChange({speed: s.key})}/>
                {s.label} (one turn in {s.secondsPerRevolution / SECONDS_PER_MINUTE} min)
            </label>)}
        </fieldset>
        <fieldset onKeyDown={walkGroup}>
            <legend>Earthquakes shown (USGS minimum magnitude)</legend>
            {choices.magnitudes.map(m => <label key={m.key} className="field-row">
                <input data-stop={settings.magnitude === m.key ? undefined : true} type="radio" name="magnitude"
                    checked={settings.magnitude === m.key} onChange={() => onChange({magnitude: m.key})}/>
                {m.label}
            </label>)}
        </fieldset>
    </Dialog>
}

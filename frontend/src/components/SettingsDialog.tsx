// The settings dialog (FR-SET-001 to 003). Every control applies and saves at
// once; the dialog and the rail's rotation button read the one settings value,
// so they cannot disagree (FR-GLB-011).
import {useEffect} from 'react'
import type {SettingChoicesDTO, SettingsDTO} from '../types'

const SECONDS_PER_MINUTE = 60

interface Props {
    settings: SettingsDTO
    choices: SettingChoicesDTO
    onChange: (patch: Partial<SettingsDTO>) => void
    onClose: () => void
}

export function SettingsDialog({settings, choices, onChange, onClose}: Props) {
    useEffect(() => {
        const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
        window.addEventListener('keydown', onKey)
        return () => window.removeEventListener('keydown', onKey)
    }, [onClose])

    return <div className="scrim" onClick={onClose}>
        <div className="dialog" role="dialog" aria-modal="true" aria-labelledby="settings-title" onClick={e => e.stopPropagation()}>
            <h2 id="settings-title">Settings</h2>
            <label className="field-row">
                <input type="checkbox" checked={settings.autoRotate}
                    onChange={e => onChange({autoRotate: e.target.checked})}/>
                Rotate the globe when idle
            </label>
            <fieldset>
                <legend>Idle rotation speed</legend>
                {choices.speeds.map(s => <label key={s.key} className="field-row">
                    <input type="radio" name="speed" checked={settings.speed === s.key}
                        onChange={() => onChange({speed: s.key})}/>
                    {s.label} (one turn in {s.secondsPerRevolution / SECONDS_PER_MINUTE} min)
                </label>)}
            </fieldset>
            <fieldset>
                <legend>Earthquakes shown (USGS minimum magnitude)</legend>
                {choices.magnitudes.map(m => <label key={m.key} className="field-row">
                    <input type="radio" name="magnitude" checked={settings.magnitude === m.key}
                        onChange={() => onChange({magnitude: m.key})}/>
                    {m.label}
                </label>)}
            </fieldset>
            <div className="dialog-actions"><button onClick={onClose}>Close</button></div>
        </div>
    </div>
}

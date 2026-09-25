// The replay control (FR-RPL-008, FR-RPL-020, FR-RPL-024, FR-RPL-025,
// NFR-KBD-009): Play/Pause, the scrubber, Now and the speed, after the time
// window in the top bar. The scrubber's right end is the span's end; only Now
// returns to the present. Each is one ring stop; the scrubber's arrows step a
// hundredth of the span and Home and End reach the ends, as a range input does,
// while Space plays or pauses.
import {icons} from '../icons'
import type {ReplaySpeedDTO} from '../types'

// The scrubber's resolution: positions 0 to SCRUB_STEPS, stepped by a hundredth.
export const SCRUB_STEPS = 1000
const SCRUB_STEP = SCRUB_STEPS / 100

export const REPLAY_LABELS = {
    play: 'Play replay',
    pause: 'Pause replay',
    position: 'Replay position',
    now: 'Now',
    nowName: 'Back to now',
    speed: (current: string, next: string) => `Replay speed ${current}; press for ${next}`,
}

interface Props {
    position: number
    playing: boolean
    replaying: boolean
    // speed is the chosen speed and next the one a press moves to; null until
    // the speeds are known, when the speed button is not shown.
    speed: ReplaySpeedDTO | null
    next: ReplaySpeedDTO | null
    onPlay: () => void
    onPause: () => void
    onSeek: (position: number) => void
    onNow: () => void
    onSpeed: () => void
}

export function ReplayControls({position, playing, replaying, speed, next, onPlay, onPause, onSeek, onNow, onSpeed}: Props) {
    const toggle = playing ? onPause : onPlay
    const label = playing ? REPLAY_LABELS.pause : REPLAY_LABELS.play
    const speedName = speed && next ? REPLAY_LABELS.speed(speed.label, next.label) : ''
    return <div className="replay" role="group" aria-label="Replay">
        <button data-stop className="replay-btn" aria-label={label} title={label} onClick={toggle}>
            <img src={playing ? icons.pause : icons.play} alt="" draggable={false}/>
        </button>
        <input data-stop type="range" className="scrubber" aria-label={REPLAY_LABELS.position}
            min={0} max={SCRUB_STEPS} step={SCRUB_STEP} value={Math.round(position * SCRUB_STEPS)}
            onChange={e => onSeek(Number(e.target.value) / SCRUB_STEPS)}
            onKeyDown={e => {
                if (e.key !== ' ') return
                e.preventDefault()
                toggle()
            }}/>
        <button data-stop className="replay-text" aria-label={REPLAY_LABELS.nowName} title={REPLAY_LABELS.nowName}
            disabled={!replaying} onClick={onNow}>{REPLAY_LABELS.now}</button>
        {speed && <button data-stop className="replay-text" aria-label={speedName} title={speedName}
            onClick={onSpeed}>{speed.label}</button>}
    </div>
}

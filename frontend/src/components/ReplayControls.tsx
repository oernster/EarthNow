// The replay control (FR-RPL-008, FR-RPL-020, NFR-KBD-009): Play/Pause and the
// scrubber, after the time window in the top bar. The scrubber's right end is
// now. Each is one ring stop; the scrubber's arrows step a hundredth of the
// span and Home and End reach the ends, as a range input does, while Space
// plays or pauses.
import {icons} from '../icons'

// The scrubber's resolution: positions 0 to SCRUB_STEPS, stepped by a hundredth.
export const SCRUB_STEPS = 1000
const SCRUB_STEP = SCRUB_STEPS / 100

export const REPLAY_LABELS = {play: 'Play replay', pause: 'Pause replay', position: 'Replay position'}

interface Props {
    position: number
    playing: boolean
    onPlay: () => void
    onPause: () => void
    onSeek: (position: number) => void
}

export function ReplayControls({position, playing, onPlay, onPause, onSeek}: Props) {
    const toggle = playing ? onPause : onPlay
    const label = playing ? REPLAY_LABELS.pause : REPLAY_LABELS.play
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
    </div>
}

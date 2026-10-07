// The idle countdown (FR-GLB-019): a thin bar along the foot of the globe area
// that empties as the idle delay runs out, so a stopped globe reads as waiting
// rather than stuck. Keyed by resumeAt, so each new pause draws it afresh.
import {useState} from 'react'

export function ResumeBar({resumeAt}: {resumeAt: number}) {
    // Read once at mount: a bar drawn part way through a delay (rotation switched
    // on while one runs) empties over what is left of it.
    const [remainingMs] = useState(() => Math.max(0, resumeAt - Date.now()))
    return <div className="resume-bar" aria-hidden="true" style={{animationDuration: `${remainingMs}ms`}}/>
}

// The one modal shell (NFR-KBD-006, FR-HLP-005): every dialog in the window is
// this frame round its own content, so the ring, the opening focus, Escape, the
// scrim and Close are written once.
//
// It is its own ring (keeb invariant 6): it opens on its first stop, Escape
// closes it and focus returns to whatever opened it. The window's ring is made
// inert by the caller while one is showing.
//
// A reading dialog (About, the licence, the notices, the guide) puts its content
// in a body that reads itself (the house auto-scroll) with Close pinned beneath
// it, as PigeonPost's LicenceModal does, so the way out never scrolls away. That
// body is a stop only while it overflows, is never the stop a dialog opens on
// and paints no ring in any state (noborderfocus: a text view never rings).
import {useEffect, useId, useRef, type ReactNode} from 'react'
import {noClickFocus, useFirstStop, useRing} from '../ring'
import {useAutoScroll} from '../useAutoScroll'

interface Props {
    title: string
    onClose: () => void
    // reading gives the content a self-reading body and the wider reading frame.
    reading?: boolean
    children: ReactNode
}

export function Dialog({title, onClose, reading = false, children}: Props) {
    const frame = useRef<HTMLDivElement>(null)
    const titleId = useId()
    const autoScroll = useAutoScroll()
    useRing(frame)
    useFirstStop(frame)

    useEffect(() => {
        const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
        window.addEventListener('keydown', onKey)
        return () => window.removeEventListener('keydown', onKey)
    }, [onClose])

    return <div className="scrim" onClick={onClose}>
        <div ref={frame} className={reading ? 'dialog reading' : 'dialog'} role="dialog" aria-modal="true"
            aria-labelledby={titleId} onClick={e => e.stopPropagation()}>
            <h2 id={titleId}>{title}</h2>
            {reading
                ? <div ref={autoScroll} className="reading-body" data-stop data-reading tabIndex={-1} onMouseDown={noClickFocus}>{children}</div>
                : children}
            <div className="dialog-actions"><button data-stop onClick={onClose}>Close</button></div>
        </div>
    </div>
}

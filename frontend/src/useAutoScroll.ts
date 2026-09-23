// The self-reading cycle on a React surface (FR-HLP-005, FR-HLP-006), ported from
// PigeonPost's hooks/useAutoScroll.ts. The machine and its DOM driver are NOT
// here: they have one home, shared with the setup page; this hook only hands
// them the element and says when it is the one being looked at.
//
// The cycle is not gated on prefers-reduced-motion. On Windows that query follows
// the general Animation effects switch, which people turn off for performance, so
// gating on it silently removes the feature; a reader who does not want it
// touches the pane, which is what the suspension is for (PigeonPost's ruling).
import {useCallback, useEffect, useState} from 'react'
import {startAutoScroll} from '../../installer/frontend/dist/auto-scroll.js'

// SCRIM is the class every modal's backdrop wears (components/Dialog.tsx).
const SCRIM = '.scrim'

/**
 * useAutoScroll answers a ref callback for the element that actually scrolls, so
 * a surface mounted again with its dialog starts a fresh cycle each time.
 */
export function useAutoScroll(): (node: HTMLElement | null) => void {
    const [node, setNode] = useState<HTMLElement | null>(null)
    useEffect(() => {
        if (!node) return
        return startAutoScroll(node, () => isTopmostSurface(node))
    }, [node])
    return useCallback((next: HTMLElement | null) => setNode(next), [])
}

/**
 * isTopmostSurface reports whether the surface is the one the reader is looking
 * at. Two surfaces reading at once compete for the eye, so a surface under a
 * modal freezes until that modal closes; one outside any modal reads only while
 * no modal is open at all.
 */
export function isTopmostSurface(node: HTMLElement): boolean {
    const scrims = document.querySelectorAll(SCRIM)
    const own = node.closest(SCRIM)
    if (!own) return scrims.length === 0
    return scrims[scrims.length - 1] === own
}

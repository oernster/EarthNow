// An image layer's state and image (FR-CLD-009, FR-CLD-016, FR-BA-006,
// FR-BA-008): the state is read again on each named event; the image, being
// large, only when the state's key says it has changed.
import {useCallback, useEffect, useRef, useState} from 'react'
import {on} from './api'

type Refused = (reason: string) => void

export interface LayerSource<T> {
    state: (onRefused: Refused) => Promise<T | null>
    image: (onRefused: Refused) => Promise<string | null>
    // key names the image the state describes; empty while there is none.
    key: (state: T) => string
    // events are the backend events after which the state is read again.
    events: readonly string[]
}

export function useLayer<T>(source: LayerSource<T>, ready: boolean, onProblem: Refused): [T | null, string] {
    const [state, setState] = useState<T | null>(null)
    const [image, setImage] = useState('')
    const held = useRef('')
    const load = useCallback(() => {
        void source.state(onProblem).then(s => {
            if (!s) return
            setState(s)
            const key = source.key(s)
            if (key === held.current) return
            held.current = key
            void source.image(onProblem).then(url => { if (url !== null) setImage(url) })
        })
    }, [source, onProblem])
    useEffect(() => {
        if (!ready) return
        load()
        const offs = source.events.map(name => on(name, load))
        return () => offs.forEach(off => off())
    }, [load, ready, source])
    return [state, image]
}

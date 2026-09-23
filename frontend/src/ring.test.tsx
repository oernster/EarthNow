// The keyboard ring, ported from ED Voyage Companion's hooks.test.tsx.
//
// jsdom performs no layout, so every element reports no offset parent and the
// ring would find no stops. layOut states the shape of the page explicitly; what
// is asserted is what the ring then does with it.
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {fireEvent, render, screen} from '@testing-library/react'
import {useRef, useState} from 'react'
import {useFirstStop, useRing, walkGroup} from './ring'

function layOut() {
    Object.defineProperty(HTMLElement.prototype, 'offsetParent', {
        configurable: true,
        get(this: HTMLElement) { return this.parentElement },
    })
}

function unlayOut() {
    delete (HTMLElement.prototype as unknown as Record<string, unknown>).offsetParent
}

/** named identifies whatever holds focus; null when nothing does. */
function named(): string | null {
    const active = document.activeElement as HTMLElement | null
    if (active === null || active === document.body) return null
    return active.getAttribute('aria-label') ?? active.textContent
}

function Ringed({enabled = true, onToggle = () => undefined}: {enabled?: boolean; onToggle?: () => void}) {
    const shell = useRef<HTMLDivElement>(null)
    useRing(shell, enabled)
    return <div ref={shell}>
        <button data-stop type="button">first</button>
        <input data-stop aria-label="text" type="text"/>
        <input data-stop aria-label="check" type="checkbox" onChange={onToggle}/>
        <button data-stop type="button" disabled>skipped</button>
        <button type="button">not a stop</button>
        <button data-stop type="button">last</button>
    </div>
}

describe('the main ring', () => {
    beforeEach(layOut)
    afterEach(unlayOut)

    it('NFR-KBD-002 starts neutral and NFR-KBD-001 steps forward on Tab and on Right', () => {
        render(<Ringed/>)
        expect(named()).toBeNull()
        fireEvent.keyDown(document, {key: 'Tab'})
        expect(named()).toBe('first')
        fireEvent.keyDown(document, {key: 'ArrowRight'})
        expect(named()).toBe('text')
    })

    it('NFR-KBD-001 steps back on Shift+Tab and on Left from a neutral start', () => {
        render(<Ringed/>)
        fireEvent.keyDown(document, {key: 'ArrowLeft'})
        expect(named()).toBe('last')
        fireEvent.keyDown(document, {key: 'Tab', shiftKey: true})
        expect(named()).toBe('check')
    })

    it('NFR-KBD-001 wraps at both ends', () => {
        render(<Ringed/>)
        screen.getByText('last').focus()
        fireEvent.keyDown(document, {key: 'Tab'})
        expect(named()).toBe('first')
        fireEvent.keyDown(document, {key: 'Tab', shiftKey: true})
        expect(named()).toBe('last')
    })

    it('skips a disabled stop and anything not marked as one', () => {
        render(<Ringed/>)
        screen.getByLabelText('check').focus()
        fireEvent.keyDown(document, {key: 'Tab'})
        expect(named()).toBe('last')
    })

    it('leaves a text field its own arrows and still lets Tab out', () => {
        render(<Ringed/>)
        const field = screen.getByLabelText('text')
        field.focus()
        fireEvent.keyDown(field, {key: 'ArrowRight'})
        expect(named()).toBe('text')
        fireEvent.keyDown(field, {key: 'Tab'})
        expect(named()).toBe('check')
    })

    // A checkbox is not a text field, so the horizontal arrows step the ring off
    // it; Enter equals Space on it.
    it('steps off a checkbox on Right and toggles it on Enter', () => {
        const toggled = vi.fn()
        render(<Ringed onToggle={toggled}/>)
        const check = screen.getByLabelText('check')
        check.focus()
        fireEvent.keyDown(check, {key: 'Enter'})
        expect(toggled).toHaveBeenCalledTimes(1)
        fireEvent.keyDown(check, {key: 'ArrowRight'})
        expect(named()).toBe('last')
    })

    it('does nothing while switched off, as under a modal', () => {
        render(<Ringed enabled={false}/>)
        fireEvent.keyDown(document, {key: 'Tab'})
        expect(named()).toBeNull()
    })

    it('leaves a key it has no meaning for alone', () => {
        render(<Ringed/>)
        fireEvent.keyDown(document, {key: 'q'})
        expect(named()).toBeNull()
    })

    it('is quiet over a container holding no stops', () => {
        function Empty() {
            const shell = useRef<HTMLDivElement>(null)
            useRing(shell)
            return <div ref={shell}/>
        }
        render(<Empty/>)
        fireEvent.keyDown(document, {key: 'Tab'})
        expect(named()).toBeNull()
    })
})

describe('a strip of peer options', () => {
    beforeEach(layOut)
    afterEach(unlayOut)

    function Strip() {
        return <div onKeyDown={walkGroup}>
            <button data-stop>1 h</button>
            <button>24 h</button>
            <button data-stop>7 days</button>
        </div>
    }

    it('NFR-KBD-005 walks its options on Up and Down, wrapping, without choosing', () => {
        render(<Strip/>)
        screen.getByText('1 h').focus()
        fireEvent.keyDown(screen.getByText('1 h'), {key: 'ArrowDown'})
        expect(named()).toBe('7 days')
        fireEvent.keyDown(screen.getByText('7 days'), {key: 'ArrowDown'})
        expect(named()).toBe('1 h')
        fireEvent.keyDown(screen.getByText('1 h'), {key: 'ArrowUp'})
        expect(named()).toBe('7 days')
        fireEvent.keyDown(screen.getByText('7 days'), {key: 'q'})
        expect(named()).toBe('7 days')
    })
})

describe('a dialog', () => {
    beforeEach(layOut)
    afterEach(unlayOut)

    function Framed() {
        const frame = useRef<HTMLDivElement>(null)
        useFirstStop(frame)
        return <div ref={frame}>
            <button data-stop disabled>skipped</button>
            <button data-stop>usable</button>
        </div>
    }

    function Opener() {
        const [open, setOpen] = useState(false)
        return <>
            <button onClick={() => setOpen(true)}>opener</button>
            <button onClick={() => setOpen(false)}>shut</button>
            {open && <Framed/>}
        </>
    }

    it('NFR-KBD-006 opens on its first usable stop and hands focus back on close', () => {
        render(<Opener/>)
        const opener = screen.getByText('opener')
        opener.focus()
        fireEvent.click(opener)
        expect(named()).toBe('usable')
        fireEvent.click(screen.getByText('shut'))
        expect(named()).toBe('opener')
    })
})

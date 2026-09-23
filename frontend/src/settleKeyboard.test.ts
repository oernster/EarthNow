// The keyboard repair shared by the application and the setup page
// (installer/frontend/dist/settle-keyboard.js; keeb hosted webview rule).
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {KEYBOARD_SETTLE_MS, settleKeyboard} from '../../installer/frontend/dist/settle-keyboard.js'

describe('settleKeyboard', () => {
    beforeEach(() => {
        vi.useFakeTimers()
        // jsdom does not implement window.focus and says so on every call.
        vi.spyOn(window, 'focus').mockImplementation(() => undefined)
    })
    afterEach(() => { vi.useRealTimers(); vi.restoreAllMocks() })

    it('asks for the keyboard once the settle time shows the page has none', () => {
        vi.spyOn(document, 'hasFocus').mockReturnValue(false)
        const take = vi.fn(() => Promise.resolve())
        settleKeyboard(take)
        vi.advanceTimersByTime(KEYBOARD_SETTLE_MS - 1)
        expect(take).not.toHaveBeenCalled()
        vi.advanceTimersByTime(1)
        expect(take).toHaveBeenCalledOnce()
    })

    it('leaves a page that already holds the keyboard alone, focusing nothing', () => {
        vi.spyOn(document, 'hasFocus').mockReturnValue(true)
        const take = vi.fn(() => Promise.resolve())
        settleKeyboard(take)
        vi.advanceTimersByTime(KEYBOARD_SETTLE_MS)
        expect(take).not.toHaveBeenCalled()
        expect(document.activeElement).toBe(document.body)
    })
})

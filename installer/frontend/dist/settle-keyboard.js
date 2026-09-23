// Repairing a launch that came up with no keyboard at all, in its one home for
// both pages (keeb: hosted webview rule; NFR-KBD-002). The application imports
// this file (frontend/src/App.tsx) as the setup page does, so neither keeps a copy.
//
// Wails hands the webview its keyboard from the main window's WM_SETFOCUS, which
// Windows raises only on a change of focus; the handler for it is bound inside an
// asynchronous callback, so whether the first focus arrives before there is a
// handler for it is a race. Focusing an element is not the same as the document
// HAVING focus. The page is the only thing that can tell, so it checks once and
// asks the Go side for the keyboard back (TakeKeyboard, which focuses the WebView2
// child as a click would).

// KEYBOARD_SETTLE_MS is how long the webview is given to come up holding the
// keyboard before the page decides it has not.
export const KEYBOARD_SETTLE_MS = 400

/**
 * settleKeyboard checks once, after the settle time, that the page holds the
 * keyboard; when it does not, it asks for it and focuses the window again.
 * Nothing on the page is focused either way, so a neutral start stays neutral.
 *
 * @param {() => Promise<unknown>} takeKeyboard asks the Go side for the keyboard
 */
export function settleKeyboard(takeKeyboard) {
    window.focus()
    window.setTimeout(() => {
        if (document.hasFocus()) return
        // A refused request leaves the page as it was, where one click still
        // hands it the keyboard; there is nothing better to say or do.
        void takeKeyboard().then(() => window.focus(), () => undefined)
    }, KEYBOARD_SETTLE_MS)
}

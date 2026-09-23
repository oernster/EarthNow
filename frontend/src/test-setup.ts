// What jsdom does not provide, provided. Ported from ED Voyage Companion's
// frontend/src/test-setup.ts.
//
// jsdom performs no layout, so ResizeObserver is absent. It is stubbed inert
// rather than worked around in each test: a component that constructs one would
// otherwise throw the moment it renders. Nothing here fakes a measurement; a test
// that needs geometry states it.

class InertObserver {
    observe(): void {}
    unobserve(): void {}
    disconnect(): void {}
}

if (!('ResizeObserver' in globalThis)) {
    ;(globalThis as unknown as {ResizeObserver: unknown}).ResizeObserver = InertObserver
}

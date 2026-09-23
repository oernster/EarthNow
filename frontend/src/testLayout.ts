// Layout for tests. jsdom performs no layout, so every element reports no offset
// parent and the ring would find no stops. layOut states the shape of the page
// explicitly; what is asserted is what the ring then does with it.

export function layOut() {
    Object.defineProperty(HTMLElement.prototype, 'offsetParent', {
        configurable: true,
        get(this: HTMLElement) { return this.parentElement },
    })
}

export function unlayOut() {
    delete (HTMLElement.prototype as unknown as Record<string, unknown>).offsetParent
}

/** named identifies whatever holds focus; null when nothing does. */
export function named(): string | null {
    const active = document.activeElement as HTMLElement | null
    if (active === null || active === document.body) return null
    return active.getAttribute('aria-label') ?? active.textContent
}

/** overflowing makes an element report more content than it shows; false gives a fit. */
export function overflowing(element: HTMLElement, overflows: boolean) {
    Object.defineProperty(element, 'clientHeight', {configurable: true, value: 100})
    Object.defineProperty(element, 'scrollHeight', {configurable: true, value: overflows ? 1000 : 100})
}

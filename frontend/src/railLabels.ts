// The rail's names, in their one home: the buttons' tooltips and accessible names
// (FR-RAIL-003, FR-DON-005) and the guide's entries (FR-HLP-004) all read these,
// so the guide can never name a button differently from the button itself.
export const RAIL_LABELS = {
    startRotating: 'Start rotating',
    stopRotating: 'Stop rotating',
    resetView: 'Reset view',
    zoomIn: 'Zoom in',
    zoomOut: 'Zoom out',
    refresh: 'Refresh now',
    status: 'Provider status',
    statusAttention: 'Provider status: something to report',
    settings: 'Settings',
    help: 'Help',
} as const

/** donateLabel is the donate button's name (FR-DON-005), built from the product's. */
export function donateLabel(product: string): string {
    return `Donate to support ${product}`
}

// The category table: the one home of every category's emoji and name
// (Appendix D.3, FR-KEY-002). The key, the markers and the tooltip all read it.
// The order is DATA-002's.

export interface CategoryInfo {
    key: string
    emoji: string
    name: string
}

export const CATEGORIES: readonly CategoryInfo[] = [
    {key: 'EARTHQUAKE', emoji: '〰️', name: 'Earthquake'},
    {key: 'VOLCANO', emoji: '🌋', name: 'Volcano'},
    {key: 'WILDFIRE', emoji: '🔥', name: 'Wildfire'},
    {key: 'SEVERE_STORM', emoji: '🌀', name: 'Severe storm'},
    {key: 'FLOOD', emoji: '🌊', name: 'Flood'},
    {key: 'LANDSLIDE', emoji: '🪨', name: 'Landslide'},
    {key: 'DROUGHT', emoji: '🏜️', name: 'Drought'},
    {key: 'DUST', emoji: '💨', name: 'Dust and haze'},
    {key: 'ICE', emoji: '🧊', name: 'Ice'},
    {key: 'OTHER', emoji: '📍', name: 'Other'},
]

const OTHER = CATEGORIES[CATEGORIES.length - 1]

export function categoryOf(key: string): CategoryInfo {
    return CATEGORIES.find(c => c.key === key) ?? OTHER
}

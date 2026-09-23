// The guide's words (FR-HLP-004), kept apart from the markup as PigeonPost keeps
// guideContent.ts, so a change to what the app says is a change to one document.
// Category emoji and names are NOT written here: they come from the category
// table, their one home; this file holds only what each category covers.
import {CATEGORIES, type CategoryInfo} from './categories'

export interface GuideSection {
    heading: string
    paragraphs: string[]
}

export const GUIDE_SECTIONS: readonly GuideSection[] = [
    {
        heading: 'The time window',
        paragraphs: [
            'The buttons above the globe choose how far back to look. An event is shown when its own time falls inside the chosen window, ending now.',
            'The line beneath the globe counts the events shown and names the window it applies to.',
        ],
    },
    {
        heading: 'How times are worded',
        paragraphs: [
            'Each event carries the time its source gives it. An exact time reads as "Observed 14 min ago"; a source that gives only a date reads as "Reported for 18 Sep 2026, 5 days ago".',
            'Separately, each source shows when it was last retrieved, for example "retrieved 3 min ago". The detail panel gives every time three ways: in words, in UTC and in your own time zone.',
            'Nothing is labelled live. Each source publishes with its own delay and is fetched on a schedule, so what you see is what the sources had said by the time shown.',
        ],
    },
    {
        heading: 'Stale and failed sources',
        paragraphs: [
            'A source is marked stale when its last successful fetch is several of its refresh intervals old, so its events may be out of date.',
            'When a fetch fails, the globe keeps showing what that source last gave and tries again after a wait that grows with each failure. The status button on the left gives the reason and the time of the next attempt.',
            'Refresh fetches every source at once; it can be used again after a short pause.',
        ],
    },
]

// COVERS says what each category holds, keyed by the category table's keys.
export const COVERS: Readonly<Record<string, string>> = {
    EARTHQUAKE: 'Earthquakes from the USGS, at or above the minimum magnitude chosen in Settings, plus any earthquake EONET reports.',
    VOLCANO: 'Volcanic activity tracked by EONET.',
    WILDFIRE: 'Wildfires tracked by EONET.',
    SEVERE_STORM: 'Tropical cyclones and other severe storms tracked by EONET.',
    FLOOD: 'Floods tracked by EONET.',
    LANDSLIDE: 'Landslides tracked by EONET.',
    DROUGHT: 'Droughts tracked by EONET.',
    DUST: 'Dust storms and haze tracked by EONET.',
    ICE: 'Sea and lake ice, such as icebergs, tracked by EONET.',
    OTHER: 'Any EONET event in a category not listed above.',
}

/** guideCategories answers each category, in the table's order, with what it covers. */
export function guideCategories(): {category: CategoryInfo; covers: string}[] {
    return CATEGORIES.map(category => ({category, covers: COVERS[category.key] ?? ''}))
}

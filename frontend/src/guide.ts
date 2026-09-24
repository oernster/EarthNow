// The guide's words (FR-HLP-004), ported from PigeonPost's guideContent.ts: held
// apart from the dialog that draws them, so the component stays a renderer and
// the words stay one readable document. It NAMES the furniture first, each entry
// carrying the REAL picture the control draws (the rail's own icon, the key's own
// emoji), so a control that is a picture can be recognised by someone who has
// just met it. Then it states the rules the window cannot say for itself.
//
// Names come from their one homes: the rail's from railLabels, the categories'
// emoji and names from the category table. This file holds only what each does.
import {CATEGORIES} from './categories'
import {icons} from './icons'
import {donateLabel, RAIL_LABELS} from './railLabels'

// GuideEntry is one named piece of furniture: its picture (an image; an emoji for a
// category), what it is called and what it does.
export interface GuideEntry {
    icon?: string
    emoji?: string
    name: string
    text: string
}

// GuideSection is one block of the document; the dialog draws its entries, then
// its paragraphs.
export interface GuideSection {
    heading: string
    intro?: string
    entries?: readonly GuideEntry[]
    paragraphs?: readonly string[]
}

// COVERS says what each category holds, keyed by the category table's keys.
export const COVERS: Readonly<Record<string, string>> = {
    EARTHQUAKE: 'earthquakes from the USGS, at or above the minimum magnitude chosen in Settings, plus any earthquake EONET reports.',
    VOLCANO: 'volcanoes in the Smithsonian and USGS weekly volcanic activity report (GVP), dated by the day the report was issued, plus any EONET tracks.',
    WILDFIRE: 'wildfires tracked by EONET.',
    SEVERE_STORM: 'tropical cyclones and other severe storms tracked by EONET.',
    FLOOD: 'floods tracked by EONET.',
    LANDSLIDE: 'landslides tracked by EONET.',
    DROUGHT: 'droughts tracked by EONET.',
    DUST: 'dust storms and haze tracked by EONET.',
    ICE: 'sea and lake ice, such as icebergs, tracked by EONET.',
    OTHER: 'any EONET event in a category not listed above.',
}

/** guideSections answers the guide, the donate entry named for product. */
export function guideSections(product: string): readonly GuideSection[] {
    return [
        {
            heading: 'The buttons down the left',
            intro: 'Hover any of them or reach it with Tab to see its name.',
            entries: [
                {icon: icons.rotate, name: RAIL_LABELS.startRotating, text: 'turns the globe slowly while you are not using it. While it turns, the button shows a cross and stops it.'},
                {icon: icons.resetView, name: RAIL_LABELS.resetView, text: 'brings the whole globe back into view.'},
                {icon: icons.zoomIn, name: RAIL_LABELS.zoomIn, text: 'moves the camera closer; the mouse wheel does the same.'},
                {icon: icons.zoomOut, name: RAIL_LABELS.zoomOut, text: 'moves the camera further away.'},
                {icon: icons.refresh, name: RAIL_LABELS.refresh, text: 'fetches every source at once; it can be used again after a short pause.'},
                {icon: icons.status, name: RAIL_LABELS.status, text: 'each source\'s state, why a fetch failed and when the next attempt is. A dot on the button means there is something to read.'},
                {icon: icons.settings, name: RAIL_LABELS.settings, text: 'idle rotation, its speed and the smallest earthquake shown.'},
                {icon: icons.help, name: RAIL_LABELS.help, text: 'this guide, About, the licence and the third-party notices.'},
                {icon: icons.donate, name: donateLabel(product), text: 'opens the donation page in your browser. Nothing is held back without a donation.'},
            ],
        },
        {
            heading: 'The categories',
            intro: 'The key on the right names each one; pressing a row hides or shows its events.',
            entries: CATEGORIES.map(c => ({emoji: c.emoji, name: c.name, text: COVERS[c.key] ?? ''})),
        },
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
                'A volcano\'s report covers the week before it is issued, so it reads as reported for that day.',
                'Separately, each source shows when it was last retrieved, for example "retrieved 3 min ago". The detail panel gives every time three ways: in words, in UTC and in your own time zone.',
                'Nothing is shown as happening this instant. Each source publishes with its own delay and is fetched on a schedule, so what you see is what the sources had said by the time shown.',
            ],
        },
        {
            heading: 'Stale and failed sources',
            paragraphs: [
                'A source is marked stale when its last successful fetch is several of its refresh intervals old, so its events may be out of date.',
                'When a fetch fails, the globe keeps showing what that source last gave and tries again after a wait that grows with each failure. The status button gives the reason and the time of the next attempt.',
            ],
        },
    ]
}

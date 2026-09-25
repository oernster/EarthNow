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
                {icon: icons.rotate, name: RAIL_LABELS.startRotating, text: 'turns the globe slowly while you are not using it, first bringing the whole globe back into view. While it turns, the button shows a cross and stops it.'},
                {icon: icons.cloudCover, name: RAIL_LABELS.showClouds, text: 'lays the clouds over the globe as weather satellites last saw them. While they show, the button shows a cross and hides them.'},
                {icon: icons.dayNight, name: RAIL_LABELS.showDayNight, text: 'shades the side of the Earth where it is night and lights its cities. It shows from the first run; while it shows, the button shows a cross and hides it.'},
                {icon: icons.resetView, name: RAIL_LABELS.resetView, text: 'brings the whole globe back into view.'},
                {icon: icons.zoomIn, name: RAIL_LABELS.zoomIn, text: 'moves the camera closer; the mouse wheel does the same.'},
                {icon: icons.zoomOut, name: RAIL_LABELS.zoomOut, text: 'moves the camera further away.'},
                {icon: icons.refresh, name: RAIL_LABELS.refresh, text: 'fetches every source at once; it can be used again after a short pause.'},
                {icon: icons.status, name: RAIL_LABELS.status, text: 'each source\'s state, why a fetch failed and when the next attempt is. A dot on the button means there is something to read.'},
                {icon: icons.settings, name: RAIL_LABELS.settings, text: 'idle rotation, its speed, the smallest earthquake shown and whether storm tracks and burnt areas show.'},
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
                'An earthquake\'s detail gives its depth with USGS\'s band: shallow above 70 km, intermediate to 300 km, deep beyond. When USGS cannot compute a depth it assigns 10 km, a fixed depth, so a depth of exactly 10 km is marked as often fixed rather than measured.',
                'Separately, each source shows when it was last retrieved, for example "retrieved 3 min ago". The detail panel gives every time three ways: in words, in UTC and in your own time zone.',
                'A storm draws a faint track behind its marker through the positions its source gave inside the chosen time window, fading with age. It shows where the storm has been, never a forecast; Settings can hide it.',
                'Nothing is shown as happening this instant. Each source publishes with its own delay and is fetched on a schedule, so what you see is what the sources had said by the time shown.',
            ],
        },
        {
            heading: 'The clouds',
            paragraphs: [
                'The clouds come from EUMETSAT\'s world cloud map, a mosaic of weather satellites\' infrared images made every three hours. The line beneath the globe gives the time the image shows in UTC with how long ago that was.',
                'An infrared image sees temperature, not colour: the colder a surface, the whiter the cloud drawn. Snow, ice and cold high ground can therefore read as cloud.',
                'A grey veil marks where no satellite sees, mostly near the poles, so an unseen region never reads as a clear sky.',
                'The clouds are fetched only while they are shown. The last image is kept, so they still show without a connection, marked with their age.',
            ],
        },
        {
            heading: 'Day and night',
            paragraphs: [
                'The sun\'s place is worked out on this computer from the time, so the layer needs no connection. The line between day and night moves with the real sun, a quarter of a degree a minute.',
                'Dawn and dusk fade across a band either side of that line rather than cutting off, as twilight does.',
                'The lights on the night side are NASA\'s Black Marble picture of the Earth at night, a composite from 2016: they show where cities are, not which lights are on now.',
                'Clouds over the night side are drawn fainter, so they never glow white in the dark.',
            ],
        },
        {
            heading: 'Replay',
            paragraphs: [
                'Play, beside the time window, replays the chosen window from its start, in 30 seconds at normal speed: events appear at their own times, storm tracks grow, the sun sweeps round and the burnt days build up. The slider beside it moves through the window by hand. The replay holds at the end of the window until Now returns to the present. The speed button after the slider plays it at half, normal or double speed, kept for next time.',
                'Replay shows what the sources hold now about those days. A source may since have revised or withdrawn a report, so it is not always what was known at the time.',
                'Its clouds are softer images than the ordinary layer, fetched for the replay and let go when it ends. They arrive while it plays; the line beneath the globe counts them in.',
            ],
        },
        {
            heading: 'Burnt areas',
            paragraphs: [
                'Settings can lay burnt ground over the globe in red, from the Global Wildfire Information System (GWIS), which maps it from satellites one UTC day at a time. Every day the chosen time window touches is drawn, so 7 days shows eight days of maps.',
                'A burn shows on the day it was mapped, which may be after the day the fire began.',
                'Each point of the map covers about 20 km at the equator, so a small burn may not show at all.',
                'A short window may show none yet, when nothing has been mapped for the days it touches; the line beneath the globe says so.',
                'The maps are fetched only while they are shown. The last ones are kept, so they still show without a connection, marked with their age.',
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

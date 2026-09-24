# EarthNow: Software Requirements Specification

Status: **Baselined, version 1.0 (2026-09-23), with amendments 1 to 32.**
Changes from here arrive as numbered amendments with a reason, never as silent
edits.

| No. | Date | Change | Reason |
|---|---|---|---|
| 1 | 2026-09-23 | FR-SPK-006 now requires the built executable to run standalone from another folder; the second-account install is removed. FR-SPK-003's Greenwich coordinate corrected to (51.4778, -0.0014). | The second-account install was not house style (owner). The earlier longitude carried the wrong sign; Greenwich lies just west of 0. |
| 2 | 2026-09-23 | FR-GLB-013 added: the camera altitude fits the whole globe to the globe area at launch and on resize. FR-GLB-007 keeps the altitude on focus; FR-GLB-008 resets to the fit altitude; NFR-UX-001 reads "globe area". | The spike's fixed altitude clipped the globe top and bottom at 1264 x 761 (owner's screenshot). |
| 3 | 2026-09-23 | FR-KEY-001 to 004 added: a key down the left side naming each category's emoji. | Owner request: the emoji need a key. |
| 4 | 2026-09-23 | FR-GEO-001 to 008 added: hovering a marker shows the nearest populated place with its country, distance and direction, resolved offline from embedded Natural Earth data; the keyboard cursor and the detail panel show the same line. | Owner request, hover rather than click. Offline because NFR-PRIV-001 allows no host beyond the two providers. |
| 5 | 2026-09-23 | DATA-005 now marks day precision per event, only when every date is midnight; DATA-012 added for USGS types and withdrawn events. | The captured EONET fixture showed a storm track with a genuine 00:00Z fix, which the per-date rule would have misread. USGS feeds carry quarry blasts and explosions, which are not earthquakes. |
| 6 | 2026-09-23 | FR-RAIL-001 to 003 added: an action rail down the left side, 68 px wide, carrying the action icons at 48 px with the donate button at its foot. The key moves to the right side (FR-KEY-001, 003); FR-GLB-013's globe area is the window less the rail and the key; FR-DON-001, 002 and 007 follow the donate button into the rail. | NFR-UX-001 leaves 55 px of height for full-width bars at 960 x 600, less than one PigeonPost header (61 px glyphs); a rail spends width instead. 960 less the 220 px key less the 68 px rail leaves 672 px, which at 600 px high is exactly 70% (owner's choice of layout). |
| 7 | 2026-09-23 | FR-RAIL-002 names the full rail: zoom in and zoom out after Reset view, provider status after Refresh, Help last as a menu of the four help dialogs. The status button carries a dot while FR-STS-003, FR-STS-005 or FR-SET-004 has something to report; standing notices move from the status line to the popover. The filter and time window icons of Appendix D.2 are not rail buttons. | The rail lacked the D.2 icons for status, help and zoom (owner). The key is the filter and the time window's control carries no icon, so neither needs a button (owner, 2026-09-23). Eight buttons plus the donate tray measure 574 px against the 600 px minimum height. |
| 8 | 2026-09-23 | FR-DON-001, 002 and 008 now seat the donate button in the rail itself, pinned to its foot, with no tray or border of its own; it is the last stop of the rail, after Help. FR-DON-003 names the Go side as the one holder of the address, with the page asking it to open the page; FR-DON-006 covers a refusal the application can observe. | Owner: the button at the bottom left without an additional tray. Ring order is reading order (keeb invariant 1), so a button drawn in the rail is reached with the rail. Wails hands a link to the desktop without reporting whether a browser opened, so the observable refusals are the https allowlist and an unconnected backend. |
| 9 | 2026-09-23 | NFR-KBD-003 now focuses the WebView2 child directly, with the page asking again through TakeKeyboard when it finds it has no keyboard. NFR-KBD-004 adds that arriving on the globe moves the cursor to an event at once and the tooltip names the keys. FR-HLP-001 adds the copyright notice. FR-HLP-004's guide leads each entry with the control's own picture. The key column's heading is the application mark alone. | `runtime.Show` alone lost the race inside Wails; the log measured the first focus failing and the page's request succeeding. The owner reported Tab after 7 d going nowhere: the globe was a stop painting nothing. The owner asked for the copyright, the guide's pictures as in ClearBudget and PigeonPost and a heading that is the mark alone, since its artwork reads EarthNow. |
| 10 | 2026-09-23 | FR-SEL-009 added: the source link is a page, never a data file; the EONET adapter takes the first source that is a page; a source that is only a file is shown as text. | The owner opened a hurricane's source and received a download. Measured in the live EONET feed: three storms list a JTWC `.tcw` warning file first (two with an NHC page second, one with nothing else) and one iceberg source is a `.csv`. |
| 11 | 2026-09-23 | FR-PRV-001 now retrieves every EONET event of the week, open and closed, a closed one marked as ended in its detail panel. FR-SET-002 and OQ-003 lower the default USGS minimum magnitude to 2.5. | The owner doubted how little the globe showed. Measured that day: EONET held 80 events for the week while the open-only request returned 17, dropping most wildfires and floods though each happened inside the window; USGS held 362 earthquakes at 2.5 and above against 243 at 3.0 (owner's choice). |
| 12 | 2026-09-23 | FR-PRV-015 added: a third provider, the Smithsonian / USGS Weekly Volcanic Activity Report (GVP), each volcano dated by the report's issue day at day precision. NFR-PRIV-001 allows its host, `volcano.si.edu`. FR-SEL-009 counts `.cfm` as a page ending. DATA-003 reads a GDACS-sourced EONET polygon latitude first. | EONET tracked no volcano in the 30 days to 2026-09-23 (25 in a year) while that week's report listed 20 erupting volcanoes; the feed carries a georss point per item (measured). The owner accepted the recommendation. All 14 GDACS flood polygons of the week arrived [lat, lng] against GeoJSON order while GDACS points did not: five were dropped as out of range and nine were drawn in the wrong place, Honduras in Antarctica. The Smithsonian feed answers 403 to a request asking only for JSON, so each adapter states what it accepts. |
| 13 | 2026-09-24 | CON-003 and the architecture read a JSON file per provider, not SQLite. FR-PRV-003 states the 2.5 default of amendment 11. NFR-LEG-001 is verified by the notices generator's check. Scope names the third provider. OQ-011 names the rail. | The code measured against the document during the documentation pass; the owner chose to bring the document to the code. |
| 14 | 2026-09-24 | FR-DON-010 is verified against the button's font size rather than the line box; it states that the donate button stands taller than a text-only button. Appendix D.1's donate row names the site's copy as written. | Measured on the built site: the mark is 45.9 px on 17 px type, exactly 2.7em, which is 1.69 times the 27.2 px line box at line height 1.6. The 1.8 figure was the step up from the earlier 1.5em mark, not a ratio to the line box. The site now exists, so `docs/donate.png` is written rather than pending (owner). |
| 15 | 2026-09-24 | FR-MRK-010 added: every marker keeps its fit-altitude size on screen while the camera zooms. FR-FLT-001 offers a toggle for every category, present or not, as the key lists them. Appendix D.1 names the page's mark and the site's icon among the application icon's uses. | Markers were built to hold their launch size on screen, so that two overlapping markers separate as the camera closes in (FR-MRK-008); at a fixed size in globe units they grow with the gap between them. The owner approved writing it down. FR-KEY-001 and FR-FLT-001 described one control two ways; the owner chose the key's reading, which keeps it steady as events come and go. `tools/genicons.py` writes both from one render. |
| 16 | 2026-09-24 | FR-GEO-002 and 003 count a position on an Antarctic ice shelf as Antarctica, never as sea. FR-GEO-008 adds Natural Earth's 1:10m Antarctic ice shelves (159 shelves) to the embedded data. | Natural Earth draws Antarctica to its grounded coast and the ice shelves as a layer of their own, so points on the Ross and Ronne shelves read "At sea" (measured). The owner chose to count a shelf as Antarctica over naming it as ice; the owner approved the download. |
| 17 | 2026-09-24 | DATA-003 reads a GDACS polygon in the order its own vertices prove, else in the order the feed's proven polygons show, in place of amendment 12's fixed latitude-first reading. | Amendment 12's rule rested on one week and would draw every flood swapped if the source corrected its order. Measured over the 30 days to 2026-09-24: 15 of 57 GDACS polygons proved their order, all latitude first, at least one in each week; the other 42 read as valid either way round. Deciding by which reading lands in a country failed: 27 landed either way and a Kenya flood landed in Spain. |
| 18 | 2026-09-24 | NFR-SEC-002 states its purpose, no network origin, rather than `'self'` on every fetch directive. DATA-001 describes the domain as built: an `Event` holding its sightings as observations, with the retrieved-at instant on the provider's snapshot. | The page's `img-src` allows `data:` and `blob:` beside `'self'`; neither reaches the network. The code grouped each sighting's time, position and measurement into an observation and kept one retrieved-at per provider, so occurred at and observed at are one time. The owner chose to bring both requirements to the code. |
| 19 | 2026-09-24 | 1.4 and 2.1 name three providers; the time window reads "3 days" and "7 days"; day precision covers a volcano dated by its report; `EarthEvent` reads `Event`. FR-GEO-001's example carries the region. FR-PRV-008 reads "any other provider's". FR-STS-005 and NFR-REL-005 state the checks the tests make. DATA-003 places a polygon at the mean of its outer ring's vertices. 4.1 names every source. RSK-001 reads the 2.5 default; RSK-002 the Wails version in use. Appendix A names the GVP adapter; Appendix C the real test and table names; Appendix D the two masters that are not square, the macOS and Linux icons as release 2 and the zoom icons as FR-RAIL-002's. | The code measured against the document during the documentation pass: each statement was stale or read differently in the tree. No behaviour changes. |
| 20 | 2026-09-24 | RSK-003 states the real ceiling of EONET use: 6 scheduled requests an hour, up to 120 with manual refresh pressed at every chance. FR-PRV-006 states that a provider whose interval exceeds the backoff ceiling retries at the ceiling. CON-008's danger band reads 381 to 400. | Measured in the code: a manual refresh marks every provider due and its cooldown is 30 s; GVP's 60 min interval is cut to the 30 min ceiling on its first failure; the danger band test flags 381 to 400 inclusive, one line stricter than the house rule's 381 to 399; a file at the cap is no further from breaking it than one a line below, so the stricter reading stands. The owner accepted the first two as they stand and left the third to be settled. No behaviour changes. |
| 21 | 2026-09-24 | FR-STS-007 added: while a provider with events held is fetching, the status area says it is refreshing and the Refresh button's icon turns for at least one turn. | The owner pressed Refresh and saw nothing happen. The page learned of a fetch only when it finished, while the status line read "retrieved under a minute ago" before and after, so a refresh that worked looked like one that did nothing. |
| 22 | 2026-09-24 | FR-PRV-010: after every press of Refresh the status area states the local time of the last manual refresh made, in place of when the next becomes available. | The owner read "Refresh available now" after a refused press and could not tell what it meant. The wait was worded in whole minutes, so every wait inside the 30 s cooldown read "now"; the line also stayed up after the cooldown ended. The owner asked for the time of the last refresh instead. |
| 23 | 2026-09-24 | CON-003 limits "no CGO" to the Windows build. FR-GLB-002 rotates from launch, resuming after the idle delay. FR-MRK-005's tooltip holds the category's emoji, the title and the place line. DATA-004 names GVP's event time. Appendix D.1 says how the release 2 icons are made; D.4 says the space behind the globe is plain black. | The code measured against the document in the final documentation pass: the Linux and macOS builds need cgo; rotation starts at launch; the tooltip carries the emoji rather than the category's name; no starfield is drawn. No behaviour changes. |
| 24 | 2026-09-24 | The cloud layer (3.2.10, FR-CLD-001 to 016): a rail button showing or hiding EUMETSAT's world cloud map over the globe, its infrared brightness mapped to opacity, a veil where no satellite sees, the image's valid time and age in the status area. It is fetched only while shown and cached for offline use. NFR-UX-002's minimum window rises to 960 by 640. NFR-PRIV-001 allows `view.eumetsat.int`; NFR-LEG-002 credits EUMETSAT; NFR-PERF-005, ASM-007, ASM-008 and RSK-006 added. Scope drops satellite cloud imagery from the out-of-scope list; weather stays out. | Owner request: toggle global cloud cover as satellite imagery shows it, from a free keyless source. Measured: NASA GIBS carries no Meteosat, so Europe and Africa would show none; EUMETSAT's layer covers the whole ring. At the old minimum the rail held 33 px spare against the 62 a ninth button needs; the owner chose the larger window over tighter gaps or a button off the rail, a veil over unseen regions and the layer hidden on a first run. |
| 25 | 2026-09-24 | 4.1 names EUMETSAT among the sources that never suggest an endorsement. NFR-UX-004 names the cloud button among the toggles it covers. | The code measured against the document in the documentation pass: About credits EUMETSAT and says it does not endorse EarthNow; the cloud button shows the state it switches to, as FR-CLD-001 requires. No behaviour changes. |
| 26 | 2026-09-24 | ASM-007 is confirmed: the owner accepts the credit "Cloud images: EUMETSAT, world cloud map (EUMETView)." with EUMETSAT named in the non-endorsement line. NFR-LEG-002 states that wording. | The owner's ruling; the wording was provisional until then. No behaviour changes. |
| 27 | 2026-09-24 | The release obligations left open are closed by the owner's acceptance. ASM-001, ASM-003 and ASM-006 are confirmed by the shipped application; ASM-002 and ASM-004 are accepted as they stand. NFR-PERF-001 takes Phase 0's three starts as its result; NFR-PERF-003 is accepted unmeasured. | The owner judged the application fine as built and asked the document to match it. Nothing here is a new measurement: the 20-start sample and the 24 h working set were never taken, the EONET rate-limit window is still unknown and the NASA imagery terms were not re-read. No behaviour changes. |
| 28 | 2026-09-24 | The Linux Flatpak, its cleanup script and the macOS DMG move from release 2 into the released product: 1.3 takes them out of scope's exclusions, 2.3 lists the three platforms side by side, DEL-005 to DEL-007 read release 1, CON-003 and D.1 drop "release 2" and RSK-002 is closed. | The owner released v1.0.0 with `EarthNowSetup.exe`, `EarthNow.dmg` and `earthnow.flatpak` (the release's assets, read) and tested the notarised Apple Silicon DMG and the Flatpak on the latest Ubuntu LTS, where the globe drew with WebGL2. Emoji coverage under Linux and macOS fonts stays unmeasured (D.3). No behaviour changes. |
| 29 | 2026-09-24 | The day and night layer (3.2.11, FR-DAY-001 to 009): the globe lit by the real sun, NASA's Black Marble city lights on the night side, a rail button after the clouds, shown on a first run. The day and night terminator and the night-lights texture leave the out-of-scope list and the Won't list. NFR-UX-002's minimum window rises to 960 by 700; NFR-PERF-006, ASM-009 and the Black Marble credit (NFR-LEG-002) added. | Owner request: show where it is daytime on Earth now, from astronomy alone with no service. The owner chose city lights over a plain darkened side, a rail button over a setting and the layer on by default. The rail's own figures (54 px buttons, 8 px gaps, about 11 px spare at 640) leave a tenth button 51 px short. |
| 30 | 2026-09-24 | The start view (FR-GLB-014 to 017): at launch the globe opens facing the country the operating system's country or region setting names, at Natural Earth's label point for it, before idle rotation begins; with no usable region it opens as before and logs why. ASM-010 added. The Won't "favourites and home location" stands: nothing is chosen or kept inside EarthNow; the region is the operating system's own setting, read on the machine and sent nowhere. | Owner request: open over the reader's own part of the world. The owner chose the country or region setting over the time zone and the display language (measured on the reference machine: all three read the United Kingdom there; a UK machine often runs an en-US display language). It chose Natural Earth's label point over the capital or the outline's centre, with 1.0.0's view as the fallback. The owner also set the order of the work before the next release: this, then earthquake depth, recent-event trails and Replay the Earth, each its own amendment; no release is cut until all four are in. |
| 31 | 2026-09-24 | Earthquake depth (FR-SEL-010 to 014): the detail panel gains a Depth row for an event whose source gives one, to one decimal with USGS's band word; a depth above sea level is worded as such; exactly 10 km is marked as often a fixed depth. Nothing changes on the globe or in the tooltip. R10 and R11 added. | Owner request, the detail panel only: "I'd resist turning the globe into a Christmas tree." DATA-010 already keeps USGS depth in kilometres (measured: `Extras.DepthKm`); nothing carried it past the domain. Measured over the cached week of 244 USGS quakes at 2.5 and above: 165 shallow, 62 intermediate, 17 deep, 77 at exactly 10 km. USGS states that 10 km is a fixed depth assigned when the data are too poor to compute one (R11); its summary feed does not say which ones, so the row says often rather than claiming this one is. |
| 32 | 2026-09-24 | Storm trails (3.2.12, FR-TRL-001 to 005): a severe storm's fixes inside the time window are drawn as a line to its marker, oldest faintest, switched by a Settings box that starts on. The Won't "storm tracks drawn as lines" leaves 1.3 and 3.6; iceberg drift and earthquake swarms are recorded there instead. | Owner request: show where a storm has been. The owner chose storms only (measured on 2026-09-23: five storms with 6 to 14 fixes over one to three days; icebergs held 41 to 65 fixes back to 2021, which a window clips to a stub; swarms need a rule, measured at 8 groups of 5 or more within 50 km in 48 h and none of 10 or more within 20 km in a day) and a Settings box over a rail button, since the rail is full at the minimum height. Wildfire perimeters were researched apart: FIRMS gives points, NIFC's are the United States only (rejected by the owner as inconsistent) and GWIS's global burnt areas come as a daily image, a candidate for its own amendment. |

Source: `EarthWatch-Implementation-Plan.md` (Oliver Ernster, supplied
2026-09-23), renamed to EarthNow by the owner. Section references of the form
"Plan 9" point into that document.

---

## 1. Introduction

### 1.1 Purpose

EarthNow is a desktop application that shows what is happening on Earth right
now: an interactive, slowly rotating 3D globe carrying recent natural events at
their geographic positions, each traceable to the public source that reported
it and honest about how fresh that report is.

The product sentence, which settles any unclear choice:

> **EarthNow lets you open a globe and see what is happening on Earth right now.**

### 1.2 Intended audience

| Reader | Uses this document to |
|---|---|
| Oliver (owner) | decide scope, answer the open questions, accept each phase |
| Implementer (Claude Code) | build each phase against numbered requirements and name the test for each |
| Artwork producer (Oliver) | produce the assets listed in Appendix D |

### 1.3 Scope

**In scope for V1 (Windows, macOS and Linux):** the globe, the three providers (NASA EONET,
USGS earthquakes, the Smithsonian / USGS weekly volcano report), the cloud layer from EUMETSAT's world cloud map (3.2.10), the day and night layer (3.2.11), the provider-neutral event model, refresh with local caching,
filters, the time window, event selection with a detail panel, source and
freshness display, keyboard navigation to the house model, the self-reading
help surfaces, the Windows build script and the bespoke setup program, the
Linux Flatpak with its cleanup script and the macOS DMG (3.5).

**Out of scope for V1 (decided; each is a Won't in 3.6):**

- user accounts, cloud sync, social features, telemetry of any kind;
- AI summaries, interpretation, predictions or forecasts;
- push or desktop notifications;
- a historical archive beyond the 7-day window, timeline playback;
- NASA FIRMS hotspots, event polygons drawn as areas, iceberg drift and earthquake swarms drawn as trails;
- weather (precipitation, temperature, wind, forecasts), satellite imagery beyond the cloud layer, cloud history or animation, aurora;
- a server or backend controlled by EarthNow;
- GIS tooling (measurement, projections, layer management);
- an update check, a system tray icon, start with Windows.

### 1.4 Definitions

One term, one meaning, throughout.

| Term | Meaning |
|---|---|
| **Event** | One `Event`: a single phenomenon reported by one provider, with one marker position. |
| **Provider** | An adapter that retrieves one public source and maps it to events. V1 has three: `EONET`, `USGS`, `GVP`. |
| **Category** | A value from the internal vocabulary in 3.4 (DATA-002), never a provider's own term. |
| **Time window** | The user-selected span ending at the current instant: 1 h, 6 h, 24 h, 3 days or 7 days. |
| **Event time** | The instant used for window membership and freshness wording: see DATA-004. |
| **Retrieved at** | The instant EarthNow last received a successful response from a provider. |
| **Refresh interval** | The delay between scheduled fetches of one provider while it is healthy. |
| **Stale** | A provider whose last successful retrieval is older than its stale threshold (NFR-FRESH-001). |
| **Marker** | The drawn representation of one event on the globe. |
| **Ring** | The keyboard focus cycle of the house keeb model. |
| **Stop** | One position on the ring. |
| **Reference machine** | Oliver's Windows 11 desktop, on which every performance figure is measured. |
| **Cloud image** | One image of EUMETSAT's world cloud map: the 10.8 µm infrared brightness of the whole geostationary ring at one valid time, in plate carrée. |
| **Valid time** | The instant a cloud image shows, as the service lists it (a new one every 3 h). |
| **Cloud layer** | The cloud image drawn over the globe texture, its brightness mapped to opacity (FR-CLD-006). |
| **Day precision** | An event dated by day alone: an EONET event whose every geometry date is exactly 00:00:00Z (DATA-005) or a volcano dated by its report's issue day (FR-PRV-015). |

### 1.5 References

| Ref | Document |
|---|---|
| R1 | EarthWatch Implementation Plan (the source brief) |
| R2 | NASA EONET v3 API, https://eonet.gsfc.nasa.gov/docs/v3 |
| R3 | USGS GeoJSON summary feeds, https://earthquake.usgs.gov/earthquakes/feed/v1.0/geojson.php |
| R4 | USGS ComCat field definitions, https://earthquake.usgs.gov/data/comcat/index.php |
| R5 | NASA media usage guidelines, https://www.nasa.gov/nasa-brand-center/images-and-media/ |
| R6 | Natural Earth terms of use, https://www.naturalearthdata.com/about/terms-of-use/ |
| R7 | House skills: `installer`, `scroll`, `keeb`, `noborderfocus` |
| R8 | ED Voyage Companion (Go + Wails Windows delivery reference) |
| R9 | PigeonPost, SymDiary (Go + Wails flatpak and DMG script references) |
| R10 | USGS, Determining the Depth of an Earthquake, https://www.usgs.gov/programs/earthquake-hazards/determining-depth-earthquake |
| R11 | USGS FAQ, Why do so many earthquakes occur at a depth of 10km?, https://www.usgs.gov/faqs/why-are-so-many-earthquakes-located-10-km-deep |

---

## 2. Overall description

### 2.1 Product perspective

A new, standalone desktop build. No existing system is replaced. EarthNow talks
to exactly four external systems in V1, all public and keyless; it talks to
nothing else. The fourth, EUMETSAT, is asked only while the cloud layer is shown.

```
   NASA EONET v3 ──┐                         ┌── Globe (WebGL, React)
   USGS feeds ─────┤                         │
   Smithsonian GVP ┼── Go backend ── bound ──┤
   EUMETSAT WMS ───┘   (providers,  methods  └── Controls, detail panel
                        store, cache)
                          │
                    JSON files in %LOCALAPPDATA%
```

All network traffic originates in the Go backend. The web frontend makes no
network request of its own (NFR-SEC-002).

### 2.2 User classes

| Class | Needs | May not |
|---|---|---|
| Viewer | open the app and see recent events with no configuration | nothing is restricted; there are no roles |
| Keyboard user | reach every control and every visible event without a pointer | |

### 2.3 Operating environment

| Item | Windows | Linux and macOS |
|---|---|---|
| OS | Windows 10 and 11, x64 | Linux (Flatpak, GNOME runtime; tested on the latest Ubuntu LTS) and macOS arm64 |
| Web runtime | WebView2 (Evergreen) | WebKitGTK 4.1; WKWebView |
| GPU | WebGL2 required | WebGL2 required (RSK-002, closed) |
| Network | intermittent is normal; offline must be survivable | same |
| Install | per user, no administrator rights | Flatpak user install; DMG drag-install |

### 2.4 Constraints

| ID | Constraint | Rationale |
|---|---|---|
| CON-001 | Backend in Go; desktop shell Wails v2; frontend React with TypeScript built by Vite. | Owner decision. |
| CON-002 | Architecture `UI → Application → Domain ← Infrastructure` under `internal/`, enforced by `tests/structural`. | House invariant. |
| CON-003 | No CGO in the Windows build; the Linux and macOS builds need it, since Wails renders there through a C web view. The cache is one JSON file per provider in the data folder, written atomically. | House rule; single static binary. The event sets are small (the largest measured feed 1.51 MB) and read whole, so a database buys nothing. |
| CON-004 | Licence GPL-3.0 for EarthNow's own code (the `LICENSE` already committed). Every bundled dependency and data asset must carry a licence compatible with shipping inside a GPL-3.0 application, recorded in a third-party notices file. | Owner's committed licence. |
| CON-005 | `VERSION` at repo root is the only version literal; `build.ps1` passes it through `-ldflags -X` against a `var`. | House versioning rule. |
| CON-006 | Delivery follows the house Go + Wails Windows checklist: `build.ps1` plus an unskippable `test.ps1` gate; the setup program is a second Wails app under `installer/` built to the `installer` skill. | Owner request; house rule. |
| CON-007 | Keyboard navigation follows the `keeb` skill with its `noborderfocus` sub-skill; self-reading surfaces follow the `scroll` skill, ported from PigeonPost's `autoScroll.ts` and `useAutoScroll.ts`, never written from the skill's tables. | Owner request. |
| CON-008 | No module over 400 lines; a file in 381 to 400 is reduced to 350 or fewer. Build and packaging scripts exempt. | House rule. |
| CON-009 | Earth imagery is real observational data from NASA or Natural Earth, never generated artwork. | Source honesty (Plan 2.4); a generated Earth would be a fabricated record of the planet. |

### 2.5 Technology decision: the globe (Plan 20 items 2 and 3)

**Decision (owner, 2026-09-23): globe.gl (on three.js), with a bundled NASA
Blue Marble texture.** The decision rests on published facts; its runtime
claims are HYPOTHESES until the Phase 0 spike measures them (FR-SPK-001 to 007).

| Criterion (Plan 9) | globe.gl 2.46.2 | CesiumJS 1.145.0 |
|---|---|---|
| Licence | MIT (read) | Apache-2.0 (read) |
| Shipped size, measured on jsDelivr | 1.89 MB min, 526 KB gzip, three.js included | 6.0 MB min, 1.77 MB gzip; 79 MB npm unpacked |
| Lat/long markers | built-in point, HTML and object layers | built-in entities |
| Camera focus animation | `pointOfView(coords, ms)` | `camera.flyTo` |
| Offline texture | any equirectangular image | needs a tile-map imagery provider; default imagery calls `api.cesium.com` with an evaluation-only token |
| Scope fit | a globe | a GIS engine; Plan 13 warns against becoming one |
| WebGL | WebGL2 (three.js dropped WebGL1 at r163) | WebGL2 by default since 1.102 |
| Clustering | none built in (unconfirmed; the spike checks) | built-in entity clustering |

The weight, the default network dependency and the GIS orientation all count
against Cesium for a product whose centre is one calm globe. The one point in
Cesium's favour, clustering, is small enough to write against the marker count
EarthNow actually meets (see NFR-PERF-002). React integration uses
`react-globe.gl` (MIT, a wrapper over globe.gl) or globe.gl directly behind one
component; the spike picks whichever keeps the globe instance under the app's control.

**Linux risk carried forward (RSK-002):** Wails v2 on Linux defaults
`WebviewGpuPolicy` to Never when `options.Linux` is nil (read in Wails v2.12.0
source); WebGL2 availability in distribution WebKitGTK builds is
unconfirmed. Release 2 must measure it on a real Linux machine before the
Flatpak is promised. Measured at release: with the policy set to Always the
globe drew on the latest Ubuntu LTS (amendment 28).

### 2.6 Assumptions and dependencies

Every assumption has an owner and a confirm-by point.

| ID | Assumption | Owner | Confirm by |
|---|---|---|---|
| ASM-001 | EONET v3 stays keyless and reachable at its current URLs. | Oliver | Confirmed: the released application fetches it keyless (amendment 27) |
| ASM-002 | EONET's `X-RateLimit-Limit: 60` (measured 2026-09-23) permits one request per provider interval with room for manual refreshes. The window length is unconfirmed. | Implementer | Accepted by the owner, 2026-09-24 (amendment 27); the window length was never read |
| ASM-003 | USGS summary feeds keep their URLs; USGS states 30 days' notice before removal (read in a search summary, page not opened). | Implementer | Confirmed: the released application fetches them (amendment 27) |
| ASM-004 | NASA Blue Marble imagery may ship inside the installer with a credit line, under the NASA media terms (R5). GPL compatibility of bundling public-domain-like imagery is an inference, unconfirmed. | Oliver | Accepted by the owner, 2026-09-24 (amendment 27) |
| ASM-005 | WebView2 on the reference machine provides WebGL2. | Implementer | Phase 0 exit (measured, FR-SPK-002) |
| ASM-006 | The GitHub repository `oernster/EarthNow` is the release surface. | Oliver | Confirmed: releases are published there (amendment 27) |
| ASM-007 | EUMETSAT's terms allow a free application to fetch and display the world cloud map with a credit line. EUMETView's capabilities state no fees and no access constraints (read 2026-09-24); the licence page carries no readable terms without a script, so the credit wording and any condition are unconfirmed. | Oliver | Confirmed by the owner on 2026-09-24 (amendment 26) |
| ASM-008 | The `mumi:worldcloudmap_ir108` layer keeps its name, extent and 3-hourly time dimension at `https://view.eumetsat.int/geoserver/wms`. | Implementer | Cloud spike exit (FR-CLD-015) |
| ASM-009 | NASA's Black Marble 2016 may ship inside the application with a credit line under the NASA media terms (R5), as the Blue Marble does (ASM-004). | Oliver | Before the release carrying the day and night layer |
| ASM-010 | macOS keeps its Region setting as the region part of the `AppleLocale` preference; a Linux desktop's region is the territory of `LC_ALL`, else of `LANG` (FR-GLB-014). Only Windows is measured (the reference machine reads GB). | Oliver | On a Mac and on Linux, before the next release |

---

## 3. System features and requirements

Priority is MoSCoW. Verify: **T** automated test (the named suite), **D**
demonstration on the reference machine, **I** inspection. A number marked
"a target" is replaced by the measured value once Phase 0 or release testing
has taken it.

### 3.1 Phase 0: technical spike (Plan 18)

Nothing past Phase 0 is built until every Must here is demonstrated.

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-SPK-001 | Must | When the spike build is launched, the application shall open a window showing the textured globe against a black background. | Given a fresh build, when launched, then the globe is visible with the Blue Marble texture. | D |
| FR-SPK-002 | Must | The spike shall report the WebGL context version and the renderer string in its log. | Log line reads `webgl2` on the reference machine. | D |
| FR-SPK-003 | Must | When given the coordinate (51.4778, -0.0014), the spike shall draw a marker at the Greenwich Observatory. | Visual check against the texture's coastline at maximum zoom; error under one marker diameter. | D |
| FR-SPK-004 | Must | When a marker is clicked, the spike shall log that marker's identifier. | Clicking the Greenwich marker logs its id. | D |
| FR-SPK-005 | Must | When a marker is selected, the spike shall animate the camera to centre that marker. | Camera centres the marker within the focus duration of NFR-UX-003. | D |
| FR-SPK-006 | Must | The spike's built executable shall run standalone from a folder outside the repository, with the globe library and texture embedded. | Copied alone to another folder and launched; the globe draws and the log records its frames. | D |
| FR-SPK-007 | Must | The spike shall render the measured worst-case marker count (2,500 markers) at the frame rate of NFR-PERF-002. | Frame-time log over 60 s of idle rotation. | D |

**Phase 0 result (2026-09-23), measured in the spike's real WebView2 window at
1264 x 761 on the reference machine (RTX 4060):**

| ID | Result |
|---|---|
| FR-SPK-001 | Met: textured globe on black, confirmed by the owner's screenshot. |
| FR-SPK-002 | Met: `webgl2`, ANGLE Direct3D 11; maximum texture size 16,384 px, so the 21,600 px texture was never an option. |
| FR-SPK-003 | Met: the Greenwich marker drew on London (owner's check). |
| FR-SPK-004 | Met: markers respond to the pointer; 24 hover lookups logged. |
| FR-SPK-005 | Met: camera asked for 51.4778, -0.0014 and arrived at 51.4778, -0.0014 within 1.2 s. |
| FR-SPK-006 | Met: the 14.9 MB executable ran alone from a temporary folder. |
| FR-SPK-007 | Met: 2,501 emoji sprites, median frame 10.00 ms, 99th percentile 10.10 ms, worst single frame 30 to 40 ms across runs. |
| NFR-PERF-001 | First globe frame 627 to 654 ms from process start across three runs, accepted by the owner as the result (amendment 27); the 20-start sample was not taken. |
| FR-GLB-013 | Globe diameter 671 px in a 1064 x 761 globe area (88.2%). |
| FR-GEO | Nearest-place lookup 503 to 548 us per hover, 154 ms to load the data at start. Natural Earth spells French Guiana's region "Guinaa" (raw bytes checked); the product carries a documented correction table for such source errors. |
| Clustering | globe.gl and three-globe offer none (their typings checked); FR-MRK-007 is written by EarthNow. |

### 3.2 Functional requirements

#### 3.2.1 Globe and camera

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-GLB-001 | Must | The globe view shall render Earth with the bundled daytime texture on a black space background. | Screenshot inspection. | D |
| FR-GLB-002 | Must | The globe view shall rotate eastward at the idle speed (one revolution per 240 s) from launch; once input has stopped it (FR-GLB-003), it shall rotate again when no user input has arrived for the idle delay (10 s). | Rotation runs at launch; after a drag, rotation resumes 10 s after the last input; measured period 240 s plus or minus 5%. | T (application idle timer) + D |
| FR-GLB-003 | Must | When pointer drag, wheel or keyboard input targets the globe, the globe view shall stop idle rotation. | Rotation halts on the first input event. | T |
| FR-GLB-004 | Must | While auto-rotate is switched off in settings, the globe view shall not rotate on idle. | Toggle off, wait 30 s, no rotation. | T |
| FR-GLB-005 | Must | When the user drags on the globe, the globe view shall rotate the globe with the drag. | D | D |
| FR-GLB-006 | Must | When the user scrolls the wheel over the globe, the globe view shall zoom between the minimum and maximum altitudes (0.15 and 4.0 globe radii). | Zoom clamps at both limits. | T + D |
| FR-GLB-007 | Must | When an event is selected, the globe view shall animate the camera to centre that event over the focus duration (NFR-UX-003), keeping the current altitude. | Given a selected event on the far side, the camera arrives centred on it at unchanged altitude. | D |
| FR-GLB-008 | Should | When "Reset view" is activated, the globe view shall return the camera to the fit altitude (FR-GLB-013). | D | D |
| FR-GLB-013 | Must | When the application starts or the globe area is resized, the globe view shall set the camera altitude so the whole globe fits the globe area with its drawn diameter between 85% and 92% of the area's shorter side. The globe area is the window less the action rail (FR-RAIL-001) and the key (FR-KEY-001). | Spike measured 2026-09-23 at 1264 x 761: altitude 1.629, diameter 671 px, 88.2% of the shorter side; the whole globe visible. | T (fit maths) + D |
| FR-GLB-010 | Must | While auto-rotate is on, the rotation button shall show `rotate` with the `negative` overlay and the tooltip "Stop rotating"; while auto-rotate is off, it shall show plain `rotate` with the tooltip "Start rotating" (NFR-UX-004). | Given rotation on, the button shows the crossed icon; one press stops rotation and the button shows the plain icon. | T |
| FR-GLB-011 | Must | When the rotation button is activated, the application shall switch auto-rotate and persist the new value as the FR-SET-001 setting. | The button and the settings dialog never disagree. | T |
| FR-GLB-012 | Must | The globe view shall rotate on idle whatever the operating system's reduced-motion or animation-effects setting; the auto-rotate setting is the only switch. | With Windows Animation effects off, rotation still starts after the idle delay. Rationale: on Windows that switch is commonly turned off for performance, the same ground on which the house scroll ruling declined to gate on it. | T |
| FR-GLB-009 | Must | If WebGL2 is unavailable, then the application shall show a message naming WebGL2 as the missing requirement, in place of the globe. | Force-disable WebGL in a test harness; the message appears; the app does not exit. | T (frontend) |
| FR-GLB-014 | Must | When the application starts, the application shall read the operating system's country or region setting as an ISO 3166-1 alpha-2 code: on Windows the user's home location, on macOS the region of `AppleLocale`, on Linux the territory of `LC_ALL`, else of `LANG` (amendment 30, ASM-010). | Each platform's reading from a recorded value: `en_GB.UTF-8` gives GB; `C.UTF-8`, `POSIX` and an empty value give none. The reference machine reads GB. | T |
| FR-GLB-015 | Must | When the region read by FR-GLB-014 is in the label table (FR-GLB-017), the globe view shall open with that country's label point facing the viewer at the fit altitude (FR-GLB-013), before idle rotation begins (FR-GLB-002). | Given the region GB, the first view faces 54.40 N, 2.12 W; rotation then starts from there. | T |
| FR-GLB-016 | Must | If no region can be read or the label table holds none for the region read, then the globe view shall open with 1.0.0's launch view and the application shall log the reason (NFR-OBS-001). | Given no region (or the region 001, the world) the first view sets the altitude alone; the log names why. | T |
| FR-GLB-017 | Must | The label table shall be generated from Natural Earth's admin-0 countries, the layer FR-GEO-008 embeds: each `ISO_A2_EH` code to its `LABEL_Y` and `LABEL_X`. Where a code names several rows, the generator shall keep the row whose `ISO_A2` is that code, else the row that is its own sovereign; if a code is still left with more than one row, then the generator shall refuse to write the table. | Natural Earth 5.1.1 gives one row each for AU, BR, FR and KZ (Australia, Brazil, France, Kazakhstan, measured); a planted second row stops the generator. | T |

#### 3.2.2 Markers

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-MRK-001 | Must | The globe view shall draw one marker per displayed event at the event's marker position (DATA-003). | A USGS fixture event at (61.899, -150.919) draws at that position. | T (projection) + D |
| FR-MRK-002 | Must | The globe view shall draw each marker with its category's emoji from the category table (Appendix D.3). | Each of the ten categories renders its own glyph. | D |
| FR-MRK-003 | Must | The globe view shall size an earthquake marker by magnitude band (bands: below 3.0, 3.0 to 4.5, 4.5 to 6, 6 and above). | Fixture quakes of 1.8, 3.1, 5.0 and 6.7 draw at four distinct sizes. | T |
| FR-MRK-004 | Must | The globe view shall draw every non-earthquake marker at one fixed size. | EONET markers carry equal size whatever their magnitude value. | T |
| FR-MRK-005 | Must | When the pointer hovers a marker, the globe view shall show a tooltip holding the category's emoji and the event title, with the place line beneath (FR-GEO-001). | D | D |
| FR-MRK-006 | Must | While an event is selected, the globe view shall draw that event's marker with the selection treatment (a ring around the emoji in the selection colour). | D | D |
| FR-MRK-007 | Must | Where markers overlap on screen at the current zoom, the globe view shall draw them as one cluster marker showing their count. | Two fixture events 1 km apart at launch altitude draw as one cluster reading 2. | T (clustering) + D |
| FR-MRK-008 | Must | When a cluster marker is activated, the globe view shall zoom toward the cluster until its members separate or maximum zoom is reached. | D | D |
| FR-MRK-009 | Must | The globe view shall draw no marker animation other than the selection treatment and hover state. | Inspection: no pulsing, flashing or looping animation on any marker. | I |
| FR-MRK-010 | Must | While the camera zooms, the globe view shall keep every marker, cluster marker included, at the size on screen it has at the fit altitude (FR-GLB-013): its size in globe units is its fit-altitude size times the camera altitude over the fit altitude. The sizes of FR-MRK-003 and FR-MRK-004 are those fit-altitude sizes. | Vitest: a marker's scale is its fit-altitude size times the altitude ratio at 0.25, 1 and 3. | T |

#### 3.2.3 Providers and refresh

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-PRV-001 | Must | The EONET provider shall retrieve every event of the widest time window (7 days), open and closed, from `/api/v3/events?status=all&days=7`; the detail panel shall say when the source has marked an event as ended. | Request URL asserted in a test with a fake HTTP client. | T |
| FR-PRV-002 | Must | The EONET provider shall parse the response body as JSON whatever `Content-Type` the server declares. | Measured 2026-09-23: EONET labels a JSON body `application/rss+xml` even when `Accept: application/json` is sent. Fixture served with that header parses. | T |
| FR-PRV-015 | Must | The GVP provider shall retrieve `https://volcano.si.edu/news/WeeklyVolcanoRSS.xml` hourly and map each item to a Volcano event at its `georss:point`, dated by the item's publish date at day precision, titled by the volcano and the week's activity and linked to the item's `guid` page; the feed's declared ISO-8859-1 shall be decoded. | The captured feed yields 20 events, Krakatau at (-6.1009, 105.4233) dated 17 Sep 2026; an unusable item is dropped and counted. | T |
| FR-PRV-003 | Must | The USGS provider shall retrieve the week feed at the highest published threshold not above the configured minimum magnitude, then keep only events at or above that minimum (default 2.5, amendment 11: `2.5_week.geojson` taken whole). USGS publishes feeds only at all, 1.0, 2.5, 4.5 and significant. | Request URL asserted. | T |
| FR-PRV-004 | Must | The USGS provider shall send `If-Modified-Since` carrying the `Last-Modified` value of its previous successful response. | Measured: USGS sends `Last-Modified`. Second request carries the header; a 304 keeps the stored events. | T |
| FR-PRV-005 | Must | The refresh scheduler shall fetch each provider on its own refresh interval (USGS 60 s, matching its measured `max-age=60`; EONET 10 min). | Fake clock advances 60 s; exactly one USGS fetch occurs. | T |
| FR-PRV-006 | Must | If a provider fetch fails, then the refresh scheduler shall retry that provider with the delay doubling from its refresh interval up to the backoff ceiling (30 min). A provider whose interval exceeds the ceiling retries at the ceiling, sooner than it is asked while healthy. | Fake clock: failures at 60, 120, 240 s, then capped at 1800 s. | T |
| FR-PRV-007 | Must | When a provider fetch succeeds after failures, the refresh scheduler shall restore that provider's normal refresh interval. | T | T |
| FR-PRV-008 | Must | If one provider fails, then the event store shall keep serving any other provider's events unchanged. | USGS fake returns 500; EONET events remain displayed. | T |
| FR-PRV-009 | Must | When the user activates "Refresh now", the refresh scheduler shall fetch every provider not already fetching. | T | T |
| FR-PRV-010 | Must | If "Refresh now" is activated within the manual refresh cooldown (30 s) of the previous manual refresh, then the refresh scheduler shall skip the fetch. After every activation the status area shall state the local time of the last manual refresh made ("Last refreshed at 17:52:10"); it never says whether a refresh is available. | Protects the EONET rate limit (ASM-002). | T |
| FR-PRV-011 | Must | If a provider response exceeds the response size cap (16 MB; the largest measured feed, USGS `all_week`, was 1.51 MB), then the provider shall discard it and report a size refusal naming the provider and the cap. | Oversized fixture refused; nothing allocated beyond the cap. | T |
| FR-PRV-012 | Must | If a provider response fails to parse, then the provider shall report a parse failure naming the provider and keep the previously stored events. | Malformed fixture; stored events survive. | T |
| FR-PRV-013 | Must | If a single feature or event inside an otherwise valid response is malformed, then the provider shall drop that item, count it and keep the rest. | Fixture with one bad feature among ten yields nine events and a dropped count of 1. | T |
| FR-PRV-014 | Must | The application shall register providers through one provider interface, so adding a provider changes only infrastructure and the composition root. | Structural test: no file outside `internal/infrastructure/providers/<name>` and `main.go` names a concrete provider. | T (structural) |

#### 3.2.4 Time window, filters and count

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-TW-001 | Must | The time window control shall offer 1 h, 6 h, 24 h, 3 days and 7 days. | T | T |
| FR-TW-002 | Must | While a time window is selected, the event query shall include an event only if its event time falls within that window, ending at the current instant. | Fake clock at 12:00; events at 11:30 and 10:30 with 1 h selected yield only the first. | T |
| FR-TW-003 | Must | When the application starts with no saved choice, the time window control shall select 24 h. | T | T |
| FR-FLT-001 | Must | The filter control shall offer one toggle per category in DATA-002, in DATA-002 order, whether or not the store holds an event of that category; the key's rows are those toggles (FR-KEY-001). | A store with only quakes and storms still offers all ten toggles, the other eight reading 0 (FR-KEY-004), plus "All events". | T |
| FR-FLT-002 | Must | When a category toggle is switched off, the globe view shall hide that category's markers. | T | T |
| FR-FLT-003 | Must | When "All events" is activated, the filter control shall switch every category toggle on. | T | T |
| FR-FLT-004 | Must | The filter control shall offer one toggle per provider. | Switching off USGS hides every USGS event. | T |
| FR-FLT-005 | Should | When the application closes, the settings store shall persist the time window and filter state; when the application starts, the controls shall restore them. | T | T |
| FR-KEY-001 | Must | While the globe view is showing, the main window shall show a key down its right side listing each category in DATA-002 order as its emoji beside its category name. | Every category in the category table renders one key row, emoji first. | T |
| FR-KEY-002 | Must | The key shall read its emoji and names from the category table (Appendix D.3), the same single home the markers read. | Structural test: no emoji literal outside the category table. | T |
| FR-KEY-003 | Must | The key shall never overlap the globe: the globe area ends at the key's left edge (FR-GLB-013). | D at the minimum window size. | D |
| FR-RAIL-001 | Must | The main window shall carry an action rail down its left side, 68 px wide, holding the action buttons one above another, each drawing its artwork at the rail's glyph size of 48 px (`--rail-glyph-size`); the globe area begins at the rail's right edge. | T (render) + D at the minimum window size. | T + D |
| FR-RAIL-002 | Must | The action rail's buttons shall be, top to bottom: rotation (FR-GLB-010), clouds (FR-CLD-001), day and night (FR-DAY-005), Reset view (FR-GLB-008), zoom in and zoom out (FR-GLB-006), Refresh (FR-PRV-009), provider status (FR-STS-003), Settings (FR-SET), then Help, a menu opening the guide, About, the licence and the third-party notices (FR-HLP). | Vitest reads the rail's accessible names in that order. | T |
| FR-RAIL-003 | Must | Every rail button's tooltip shall open to the right of the button, so it is not clipped at the window's left edge. | D at the minimum window size. | D |
| FR-KEY-004 | Should | Each key row shall show the count of that category's displayed events. | 3 displayed quakes read "〰️ Earthquake 3". | T |
| FR-CNT-001 | Must | The status area shall show the count of events currently displayed with the window it applies to, worded "N events in the last <window>". | 3 displayed events with 24 h selected reads "3 events in the last 24 h". | T (wording) |

#### 3.2.5 Selection and detail

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-SEL-001 | Must | When a marker is activated, the application shall select its event and open the detail panel for it. | D | D |
| FR-SEL-002 | Must | The detail panel shall show title, category, provider, latitude and longitude, event time, retrieved-at time, measurement (when present) and the source link (when present). | Fixture event with every field renders every row; one missing magnitude omits that row. | T |
| FR-GEO-001 | Must | When the pointer hovers a marker, the marker tooltip shall show, beneath the event title, the nearest populated place to the event's marker position, with its country, distance and compass direction, worded "23 km NE of Tromsø, Troms, Norway" (place, region, country). | Fixture event at (69.70, 19.10) reads a distance and direction from Tromsø, Troms, Norway, computed by great-circle distance. | T |
| FR-GEO-002 | Must | Where the event's position lies inside a country's boundary, the place line shall name that country as the event's country, even when the nearest populated place lies across a border. A position on an Antarctic ice shelf lies in Antarctica. | Fixture point just inside one country, nearer a city across the border, names the country it lies in and the city with its own country. Points on the Ross and Ronne ice shelves read as Antarctica, with no "At sea". | T |
| FR-GEO-003 | Must | Where the event's position lies inside no country's boundary and on no Antarctic ice shelf, the place line shall word it as at sea, still naming the nearest populated place with its distance and direction. | Mid-Atlantic fixture reads "At sea; 1,240 km W of ..."; a Weddell Sea point still reads at sea. | T |
| FR-GEO-004 | Must | The application shall resolve places from data embedded in the application, never from a network service (NFR-PRIV-001). | Structural test: the geocoder package imports no network package. | T |
| FR-GEO-005 | Must | Distances shall be rounded to whole kilometres and directions to the eight compass points, so no precision beyond the source's is claimed. | T | T |
| FR-GEO-006 | Must | While the globe's keyboard cursor rests on an event (NFR-KBD-004), the globe view shall show that event's tooltip, place line included, as hover does. | Down arrow onto a fixture event shows its tooltip with the place line. | T |
| FR-GEO-007 | Must | The detail panel shall repeat the place line for the selected event. | T | T |
| FR-GEO-008 | Must | The populated places, country boundaries and Antarctic ice shelves shall come from Natural Earth (public domain, R6), embedded in the application and credited in the third-party notices. | I on `THIRD_PARTY_NOTICES`. | I |
| FR-SEL-003 | Must | The detail panel shall show each time as freshness wording (NFR-FRESH-002), as an exact UTC timestamp and as the same instant in the machine's local time zone. | T | T |
| FR-SEL-004 | Must | Where the event time has day precision (DATA-005), the detail panel shall show the date alone and word freshness in days. | Sea-ice fixture dated 2026-09-18T00:00:00Z on 2026-09-23 reads "Reported for 18 Sep 2026, 5 days ago", never "Observed 5 days 11 h ago". | T |
| FR-SEL-005 | Must | When the source link is activated, the application shall open it in the system browser. | T (fake opener receives the URL) | T |
| FR-SEL-006 | Must | If an event's source URL is not an absolute `https` URL, then the detail panel shall show it as text rather than a link. | `javascript:` and `http:` fixtures render as plain text. | T |
| FR-SEL-009 | Must | Where a provider gives several sources for an event, the detail panel shall link the first that names a page (no file ending; else a web page ending such as `.html` or `.shtml`); if none does, it shall show the first source as text rather than as a link. | Polo's fixture (JTWC `.tcw` then NHC `.shtml`) links the NHC page; a `.tcw` alone renders as text. | T |
| FR-SEL-010 | Must | Where the selected event carries a depth (DATA-010), the detail panel shall show a Depth row: the depth in kilometres to one decimal, then USGS's band for that value (R10): shallow below 70 km, intermediate from 70 to below 300 km, deep from 300 km. | 18.44 reads "18.4 km, shallow"; 70 reads "70.0 km, intermediate"; 583.682 reads "583.7 km, deep"; 69.96 reads "70.0 km, intermediate" (the band follows the value shown). | T |
| FR-SEL-011 | Must | Where the depth is below 0 km (USGS measures depth from the geoid, so a quake beneath high ground can read negative), the Depth row shall read the distance above sea level instead: "1.2 km above sea level, shallow". | -1.2 reads "1.2 km above sea level, shallow"; 0 reads "0.0 km, shallow". | T |
| FR-SEL-012 | Must | Where the depth is exactly 10 km, the Depth row shall add "often a fixed depth: USGS assigns 10 km when it cannot compute one" (R11). | 10 reads "10.0 km, shallow (often a fixed depth: USGS assigns 10 km when it cannot compute one)"; 10.04 reads "10.0 km, shallow" with no note. | T |
| FR-SEL-013 | Must | Where the selected event carries no depth, the detail panel shall show no Depth row. | An EONET event and a USGS event with no third coordinate show none. | T |
| FR-SEL-014 | Must | The guide shall say that USGS assigns 10 km as a fixed depth when it cannot compute one (R11), so a depth of 10 km may not be measured. | The guide names the fixed depth of 10 km. | T |
| FR-SEL-007 | Must | When the detail panel is dismissed, the application shall clear the selection and return focus to the opener (NFR-KBD-006). | T | T |
| FR-SEL-008 | Must | If the selected event leaves the displayed set on refresh or filtering, then the detail panel shall stay open with a line stating the event is no longer in the current view. | T | T |

#### 3.2.6 Status, staleness and errors

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-STS-001 | Must | While a provider's first fetch of the session is in progress with no cached events, the status area shall show that provider as loading. | T | T |
| FR-STS-002 | Must | While a provider is stale, the status area shall show that provider's retrieved-at freshness wording marked as stale. | Fake clock past the threshold; status reads "USGS: retrieved 4 min ago (stale)". | T |
| FR-STS-003 | Must | If a provider's latest fetch failed, then the provider status popover shall show the failure reason in words and the time of the next attempt. | T | T |
| FR-STS-004 | Must | When the application starts with cached events, the globe view shall draw them before the first fetch completes, marked with their retrieved-at freshness. | Start offline with a populated cache; markers appear; status says "retrieved 3 h ago (stale)". | T + D |
| FR-STS-005 | Must | If the cache cannot be opened, then the application shall run with an in-memory store and state in the status popover that events will not survive a restart. | A globe built with no cache states the notice. | T |
| FR-STS-006 | Must | The application shall label no data as "live". | Structural test: the word `live` appears in no user-facing wording table. | T (structural) |
| FR-STS-007 | Must | While a provider's fetch is in progress and its events are held, the status area shall show that provider as refreshing and the Refresh button's icon shall turn. The icon shall complete at least one turn (1 s), so a fetch faster than that is still seen; under the reduced-motion setting it does not turn and the status area alone says so. | Go: a running provider with events held reads refreshing, one loading or idle does not. Vitest: the status line reads "USGS: refreshing"; the button turns while any provider refreshes and for one turn after a fast one. | T |

#### 3.2.7 Settings

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-SET-001 | Must | The settings dialog shall offer auto-rotate on or off. | T | T |
| FR-SET-002 | Should | The settings dialog shall offer the USGS minimum magnitude: all, 1.0, 2.5 (default), 3.0, 4.5. | Changing to 4.5 switches the next fetch to `4.5_week.geojson`. | T |
| FR-SET-003 | Could | The settings dialog shall offer the idle rotation speed from three presets. | T | T |
| FR-SET-004 | Must | If the settings file is missing or unreadable, then the application shall start with defaults and state which file was not read in the status popover. | T | T |

#### 3.2.8 Help surfaces

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-HLP-001 | Must | The About dialog shall show the product name, version from `VERSION`, the copyright notice "© Oliver Ernster 2026", the licence and the data attributions of NFR-LEG-002. | T | T |
| FR-HLP-002 | Must | The licence dialog shall show the full GPL-3.0 text. | T | T |
| FR-HLP-003 | Must | The third-party notices dialog shall show each bundled dependency and data asset with its licence or terms. | Every entry in `THIRD_PARTY_NOTICES` renders. | T |
| FR-HLP-004 | Should | The guide dialog shall name every rail button and every category, each entry led by the control's own picture (the rail icon, the category emoji), then explain the time window, freshness wording and staleness, plus what the cloud layer shows, its veil and that cold ground reads as cloud, plus what the day and night layer shows. | I | I |
| FR-HLP-005 | Must | While the About, licence, notices or guide dialog overflows, its reading body shall read itself using the house auto-scroll cycle (CON-007). | Tick-driven Vitest over the ported state machine; hook test under jsdom. | T |
| FR-HLP-006 | Must | While the detail panel's body overflows, the detail panel shall read itself using the same cycle. | T | T |

#### 3.2.9 Donations (the house `donate` skill; PigeonPost is the reference)

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-DON-001 | Must | The action rail (FR-RAIL-001) shall end in the donate button, pinned to the rail's foot, built as every other rail button is (`RailButton`) with no tray, border or second set of dimensions. | Inspection: no second set of button dimensions in the stylesheet; Vitest reads the button outside the rail's action group with the rail button class. | T + I |
| FR-DON-002 | Must | The donate button shall draw its artwork inside the rail's glyph box (`--rail-glyph-size`), cropped to the artwork and scaled by height by `tools/genicons.py` from `assets/donate.png`. | T (render) | T |
| FR-DON-003 | Must | When the donate button is activated, the application shall hand `https://www.paypal.com/ncp/payment/9LWU8TKV2MSRE` to the system browser through the same external-open facade and scheme allowlist as FR-SEL-005; the page asks the Go side to open it and never holds the address. | Go test asserts the address literally and its https scheme, proved to bite by altering one character; Vitest asserts the page calls the facade. | T |
| FR-DON-004 | Must | The donate address shall have exactly one home: a named constant beside the product identity. | Structural test: the address string appears exactly once across `frontend/src` and `internal` combined. | T |
| FR-DON-005 | Must | The donate button shall carry the tooltip and accessible name "Donate to support EarthNow", as PigeonPost's does. | T | T |
| FR-DON-006 | Must | If opening the donation page is refused (the allowlist or an unconnected backend), then the status area shall say the donation page could not be opened. | T with a refusing fake facade. | T |
| FR-DON-007 | Must | The donate tooltip shall open to the right of the button and upwards, so it is clipped neither at the window's left edge nor at its foot. | D at the minimum window size. | D |
| FR-DON-008 | Must | The donate button shall be the rail's last stop, reached after Help. | Vitest reads it last among the rail's buttons. | T |
| FR-DON-009 | Must | The application shall fetch nothing from the donate address itself; no feature shall depend on a donation. | I | I |
| FR-DON-010 | Should | When the GitHub Pages site is built, its home page shall end with a "Supporting EarthNow" section after the download call to action, with a donate button whose mark is 2.7em high. | The mark's computed height measured at 2.7 times the button's computed font size, at desktop width and at 375 px. The button is taller than a text-only button beside it by design. | D |

#### 3.2.10 Cloud layer (amendment 24)

The source is EUMETSAT's world cloud map, measured on 2026-09-24: WMS layer
`mumi:worldcloudmap_ir108` at `https://view.eumetsat.int/geoserver/wms`, a
global mosaic of the geostationary ring's 10.8 µm infrared channel, extent
-180 to 180 by -90 to 90, a new valid time every 3 h (the newest was 15:00 UTC
at 18:09 UTC). A 2048 x 1024 PNG was 1.6 MB and arrived in 1.1 s. The service
sends no `Last-Modified`; it answers an error with HTTP 200 and an XML
exception report. Warm ground is opaque dark grey in the image, cold cloud
tops white; pixels the satellites do not see are transparent (beyond about 70
to 80 degrees of latitude and a sliver at 180 degrees).

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-CLD-001 | Must | While the cloud layer is hidden, the cloud button shall show the `cloud-cover` artwork alone with the tooltip and accessible name "Show clouds"; while it is shown, the `cloud-cover` artwork with the `negative` overlay and "Hide clouds" (NFR-UX-004). | Given the layer hidden, the button reads "Show clouds"; one press shows the layer and the button reads "Hide clouds". | T |
| FR-CLD-002 | Must | When the cloud button is activated, the globe view shall switch the cloud layer between shown and hidden. | T | T |
| FR-CLD-003 | Must | The settings store shall persist whether the cloud layer is shown; when no saved choice exists, the cloud layer shall start hidden. | A first run starts hidden; showing it, closing and reopening starts shown. | T |
| FR-CLD-004 | Must | While the cloud layer is shown, the cloud provider shall read the newest valid time the service lists for the layer once per cloud interval (60 min). | Fake clock over two intervals with the layer shown: two capabilities requests. | T |
| FR-CLD-016 | Must | When the newest listed valid time differs from the held image's, the cloud provider shall retrieve that time's image as a 2048 x 1024 PNG over (-180, -90, 180, 90) in CRS:84. | Fake fetcher: the GetMap URL carries `time=` the listed default; a check listing the held time fetches no image. | T |
| FR-CLD-005 | Must | While the cloud layer is hidden, the cloud provider shall make no request. | Fake clock over two intervals with the layer hidden: zero requests. | T |
| FR-CLD-006 | Must | The application shall draw each cloud image pixel as white whose opacity follows its infrared brightness: none at or below the clear threshold (65 of 255), full at or above the cloud threshold (90), linear between; both measured by the cloud spike (FR-CLD-015). | Brightness 65 maps to opacity 0, 90 to 1, 77.5 to 0.5. | T |
| FR-CLD-007 | Must | Where the cloud image holds no data (a transparent source pixel), the application shall draw the no-data veil: grey at 20% opacity (a target the spike confirms by eye), so an unseen region never reads as a clear sky. | A transparent source pixel maps to the veil colour and opacity. | T |
| FR-CLD-008 | Must | The globe view shall draw the cloud layer over the globe texture, turning with it, beneath every marker. | Inspection with the layer shown: markers and clusters stay above the cloud. | D |
| FR-CLD-009 | Must | While the cloud layer is shown, the status area shall state the cloud image's valid time in UTC and its age, worded "Clouds: image of 15:00 UTC, 3 h ago" (NFR-FRESH-002), never "live" (FR-STS-006). | Fake clock three hours after a 15:00 image reads exactly that. | T |
| FR-CLD-010 | Should | While the held cloud image's valid time is older than three image intervals (9 h), the status area shall mark it stale. | Fake clock 9 h 1 min after the valid time appends "(stale)". | T |
| FR-CLD-011 | Must | If a cloud fetch fails, then the cloud provider shall keep drawing the held image, retry with FR-PRV-006's backoff and state the reason in the provider status popover under "EUMETSAT". | Fake fetcher failing: the held image stays; the popover names the reason and the next attempt. | T |
| FR-CLD-012 | Must | If the service answers with anything other than a PNG of the requested size (its errors arrive as HTTP 200 carrying XML), then the cloud provider shall treat the answer as a failed fetch (FR-CLD-011). | An XML exception report served with status 200 is refused and the held image kept. | T |
| FR-CLD-013 | Must | If the cloud layer is shown with no image held and the first fetch fails, then the status area shall say the cloud image could not be retrieved, with the reason in the provider status popover. | Fake fetcher failing on a first run: the status line says so; the globe draws no veil and no cloud. | T |
| FR-CLD-014 | Must | The cache shall keep the last good cloud image with its valid time, so the cloud layer draws offline at start, marked with its age (FR-STS-004). | Start offline with a held image: the layer draws and the status line gives its age. | T |
| FR-CLD-015 | Must | Before the cloud layer is built, a cloud spike shall measure, on the reference machine: the two opacity thresholds over the images of one week, chosen so land at night and warm ocean read clear; the veil by eye; the frame time of NFR-PERF-005. | Its measured results recorded in this section, as Phase 0's are in 3.1. | D |

**Cloud spike, thresholds (measured 2026-09-24).** Reference: EUMETSAT's
Meteosat cloud mask (`msg_fes:clm`: cloud, clear land, clear sea), fetched with
the world cloud map for 21 valid times across 17 to 24 September at 8 h steps,
over (-70, -70, 70, 70) at 560 x 560; 3,788,580 cleanly labelled pixels, 68.0%
cloud. The share the mask calls cloud rises from 0.6% at brightness 40 to 49,
through 14.4% at 50 to 59 and 33.4% at 70 to 79, to 95.5% at 100 to 109. The
drafted targets 90 and 200 agreed with the mask on 45.3% of pixels. The pair 65
and 90 agreed on 87.5%, with the drawn opacity a mean 0.052 from the mask's
cloud share at each brightness, the closest of the pairs tried (61/78, 55/95,
55/105, 60/100, 50/110). The reference covers Meteosat's disk alone, in one
September week; the other satellites of the mosaic are assumed to be calibrated
alike (ASM-008).

**Cloud spike, the real window (checked by the owner, 2026-09-24).** In a
packaged build the cloud layer lines up with the texture, turns with the globe
and stays beneath every marker and cluster (FR-CLD-008). The 20% grey veil reads
as unseen rather than as cloud or clear sky, so it stands (FR-CLD-007). Idle
rotation with the layer shown looked smooth; the owner accepted that as
NFR-PERF-005's result for this release. No frame time was recorded, since the
application carries no frame-time log: the figure stays unmeasured.

The infrared channel cannot tell cold ground from cloud: Antarctica,
Greenland and high mountains in winter read as cloud. The guide says so
(FR-HLP-004); no requirement here claims otherwise.

#### 3.2.11 Day and night (amendment 29)

The sun's position is astronomy: a pure function of the UTC instant, so the
layer needs no service and makes no request. The night side shows NASA Earth
Observatory's Black Marble 2016 city lights (D.4). globe.gl's own
`day-night-cycle` example (read in `node_modules/globe.gl/example`) blends a day
and a night texture by the angle to the sun; EarthNow takes that approach, with
the sun's position computed in the domain rather than by a page library and
both textures bundled rather than fetched.

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-DAY-001 | Must | The domain shall compute, for any UTC instant, the subsolar point: its latitude the solar declination and its longitude from the hour angle corrected by the equation of time, by NOAA's solar position equations. | Within 0.1 degrees of NOAA's solar calculator at the four instants recorded in the test (2026's solstices and equinoxes at 12:00 UTC). | T |
| FR-DAY-002 | Must | The domain shall give the light at a point as a factor from 0 (night) to 1 (day) of the sun's elevation there: 0 at or below the twilight limit (6 degrees below the horizon), 1 at or above 6 degrees above it, linear between (a target the real window confirms by eye). | Elevation -6 gives 0, 0 gives 0.5, 6 gives 1. | T |
| FR-DAY-003 | Must | While the day and night layer is shown, the globe view shall draw each point of the globe as the day texture and the night texture mixed by FR-DAY-002's light factor, the sun placed at FR-DAY-001's subsolar point. | Inspection with the layer shown: the lit side faces the subsolar point; the terminator runs through dawn and dusk. | D |
| FR-DAY-004 | Must | While the day and night layer is shown, the globe view shall move the sun at least once a minute (it moves a quarter of a degree a minute). | Fake timer: the sun's position is asked for again within 60 s. | T |
| FR-DAY-005 | Must | While the day and night layer is hidden, the day and night button shall show the `day-night` artwork alone with the tooltip and accessible name "Show day and night"; while it is shown, the `day-night` artwork with the `negative` overlay and "Hide day and night" (NFR-UX-004). | Given the layer hidden, the button reads "Show day and night"; one press shows the layer and the button reads "Hide day and night". | T |
| FR-DAY-006 | Must | When the day and night button is activated, the globe view shall switch the layer between shown and hidden. | T | T |
| FR-DAY-007 | Must | The settings store shall persist whether the day and night layer is shown; when no saved choice exists, the layer shall start shown. | A first run starts shown; hiding it, closing and reopening starts hidden. | T |
| FR-DAY-008 | Must | While the day and night layer is hidden, the globe view shall draw the day texture over the whole globe, as before this amendment. | The light factor handed to the globe is 1 everywhere while hidden. | T |
| FR-DAY-009 | Must | While the day and night layer and the cloud layer are both shown, the globe view shall dim each cloud by its point's light factor, down to a night floor of 25% of its opacity (a target the real window confirms by eye), so no cloud glows white over a dark side. | Light factor 0 gives 25% of the cloud's opacity, 1 gives all of it. | T + D |

#### 3.2.12 Storm trails (amendment 32)

A storm's source gives a fix every few hours (measured: the five storms cached
on 2026-09-23 held 6 to 14 fixes over one to three days). The trail joins the
fixes that fall inside the chosen time window, so it is recent movement and
never an archive. Only severe storms carry one this time: iceberg drift and
earthquake swarms stay Won'ts (3.6).

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-TRL-001 | Must | The domain shall give a severe storm's trail as the positions of its observations inside the time window, oldest first, the newest being its marker position (DATA-003); where fewer than two fall inside, the storm shall have no trail. | A storm with fixes at -30 h, -20 h and -2 h gives two points in the 24 h window, three in 3 days; one fix gives none; an iceberg gives none. | T |
| FR-TRL-002 | Must | While trails are switched on, the globe view shall draw each trail as a line through its points at the marker altitude, fading from faintest at the oldest point to strongest at the marker (a look the real window confirms by eye). | The globe is handed each storm's points with a colour ramp from faint to strong. | T + D |
| FR-TRL-003 | Must | While trails are switched off, the globe view shall draw no trail. | No path data is handed to the globe while off. | T |
| FR-TRL-004 | Must | The settings dialog shall offer "Show storm tracks", applied at once and kept by the settings store (FR-SET-004). | Clearing it hides the tracks; reopening keeps it cleared. | T |
| FR-TRL-005 | Must | When no saved choice exists, including a settings file from before this amendment, trails shall be switched on. | A first run and a 1.0.0 settings file both start with trails on. | T |

### 3.3 Non-functional requirements

Every performance figure is measured on the reference machine.

| ID | Pri | Requirement | Method |
|---|---|---|---|
| NFR-PERF-001 | Must | When launched, the application shall show the globe within 3 s (a target; Phase 0 measures it) at the 95th percentile of 20 cold starts. | Timestamp from process start to first globe frame, logged. Phase 0's three starts stand as the result (amendment 27). |
| NFR-PERF-002 | Must | While idle-rotating with 2,500 markers, the globe view shall hold a median frame time of 16.7 ms or less and a 99th percentile of 33 ms or less (a target; Phase 0 measures it). 2,500 is the measured USGS all-magnitude week (2,126 on 2026-09-23) plus EONET plus headroom. | Frame-time log over 60 s. |
| NFR-PERF-003 | Must | While running for 24 h with default settings, the application's working set shall stay below 500 MB (a target). | Process working set sampled hourly. Accepted by the owner unmeasured (amendment 27). |
| NFR-PERF-004 | Must | When a refresh completes, the globe view shall remain interactive throughout, with no frame over 100 ms attributable to applying the new event set. | Frame-time log across 20 refreshes. |
| NFR-PERF-006 | Must | While idle-rotating with the day and night layer and the cloud layer shown and 2,500 markers, the globe view shall hold NFR-PERF-002's median of 16.7 ms and 99th percentile of 33 ms (a target). | Frame-time log over 60 s with both layers shown; where the application carries none, the owner's judgement by eye, as for NFR-PERF-005 (amendment 27). |
| NFR-PERF-005 | Must | While idle-rotating with the cloud layer shown and 2,500 markers, the globe view shall hold NFR-PERF-002's median of 16.7 ms and 99th percentile of 33 ms (a target; the cloud spike, FR-CLD-015, measures it). | Frame-time log over 60 s with the layer shown. |
| NFR-FRESH-001 | Must | The status model shall mark a provider stale when its last successful retrieval is older than three times its refresh interval. | T |
| NFR-FRESH-002 | Must | The wording component shall render ages as: under 60 s "under a minute ago"; under 60 min "N min ago"; under 48 h "N h ago"; otherwise "N days ago", each rounded down. | T, table-driven. |
| NFR-UX-001 | Must | The globe area (FR-GLB-013) shall occupy at least 70% of the window area at every window size from the minimum size upwards. | T (layout) at three sizes. |
| NFR-UX-002 | Must | The main window shall have a minimum size of 960 by 700 pixels, the height that holds the action rail's ten buttons and the donate button (measured at 600: eight buttons of 54 px with 8 px gaps left 33 px above the donate button; each further button needs 62, so ten need 691; 700 is a target the real window confirms). | I; the rail's buttons all visible at the minimum size (D). |
| NFR-UX-003 | Must | The camera focus animation shall last 1,000 ms. | T (constant) + D |
| NFR-UX-004 | Must | Every two-state toggle button shall show the state it switches TO, never the current state; its tooltip and accessible name shall name that action. This covers the rotation button, the cloud button and the setup program's theme toggle. | T per toggle: the icon and label after a press are the opposite pair. |
| NFR-UX-005 | Must | The main window shall use one dark palette; it shall offer no light theme in V1. The setup program keeps the house light and dark toggle. | I |
| NFR-UX-006 | Must | The main window's heading shall read "EarthNow"; the event count line (FR-CNT-001) carries the sense of now. | T |
| NFR-KBD-001 | Must | The main window shall implement the keeb ring: Tab and Right forward, Shift+Tab and Left back, wrapping at both ends. | Vitest ring walk. |
| NFR-KBD-002 | Must | When the main window opens, the application shall focus a neutral sink so no control shows a ring until the first Tab. | T + D |
| NFR-KBD-003 | Must | When the page's DOM is ready, the host shall focus the WebView2 child directly (falling back to `runtime.Show`); if the page finds it holds no keyboard after the settle time, it shall ask the host again. | Vitest over the settle step; the log records each focus attempt; first Tab steps the ring in the real window with no click (D). |
| NFR-KBD-004 | Must | The globe shall be one canvas stop on the ring: while it holds focus, Up and Down shall walk the displayed events newest first (wrapping) and move the camera to each; Enter or Space shall open the detail panel; plus and minus shall zoom. Arriving on the globe shall move the cursor to an event at once and show its tooltip with a line naming these keys, since the globe paints no ring. | T (cursor walk) + D |
| NFR-KBD-005 | Must | The time window control shall be one stop per option, bounded for Tab and wrapping for Up and Down. | T |
| NFR-KBD-006 | Must | Every dialog and the detail panel shall open focused on its first actionable stop, close on Escape and return focus to its opener. | T per dialog. |
| NFR-KBD-007 | Must | No container, reading body or scroll region shall take focus from a click or paint a ring; a reading body is reachable by Tab only while it overflows and never rings (noborderfocus). | Static CSS scan plus a runtime check, each proved by a planted violation. |
| NFR-KBD-008 | Must | Rings shall follow the three-state model: none at rest, the ring green while an enabled control is hovered or focused, a permanent danger ring while disabled; the accent never rings. | Static CSS scan. |
| NFR-A11Y-001 | Must | Every category shall differ from every other by its emoji, never by colour alone. | I against Appendix D. |
| NFR-A11Y-002 | Must | Text and icon contrast against its surface shall be at least 4.5:1 for body text and 3:1 for glyphs and large text. | Contrast check over the palette tokens. |
| NFR-A11Y-003 | Must | Every icon-only control shall carry a text tooltip and an accessible name. | T |
| NFR-SEC-001 | Must | The frontend shall render provider text as text only. | Structural test: `dangerouslySetInnerHTML` and `innerHTML` appear nowhere in `frontend/src`. |
| NFR-SEC-002 | Must | The page's Content-Security-Policy shall let no fetch directive reach a network origin: `default-src` and `connect-src` are `'self'`, no directive names a network scheme or a wildcard. Images may also come from `data:` and `blob:`, which fetch nothing. | T reading the meta tag in `index.html`. |
| NFR-SEC-003 | Must | The backend shall validate every coordinate (latitude in [-90, 90], longitude in [-180, 180], finite) before an event is stored. | T with out-of-range fixtures. |
| NFR-PRIV-001 | Must | EarthNow's own code shall send no telemetry and contact no host other than the three provider hosts (EONET, USGS, GVP), plus `view.eumetsat.int` while the cloud layer is shown (FR-CLD-005). (What the WebView2 runtime itself contacts is Microsoft's and is not measured here.) | I on the provider registry; T asserting the host allowlist in the HTTP client. |
| NFR-PRIV-002 | Must | The application shall store settings, cache and log under `%LOCALAPPDATA%\EarthNow` only. | T on the path helper. |
| NFR-REL-001 | Must | The application shall not end the run before its window opens for any runtime failure; a failure found at startup is carried into the window as a stated problem. | T on startup paths with injected failures. |
| NFR-REL-002 | Must | The application shall point the standard error handle at the log file as the first act of `main`, so a panic leaves a record. | I + D (planted panic in a debug build). |
| NFR-REL-003 | Must | Every goroutine the application starts shall recover a panic at its top, log it and surface it in the status popover. | T per goroutine entry, planted panic. |
| NFR-REL-004 | Must | Every bound backend call shall take a refusal handler on the frontend, so an unhandled rejection cannot compile. | `tsc --noEmit` in the build. |
| NFR-REL-005 | Must | Cache writes shall be atomic per provider, so an interrupted write leaves the previous event set intact. | T: a save writes a temporary file beside the old one and renames it over, so the old file is untouched until the rename; the test round-trips a save and finds no temporary file left. |
| NFR-OBS-001 | Must | The log shall record, per provider fetch: start, completion or failure, HTTP status, cache status (200, 304), event count and dropped-item count. | T on the log lines with a fake sink. |
| NFR-OBS-002 | Must | The log shall rotate at 5 MB, keeping one previous file. | T |
| NFR-MNT-001 | Must | Coverage over `internal/domain` and `internal/application` shall be 100%; infrastructure floors are set at their measured value when first written and never lowered. | `test.ps1` |
| NFR-MNT-002 | Must | `test.ps1` shall run gofmt, go vet, staticcheck, go test with coverage, then the frontend's eslint, `tsc --noEmit` and Vitest, failing on any non-zero exit. | Run the script against a planted gofmt violation. |
| NFR-MNT-003 | Must | The Go DTOs and the hand-written TypeScript interfaces shall be compared by a structural test. | Planted field rename fails the test. |
| NFR-MNT-004 | Must | The repository shall carry README.md, ARCHITECTURE.md, TESTING.md and DEVELOPMENT.md, each ported in shape from the nearest house reference. | I |
| NFR-LEG-001 | Must | Each bundled third-party component shall appear in `THIRD_PARTY_NOTICES` with its licence and the licence text in full. | `tools/notices.py --check` in `test.ps1`: the file must equal what the shipped Go modules (`go list -deps`) and page packages (`npm ls --omit=dev`) call for. |
| NFR-LEG-002 | Must | The About dialog shall credit "NASA Earth Observatory" for the imagery, NASA EONET and the USGS Earthquake Hazards Program for event data, EUMETSAT for the cloud images ("Cloud images: EUMETSAT, world cloud map (EUMETView).", ASM-007), NASA Earth Observatory for the night lights ("Night lights: NASA Earth Observatory (Black Marble 2016)."), without implying endorsement and without the NASA insignia. | I against R5. |

### 3.4 Data requirements

| ID | Pri | Requirement |
|---|---|---|
| DATA-001 | Must | The domain shall define `Event` with: provider and provider event id (together its id), category, title, description (source text only), status, updated at, source URL, extras (a closed set of named provider details: source category, depth in kilometres, tsunami flag, magnitude type) and its observations, oldest first, every source point retained, each with its time, that time's precision, its position and its measurement (value and unit) where the source gave one. The retrieved-at instant belongs to the provider's snapshot in the store, not to each event. Absent source fields stay absent; none is defaulted to a plausible value. |
| DATA-002 | Must | The category vocabulary shall be: EARTHQUAKE, VOLCANO, WILDFIRE, SEVERE_STORM, FLOOD, LANDSLIDE, DROUGHT, DUST, ICE, OTHER. |
| DATA-003 | Must | An event's marker position shall be its latest source point within the time window; for a polygon, the mean of its outer ring's vertices. A GDACS polygon holding a vertex value beyond 90 is read in the order that value proves (only a longitude exceeds 90); any other GDACS polygon is read in the order the feed's proven GDACS polygons more often show, latitude first when they show none or tie. |
| DATA-004 | Must | Event time shall be, per provider: USGS `properties.time` (ms since epoch, UTC); EONET the date of the latest geometry within the window; GVP the report's issue day (FR-PRV-015). |
| DATA-005 | Must | An EONET event whose every geometry date is exactly 00:00:00Z shall have all its observations marked day precision; an event with any other time of day keeps instant precision throughout. Measured 2026-09-23: every sea-ice date in the 7-day set was 00:00Z; wildfire times carried minutes; Hurricane Polo's 6-hourly track held a genuine 00:00Z fix beside 06:00, 12:00 and 18:00. A single-point event reported at exactly midnight is still read as a date; the error falls on the side of less claimed precision. |
| DATA-012 | Must | A USGS feature shall map to EARTHQUAKE only when its `type` is `earthquake`; any other type (quarry blast, explosion, ice quake) shall map to OTHER with the type kept as the source category. A feature whose status is `deleted` shall not be shown. |
| DATA-006 | Must | EONET categories shall map: earthquakes to EARTHQUAKE, volcanoes to VOLCANO, wildfires to WILDFIRE, severeStorms to SEVERE_STORM, floods to FLOOD, landslides to LANDSLIDE, drought to DROUGHT, dustHaze to DUST, seaLakeIce to ICE; snow, tempExtremes, waterColor, manmade and any unknown id to OTHER, the original id kept in metadata. The 13 source ids were read from `/api/v3/categories` on 2026-09-23. |
| DATA-007 | Must | The mapping shall be data in the EONET adapter, never conditionals in the UI or the domain. |
| DATA-008 | Must | A re-fetched event with an existing id shall replace the stored one; no two stored events shall share an id. |
| DATA-009 | Must | The cache shall hold, per provider: the last successful event set, its retrieved-at instant and its `Last-Modified` value. It shall discard events older than the widest window (7 days). |
| DATA-010 | Must | USGS earthquake depth shall be kept in metadata in kilometres; USGS `tsunami` shall be kept but shall never be worded as "tsunami occurred" (R4: it flags large oceanic events only). |
| DATA-011 | Must | Measurements shall be shown with their source unit as given; no cross-category severity scale shall exist. |

### 3.5 Delivery requirements

| ID | Pri | Release | Requirement |
|---|---|---|---|
| DEL-001 | Must | 1 | `build.ps1` at repo root shall read `VERSION`, run `test.ps1` and throw on failure, build the app, zip it into the setup program's payload, build the setup program with the version in `-ldflags` and write `dist-installer\EarthNowSetup.exe`; `-SkipInstaller` shall stop after the app. Ported from ED Voyage Companion. |
| DEL-002 | Must | 1 | The setup program shall follow the `installer` skill: screen stack (route, uninstall, running, progress, verdict), footer rebuilt per screen, route read once (install, update, downgrade, manage), 126 px mark with no version in the header, three-state ring, per-user install under `%LOCALAPPDATA%\Programs\EarthNow` and `HKCU`, running-app check before any file is touched, fenced extraction, a step log, a verdict at the end of every path. |
| DEL-003 | Must | 1 | The install policy shall live in `internal/infrastructure/setup`; `installer/app.go` shall be a facade owning no install logic. |
| DEL-004 | Must | 1 | One committed `.ico`, generated from the master by `tools/genicons.py`, shall be placed on both executables. |
| DEL-005 | Should | 1 | `build_flatpak.sh` shall build a user Flatpak with application id `uk.codecrafter.EarthNow` on the GNOME runtime (webkit2gtk-4.1), passing `options.Linux` with a GPU policy that enables WebGL (RSK-002), ported from PigeonPost or SymDiary. |
| DEL-006 | Should | 1 | `cleanup_flatpak.sh` shall uninstall the user Flatpak and remove only flatpak artefacts, in the house shape (bash, `set -euo pipefail`, `APP_ID`, `section()` helper, header stating it touches no other build outputs), ported from PigeonPost. |
| DEL-007 | Should | 1 | `builddmg.sh` shall build a signed, notarised arm64 DMG stamped from `VERSION`, ported from SymDiary's port of PigeonPost's `builddmg.sh`; `ALLOW_UNNOTARIZED=1` for local builds only. |

### 3.6 Won't this time (recorded so they are not re-proposed)

Timeline playback; NASA FIRMS; polygons drawn as shapes; iceberg drift and earthquake swarms drawn as trails (amendment 32: no agreed swarm rule yet); the
notifications; favourites and
home location; screenshots and export; an event list or search view (Plan 16
allows deferring it; NFR-KBD-004 makes every event reachable from the keyboard
meanwhile); cross-provider duplicate merging (measured 2026-09-23: EONET holds 15 earthquakes in its whole history, the newest dated 2018-09-28, none in the last 365 days, so an EONET and a USGS marker for the same quake does not arise in practice; revisit if it is ever observed); light theme for the
main window (NFR-UX-005); update check; system tray icon and start with
Windows.

---

## 4. Other requirements

### 4.1 Legal

Covered by CON-004, NFR-LEG-001 and NFR-LEG-002. EarthNow names NASA, the
USGS, the Smithsonian's Global Volcanism Program, EUMETSAT and Natural Earth
only as sources; the product name, icon and site never suggest any of them endorses it.

### 4.2 Internationalisation

V1 is English only. All user-facing text lives in one wording module per
surface so a locale system can be added without touching components; times
are shown in UTC and local time in the detail panel (exact) and as relative
wording elsewhere.

### 4.3 Risk register

A personal desktop utility with no safety consequence; a full FMEA would be
disproportionate. The risks that could stop delivery:

| ID | Risk | Consequence | Response |
|---|---|---|---|
| RSK-001 | Marker count at "all magnitudes" costs frame rate. | Janky globe, the core experience. | FR-SPK-007 measures it in Phase 0; default minimum magnitude 2.5 (362 quakes measured for a week, amendment 11) keeps the default far below the worst case. |
| RSK-002 | WebGL2 missing or disabled under Wails on Linux. | The Flatpak shows no globe. | Closed (amendment 28): the globe drew on the latest Ubuntu LTS. As written before release: measure on real hardware before promising DEL-005; FR-GLB-009 guarantees a stated failure rather than a blank window. |
| RSK-003 | EONET rate limit window unknown. | Throttled refreshes. | ASM-002; the 10 min interval keeps scheduled use at 6 requests per hour. A manual refresh asks every provider and is allowed every 30 s (FR-PRV-010), so a user pressing it at every chance reaches 120 per hour; accepted by the owner. |
| RSK-004 | EONET `Content-Type` mislabels JSON. | A strict parser rejects valid data. | FR-PRV-002. |
| RSK-006 | The infrared cloud image reads cold ground as cloud and holds no data at the poles. | A reader takes snow for cloud or a pole for clear sky. | FR-CLD-007's veil marks the unseen; the guide states the limit (FR-HLP-004); the spike's thresholds (FR-CLD-015) are chosen against a week of images. |
| RSK-005 | Imagery terms change or GPL bundling is challenged. | Re-cut release. | ASM-004; imagery ships as a separate asset with its own notice so it can be swapped. |

---

## Appendix A. Build order (Plan 18, inside-out)

1. **Phase 0 spike**: FR-SPK-001 to 007. Throwaway code allowed; the measured
   numbers are kept and replace the targets they test.
2. **Domain**: `Event`, category vocabulary, time window, freshness
   wording, staleness, clustering maths, ring walk logic. 100% coverage, no I/O.
3. **Application**: use cases, one per user action (select event, set window,
   toggle filter, refresh now, read status), the refresh scheduler on an
   injected clock, the provider port.
4. **Infrastructure**: EONET, USGS and GVP adapters against captured fixtures,
   JSON cache, settings file, log, the HTTP client with host allowlist and size cap.
5. **UI**: globe component, controls, detail panel, dialogs, ring.
6. **Delivery**: `build.ps1`, `test.ps1`, setup program, genicons, the four documents.

The headless check before any UI exists: every use case in step 3 runs from a
Go test with named inputs and asserted outputs.

## Appendix B. Open questions register

There are no open questions. Every question raised in drafting was answered by
the owner on 2026-09-23; each answer now lives in the requirement it settled:

| ID | Settled | Where it lives |
|---|---|---|
| OQ-001 | globe.gl, not CesiumJS | 2.5 |
| OQ-002 | every drafted number accepted; the performance ones are targets Phase 0 measures | FR-GLB-002, 006; FR-PRV-005, 006, 010, 011; NFR-PERF, NFR-FRESH-001, NFR-UX-002, 003; NFR-OBS-002 |
| OQ-003 | default USGS minimum magnitude 2.5 (was 3.0; amendment 11) | FR-PRV-003, FR-SET-002 |
| OQ-004 | idle rotation is not gated on the OS reduced-motion setting | FR-GLB-012 |
| OQ-005 | magnitude bands below 3.0, 3.0 to 4.5, 4.5 to 6, 6 and above | FR-MRK-003 |
| OQ-006 | Blue Marble at 5400 x 2700, reviewed after the spike | D.4 |
| OQ-007 | no duplicate suppression; measured unnecessary | 3.6 |
| OQ-008 | main window dark only; setup program keeps light and dark | NFR-UX-005 |
| OQ-009 | detail panel shows local time beside UTC | FR-SEL-003 |
| OQ-010 | heading reads "EarthNow" | NFR-UX-006 |
| OQ-011 | the donate button at the foot of the action rail (amendment 8) | FR-DON-001 |
| OQ-012 | earthquake emoji 〰️ | D.3 |

The assumptions in 2.6 remain open until confirmed at their stated points; they
are dated checks, not unanswered decisions.

## Appendix C. Traceability

Each requirement ID appears in the name or a comment of the test that verifies
it (`TestFRPRV006_FRPRV007_BackoffDoublesToTheCeilingThenResets`,
`it("FR-TW-002 ...")`). A structural test lists every `Must` ID in this
document and fails when one has no test naming it. Requirements verified by D
or I are listed in TESTING.md's "Checked by a person" table instead.

## Appendix D. Artwork register

Masters are square PNGs on a transparent background at 1254 x 1254 pixels (the
size the SymDiary masters use), except `donate.png` at 1312 x 1199 and the
`negative.png` overlay at 1278 x 1230; `tools/genicons.py` derives every smaller
size.
No artwork may depict the Earth's surface in place of the NASA texture (CON-009).

### D.1 Application identity

| File | Used for | Notes |
|---|---|---|
| `assets/application-icon.png` | the `.ico` on both executables, the setup header mark (256 px), the page's mark and the site's icon (`docs/icon.png`, 208 px, one render); Linux hicolor 16 to 512 (`genicons.py --hicolor`) and the macOS `.icns`, which Wails makes from the master copied to `build/appicon.png` | Must read at 16 px. |
| `assets/light-mode.png` | theme toggle in the setup program (shown while dark, per the installer skill) | Sun. |
| `assets/dark-mode.png` | same, shown while light | Moon. |
| `assets/donate.png` | the donate button (FR-DON) and the site's donate button | Not squared like an icon: `genicons.py` crops to the artwork and scales by height to four times the drawn glyph height, writing every destination in one loop so they cannot drift: `frontend/src/assets/donate.png` for the rail and `docs/donate.png` for the site (FR-DON-010). |

### D.2 Action icons (generated to 208 px, house nav-band style)

| File | Action |
|---|---|
| `refresh.png` | Refresh now |
| `rotate.png` | Shown while rotation is off: pressing it starts rotation |
| `negative.png` | Overlay only. `tools/genicons.py` composites it over `rotate.png` to make `rotate-stop.png`, shown while rotation is on: pressing it stops rotation (NFR-UX-004). The overlay is never shown alone. |
| `reset-view.png` | Reset view |
| `filter.png` | Category and provider filters |
| `time-window.png` | Time window (only if the segmented control carries an icon) |
| `status.png` | Provider status popover |
| `settings.png` | Settings |
| `help-info.png` | About, guide, licences |
| `zoom-in.png`, `zoom-out.png` | Zoom buttons (FR-RAIL-002) |
| `day-night.png` | Day and night button (FR-DAY-005), supplied by the owner: shown alone while the layer is hidden. `tools/genicons.py` composites `negative.png` over it to make `day-night-hide.png`, shown while the layer is shown, the default. |
| `cloud-cover.png` | Cloud button (FR-CLD-001), supplied by the owner: shown alone while the layer is hidden, the default. `tools/genicons.py` composites `negative.png` over it to make `cloud-cover-hide.png`, shown while the layer is shown, as it does for rotation. |

### D.3 Category markers: emoji, not artwork (decided by the owner 2026-09-23)

Each category is drawn with one emoji, held in a single category table in the
frontend (its one home; the filter control and the marker both read it). Every
code point below was confirmed present in `C:\Windows\Fonts\seguiemj.ttf` on
the reference machine (Windows 11) on 2026-09-23. Windows 10 coverage and the
look under Linux and macOS fonts are unmeasured.

| Category | Emoji | Code point |
|---|---|---|
| EARTHQUAKE | 〰️ (a seismograph trace; Unicode has no earthquake emoji) | U+3030 |
| VOLCANO | 🌋 | U+1F30B |
| WILDFIRE | 🔥 | U+1F525 |
| SEVERE_STORM | 🌀 | U+1F300 |
| FLOOD | 🌊 | U+1F30A |
| LANDSLIDE | 🪨 | U+1FAA8 |
| DROUGHT | 🏜️ | U+1F3DC |
| DUST | 💨 | U+1F4A8 |
| ICE | 🧊 | U+1F9CA |
| OTHER | 📍 | U+1F4CD |

An emoji carries its own colour, so there is no per-category tint; categories
are told apart by the emoji itself (NFR-A11Y-001). Emoji are drawn to textures
once per category at start rather than as live page elements, since 2,500 page
elements moved every frame is a frame-rate risk (a hypothesis; FR-SPK-007
measures the chosen form).

### D.4 Not artwork

Night texture: NASA Earth Observatory's Black Marble 2016 (`BlackMarble_2016_3km.jpg`,
13500 x 6750, 8.1 MB, measured by its headers) resampled once to 5400 x 2700 so it
matches the day texture pixel for pixel (both equirectangular from 180 degrees west, checked
side by side), committed as `frontend/src/assets/earth-night.jpg` (1.04 MB) and credited
(NFR-LEG-002).

Earth texture: NASA Blue Marble Next Generation at 5400 x 2700 (about 7.4 km
per pixel at the equator), downloaded and credited (R5); the spike reviews
the sharpness at maximum zoom before the size is fixed. The space behind the globe is plain black (FR-GLB-001); no image is needed.

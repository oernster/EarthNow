# EarthNow: Software Requirements Specification

Changes are made through git; the commit says why.

## 1. Introduction

### 1.1 Purpose

EarthNow is a desktop application showing what is happening on Earth now: a
slowly rotating 3D globe carrying recent natural events, each traceable to its
public source and honest about how fresh it is.

> **EarthNow lets you open a globe and see what is happening on Earth right now.**

That sentence settles any unclear choice.

### 1.2 Scope

In scope, on Windows, macOS and Linux: the globe; three event providers (NASA
EONET, USGS earthquakes, the Smithsonian / USGS weekly volcano report); the
cloud, day and night, storm trail and burnt-area layers; Replay; caching;
filters; the time window; the detail panel; keyboard navigation; help; the
setup program, Flatpak and DMG. Out of scope: the Won't list (3.6).

### 1.3 Definitions

| Term | Meaning |
|---|---|
| Event | One phenomenon reported by one provider, with one marker position. |
| Provider | An adapter mapping one public source to events: `EONET`, `USGS`, `GVP`. |
| Time window | The chosen span ending now: 1 h, 6 h, 24 h, 3 days or 7 days. |
| Event time | The instant used for window membership and freshness (DATA-004). |
| Retrieved at | When EarthNow last had a successful response from a provider. |
| Stale | Last successful retrieval older than the stale threshold (NFR-FRESH-001). |
| Globe area | The window less the action rail and the key. |
| Fit altitude | The camera altitude at which the whole globe fits the globe area (FR-GLB-013). |
| Ring, stop | The keeb keyboard focus cycle; one position on it. |
| Reference machine | The owner's Windows 11 desktop, where performance is judged. |
| Valid time | The instant a cloud image shows; a new one every 3 h. |
| Burnt-area day | One UTC day of GWIS mapping, as layer `nrt.ba` answers for `time=` that day. |
| Span | The chosen window, ending when a replay started (FR-RPL-003). |
| Replay instant | The moment Replay shows (FR-RPL-001). |
| Day precision | Dated by day alone: an EONET event under DATA-005; a volcano dated by its report. |
| Report week | The week a GVP report covers, read from its item titles. |
| Ongoing | Reported as continuing: a volcano in a current GVP report, in progress from its report week's first day until now (FR-PRV-016). |

### 1.4 References

| Ref | Document |
|---|---|
| R2 | NASA EONET API, https://eonet.gsfc.nasa.gov/docs/v3 |
| R3 | USGS GeoJSON feeds, https://earthquake.usgs.gov/earthquakes/feed/v1.0/geojson.php |
| R4 | USGS ComCat fields, https://earthquake.usgs.gov/data/comcat/index.php |
| R5 | NASA media guidelines, https://www.nasa.gov/nasa-brand-center/images-and-media/ |
| R6 | Natural Earth terms, https://www.naturalearthdata.com/about/terms-of-use/ |
| R7 | House skills: `installer`, `scroll`, `keeb`, `noborderfocus` |
| R8 | ED Voyage Companion (Windows delivery reference) |
| R9 | PigeonPost, SymDiary (Flatpak and DMG references) |
| R10 | USGS, Determining the Depth of an Earthquake, https://www.usgs.gov/programs/earthquake-hazards/determining-depth-earthquake |
| R11 | USGS FAQ on 10 km depths, https://www.usgs.gov/faqs/why-are-so-many-earthquakes-located-10-km-deep |
| R12 | GWIS WMS, https://maps.effis.emergency.copernicus.eu/gwis |
| R13 | GWIS licence (CC BY 4.0), https://gwis.jrc.ec.europa.eu/about-gwis/data-license |

---

## 2. Overall description

### 2.1 Product perspective

A standalone desktop application talking to five public keyless services:
EONET, USGS and GVP always; EUMETSAT only for clouds; GWIS only for burnt areas.

```
   NASA EONET ─────┐                         ┌── Globe (WebGL, React)
   USGS feeds ─────┤                         │
   Smithsonian GVP ┼── Go backend ── bound ──┤
   EUMETSAT WMS ───┤   (providers,  methods  └── Controls, detail panel
   GWIS WMS ───────┘    store, cache)
                          │
                 JSON files in the data folder
```

Only the Go backend touches the network (NFR-SEC-002).

### 2.2 Users and environment

Two user classes: a viewer who wants events with no configuration and a
keyboard user who reaches everything without a pointer. There are no roles.

| Item | Windows | macOS and Linux |
|---|---|---|
| OS | Windows 10 and 11, x64 | macOS arm64; Linux Flatpak on the GNOME runtime |
| Web runtime | WebView2 | WKWebView; WebKitGTK |
| GPU | WebGL2 | WebGL2 |
| Network | intermittent; offline survivable | same |
| Install | per user, no administrator rights | DMG; Flatpak user install |

### 2.3 Constraints

| ID | Constraint |
|---|---|
| CON-001 | Go backend; Wails desktop shell; React and TypeScript frontend built by Vite. |
| CON-002 | Architecture `UI → Application → Domain ← Infrastructure` under `internal/`, enforced by `tests/structural`. |
| CON-003 | No CGO in the Windows build (macOS and Linux need it). The cache is one JSON file per provider, written atomically. |
| CON-004 | GPL-3.0 for EarthNow's code; every bundled component compatible and recorded in `THIRD_PARTY_NOTICES`. |
| CON-005 | `VERSION` is the only version literal; `build.ps1` passes it through `-ldflags -X`. |
| CON-006 | `build.ps1` with an unskippable `test.ps1` gate; the setup program is a Wails app under `installer/` built to the `installer` skill. |
| CON-007 | Keyboard navigation follows `keeb` with `noborderfocus`; self-reading surfaces port PigeonPost's `autoScroll.ts` and `useAutoScroll.ts`. |
| CON-008 | No module over 400 lines; one of 381 to 400 lines is cut to 350 or fewer. Build scripts exempt. |
| CON-009 | Earth imagery is real data from NASA or Natural Earth, never generated. |

### 2.4 Assumptions

| ID | Assumption | Status |
|---|---|---|
| ASM-001 | EONET stays keyless at its current URLs. | Holds |
| ASM-002 | EONET's limit (`X-RateLimit-Limit: 60`) allows scheduled plus manual fetches; its window is unknown. | Accepted |
| ASM-003 | USGS feeds keep their URLs. | Holds |
| ASM-004 | NASA Blue Marble may ship with a credit line (R5). | Accepted |
| ASM-005 | WebView2 provides WebGL2. | Confirmed (FR-SPK-002) |
| ASM-006 | `oernster/EarthNow` on GitHub is the release surface. | Confirmed |
| ASM-007 | EUMETSAT allows the cloud map with NFR-LEG-002's credit. | Confirmed |
| ASM-008 | `mumi:worldcloudmap_ir108` keeps its name, extent and 3-hourly times at `https://view.eumetsat.int/geoserver/wms`. | Open |
| ASM-009 | NASA Black Marble 2016 may ship with a credit line (R5). | Confirmed |
| ASM-010 | macOS region is the region of `AppleLocale`; Linux's is the territory of `LC_ALL`, else `LANG`. | Confirmed |
| ASM-011 | NFR-LEG-002's credit meets GWIS's CC BY 4.0 (R13). | Confirmed |
| ASM-012 | `nrt.ba` keeps its name, EPSG:4326 extent and one-day `time` at R12. | Open |

### 2.5 The globe library

globe.gl (MIT, on three.js) with a bundled Blue Marble texture, chosen over
CesiumJS: Cesium is several times the size, defaults to network imagery with an
evaluation token and is a GIS engine. globe.gl has no clustering, so EarthNow
writes its own (FR-MRK-007).

---

## 3. Requirements

Priority is MoSCoW. Verify: T test, D demonstration, I inspection.

### 3.1 Technical spike (closed)

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-SPK-001 | Must | The spike shall show the textured globe on black. | Met. | D |
| FR-SPK-002 | Must | The spike shall log its WebGL version and renderer. | Met: `webgl2`, maximum texture 16,384 px. | D |
| FR-SPK-003 | Must | The spike shall draw a marker at Greenwich (51.4778, -0.0014). | Met. | D |
| FR-SPK-004 | Must | When a marker is clicked, the spike shall log its id. | Met. | D |
| FR-SPK-005 | Must | When a marker is selected, the spike shall centre the camera on it. | Met within 1.2 s. | D |
| FR-SPK-006 | Must | The built spike shall run alone from outside the repository. | Met. | D |
| FR-SPK-007 | Must | The spike shall draw 2,500 markers at NFR-PERF-002's frame rate. | Met: median 10.00 ms, 99th percentile 10.10 ms. | D |

### 3.2 Functional requirements

#### 3.2.1 Globe and camera

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-GLB-001 | Must | The globe view shall render Earth with the day texture on black. | Seen. | D |
| FR-GLB-002 | Must | The globe view shall rotate eastward from launch at one revolution per 240 s; after input stops it, rotation shall resume once no input has arrived for 10 s. | Resumes 10 s after the last input. | T + D |
| FR-GLB-003 | Must | When drag, wheel or key input targets the globe, the globe view shall stop rotating. | Stops on the first input. | T |
| FR-GLB-004 | Must | While auto-rotate is off, the globe shall not rotate on idle. | No rotation after 30 s. | T |
| FR-GLB-005 | Must | When the user drags, the globe shall rotate with the drag. | Seen. | D |
| FR-GLB-006 | Must | When the wheel turns, the globe view shall zoom between altitudes 0.15 and 4.0 globe radii. | Clamped at both. | T + D |
| FR-GLB-007 | Must | When an event is selected, the camera shall centre it over the focus duration at the same altitude. | Arrives centred. | D |
| FR-GLB-008 | Should | When Reset view is activated, the camera shall return to the fit altitude. | Seen. | D |
| FR-GLB-009 | Must | If WebGL2 is unavailable, then the application shall show a message naming it in place of the globe. | Message shown; no exit. | T |
| FR-GLB-010 | Must | The rotation button shall show the state it switches to: "Stop rotating" with the `negative` overlay while rotating, "Start rotating" otherwise. | One press flips it. | T |
| FR-GLB-011 | Must | When the rotation button is activated, auto-rotate shall switch and persist (FR-SET-001). | Button and settings agree. | T |
| FR-GLB-012 | Must | The globe shall rotate on idle whatever the OS reduced-motion setting. | Rotates with animation effects off. | T |
| FR-GLB-013 | Must | When the app starts or the globe area resizes, the fit altitude shall make the globe's diameter 85% to 92% of the area's shorter side. | Fit share 0.9. | T + D |
| FR-GLB-014 | Must | At start the application shall read the OS region as ISO 3166-1 alpha-2: Windows home location; macOS `AppleLocale`; Linux `LC_ALL`, else `LANG`. | `en_GB.UTF-8` gives GB; `POSIX` gives none. | T |
| FR-GLB-015 | Must | When the region has a label point, the globe shall open facing it at the fit altitude before rotating. | GB faces 54.40 N, 2.12 W. | T |
| FR-GLB-016 | Must | If no region or label is found, then the globe shall open at the default view and the log shall say why. | Region 001 sets altitude alone. | T |
| FR-GLB-017 | Must | The label table shall be generated from Natural Earth admin-0 (`ISO_A2_EH` to `LABEL_Y`, `LABEL_X`), preferring the row whose `ISO_A2` matches and then the sovereign row; if a code is still ambiguous, then generation shall fail. | A planted duplicate stops it. | T |
| FR-GLB-018 | Must | When rotation is due to resume with the altitude over half a zoom step from the fit altitude, the camera shall first return to the fit altitude over the focus duration; input during the return shall stop it and restart the idle delay. | Returns over 1,000 ms, then rotates. | T + D |
| FR-GLB-019 | Must | While the idle delay runs with auto-rotate on, a bar along the foot of the globe area shall empty steadily until the delay ends; input shall refill it. | Empties over 10 s; switched on 2.5 s in, over 7.5 s. | T + D |

#### 3.2.2 Markers

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-MRK-001 | Must | The globe shall draw one marker per displayed event at its position (DATA-003). | Fixture lands at its coordinates. | T + D |
| FR-MRK-002 | Must | Each marker shall show its category's emoji (D.3). | Seen. | D |
| FR-MRK-003 | Must | Earthquake markers shall be sized by band: below 3.0, 3.0 to 4.5, 4.5 to 6, 6 and above. | Four sizes. | T |
| FR-MRK-004 | Must | Every other marker shall have one fixed size. | Equal sizes. | T |
| FR-MRK-005 | Must | When a marker is hovered, a tooltip shall show its emoji, title and place line. | Seen. | D |
| FR-MRK-006 | Must | While selected, a marker shall carry a ring in the selection colour. | Seen. | D |
| FR-MRK-007 | Must | Where markers overlap on screen, the globe shall draw one cluster showing the count. | Two events 1 km apart read 2. | T + D |
| FR-MRK-008 | Must | When a cluster is activated, the camera shall zoom until it parts or reaches maximum zoom. | Seen. | D |
| FR-MRK-009 | Must | No marker shall animate beyond selection and hover. | None does. | I |
| FR-MRK-010 | Must | While zooming, every marker shall keep its fit-altitude size on screen. | Scale correct at 0.25, 1 and 3. | T |
| FR-MRK-011 | Must | When a cluster is activated within half a zoom step of the minimum altitude, its members shall be listed instead. | List opens at the minimum. | T |
| FR-MRK-012 | Must | The list shall show each member's emoji, title and freshness; choosing one shall select it. | Selects the chosen one. | T |

#### 3.2.3 Providers and refresh

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-PRV-001 | Must | The EONET provider shall fetch `/api/v3/events?status=all&days=7`; the panel shall say when an event has ended. | URL asserted. | T |
| FR-PRV-002 | Must | The EONET provider shall parse JSON whatever `Content-Type` is declared. | `application/rss+xml` parses. | T |
| FR-PRV-003 | Must | The USGS provider shall fetch the week feed at the highest of all, 1.0, 2.5 and 4.5 not above the minimum magnitude (default 2.5), keeping events at or above it. | URL asserted. | T |
| FR-PRV-004 | Must | The USGS provider shall send `If-Modified-Since` with its last `Last-Modified`. | A 304 keeps events. | T |
| FR-PRV-005 | Must | Each provider shall be fetched on its own interval: USGS 60 s, EONET 10 min, GVP 60 min. | One USGS fetch per 60 s. | T |
| FR-PRV-006 | Must | If a fetch fails, then retries shall double from the interval up to 30 min; a longer interval retries at 30 min. | 60, 120, 240 s then 1,800 s. | T |
| FR-PRV-007 | Must | When a fetch succeeds after failures, the normal interval shall return. | Restored. | T |
| FR-PRV-008 | Must | If one provider fails, then the others' events shall stay. | Unchanged. | T |
| FR-PRV-009 | Must | When Refresh is activated, every provider not already fetching shall fetch. | All fetch. | T |
| FR-PRV-010 | Must | If Refresh is pressed within 30 s of the last manual refresh, then it shall be skipped; after each press the status area shall show the local time of the last manual refresh made. | "Last refreshed at 17:52:10". | T |
| FR-PRV-011 | Must | If a response exceeds 16 MB, then it shall be refused, naming the provider and cap. | Refused. | T |
| FR-PRV-012 | Must | If a response fails to parse, then the failure shall be reported and stored events kept. | Kept. | T |
| FR-PRV-013 | Must | If one item is malformed, then it shall be dropped and counted. | Nine of ten kept, 1 dropped. | T |
| FR-PRV-014 | Must | Providers shall register through one interface; only infrastructure and the composition root name one. | Structural test. | T |
| FR-PRV-015 | Must | The GVP provider shall fetch `https://volcano.si.edu/news/WeeklyVolcanoRSS.xml` hourly as ISO-8859-1, mapping each item to an ongoing volcano at its `georss:point` with report week, issue day and `guid` link; an item with no readable week is dropped. | "31 December-6 January 2027" reads 31 Dec 2026 to 6 Jan 2027. | T |
| FR-PRV-016 | Must | If the GVP report was issued over 14 days ago, then none of its volcanoes shall show and the popover shall say it is too old. | Issued 17 Sep: none on 2 Oct. | T |

#### 3.2.4 Time window, filters, key and rail

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-TW-001 | Must | The time window shall offer 1 h, 6 h, 24 h, 3 days and 7 days. | Five options. | T |
| FR-TW-002 | Must | An event shall show only if its event time is in the window ending now; an ongoing event is in every window. | 1 h at 12:00 excludes 10:30. | T |
| FR-TW-003 | Must | With no saved choice, the window shall be 24 h. | 24 h. | T |
| FR-FLT-001 | Must | The filter shall offer a toggle per category in DATA-002 order, present or not, plus "All events"; the key rows are those toggles. | Seven toggles always. | T |
| FR-FLT-002 | Must | When a category is off, its markers shall be hidden. | Hidden. | T |
| FR-FLT-003 | Must | When "All events" is activated, every category shall switch on. | All on. | T |
| FR-FLT-004 | Must | The filter shall offer a toggle per provider. | USGS off hides its events. | T |
| FR-FLT-005 | Should | The window and filters shall persist across restarts. | Restored. | T |
| FR-KEY-001 | Must | A key down the right side shall list each category in DATA-002 order, emoji then name. | One row each. | T |
| FR-KEY-002 | Must | The key shall read the category table, the markers' single home. | No emoji literal elsewhere. | T |
| FR-KEY-003 | Must | The key shall never overlap the globe. | At the minimum size. | D |
| FR-KEY-004 | Should | Each key row shall show its displayed count. | "〰️ Earthquake 3". | T |
| FR-RAIL-001 | Must | An action rail 68 px wide shall run down the left side, its artwork at 48 px (`--rail-glyph-size`). | Rendered. | T + D |
| FR-RAIL-002 | Must | The rail shall hold, top to bottom: rotation, clouds, day and night, Reset view, zoom in, zoom out, Refresh, status, Settings, Help (guide, About, licence, notices). | Names in order. | T |
| FR-RAIL-003 | Must | Rail tooltips shall open to the right. | At the minimum size. | D |
| FR-CNT-001 | Must | The status area shall read "N events in the last <window>". | "3 events in the last 24 h". | T |

#### 3.2.5 Selection, place and detail

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-SEL-001 | Must | When a marker is activated, its event shall be selected and the detail panel opened. | Seen. | D |
| FR-SEL-002 | Must | The panel shall show title, category, provider, position, event time and retrieved at; also the measurement and source link when present. | Absent fields omitted. | T |
| FR-SEL-003 | Must | The panel shall show each time as freshness, exact UTC and local time. | All three. | T |
| FR-SEL-004 | Must | Where the time has day precision, the panel shall show the date with freshness in days; for an ongoing event, the report week and issue day. | "Continuing: report for 10 to 16 Sep 2026, issued 17 Sep 2026". | T |
| FR-SEL-005 | Must | When the source link is activated, it shall open in the system browser. | Fake opener gets the URL. | T |
| FR-SEL-006 | Must | If a source URL is not absolute `https`, then it shall show as text. | `http:` is text. | T |
| FR-SEL-007 | Must | When the panel closes, the selection shall clear and focus return to the opener. | Focus returns. | T |
| FR-SEL-008 | Must | If the selected event leaves the view, then the panel shall stay open and say so. | Line shown. | T |
| FR-SEL-009 | Must | Where an event has several sources, the panel shall link the first page (no file ending, `.html`, `.shtml` or `.cfm`); if none is a page, then the first shall show as text. | `.tcw` alone is text. | T |
| FR-SEL-010 | Must | Where the event has a depth, the panel shall show it in km to one decimal with USGS's band: shallow below 70, intermediate below 300, deep from 300 (R10). | 69.96 reads "70.0 km, intermediate". | T |
| FR-SEL-011 | Must | Where the depth is below 0, the row shall give the height above sea level. | "1.2 km above sea level, shallow". | T |
| FR-SEL-012 | Must | Where the depth is exactly 10 km, the row shall add "often a fixed depth: USGS assigns 10 km when it cannot compute one" (R11). | 10.04 has no note. | T |
| FR-SEL-013 | Must | Where there is no depth, no Depth row shall show. | None. | T |
| FR-SEL-014 | Must | The guide shall explain the fixed 10 km depth. | Named. | T |
| FR-GEO-001 | Must | The tooltip shall name the nearest populated place with region, country, distance and direction, as "23 km NE of Tromsø, Troms, Norway". | Great-circle distance. | T |
| FR-GEO-002 | Must | Where the position is in a country, that country shall be named even if the place is across a border; an Antarctic ice shelf counts as Antarctica. | Ross shelf reads Antarctica. | T |
| FR-GEO-003 | Must | Where the position is in no country, the line shall read at sea with the nearest place. | "At sea; 1,240 km W of ...". | T |
| FR-GEO-004 | Must | Places shall come from embedded data, never the network. | No network import. | T |
| FR-GEO-005 | Must | Distances shall round to whole km; directions to eight compass points. | Rounded. | T |
| FR-GEO-006 | Must | While the keyboard cursor is on an event, its tooltip shall show. | Shown. | T |
| FR-GEO-007 | Must | The panel shall repeat the place line. | Shown. | T |
| FR-GEO-008 | Must | Places, borders and Antarctic ice shelves shall come from Natural Earth, embedded and credited. | In the notices. | I |

#### 3.2.6 Status and errors

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-STS-001 | Must | While a first fetch runs with nothing cached, the provider shall show as loading. | Loading. | T |
| FR-STS-002 | Must | While a provider is stale, the status area shall mark its age stale. | "USGS: retrieved 4 min ago (stale)". | T |
| FR-STS-003 | Must | If a fetch failed, then the popover shall give the reason and next attempt. | Both shown. | T |
| FR-STS-004 | Must | When the app starts with cached events, they shall draw before the first fetch, with their age. | Draws offline. | T + D |
| FR-STS-005 | Must | If the cache cannot open, then the app shall run in memory and say events will not survive a restart. | Notice shown. | T |
| FR-STS-006 | Must | No data shall be labelled "live". | Not in any wording table. | T |
| FR-STS-007 | Must | While a provider with events fetches, the status area shall say refreshing and the Refresh icon shall turn at least one full turn (1 s), words alone under reduced motion. | "USGS: refreshing". | T |

#### 3.2.7 Settings

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-SET-001 | Must | Settings shall offer auto-rotate on or off. | Offered. | T |
| FR-SET-002 | Should | Settings shall offer the USGS minimum: all, 1.0, 2.5 (default), 3.0, 4.5. | 4.5 fetches `4.5_week.geojson`. | T |
| FR-SET-003 | Could | Settings shall offer idle speeds of 480, 240 and 120 s a revolution. | Offered. | T |
| FR-SET-004 | Must | If the settings file is unreadable, then defaults shall apply and the popover name the file. | Notice shown. | T |

#### 3.2.8 Help

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-HLP-001 | Must | About shall show the name, `VERSION`, "© Oliver Ernster 2026", the licence and NFR-LEG-002's credits. | All shown. | T |
| FR-HLP-002 | Must | The licence dialog shall show the full GPL-3.0. | Shown. | T |
| FR-HLP-003 | Must | The notices dialog shall list every bundled component with its licence. | Every entry renders. | T |
| FR-HLP-004 | Should | The guide shall picture and name every rail button and category; it shall explain the window, freshness, staleness, the cloud veil (cold ground reads as cloud) and day and night. | Read through. | I |
| FR-HLP-005 | Must | While a help dialog overflows, it shall read itself with the house auto-scroll cycle. | State machine tested. | T |
| FR-HLP-006 | Must | While the detail panel overflows, it shall do the same. | Hook tested. | T |

#### 3.2.9 Donations

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-DON-001 | Must | The rail shall end in a donate `RailButton` pinned to its foot, with no tray or own dimensions. | Rail class; no extra sizes. | T + I |
| FR-DON-002 | Must | Its artwork shall sit in the glyph box, cropped and scaled by `tools/genicons.py`. | Rendered. | T |
| FR-DON-003 | Must | When activated, the Go side shall open `https://www.paypal.com/ncp/payment/9LWU8TKV2MSRE` through FR-SEL-005's facade and allowlist. | Address and scheme asserted. | T |
| FR-DON-004 | Must | The address shall have one home beside the product identity. | Appears once. | T |
| FR-DON-005 | Must | Its tooltip and name shall be "Donate to support EarthNow". | Asserted. | T |
| FR-DON-006 | Must | If the page cannot open, then the status area shall say so. | Refusing fake. | T |
| FR-DON-007 | Must | Its tooltip shall open right and up. | At the minimum size. | D |
| FR-DON-008 | Must | It shall be the rail's last stop, after Help. | Last. | T |
| FR-DON-009 | Must | Nothing shall be fetched from the address; no feature shall need a donation. | Inspected. | I |
| FR-DON-010 | Should | The site's home page shall end with "Supporting EarthNow" after the download call, its donate mark 2.7em high. | Measured at desktop and 375 px. | D |

#### 3.2.10 Cloud layer

Source: EUMETSAT WMS `mumi:worldcloudmap_ir108` (ASM-008), a world 10.8 µm
infrared mosaic every 3 h. Errors come as HTTP 200 with XML; unseen pixels are
transparent; cold ground reads as cloud.

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-CLD-001 | Must | The cloud button shall read "Show clouds" while hidden and "Hide clouds" with the `negative` overlay while shown. | Flips on a press. | T |
| FR-CLD-002 | Must | When it is activated, the layer shall switch. | Switched. | T |
| FR-CLD-003 | Must | The layer's state shall persist; with no saved choice it starts hidden. | First run hidden. | T |
| FR-CLD-004 | Must | While shown, the newest valid time shall be read every 60 min. | Two reads in two hours. | T |
| FR-CLD-005 | Must | While hidden, no cloud request shall be made. | Zero requests. | T |
| FR-CLD-006 | Must | Each pixel shall be drawn white at opacity 0 up to brightness 65, 1 from 90, linear between. | 77.5 gives 0.5. | T |
| FR-CLD-007 | Must | Where no satellite sees, a grey veil at 20% shall be drawn. | Transparent maps to veil. | T |
| FR-CLD-008 | Must | The layer shall turn with the globe, over the texture and beneath markers. | Seen. | D |
| FR-CLD-009 | Must | While shown, the status area shall read "Clouds: image of 15:00 UTC, 3 h ago". | Exact. | T |
| FR-CLD-010 | Should | While the image is over 9 h old, it shall be marked stale. | "(stale)" at 9 h 1 min. | T |
| FR-CLD-011 | Must | If a fetch fails, then the held image shall stay, retries follow FR-PRV-006 and the popover gives the reason under "EUMETSAT". | Image kept. | T |
| FR-CLD-012 | Must | If the answer is not a PNG of the asked size, then it shall count as a failure. | XML refused. | T |
| FR-CLD-013 | Must | If no image is held and the first fetch fails, then the status area shall say so. | No veil drawn. | T |
| FR-CLD-014 | Must | The cache shall keep the last good image, so it draws offline. | Draws offline. | T |
| FR-CLD-015 | Must | A spike shall set the thresholds, veil and NFR-PERF-005 before the build. | Done: 65 and 90 matched EUMETSAT's cloud mask on 87.5% of pixels. | D |
| FR-CLD-016 | Must | When the newest valid time changes, a 2048 by 1024 PNG over (-180, -90, 180, 90) in CRS:84 shall be fetched. | Unchanged time fetches nothing. | T |

#### 3.2.11 Day and night

The sun is computed, so the layer makes no request. The night side shows
Black Marble 2016 city lights.

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-DAY-001 | Must | The domain shall compute the subsolar point by NOAA's equations. | Within 0.1 degrees at 2026's solstices and equinoxes. | T |
| FR-DAY-002 | Must | Light shall be 0 at sun elevation -6 degrees or below, 1 at 6 or above, linear between. | 0 gives 0.5. | T |
| FR-DAY-003 | Must | While shown, the day and night textures shall blend by the light. | Lit side faces the sun. | D |
| FR-DAY-004 | Must | While shown, the sun shall move at least once a minute. | Asked again within 60 s. | T |
| FR-DAY-005 | Must | The button shall read "Show day and night" while hidden and "Hide day and night" with the overlay while shown. | Flips on a press. | T |
| FR-DAY-006 | Must | When activated, the layer shall switch. | Switched. | T |
| FR-DAY-007 | Must | The state shall persist; with no saved choice it starts shown. | First run shown. | T |
| FR-DAY-008 | Must | While hidden, the day texture shall cover the globe. | Light 1 everywhere. | T |
| FR-DAY-009 | Must | With clouds also shown, each cloud shall dim by the light down to 25% of its opacity. | Light 0 gives 25%. | T + D |

#### 3.2.12 Storm trails

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-TRL-001 | Must | A severe storm's trail shall be its fixes in the window, oldest first, ending at its marker; fewer than two give none. | Fixes at -30, -20, -2 h: two points in 24 h. | T |
| FR-TRL-002 | Must | While on, each trail shall be a line at marker altitude, faint at the oldest end. | Colour ramp handed over. | T + D |
| FR-TRL-003 | Must | While off, no trail shall be drawn. | No path data. | T |
| FR-TRL-004 | Must | Settings shall offer "Show storm tracks", applied at once and kept. | Kept. | T |
| FR-TRL-005 | Must | With no saved choice, trails shall be on. | Older settings start on. | T |

#### 3.2.13 Burnt areas

Source: GWIS WMS `nrt.ba` (R12, ASM-012), CC BY 4.0. One UTC day per request;
each day is its own mapping, never a running total; burnt ground is red
(255, 0, 0). Spike: 2048 px chosen over 4096 (about 37.5 KB against 76 KB a day).

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-BA-001 | Must | A window's days shall be every UTC day beginning before now and ending after the window's start. | 7 days at noon gives eight days. | T |
| FR-BA-002 | Must | While shown, each missing day shall be fetched as a 2048 by 1024 PNG, one request per day. | No range requested. | T |
| FR-BA-003 | Must | While shown, every day shall be fetched again every 60 min. | Each day twice in two hours. | T |
| FR-BA-004 | Must | When the window changes, the new missing days shall be fetched. | 24 h to 3 days fetches two. | T |
| FR-BA-005 | Must | While hidden, no request shall be made. | Zero. | T |
| FR-BA-006 | Must | While shown, the window's days shall draw as a union at each pixel's highest opacity, in red. | 0.4 and 0.9 draw 0.9. | T + D |
| FR-BA-007 | Must | The layer shall turn with the globe, beneath clouds and markers. | Seen. | D |
| FR-BA-008 | Must | While shown, the status area shall read "Burnt areas: 17 to 23 Sep (UTC), retrieved 12 min ago". | Exact. | T |
| FR-BA-009 | Must | If no day draws anything, then it shall read "Burnt areas: none mapped yet for this window". | Exact. | T |
| FR-BA-010 | Must | Settings shall offer "Show burnt areas", applied at once and kept. | Kept. | T |
| FR-BA-011 | Must | With no saved choice, the layer shall start shown. | First run shown. | T |
| FR-BA-012 | Must | If a day fails, then its held image and the other days shall stay, retries follow FR-PRV-006 and the popover names the day under "GWIS". | Other days draw. | T |
| FR-BA-013 | Must | If the answer is not a PNG of the asked size, then it shall count as a failure. | Empty body refused. | T |
| FR-BA-014 | Must | The cache shall keep each day's last image and drop days older than 7 days. | Draws offline. | T |
| FR-BA-015 | Must | A spike shall measure size, fetch time, look and NFR-PERF-007 before the build. | Done; whether a finished day changes was waived. | D |
| FR-BA-016 | Must | The guide shall say GWIS maps one UTC day at a time, shows a burn on the day mapped, may miss small burns and may show none yet. | All four. | T |
| FR-BA-017 | Must | If nothing is held and a fetch failed, then the status area shall read "Burnt areas: the maps could not be retrieved". | Exact. | T |

#### 3.2.14 Replay

Replay plays the chosen window from its start using held events (DATA-009),
the computed sun, the burnt-area days and cloud images fetched every 3 h at
1024 by 512 into memory. It shows what sources hold now about the span.

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| FR-RPL-001 | Must | The replay instant at position 0 to 1 shall be span start plus that share, clamped. | 0.5 of 7 days is the midpoint. | T |
| FR-RPL-002 | Must | While not replaying, the ordinary view shall show with the scrubber at its right end; while replaying, that end is the span's end. | A seek to the end stays put. | T |
| FR-RPL-003 | Must | When a replay starts by Play or the scrubber, the span shall be fixed until Now or another window. | Still ends at 12:00 at 12:05. | T |
| FR-RPL-004 | Must | When Play is pressed, the position shall run to the end over the rest of one pass, from 0 if not replaying or at the end. | 0.5 at 15 s at normal speed. | T |
| FR-RPL-005 | Must | When play reaches the end, it shall pause there, still replaying. | Paused at the end. | T |
| FR-RPL-006 | Must | When Pause is pressed, the position shall hold. | Holds. | T |
| FR-RPL-007 | Must | When the scrubber moves, play shall pause at its position. | Paused there. | T |
| FR-RPL-008 | Must | The button shall read "Pause replay" with `pause` while playing, else "Play replay" with `play`. | Flips on a press. | T |
| FR-RPL-009 | Must | While replaying, an event shall show only if it has an observation from span start to the replay instant, at its latest; an ongoing one from its report week's start. | A 10:00 quake is absent at 09:59. | T |
| FR-RPL-010 | Must | While replaying, a trail shall join fixes from span start to the instant. | One fix gives none. | T |
| FR-RPL-011 | Must | While replaying, the count shall read "N events up to <instant> UTC in the last <window>". | Exact. | T |
| FR-RPL-012 | Must | While replaying, the sun shall sit at the replay instant. | Asserted. | T |
| FR-RPL-013 | Must | While replaying, burnt areas shall draw from span start to the instant's day. | 17 to 20 Sep at 20 Sep. | T |
| FR-RPL-014 | Must | While replaying, the latest cloud image at or before the instant shall draw; before the first, none. | 14:59 draws 12:00. | T |
| FR-RPL-015 | Must | While replaying with clouds, every valid time of the span shall be fetched at 1024 by 512, oldest first, never cached. | Eight for 24 h. | T |
| FR-RPL-016 | Must | While images arrive, the control shall read "Clouds 12 of 56"; play shall not wait. | Counts up. | T |
| FR-RPL-017 | Must | If an image fails, then the latest earlier one shall draw and the popover list the gaps. | 16:00 draws 12:00. | T |
| FR-RPL-018 | Must | When the replay ends, its images shall be released. | None held. | T |
| FR-RPL-019 | Must | While replaying, the status area shall read "Replay: 20 Sep 14:00 UTC". | Exact. | T |
| FR-RPL-020 | Must | The top bar shall hold Play/Pause, scrubber, speed and Now (while replaying) after the window, taking no globe height. | One row at 960 by 700. | T + D |
| FR-RPL-021 | Must | When the window changes, the replay shall end. | Ordinary view. | T |
| FR-RPL-022 | Must | While replaying, refresh shall continue; events after the span's end shall wait until it ends. | Absent from replay frames. | T |
| FR-RPL-023 | Must | The guide shall say Replay shows what sources hold now, with softer clouds. | Both named. | T |
| FR-RPL-024 | Must | A Now button ("Back to now") shall show only while replaying, last in the row; it shall end the replay. | No replay, no button. | T |
| FR-RPL-025 | Must | A speed button shall cycle half, normal and double (passes of 60, 30 and 15 s), kept in settings, normal by default, named with the current and next speed. | "Replay speed 1x; press for 2x". | T |

### 3.3 Non-functional requirements

Frame rates are judged by eye, since the application keeps no frame-time log.

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| NFR-PERF-001 | Must | The globe shall show within 3 s of launch. | 627 to 654 ms. | D |
| NFR-PERF-002 | Must | Idle rotation with 2,500 markers shall hold a median frame of 16.7 ms and a 99th percentile of 33 ms. | FR-SPK-007. | D |
| NFR-PERF-003 | Must | 24 h with defaults shall stay below a 500 MB working set. | Accepted unmeasured. | D |
| NFR-PERF-004 | Must | Applying a refresh shall cause no frame over 100 ms. | No stall. | D |
| NFR-PERF-005 | Must | The cloud layer shall keep NFR-PERF-002's frame times. | Smooth. | D |
| NFR-PERF-006 | Must | Clouds with day and night shall keep them. | Smooth. | D |
| NFR-PERF-007 | Must | Every layer with 7 days of burnt areas shall keep them. | Smooth. | D |
| NFR-PERF-008 | Must | A 7 day replay with every layer shall keep them. | Smooth. | D |
| NFR-FRESH-001 | Must | A provider shall be stale when its last success is older than three intervals. | Table-driven. | T |
| NFR-FRESH-002 | Must | Ages shall read "under a minute ago", "N min ago" below 60 min, "N h ago" below 48 h, else "N days ago", rounded down. | Table-driven. | T |
| NFR-UX-001 | Must | The globe area shall be at least 70% of the window from the minimum size up. | Three sizes. | T |
| NFR-UX-002 | Must | The minimum window shall be 960 by 700, holding every rail button. | All visible. | I + D |
| NFR-UX-003 | Must | The camera focus animation shall last 1,000 ms. | Constant. | T + D |
| NFR-UX-004 | Must | Every toggle (rotation, clouds, day and night, Play/Pause, the setup theme) shall show the state it switches to. | Flips on a press. | T |
| NFR-UX-005 | Must | The main window shall be dark only; the setup program keeps light and dark. | Inspected. | I |
| NFR-UX-006 | Must | The heading shall read "EarthNow". | Asserted. | T |
| NFR-KBD-001 | Must | The ring shall run Tab and Right forward, Shift+Tab and Left back, wrapping. | Ring walk. | T |
| NFR-KBD-002 | Must | When the window opens, focus shall rest on a neutral sink. | No ring at open. | T + D |
| NFR-KBD-003 | Must | When the DOM is ready, the host shall focus the WebView2 child; a page without the keyboard after the settle time shall ask again. | First Tab works with no click. | D |
| NFR-KBD-004 | Must | The globe shall be one stop: Up and Down walk events newest first; Enter or Space open the panel; plus and minus zoom; arriving shows a tooltip naming the keys. | Cursor walk. | T + D |
| NFR-KBD-005 | Must | The time window shall be one stop per option, wrapping for Up and Down. | Asserted. | T |
| NFR-KBD-006 | Must | Every dialog shall open on its first stop, close on Escape and return focus. | Per dialog. | T |
| NFR-KBD-007 | Must | No container shall take click focus or paint a ring; a reading body is a stop only while it overflows. | Planted violations caught. | T |
| NFR-KBD-008 | Must | Rings shall be none at rest, green on hover or focus, danger while disabled; the accent never rings. | CSS scan. | T |
| NFR-KBD-009 | Must | Replay's controls shall be stops after the window; on the scrubber arrows step a hundredth, Home and End reach the ends, Space plays or pauses. | Keys asserted. | T |
| NFR-A11Y-001 | Must | Categories shall differ by emoji, never colour alone. | Against D.3. | I |
| NFR-A11Y-002 | Must | Contrast shall be 4.5:1 for text and 3:1 for glyphs and large text. | Tokens checked. | T |
| NFR-A11Y-003 | Must | Every icon-only control shall have a tooltip and accessible name. | Asserted. | T |
| NFR-SEC-001 | Must | Provider text shall render as text only. | No `innerHTML` in `frontend/src`. | T |
| NFR-SEC-002 | Must | The CSP shall let nothing reach a network origin: `default-src` and `connect-src` `'self'`; images also `data:` and `blob:`. | Meta tag read. | T |
| NFR-SEC-003 | Must | Every coordinate shall be finite and in range before storing. | Out-of-range refused. | T |
| NFR-PRIV-001 | Must | EarthNow shall send no telemetry and contact only the EONET, USGS and GVP hosts, `view.eumetsat.int` while clouds show and `maps.effis.emergency.copernicus.eu` while burnt areas show. | Allowlist asserted. | T + I |
| NFR-PRIV-002 | Must | Settings, cache and log shall live only in an `EarthNow` folder under the OS user cache folder (`%LOCALAPPDATA%` on Windows). | Path asserted. | T |
| NFR-REL-001 | Must | No failure shall end the run before the window opens; startup failures show in it. | Injected failures. | T |
| NFR-REL-002 | Must | `main` shall first point standard error at the log. | Planted panic. | I + D |
| NFR-REL-003 | Must | Every goroutine shall recover, log and surface a panic. | Entries inspected. | I |
| NFR-REL-004 | Must | Every bound call shall take a refusal handler. | `tsc --noEmit`. | T |
| NFR-REL-005 | Must | Cache writes shall be atomic per provider. | Temp file renamed over. | T |
| NFR-OBS-001 | Must | The log shall record each fetch's start, outcome, HTTP and cache status, event and dropped counts. | Fake sink. | T |
| NFR-OBS-002 | Must | The log shall rotate at 5 MB, keeping one old file. | Asserted. | T |
| NFR-MNT-001 | Must | `internal/domain` and `internal/application` shall have 100% coverage; other floors never fall. | `test.ps1`. | T |
| NFR-MNT-002 | Must | `test.ps1` shall run gofmt, vet, staticcheck, go test, eslint, `tsc --noEmit` and Vitest, failing on any error. | Planted gofmt fault. | T |
| NFR-MNT-003 | Must | A structural test shall match the Go DTOs to the TypeScript types. | Planted rename. | T |
| NFR-MNT-004 | Must | The repo shall carry README.md, ARCHITECTURE.md, TESTING.md and DEVELOPMENT.md. | Present. | I |
| NFR-LEG-001 | Must | `THIRD_PARTY_NOTICES` shall list every bundled component with its full licence. | `tools/notices.py --check`. | T |
| NFR-LEG-002 | Must | About shall credit NASA Earth Observatory (Blue Marble, Black Marble 2016), EONET, USGS, the Smithsonian's Global Volcanism Program, Natural Earth, "Cloud images: EUMETSAT, world cloud map (EUMETView)." and "Burnt areas: © European Union, Global Wildfire Information System (GWIS), Copernicus Emergency Management Service, CC BY 4.0; redrawn over the globe." with a line that none endorses EarthNow; no NASA insignia. | Against R5. | I |

### 3.4 Data requirements

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| DATA-001 | Must | `Event` shall hold provider and id, category, title, source-text description, status, updated at, source URL, extras (source category, depth, tsunami flag, magnitude type) and observations oldest first (time, precision, position, measurement); an ongoing event also its report. Retrieved-at belongs to the provider snapshot; absent fields stay absent. | Shape test. | T |
| DATA-002 | Must | Categories shall be EARTHQUAKE, VOLCANO, WILDFIRE, SEVERE_STORM, FLOOD, ICE, OTHER. | Asserted. | T |
| DATA-003 | Must | The marker position shall be the latest point in the window; a polygon's, its outer ring's mean. A GDACS polygon is read in the order a value beyond 90 proves, else the order the feed's proven polygons mostly show, latitude first by default. | Fixtures placed. | T |
| DATA-004 | Must | Event time shall be USGS `properties.time`; EONET's latest geometry in the window; GVP's report week up to now, ordered by issue day. | Per provider. | T |
| DATA-005 | Must | An EONET event whose every date is exactly 00:00:00Z shall have day precision; otherwise instant. | One 00:00Z fix among others stays instant. | T |
| DATA-006 | Must | EONET earthquakes, volcanoes, wildfires, severeStorms, floods and seaLakeIce shall map to their own; every other id to OTHER, kept as source category. | Asserted. | T |
| DATA-007 | Must | The mapping shall be data in the EONET adapter. | Asserted. | T |
| DATA-008 | Must | A re-fetched event shall replace the one with its id. | No duplicates. | T |
| DATA-009 | Must | The cache shall hold each provider's events, retrieved-at and `Last-Modified`, dropping events over 7 days old unless ongoing and current. | Asserted. | T |
| DATA-010 | Must | USGS depth shall be kept in km; `tsunami` kept but never worded as "tsunami occurred" (R4). | Asserted. | T |
| DATA-011 | Must | Measurements shall keep their source unit; no cross-category severity scale. | Asserted. | T |
| DATA-012 | Must | A USGS feature shall be EARTHQUAKE only if its `type` is `earthquake`, else OTHER; a `deleted` one shall not show. | Quarry blast is OTHER. | T |
| DATA-013 | Must | GVP text shall read a letter then "?s" at a word's end as an apostrophe then "s" and "SO?" at a word's start as "SO₂"; other question marks stay. | "Karangetang?s" mended. | T |

### 3.5 Delivery requirements

| ID | Pri | Requirement | Acceptance | Verify |
|---|---|---|---|---|
| DEL-001 | Must | `build.ps1` shall read `VERSION`, gate on `test.ps1`, build the app and setup program and write `dist-installer\EarthNowSetup.exe`; `-SkipInstaller` stops after the app. | Asserted. | T |
| DEL-002 | Must | The setup program shall follow the `installer` skill: screen stack, route read once, 126 px mark with no version, per-user install under `%LOCALAPPDATA%\Programs\EarthNow` and `HKCU`, running-app check, fenced extraction, a step log and a verdict on every path. | Walked through. | D |
| DEL-003 | Must | Install policy shall live in `internal/infrastructure/setup`; `installer/app.go` is a facade. | Structural test. | T |
| DEL-004 | Must | One `.ico` from `tools/genicons.py` shall be on both executables. | Asserted. | T |
| DEL-005 | Should | `build_flatpak.sh` shall build Flatpak `uk.codecrafter.EarthNow` on the GNOME runtime with WebGL enabled. | Globe draws on Ubuntu LTS. | D |
| DEL-006 | Should | `cleanup_flatpak.sh` shall remove only Flatpak artefacts, in the house shape. | Inspected. | I |
| DEL-007 | Should | `builddmg.sh` shall build a signed, notarised arm64 DMG from `VERSION`; `ALLOW_UNNOTARIZED=1` only locally. | Installs. | D |

### 3.6 Won't have

- Accounts, sync, social features or telemetry.
- AI summaries, predictions or forecasts.
- Notifications.
- An archive beyond 7 days; replay beyond the window; replay saved as video.
- NASA FIRMS hotspots.
- Event polygons drawn as areas.
- Wildfire perimeters as shapes (GWIS has no vectors; NIFC is United States only).
- Iceberg drift or earthquake swarms as trails.
- Weather; imagery beyond clouds and burnt areas; aurora.
- A server run by EarthNow.
- GIS tooling.
- Favourites or a saved home location.
- Screenshots and export.
- An event list or search view (NFR-KBD-004 reaches every event).
- Cross-provider duplicate merging (EONET has held no quake since 2018).
- A light theme for the main window.
- An update check, a tray icon or start with Windows.

---

## 4. Other requirements

### 4.1 Legal

CON-004, NFR-LEG-001 and NFR-LEG-002 apply. NASA, the USGS, the Smithsonian,
EUMETSAT, GWIS and Natural Earth are named only as sources, never as endorsers.

### 4.2 Internationalisation

English only. Text lives in one wording module per surface.

### 4.3 Risk register

| ID | Risk | Response |
|---|---|---|
| RSK-001 | All-magnitude marker counts cost frame rate. | Default minimum 2.5 stays far below the 2,500 proved. |
| RSK-002 | No WebGL2 under Wails on Linux. | Closed: GPU policy Always; FR-GLB-009 covers failure. |
| RSK-003 | EONET's rate-limit window is unknown. | 6 scheduled requests an hour; 120 at most with manual refresh; accepted. |
| RSK-004 | EONET mislabels JSON. | FR-PRV-002. |
| RSK-005 | Imagery terms change. | Imagery ships as a swappable asset with its own notice. |
| RSK-006 | Infrared reads cold ground as cloud and misses the poles. | The veil and the guide. |

---

## Appendix B. Open questions

None is open. OQ-003, cited in code, settled the default USGS minimum at 2.5.

## Appendix C. Traceability

Each id is named in the test that verifies it (`TestFRPRV006_...`,
`it("FR-TW-002 ...")`). A structural test fails when a Must verified by T is
named by no test; a Must verified only by D or I is listed in TESTING.md's
"Checked by a person" table.

## Appendix D. Artwork register

Masters are transparent PNGs 1254 px square, except `donate.png` (1312 by 1199)
and `negative.png` (1278 by 1230); `tools/genicons.py` derives every size. No
artwork depicts the Earth's surface (CON-009).

### D.1 Application identity

| File | Used for |
|---|---|
| `application-icon.png` | `.ico`, setup header (256 px), page mark, site icon (208 px), Linux hicolor, macOS `.icns`. Must read at 16 px. |
| `light-mode.png`, `dark-mode.png` | Setup theme toggle. |
| `donate.png` | The rail's donate button, scaled to four times the glyph height. The site's `docs/donate.png` is the shared house mark, not derived. |

### D.2 Action icons (208 px)

| File | Action |
|---|---|
| `rotate.png` | Start rotating; with `negative.png`, `rotate-stop.png`. |
| `negative.png` | Overlay only. |
| `cloud-cover.png` | Show clouds; with the overlay, `cloud-cover-hide.png`. |
| `day-night.png` | Show day and night; with the overlay, `day-night-hide.png`. |
| `reset-view.png`, `zoom-in.png`, `zoom-out.png` | Reset view; zoom. |
| `refresh.png`, `status.png`, `settings.png`, `help-info.png` | Refresh; status; Settings; Help. |
| `play.png`, `pause.png` | Replay. |
| `filter.png`, `time-window.png` | Masters held; unused. |

### D.3 Category emoji

One category table in the frontend is the single home; each emoji is drawn to
a texture once. All are in Windows 11's Segoe UI Emoji.

| Category | Emoji | Code point |
|---|---|---|
| EARTHQUAKE | 〰️ | U+3030 |
| VOLCANO | 🌋 | U+1F30B |
| WILDFIRE | 🔥 | U+1F525 |
| SEVERE_STORM | 🌀 | U+1F300 |
| FLOOD | 🌊 | U+1F30A |
| ICE | 🧊 | U+1F9CA |
| OTHER | 📍 | U+1F4CD |

### D.4 Textures

Day: NASA Blue Marble Next Generation, 5400 by 2700. Night: NASA Black Marble
2016 resampled to match. Space is plain black.

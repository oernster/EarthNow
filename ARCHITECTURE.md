# Architecture

EarthNow answers one sentence: open a globe and see what is happening on Earth
right now. The Go side fetches three public sources of events, keeps their last
good sets and decides what falls inside the chosen time window. While the cloud
layer is shown it also fetches EUMETSAT's cloud image and draws it. The page
draws the globe and reads the answer. This document says how the code is divided, which rules
the tests hold it to and why each design choice was made.

[REQUIREMENTS.md](REQUIREMENTS.md) is the specification; requirement numbers
below (FR-PRV-005 and so on) refer to it.

## Invariants

`UI -> Application -> Domain <- Infrastructure`

Dependencies point inwards. Each rule below is enforced by a test or by a step
of the gate, not by convention. `tests/structural/boundary_test.go` records
that each of its assertions was proved to bite by planting a violation.

| Invariant | Enforced by |
|---|---|
| The domain imports nothing of EarthNow's own outside `internal/domain`. | [`TestDomainHasNoOutwardImports`](tests/structural/boundary_test.go) |
| The domain is pure: it imports none of `net`, `net/http`, `os`, `path/filepath`, `math/rand`, `math/rand/v2`, `database/sql`, `io/ioutil`, `os/exec`, `log`, `log/slog` or `sync`; it calls none of `time.Now`, `time.Since`, `time.Until`, `rand.Intn` or `rand.Float64`. The time arrives through the `Clock` port. | [`TestDomainIsPure`](tests/structural/boundary_test.go) |
| The application imports no infrastructure, no Wails package and not `net/http`. | [`TestApplicationDoesNotImportInfrastructure`](tests/structural/boundary_test.go) |
| Only `main.go` and `app.go` import both the application services and infrastructure. | [`TestCompositionRootIsWhitelisted`](tests/structural/boundary_test.go) |
| A provider adapter is imported only by its own package and the composition root, so adding a provider changes infrastructure and `main.go` alone (FR-PRV-014). | [`TestFRPRV014_OnlyTheCompositionRootNamesAProvider`](tests/structural/boundary_test.go) |
| No Go file, no page source file (`frontend/src`, tests included) and no setup page file (`.html`, `.css`, `.js` in `installer/frontend/dist`) exceeds 400 lines (CON-008). | [`TestCON008_NoFileExceedsTheLineLimit`](tests/structural/boundary_test.go) |
| None of those files sits in the danger band of 381 to 400 lines; one that does is reduced to 350 or fewer. | [`TestCON008_NoFileInTheDangerBand`](tests/structural/boundary_test.go) |
| Every exported Go type carries a doc comment. | [`TestEveryExportedTypeIsDocumented`](tests/structural/boundary_test.go) |
| The product's name is written only in `internal/product/product.go`: no Go string literal outside it, no page source file, `frontend/index.html` and no setup page file spells it. | [`TestTheProductIsNamedOnce`](tests/structural/name_test.go) |
| The donate address appears exactly once in shipped source (`internal` and `frontend/src`, tests aside), in `internal/product/product.go` (FR-DON-004). | [`TestFRDON004_TheDonateAddressHasOneHome`](tests/structural/identity_test.go) |
| Every struct in `internal/application/dto` is paired with an interface in `frontend/src/types.ts` declaring the same JSON field names, in both directions (NFR-MNT-003). | [`TestNFRMNT003_TheWireMatchesOnBothSides`](tests/structural/wire_test.go) |
| `internal/domain` and `internal/application` together are at 100% statement coverage (NFR-MNT-001). | [`test.ps1`](test.ps1) |
| Each infrastructure package holds its measured coverage floor; the page holds its istanbul floors. | [`test.ps1`](test.ps1), [`frontend/vite.config.ts`](frontend/vite.config.ts) |
| `THIRD_PARTY_NOTICES` is exactly what `tools/notices.py` writes from the shipped dependency tree (NFR-LEG-001). | `python tools/notices.py --check`, run by [`test.ps1`](test.ps1) |
| The geocoder imports no network package (`net`, `net/http`, `net/url`, `httpfetch`): places come from the embedded Natural Earth data (FR-GEO-004). | [`TestFRGEO004_TheGeocoderReachesNoNetwork`](tests/structural/rules_test.go) |
| Every category emoji is written only in `frontend/src/categories.ts`, the table the key, markers, clusters and tooltips read (FR-KEY-002). | [`TestFRKEY002_EmojiLiveOnlyInTheCategoryTable`](tests/structural/rules_test.go) |
| No Go string literal and no page source outside a comment uses the word "live" (FR-STS-006). | [`TestFRSTS006_NothingIsLabelledLive`](tests/structural/rules_test.go) |
| The page uses neither `innerHTML` nor `dangerouslySetInnerHTML`, so provider text is rendered as text (NFR-SEC-001). | [`TestNFRSEC001_ProviderTextIsRenderedAsText`](tests/structural/rules_test.go) |
| Every bound call on the page takes a refusal handler as its last argument, so a call without one does not compile (NFR-REL-004). | `tsc --noEmit` over [`frontend/src/api.ts`](frontend/src/api.ts), run by `test.ps1` |
| Every Must in REQUIREMENTS.md is named by the test that verifies it; failing that, it is listed in TESTING.md's "Checked by a person" table. One verified by T is named by a test whatever the table says (Appendix C). | [`TestAppendixC_EveryMustIsNamedByATest`](tests/structural/trace_test.go) |
| The gate runs every check in order, each throwing on failure; the build runs the gate before building anything and places the one icon on both executables (NFR-MNT-002, DEL-001, DEL-004). | [`delivery_test.go`](tests/structural/delivery_test.go) |
| The setup program's facade imports none of the means to install alone (DEL-003). | [`TestDEL003_TheSetupFacadeOwnsNoInstallLogic`](tests/structural/delivery_test.go) |
| The palette meets its contrast, the rings follow their three states, the page's CSP names no network origin and the globe area holds 70% of the minimum window (NFR-A11Y-002, NFR-KBD-008, NFR-SEC-002, NFR-UX-001). | [`page_test.go`](tests/structural/page_test.go) |
| No EONET category id the adapter maps appears in the page, the domain or the application (DATA-007). | [`TestDATA007_TheCategoryMappingIsDataInTheAdapter`](tests/structural/page_test.go) |
| No page source and no application source words the tsunami flag (DATA-010). | [`TestDATA010_TheTsunamiFlagIsNeverWorded`](tests/structural/page_test.go) |

### Held by the code, not yet by a test

These hold in the tree as it stands; no test fails if one is broken.

- Only `internal/infrastructure/httpfetch` and `main.go` import `net/http`; no
  Go file imports `net`.
- `frontend/src` makes no `fetch`, `XMLHttpRequest` or `WebSocket` call. The
  test above holds the Content-Security-Policy that would refuse one; nothing
  scans the source itself.

## Layers

```
main.go, app.go          composition root and the Wails facade the page calls
frontend/src             the page: React, TypeScript, globe.gl
internal/application/
    ports                what the application needs from outside
    services             the use cases: globe, store, scheduler, preferences, clouds
    dto                  the shapes that cross to the page
internal/domain/
    cloud                the cloud layer's opacity ramp, veil and wording
    event                the provider-neutral event, magnitude bands, links
    freshness            age wording and staleness
    window               the time windows and what falls inside one
internal/infrastructure/
    httpfetch            the one way to the network
    providers/eonet      NASA EONET v3
    providers/usgs       USGS earthquake GeoJSON feeds
    providers/gvp        the Weekly Volcanic Activity Report
    clouds               EUMETSAT's world cloud map, drawn for the globe
    cache                each provider's last good set and the cloud image on disk
    settings             settings.json
    geo                  nearest place and country, from embedded data
    runlog               the log, plus crash output pointed at it
    window               handing the WebView2 child the keyboard
    setup                the install policy behind the setup program
internal/product         the product's identity, in one place
installer                the setup program, a facade over setup
tests/structural         the invariants above
tools                    the icon, notices and Natural Earth data generators
docs                     the GitHub Pages site
```

### Domain

Pure Go over values handed in; no clock, no disk, no network.

- `cloud`: the brightness-to-opacity ramp between the thresholds the cloud
  spike measured, 65 and 90 of 255 (FR-CLD-006); the grey veil for a pixel
  with no data (FR-CLD-007); the status wording and the image's staleness
  after three image intervals (FR-CLD-009, FR-CLD-010).
- `event`: `Event` and its observations, the provider and category
  vocabularies (DATA-001, DATA-002), the earthquake size bands of FR-MRK-003
  and the rule that a source link is a page rather than a data file (FR-SEL-009).
- `freshness`: age wording (NFR-FRESH-002, FR-SEL-004) and the rule that a
  provider is stale three intervals after its last success (NFR-FRESH-001).
- `window`: the five windows of FR-TW-001 with 24 h the default; which
  observation of an event falls inside a window (DATA-003, DATA-004).

### Application

The use cases, behind the ports in
[`ports.go`](internal/application/ports/ports.go): `Clock`, `SnapshotCache`,
`SettingsStore`, `Geocoder`, `Provider`, `CloudSource` and `CloudCache`.

- `Globe` refreshes one provider at a time and answers the view for a window
  and a filter: the events shown, the count per category, each provider's
  status and any standing notice.
- `Store` holds each provider's latest set. A failed provider keeps its last
  set; the others are unaffected (FR-PRV-008).
- `Scheduler` decides when each provider is next due, with the backoff that
  doubles up to 30 minutes after a failure (FR-PRV-006) and the 30-second
  cooldown between manual refreshes (FR-PRV-010), answering when the last one
  was made. It also marks which providers are refreshing, fetching with events
  already held (FR-STS-007). It holds no timer: the facade
  asks what is due and when to wake, so every timing rule runs on a fake clock
  in its tests.
- `Preferences` loads, normalises and saves the settings, telling the USGS
  adapter its minimum magnitude and saying in a notice when the settings file
  was missing or unreadable (FR-SET-004).
- `Clouds` runs the cloud layer. Like the scheduler it holds no timer: the
  facade asks whether a check is due, which is never while the layer is
  hidden (FR-CLD-005). A check reads the newest listed valid time once an
  hour and fetches an image only when that time is new (FR-CLD-004,
  FR-CLD-016). A failure keeps the held image and retries on the scheduler's
  own backoff, one shared function (FR-CLD-011). It answers the status line
  and the EUMETSAT entry for the status popover, which travels apart from the
  event providers since the key filters those.

### Infrastructure

Each package implements a port or a piece of setup policy against the real
machine.

- `httpfetch` is the only network client. It makes a GET to an allowed host,
  follows a redirect only to an allowed host, refuses a body over the cap, sends `If-Modified-Since` where a validator is
  held and asks each source for the media types it serves.
- The three providers each own their source's schema; nothing outside the
  package knows it. Each names its one host and its refresh interval: USGS
  every minute, EONET every ten minutes, the volcano report every hour.
- `clouds` reads the layer's own capabilities document (6.4 KB, against 282 KB
  for the whole service) for the newest valid time, then fetches that time's
  2048 by 1024 PNG. It refuses anything that is not a PNG of that size, since
  the service reports errors as XML with status 200 (FR-CLD-012), then draws
  every pixel by the domain's ramp and veil. Measured on the real image: 238 ms,
  1.6 MB in and 652 KB out, once per new image.
- `cache` keeps one JSON file per provider, stamped with a schema version and
  written beside the old one then renamed over it, so an interrupted write
  leaves the previous set whole (NFR-REL-005, CON-003). The cloud image and its
  valid time are kept the same way in `cloud.json`, through the same reader and
  writer (FR-CLD-014).
- `geo` loads the embedded Natural Earth places, country outlines and Antarctic
  ice shelves and words the nearest place, its distance and its direction
  (FR-GEO-001 to 005). A point on an ice shelf lies in Antarctica, since
  Natural Earth draws Antarctica only to its grounded coast.
- `runlog` keeps `Log.txt`, rotating it at 5 MB while running with one
  `Log.previous.txt` beside it. It also points the process's error output at
  the file.
- `window` finds this process's WebView2 child window and focuses it, which a
  click would otherwise have to do.

### UI

The page in `frontend/src` draws the globe through globe.gl, the action rail on
the left, the key on the right, the detail panel, the dialogs and the keyboard
ring. It reaches the Go side through one module, [`api.ts`](frontend/src/api.ts).
It states the wire's shapes in [`types.ts`](frontend/src/types.ts). The facade
in `app.go` owns no rules: it forwards to the application and runs the
background work.

### Outside the layers

`internal/product` holds the name, the file-system slug, the licence line, the
copyright notice, the donate address and the credits. It is a leaf that the
composition root, the setup program, the `runlog` and `setup` packages and the
tests read, so it belongs to no layer.

## Dependency direction

```
            +---------------------------+
   page     |  frontend/src (React)     |
            +-------------+-------------+
                          | bound calls, events
            +-------------v-------------+
   UI       |  app.go facade, main.go   |
            +-------------+-------------+
                          | calls
            +-------------v-------------+
            |        application        |  ports + services + dto
            +------+-------------^------+
        depends on |             | implements
            +------v-----+       |
            |   domain   |       |
            +------------+       |
            +--------------------+------------------+
            |            infrastructure             |
            | httpfetch, providers, clouds, cache,  |
            | settings,                             |
            | geo, runlog, window, setup            |
            +---------------------------------------+
```

## How the globe fills

`main.go` is the composition root (CON-002). In order:

1. **The log first.** `keepLog` opens `%LOCALAPPDATA%\EarthNow\Log.txt`
   through `runlog`, writes the line naming the version and points the run's
   error output at the file, so a panic leaves a record (NFR-REL-002). A log
   that cannot be opened falls back to standard error rather than ending the
   run. During the `wails build` bindings pass (`binding_pass.go`) it uses
   standard error, so a build never writes to the user's log.
2. **The client and the providers.** One `httpfetch` client with a 30-second
   timeout and the 16 MB response cap of FR-PRV-011, allowed the three
   providers' hosts plus EUMETSAT's and no others; then the EONET, USGS and GVP
   adapters over it.
3. **The globe and the cache.** `Globe` over a `Store` and the system clock,
   with the cache under `%LOCALAPPDATA%\EarthNow\cache`. With no data folder
   the run carries on in memory and says so (FR-STS-005).
4. **The settings,** loaded before the first fetch so USGS is asked at the
   saved minimum.
5. **The cached sets,** put back before any fetch so the globe opens on the
   last known events marked with their age (FR-STS-004). Then the cloud layer:
   its held image restored and its shown setting applied, so a hidden layer
   asks nothing (FR-CLD-005, FR-CLD-014).
6. **The help texts:** the About details from `internal/product`, plus
   `LICENSE` and `THIRD_PARTY_NOTICES` embedded from the repository root.
7. **Wails,** with the window at 1280 by 800 and a minimum of 960 by 640, a
   black background and WebView2's data kept in
   `%LOCALAPPDATA%\EarthNow\webview`. On Linux the web view's GPU policy is set
   to Always, since Wails otherwise turns acceleration off and the globe would
   have no WebGL (RSK-002).

When Wails starts, the facade starts two goroutines, each with a recover at
its top that logs the stack and tells the page (NFR-REL-003): one loads the
gazetteer; the other drives the scheduler and the cloud layer. The driver
starts every provider that is due, each fetch in a guarded goroutine of its own, emits
`events-changed` so the page can show them refreshing (FR-STS-007), then sleeps
until the next provider falls due, a fetch finishes or a manual refresh
arrives. Each fetch logs its start and its outcome, with the status, the event
count and the dropped count (NFR-OBS-001), then emits `events-changed` again;
the page answers each by asking for the view. A cloud check runs the same
way and emits `clouds-changed`; the page then asks for the cloud state; it asks
for the image only when the valid time has changed.

When the page's DOM is ready, the facade focuses the WebView2 child directly,
falling back to asking Wails to show the window. The page calls
`TakeKeyboard` again if it finds it holds no keyboard (NFR-KBD-003).

## Data locations

| What | Where |
|---|---|
| Settings | `%LOCALAPPDATA%\EarthNow\settings.json` |
| Cache | `%LOCALAPPDATA%\EarthNow\cache`, one `<provider>.json` per provider plus `cloud.json` |
| Log | `%LOCALAPPDATA%\EarthNow\Log.txt`, rotating at 5 MB to `Log.previous.txt` |
| The window's WebView2 data | `%LOCALAPPDATA%\EarthNow\webview` |
| Installed files | `%LOCALAPPDATA%\Programs\EarthNow`, with `uninstall.exe` beside the application |
| Apps list entry | `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\EarthNow` |
| Shortcuts | the Start Menu Programs folder under `%APPDATA%` and the user's Desktop |
| Setup's step log | `%TEMP%\EarthNowSetup.log` |
| Setup's WebView2 data | `%TEMP%\EarthNowSetup` |

The data folder is `os.UserCacheDir()` joined with the product's slug, which on
Windows is `%LOCALAPPDATA%\EarthNow` (NFR-PRIV-002). On macOS it is
`~/Library/Caches/EarthNow`; inside the Flatpak it is
`~/.var/app/uk.codecrafter.EarthNow/cache/EarthNow`, since Flatpak points
`XDG_CACHE_HOME` into the sandbox.

## The setup program

`installer/` is a second Go `main` package: a Wails application embedding the
built application as `payload.zip`, with a hand-written page in
`installer/frontend/dist` and no front-end build step. It is a facade; the
install policy lives in `internal/infrastructure/setup` (DEL-003).

- **One reading of the machine decides the route.** `DetectState` compares the
  version recorded in the Apps list entry with this one: install where nothing
  is recorded, update over an older version, go back over a newer one, manage
  (Repair or Reinstall) over the same one. Started with `-uninstall`, as the
  Apps list starts it, setup opens on removal.
- **Per user.** Files under `%LOCALAPPDATA%\Programs\EarthNow`; the Apps list
  entry under `HKCU`, with Modify and Repair both reopening setup. Windows never
  asks for administrator rights.
- **A running EarthNow is refused,** before any file is touched, because
  extracting over a locked executable fails part way. The page offers to close
  it.
- **The payload is fenced:** an archive entry whose path would climb out of the
  install folder is refused.
- **A step log** in `%TEMP%\EarthNowSetup.log`, one timestamped line per step,
  kept outside the folders an uninstall removes.
- **Install, update, go back and reinstall are one act,** `write`: extract,
  register, then apply the shortcut boxes. Repair runs the same path leaving the
  shortcuts as they stand.
- **Uninstall** removes the shortcuts and the Apps list entry, removes the data
  folder when asked, then hands the install folder to a hidden command that
  waits about two seconds (`ping -n 3`) so setup's own copy there can exit,
  then deletes the folder.
- **The page names nothing.** The product's name arrives in the state the
  facade sends, so the page carries no copy for a rename to miss.

The auto-scroll machine and the keyboard repair live once, in
`installer/frontend/dist/auto-scroll.js` and `settle-keyboard.js`, because the
setup page can embed only what sits under `installer/`. The application's page
imports the same files; `frontend/vite.config.ts` lets the dev server read
that folder.

## Versioning

`VERSION` holds the only version string (CON-005). `build.ps1` passes it to
both programs through `-ldflags "-X main.appVersion=..."`; `appVersion` is a
`var` in each because `-X` does nothing to a `const`. A binary built without
the flag reports `0.0.0-dev`. The site under `docs/` cannot read `VERSION`, so
`stamp_version.py` writes it between the page's version markers; `build.ps1`
runs it before the gate.

## Decisions

Each row is stated in a code comment or in REQUIREMENTS.md.

| Decision | Chosen | Rejected: what and why |
|---|---|---|
| The globe library | globe.gl on three.js, with the Blue Marble texture bundled (REQUIREMENTS.md 2.5) | CesiumJS: several times the shipped size, default imagery from a network service with an evaluation token and a GIS engine where one calm globe is wanted. |
| Marker size across zoom | Each marker keeps its launch size on screen: it is drawn at the altitude over the fit altitude times its fit-altitude size (`markers.ts`, `rescale`). | A fixed size in globe units: two overlapping markers would grow with the gap between them and never separate (FR-MRK-008). |
| Clustering | Written in `clusters.ts`, pure, from positions, sizes and the scale | A library's: globe.gl and three-globe offer none (Phase 0, their typings checked). |
| Markers | Emoji drawn to sprite textures, one table as their home (Appendix D.3) | 2,500 page elements moved every frame. The spike measured 2,501 sprites at a median frame of 10.00 ms. |
| The page's coverage provider | istanbul (`frontend/vite.config.ts`) | v8: it reported `GlobeView.tsx` at 100% with no test importing it; istanbul read it at 0%. |
| The page's composition root | `App.tsx` and `main.tsx` excluded from the coverage floors, checked by eye | Counting them: they wire the parts together, as `main.go` does on the Go side. |
| Future-dated events | Shown only up to `window.ClockSkew` (15 minutes) ahead of the clock, so a slow machine clock hides nothing new | Showing any future date: GDACS published a flood alert dated days ahead, which is not an event that has happened (TECH_DEBT.md). |
| Third-party notices | Written by `tools/notices.py` from `go list -deps` and `npm ls --omit=dev --all`, with every licence text in full; the gate checks the file is current | Written by hand: a dependency added or bumped without its notice would ship unnoticed (NFR-LEG-001). |
| The cloud image | Drawn in Go by the domain's ramp and handed to the page as a PNG data URL, then laid on a second sphere just above the globe (`cloudLayer.ts`) | Drawing on the page: it would have to fetch the image, which its CSP forbids (NFR-SEC-002). Measured cost in Go: 238 ms once per new image. |
| Cloud times | The layer's own capabilities document, its time dimension's default | The whole service's document: 282 KB against 6.4 KB (measured). |
| The network | One client with a host allowlist and a size cap; the page makes no request | Fetching from the page: the CSP gives it no origin but its own (NFR-SEC-002). |
| What each source is asked for | Each adapter states the media types it accepts | One `Accept` for all: the Smithsonian feed answers 403 to a request asking only for JSON (measured). |
| Refresh intervals | USGS 60 s, matching its measured `max-age=60`; EONET 10 min; GVP hourly | One interval for all: the sources change at very different rates. |
| EONET scope | Every event of the week, open and closed (amendment 11) | Open events only: 17 against 80 that week, dropping most wildfires and floods. |
| Volcanoes | A third provider, the Weekly Volcanic Activity Report (amendment 12) | EONET alone: it tracked no volcano in the 30 days measured while that week's report listed 20. |
| GDACS flood polygons | A ring holding a value beyond 90 is read in the order that value proves; the rest follow the order the feed's proven rings show, latitude first when they show none (`order.go`, amendments 12 and 17) | GeoJSON order: all 14 of the week arrived latitude first; five were dropped and nine drawn in the wrong place. A fixed latitude-first rule: a correction upstream would draw every flood swapped. Deciding by which reading lands on a country: 27 of 57 rings read as land either way and a coastal Kenya flood read as Spain (measured over 30 days). |
| The source link | The first source naming a page; a data file is shown as text (FR-SEL-009) | The first source: a storm's first source was a `.tcw` warning file, which downloaded. |
| The donate address | Held by the Go side, which opens it through the same check as a source link: an https scheme, a host and a page rather than a data file (`SafeURL`, FR-DON-003) | Held by the page: a second home for a rename or a typo to miss. |
| Controls | An action rail down the left, 68 px wide (amendment 6) | Full-width bars: at 960 by 600 the 70% globe area of NFR-UX-001 leaves them 55 px of height, less than one PigeonPost header. |
| The keyboard | The WebView2 child focused directly, with the page asking again through `TakeKeyboard` (amendment 9) | `runtime.Show` alone: it lost a race inside Wails; the log measured the first focus failing. |
| WebView2 data | Inside the data folder | The default: it falls to `%APPDATA%\EarthNow.exe`, outside the one folder NFR-PRIV-002 allows (measured). |
| The log | Rotated while running at 5 MB, one previous file kept | Rotating only at start: EarthNow runs for days. |
| cgo | Pinned off in `build.ps1` for the gate and the build | Left to the machine: the first machine with a C compiler would quietly build a different binary. |
| The gate | Run by `build.ps1` with no switch to skip it | A skip switch: it is used on the day it would have caught something. |
| The icon | One committed `.ico`, made by `tools/genicons.py`, placed on both executables (DEL-004) | Letting Wails derive one: it does so only when the file is absent. |
| Setup program | A Wails application ported from ED Voyage Companion, to the `installer` skill (CON-006, DEL-002) | Designing one afresh. |

## See also

- [TESTING.md](TESTING.md) for the gate, the floors and the checks a person
  makes.
- [DEVELOPMENT.md](DEVELOPMENT.md) for the tools, the build and the release
  steps.

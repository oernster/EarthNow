# Architecture

EarthNow answers one sentence: open a globe and see what is happening on Earth
right now. The Go side fetches three public sources, keeps their last good sets
and decides what falls inside the chosen time window; the page draws the globe
and reads the answer. This document says how the code is divided, which rules
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

### Held by the code, not yet by a test

These hold in the tree as it stands; no test fails if one is broken.

- Only `internal/infrastructure/httpfetch` and `main.go` import `net/http`; no
  Go file imports `net`.
- The page's Content-Security-Policy in `frontend/index.html` limits every
  fetch to `'self'` (NFR-SEC-002); `frontend/src` makes no `fetch`,
  `XMLHttpRequest` or `WebSocket` call.

## Layers

```
main.go, app.go          composition root and the Wails facade the page calls
frontend/src             the page: React, TypeScript, globe.gl
internal/application/
    ports                what the application needs from outside
    services             the use cases: globe, store, scheduler, preferences
    dto                  the shapes that cross to the page
internal/domain/
    event                the provider-neutral event, magnitude bands, links
    freshness            age wording and staleness
    window               the time windows and what falls inside one
internal/infrastructure/
    httpfetch            the one way to the network
    providers/eonet      NASA EONET v3
    providers/usgs       USGS earthquake GeoJSON feeds
    providers/gvp        the Weekly Volcanic Activity Report
    cache                each provider's last good set on disk
    settings             settings.json
    geo                  nearest place and country, from embedded data
    runlog               the log, plus crash output pointed at it
    window               handing the WebView2 child the keyboard
    setup                the install policy behind the setup program
internal/product         the product's identity, in one place
installer                the setup program, a facade over setup
tests/structural         the invariants above
tools                    the icon and notices generators
```

### Domain

Pure Go over values handed in; no clock, no disk, no network.

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
`SettingsStore`, `Geocoder` and `Provider`.

- `Globe` refreshes one provider at a time and answers the view for a window
  and a filter: the events shown, the count per category, each provider's
  status and any standing notice.
- `Store` holds each provider's latest set. A failed provider keeps its last
  set; the others are unaffected (FR-PRV-008).
- `Scheduler` decides when each provider is next due, with the backoff that
  doubles up to 30 minutes after a failure (FR-PRV-006) and the 30-second
  cooldown between manual refreshes (FR-PRV-010). It holds no timer: the facade
  asks what is due and when to wake, so every timing rule runs on a fake clock
  in its tests.
- `Preferences` loads, normalises and saves the settings, telling the USGS
  adapter its minimum magnitude and saying in a notice when the settings file
  was missing or unreadable (FR-SET-004).

### Infrastructure

Each package implements a port or a piece of setup policy against the real
machine.

- `httpfetch` is the only network client. It makes a GET to an allowed host,
  refuses a body over the cap, sends `If-Modified-Since` where a validator is
  held and asks each source for the media types it serves.
- The three providers each own their source's schema; nothing outside the
  package knows it. Each names its one host and its refresh interval: USGS
  every minute, EONET every ten minutes, the volcano report every hour.
- `cache` keeps one JSON file per provider, stamped with a schema version and
  written beside the old one then renamed over it, so an interrupted write
  leaves the previous set whole (NFR-REL-005, CON-003).
- `geo` loads the embedded Natural Earth places and country outlines and words
  the nearest place, its distance and its direction (FR-GEO-001 to 005).
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
composition root, the setup program and the tests read, so it belongs to no
layer.

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
            | httpfetch, providers, cache, settings,|
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
   providers' hosts and no others; then the EONET, USGS and GVP adapters over it.
3. **The globe and the cache.** `Globe` over a `Store` and the system clock,
   with the cache under `%LOCALAPPDATA%\EarthNow\cache`. With no data folder
   the run carries on in memory and says so (FR-STS-005).
4. **The settings,** loaded before the first fetch so USGS is asked at the
   saved minimum.
5. **The cached sets,** put back before any fetch so the globe opens on the
   last known events marked with their age (FR-STS-004).
6. **The help texts:** the About details from `internal/product`, plus
   `LICENSE` and `THIRD_PARTY_NOTICES` embedded from the repository root.
7. **Wails,** with the window at 1280 by 800 and a minimum of 960 by 600, a
   black background and WebView2's data kept in
   `%LOCALAPPDATA%\EarthNow\webview`.

When Wails starts, the facade starts two goroutines, each with a recover at
its top that logs the stack and tells the page (NFR-REL-003): one loads the
gazetteer; the other drives the scheduler. The driver starts every provider
that is due, each fetch in a guarded goroutine of its own, then sleeps until
the next provider falls due, a fetch finishes or a manual refresh arrives.
Each fetch logs its start and its outcome, with the status, the event count
and the dropped count (NFR-OBS-001), then emits `events-changed`; the page
answers by asking for the view.

When the page's DOM is ready, the facade focuses the WebView2 child directly,
falling back to asking Wails to show the window. The page calls
`TakeKeyboard` again if it finds it holds no keyboard (NFR-KBD-003).

## Data locations

| What | Where |
|---|---|
| Settings | `%LOCALAPPDATA%\EarthNow\settings.json` |
| Cache | `%LOCALAPPDATA%\EarthNow\cache`, one `<provider>.json` per provider |
| Log | `%LOCALAPPDATA%\EarthNow\Log.txt`, rotating at 5 MB to `Log.previous.txt` |
| The window's WebView2 data | `%LOCALAPPDATA%\EarthNow\webview` |
| Installed files | `%LOCALAPPDATA%\Programs\EarthNow`, with `uninstall.exe` beside the application |
| Apps list entry | `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\EarthNow` |
| Shortcuts | the Start Menu Programs folder under `%APPDATA%` and the user's Desktop |
| Setup's step log | `%TEMP%\EarthNowSetup.log` |
| Setup's WebView2 data | `%TEMP%\EarthNowSetup` |

The data folder is `os.UserCacheDir()` joined with the product's slug, which on
Windows is `%LOCALAPPDATA%\EarthNow` (NFR-PRIV-002).

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
the flag reports `0.0.0-dev`.

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
| The network | One client with a host allowlist and a size cap; the page makes no request | Fetching from the page: the CSP gives it no origin but its own (NFR-SEC-002). |
| What each source is asked for | Each adapter states the media types it accepts | One `Accept` for all: the Smithsonian feed answers 403 to a request asking only for JSON (measured). |
| Refresh intervals | USGS 60 s, matching its measured `max-age=60`; EONET 10 min; GVP hourly | One interval for all: the sources change at very different rates. |
| EONET scope | Every event of the week, open and closed (amendment 11) | Open events only: 17 against 80 that week, dropping most wildfires and floods. |
| Volcanoes | A third provider, the Weekly Volcanic Activity Report (amendment 12) | EONET alone: it tracked no volcano in the 30 days measured while that week's report listed 20. |
| GDACS flood polygons | Read latitude first (amendment 12) | GeoJSON order: all 14 of the week arrived latitude first; five were dropped and nine drawn in the wrong place. TECH_DEBT.md item 2 keeps it under watch. |
| The source link | The first source naming a page; a data file is shown as text (FR-SEL-009) | The first source: a storm's first source was a `.tcw` warning file, which downloaded. |
| The donate address | Held by the Go side, which opens it through the same https allowlist as a source link (FR-DON-003) | Held by the page: a second home for a rename or a typo to miss. |
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

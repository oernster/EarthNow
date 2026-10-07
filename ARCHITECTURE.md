# Architecture

EarthNow opens a globe and shows what is happening on Earth now. The Go side
fetches three public event sources, keeps each one's last good set and decides
what falls inside the chosen time window. While the cloud or burnt-area layer is
shown it fetches and draws that layer's image; for day and night it works out
where the sun stands from the time alone. The page draws the globe and reads the
answers. Requirement numbers (FR-PRV-005 and so on) refer to
[REQUIREMENTS.md](REQUIREMENTS.md); the conceptual decisions are argued in
[DECISIONS-TRADEOFFS.md](DECISIONS-TRADEOFFS.md).

## Invariants

`UI -> Application -> Domain <- Infrastructure`

Dependencies point inwards. Each rule below is held by a test or a gate step.
`tests/structural/boundary_test.go` records that each of its assertions was
proved to bite by planting a violation.

| Invariant | Enforced by |
|---|---|
| The domain imports nothing of EarthNow's own outside `internal/domain`. | [`TestDomainHasNoOutwardImports`](tests/structural/boundary_test.go) |
| The domain is pure: none of `net`, `net/http`, `os`, `path/filepath`, `math/rand`, `math/rand/v2`, `database/sql`, `io/ioutil`, `os/exec`, `log`, `log/slog` or `sync`; no call to `time.Now`, `time.Since`, `time.Until`, `rand.Intn` or `rand.Float64`. Time arrives through the `Clock` port. | [`TestDomainIsPure`](tests/structural/boundary_test.go) |
| The application imports no infrastructure, no Wails package and not `net/http`. | [`TestApplicationDoesNotImportInfrastructure`](tests/structural/boundary_test.go) |
| Only `main.go` and `app.go` import both the services and infrastructure. | [`TestCompositionRootIsWhitelisted`](tests/structural/boundary_test.go) |
| A provider adapter is imported only by its own package and the composition root (FR-PRV-014). | [`TestFRPRV014_OnlyTheCompositionRootNamesAProvider`](tests/structural/boundary_test.go) |
| No Go file, page source file (`frontend/src`, tests included) or setup page file (`.html`, `.css`, `.js` in `installer/frontend/dist`) exceeds 400 lines (CON-008). | [`TestCON008_NoFileExceedsTheLineLimit`](tests/structural/boundary_test.go) |
| None of those files sits at 381 to 400 lines; one that does is cut to 350 or fewer. | [`TestCON008_NoFileInTheDangerBand`](tests/structural/boundary_test.go) |
| Every exported Go type carries a doc comment. | [`TestEveryExportedTypeIsDocumented`](tests/structural/boundary_test.go) |
| The product's name is written only in `internal/product/product.go`: no other Go string literal, page source file, `frontend/index.html` or setup page file spells it (tests aside). | [`TestTheProductIsNamedOnce`](tests/structural/name_test.go) |
| The donate address appears once in shipped source, in `internal/product/product.go` (FR-DON-004). | [`TestFRDON004_TheDonateAddressHasOneHome`](tests/structural/identity_test.go) |
| Every struct in `internal/application/dto` is paired with an interface in `frontend/src/types.ts` with the same JSON fields, both ways (NFR-MNT-003). | [`TestNFRMNT003_TheWireMatchesOnBothSides`](tests/structural/wire_test.go) |
| `internal/domain` and `internal/application` are gated at 100% statement coverage (NFR-MNT-001). | [`TestNFRMNT001_DomainAndApplicationAreGatedAt100`](tests/structural/delivery_test.go) |
| Each infrastructure package holds its measured floor; the page holds its istanbul floors. | [`test.ps1`](test.ps1), [`frontend/vite.config.ts`](frontend/vite.config.ts) |
| The gate runs gofmt, vet, staticcheck, the Go tests, lint, `tsc --noEmit` (so a bound call without a refusal handler fails to compile, NFR-REL-004), the page tests and `tools/notices.py --check` (NFR-LEG-001) in order, each throwing on failure (NFR-MNT-002). | [`TestNFRMNT002_TheGateRunsEveryCheck`](tests/structural/delivery_test.go) |
| The build reads `VERSION`, runs the gate before building anything (DEL-001) and places one committed icon on both executables (DEL-004). | [`TestDEL001_TheBuildRunsTheGateFirst`](tests/structural/delivery_test.go), [`TestDEL004_OneIconOnBothExecutables`](tests/structural/delivery_test.go) |
| The setup program's facade imports none of the means to install alone (DEL-003). | [`TestDEL003_TheSetupFacadeOwnsNoInstallLogic`](tests/structural/delivery_test.go) |
| The geocoder imports no network package: places come from embedded Natural Earth data (FR-GEO-004). | [`TestFRGEO004_TheGeocoderReachesNoNetwork`](tests/structural/rules_test.go) |
| Every `httpfetch.New` call outside the tests names exactly one host (NFR-PRIV-001). | [`TestNFRPRIV001_EveryClientHoldsOneHost`](tests/structural/client_test.go) |
| Category emoji are written only in `frontend/src/categories.ts` (FR-KEY-002). | [`TestFRKEY002_EmojiLiveOnlyInTheCategoryTable`](tests/structural/rules_test.go) |
| No Go string literal and no page source outside a comment says "live" (FR-STS-006). | [`TestFRSTS006_NothingIsLabelledLive`](tests/structural/rules_test.go) |
| The page uses neither `innerHTML` nor `dangerouslySetInnerHTML` (NFR-SEC-001). | [`TestNFRSEC001_ProviderTextIsRenderedAsText`](tests/structural/rules_test.go) |
| The palette meets its contrast, the rings follow their three states, the CSP names no network origin and the globe area holds 70% of the minimum window (NFR-A11Y-002, NFR-KBD-008, NFR-SEC-002, NFR-UX-001). | [`page_test.go`](tests/structural/page_test.go) |
| No EONET category id the adapter maps appears in the page, the domain or the application (DATA-007). | [`TestDATA007_TheCategoryMappingIsDataInTheAdapter`](tests/structural/page_test.go) |
| No page or application source words the tsunami flag (DATA-010). | [`TestDATA010_TheTsunamiFlagIsNeverWorded`](tests/structural/page_test.go) |
| Every Must is named by its verifying test, else listed in TESTING.md's "Checked by a person" table; one verified by T needs a test regardless (Appendix C). | [`TestAppendixC_EveryMustIsNamedByATest`](tests/structural/trace_test.go) |

### Held by the code, not yet by a test

- Only `internal/infrastructure/httpfetch` and `main.go` import `net/http`; no
  shipped Go file imports `net` (one test, `httpfetch/dial_test.go`, does).
- `frontend/src` makes no `fetch`, `XMLHttpRequest` or `WebSocket` call. The
  CSP test holds the policy that would refuse one; nothing scans the source.

## Layers

### Domain (`internal/domain`)

Pure Go over values handed in.

| Package | Holds |
|---|---|
| `burnt` | The union of a window's days at each pixel's highest opacity (FR-BA-006); the status wording (FR-BA-008, FR-BA-009). |
| `cloud` | The brightness ramp between `ClearThreshold` 65 and `CloudThreshold` 90 (FR-CLD-006); the veil for no data (FR-CLD-007); wording and staleness after three `ImageInterval`s of 3 h (FR-CLD-009, FR-CLD-010). |
| `event` | `Event`, the provider and category vocabularies (DATA-001, DATA-002), quake size bands (FR-MRK-003), `MeetsMinimum`, `CheckUsable` (FR-PRV-012), the source-link rule (FR-SEL-009), an ongoing event's `Report`, current for `ReportCurrency` of 14 days (FR-PRV-015, FR-PRV-016); depth wording, with bands at 70 and 300 km and 10 km marked as USGS's fixed depth (FR-SEL-010 to 012). |
| `freshness` | Age wording (NFR-FRESH-002, FR-SEL-004); stale after `StaleAfterIntervals` of 3 (NFR-FRESH-001). |
| `region` | A region code or locale name's territory as ISO alpha-2 (FR-GLB-014). |
| `sun` | The subsolar point by NOAA's equations (FR-DAY-001); light from 0 to 1 across `TwilightDegrees` 6 either side of the horizon (FR-DAY-002); `CloudNightFloor` 0.25 (FR-DAY-009). |
| `window` | The five windows, 24 h by default (FR-TW-001); `Range`, the one rule for what a window or a replay shows (DATA-003, DATA-004, FR-RPL-009); `Days` (FR-BA-001); `ClockSkew` of 15 minutes; storm trails (FR-TRL-001). |

### Application (`internal/application`)

`ports` declares `Clock`, `SnapshotCache`, `SettingsStore`, `Geocoder`,
`Provider`, `CloudSource`, `CloudCache`, `BurntSource`, `BurntCache`,
`RegionSource` and `LabelPoints` in
[`ports.go`](internal/application/ports/ports.go); `dto` holds the shapes that
cross to the page. The services hold no timer: the facade asks what is due, so
every timing rule runs on a fake clock in tests.

| Service | Does |
|---|---|
| `Globe` | Refreshes one provider at a time; answers the view for a window and filter: events, counts, each provider's status and any notice. |
| `Store` | Holds each provider's latest set; a failed provider keeps its last (FR-PRV-008). Hides a stale ongoing report (FR-PRV-016) and quakes below the minimum (FR-SET-002), filtering what is shown, never what is held. |
| `Scheduler` | When each provider is due; backoff doubling to `BackoffCeiling` of 30 minutes (FR-PRV-006); `ManualCooldown` of 30 s (FR-PRV-010); which providers are refreshing (FR-STS-007). |
| `Preferences` | Loads, normalises and saves settings; tells USGS and the globe the minimum magnitude; notices a missing or unreadable file (FR-SET-004); offers `ReplaySpeeds` (FR-RPL-025). |
| `Clouds` | Checks hourly (`CloudInterval`) for a new valid time, never while hidden (FR-CLD-004, FR-CLD-005, FR-CLD-016); keeps the held image on failure with the scheduler's backoff (FR-CLD-011). |
| `BurntAreas` | Fetches each day the window touches hourly (`BurntInterval`), only missing days when it widens; never while hidden (FR-BA-002 to 005, FR-BA-012). |
| `Replay` | Answers one replay frame: events up to the instant, the sun, the cloud image and burnt days to draw. |
| `ReplayClouds` | Fetches a replay's cloud images at half size, in memory only (FR-RPL-015 to 018). |
| `Sun` | The subsolar point with the twilight limit and night floor, so the page holds no figure (FR-DAY-001, FR-DAY-004). |
| `StartView` | The label point of the country the region setting names, else nothing (FR-GLB-015, FR-GLB-016). |

### Infrastructure (`internal/infrastructure`)

| Package | Does |
|---|---|
| `httpfetch` | The only network client: GET over https to allowed hosts, redirects only to https allowed hosts, a body cap, `If-Modified-Since` and per-source `Accept`. Go's transport honours `HTTPS_PROXY`. |
| `providers/eonet`, `usgs`, `gvp` | Each owns its source's schema, host and `Interval`: USGS a minute, EONET ten minutes, GVP an hour. Each applies `event.CheckUsable` (FR-PRV-012). |
| `clouds` | Reads the layer's capabilities document for the newest time, fetches a 2048 by 1024 PNG and draws it by the domain's ramp; replay asks for half that size. |
| `gwis` | One UTC day's 2048 by 1024 PNG per request; composes a window's days. |
| `pngcheck` | Refuses a map answer that is not a PNG of the size asked (FR-CLD-012, FR-BA-013). |
| `cache` | One JSON file per provider plus `cloud.json` and `burnt.json`, schema-stamped, written then renamed (NFR-REL-005, CON-003); holds only what the widest window can show (DATA-009). |
| `settings` | `settings.json` with a capped read and write-then-rename. |
| `oslocale` | The OS region: home location on Windows, `AppleLocale` on macOS, `LC_ALL` else `LANG` on Linux (FR-GLB-014). |
| `geo` | Nearest place, country and label points from embedded Natural Earth data (FR-GEO-001 to 005). |
| `runlog` | `Log.txt`, rotated at 5 MB to `Log.previous.txt`; error output pointed at it. |
| `window` | Focuses this process's WebView2 child. |
| `setup` | The install policy behind the setup program. |

### UI and the rest

| Path | Role |
|---|---|
| `main.go`, `app.go` | Composition root (CON-002) and the Wails facade; `layers.go` and `replay.go` hold the image layers' and Replay's part of it. The facade owns no rules. |
| `frontend/src` | The page: React, TypeScript, globe.gl. It reaches Go only through [`api.ts`](frontend/src/api.ts) and states the wire in [`types.ts`](frontend/src/types.ts). |
| `internal/product` | Name, slug, licence, copyright, donate address and credits: a leaf read by the composition root, setup, `runlog` and the tests. |
| `installer` | The setup program, a facade over `setup`. |
| `tests/structural` | The invariants above. |
| `tools` | Icon, notices and Natural Earth data generators. |
| `docs` | The GitHub Pages site. |

Idle rotation lives in [`useIdleRotation.ts`](frontend/src/useIdleRotation.ts):
when the idle delay ends it returns the camera to the fit altitude, then turns;
input during the return stops it (FR-GLB-018). It also answers when the running
delay ends, which `GlobeView.tsx` hands to
[`ResumeBar.tsx`](frontend/src/components/ResumeBar.tsx), the countdown bar
that empties over what is left of the delay (FR-GLB-019).

Replay's state lives in [`useReplay.ts`](frontend/src/useReplay.ts): the
position, playing and the span's end, fixed until Now or another window
(FR-RPL-003, FR-RPL-021, FR-RPL-024). It asks for a frame at most every
`FRAME_ASK_MS`.
[`ReplayControls.tsx`](frontend/src/components/ReplayControls.tsx) draws its
controls on the top bar (FR-RPL-020).

## Dependency direction

```
   page    frontend/src (React)
              | bound calls, events
   UI      app.go facade, main.go
              | calls
           application (ports, services, dto)
              | depends on          ^ implements
           domain               infrastructure
```

## How the globe fills

`main.go` wires the program in this order:

1. `keepLog` opens the log through `runlog` and points error output at it
   (NFR-REL-002); it falls back to standard error, as it does in the
   `wails build` bindings pass (`binding_pass.go`).
2. `clientFor` builds one `httpfetch` client per host, with `requestTimeout`
   (30 s) and `responseCap` (16 MB, FR-PRV-011).
3. `Globe` over a `Store` and the clock, with the cache in the data folder;
   without one the run carries on in memory (FR-STS-005).
4. Settings load before any fetch, so USGS is asked at the saved minimum.
5. Cached sets are restored so the globe opens on the last known events
   (FR-STS-004); the image layers restore their images and shown settings.
6. Wails starts at 1280 by 800 (minimum 960 by 700). On Linux the GPU policy is
   set to Always so the globe has WebGL (RSK-002).

On startup the facade runs two guarded goroutines, each logging a panic and
telling the page (NFR-REL-003): one loads the gazetteer; the other, `drive`,
starts each due fetch in a guarded goroutine and sleeps until the next is due,
a fetch ends or a manual refresh arrives. Each fetch logs its outcome
(NFR-OBS-001) and emits `events-changed`; the page asks for the view. The
layers emit `clouds-changed` and `burnt-changed`, read through one hook
(`useLayer.ts`) that fetches an image only when its key changes; replay images
emit `replay-changed`. The sun needs no background work: the page asks once a
minute while the layer is shown (FR-DAY-004). On DOM ready the facade focuses
the WebView2 child; the page calls `TakeKeyboard` if it holds no keyboard
(NFR-KBD-003).

## Data locations

| What | Where |
|---|---|
| Settings | `%LOCALAPPDATA%\EarthNow\settings.json` |
| Cache | `%LOCALAPPDATA%\EarthNow\cache`: `<provider>.json`, `cloud.json`, `burnt.json` |
| Log | `%LOCALAPPDATA%\EarthNow\Log.txt`, rotating at 5 MB to `Log.previous.txt` |
| WebView2 data | `%LOCALAPPDATA%\EarthNow\webview` |
| Installed files | `%LOCALAPPDATA%\Programs\EarthNow`, with `uninstall.exe` |
| Apps list entry | `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\EarthNow` |
| Shortcuts | Start Menu Programs under `%APPDATA%`; the Desktop |
| Setup's step log | `%TEMP%\EarthNowSetup.log` |
| Setup's WebView2 data | `%TEMP%\EarthNowSetup` |

The data folder is `os.UserCacheDir()` joined with the product's slug
(NFR-PRIV-002): `~/Library/Caches/EarthNow` on macOS and
`~/.var/app/uk.codecrafter.EarthNow/cache/EarthNow` inside the Flatpak.

## The setup program

`installer/` is a second Wails `main` embedding the built application as
`payload.zip`, with a hand-written page in `installer/frontend/dist`. The
install policy lives in `internal/infrastructure/setup` (DEL-003).

- `DetectState` reads the machine once: install when nothing is recorded,
  otherwise update, go back or manage (Repair or Reinstall) by comparing
  versions; `-uninstall` opens on removal.
- Per user under `HKCU`, so no administrator rights; Modify and Repair reopen
  setup.
- A running EarthNow is refused before any file is touched.
- An archive entry whose path climbs out of the install folder is refused.
- Install, update, go back and reinstall are one path, `write`: extract,
  register, apply shortcuts. Repair keeps the shortcuts as they stand.
- Uninstall removes shortcuts, the Apps list entry and the data folder if asked,
  then hands the install folder to a hidden `ping 127.0.0.1 -n 3` and `rmdir`
  so setup's own copy can exit first.
- The page names nothing; the product's name arrives in the state.

`auto-scroll.js` and `settle-keyboard.js` live once in
`installer/frontend/dist`; the application's page imports them, with
`frontend/vite.config.ts` letting the dev server read that folder.

## Versioning

`VERSION` holds the only version string (CON-005). `build.ps1` passes it to both
programs as `-X main.appVersion` (a `var`, since `-X` ignores a `const`). It
first runs `stamp_version.py` to write the version into the site. See
[DEVELOPMENT.md](DEVELOPMENT.md#versioning).

## Implementation decisions

The conceptual choices live in [DECISIONS-TRADEOFFS.md](DECISIONS-TRADEOFFS.md);
these are the rows it would not hold.

| Decision | Chosen | Rejected and why |
|---|---|---|
| Page coverage provider | istanbul | v8 reported `GlobeView.tsx` at 100% with no test importing it. |
| Page composition root | `App.tsx` and `main.tsx` excluded from the floors | They only wire parts together, as `main.go` does. |
| Clustering | Written in `clusters.ts`, pure | globe.gl and three-globe offer none. |
| Markers at the edge | Drawn over the globe; `overHorizon` hides those beyond the horizon | The depth test cut sprites in half near the edge. |
| Cloud image | Drawn in Go, sent as a PNG data URL, laid on a second sphere (`imageLayers.ts`) | Fetching on the page breaks its CSP (NFR-SEC-002). |
| Replay frames | The page plays its position and asks Go every `FRAME_ASK_MS` | A Go timer pushing frames for what is a page animation. |
| Sun position | Go domain, NOAA's equations | npm `solar-calculator` on the page: a second home for astronomy. |
| Light figures | The shader keeps the ramp shape; the limit and floor arrive from Go | Per-pixel light in Go cannot cross the wire. |
| Cloud times | The layer's own capabilities document (6.4 KB) | The whole service's (282 KB). |
| `Accept` header | Set per adapter | The Smithsonian feed answers 403 to JSON only. |
| Keyboard | Focus the WebView2 child, with `TakeKeyboard` as a retry | `runtime.Show` alone lost a race inside Wails. |
| WebView2 data | Inside the data folder | The default falls to `%APPDATA%\EarthNow.exe`. |
| cgo | Pinned off in `build.ps1` | A machine with a C compiler would quietly build a different binary. |
| Icon | One committed `.ico` from `tools/genicons.py` | Wails derives one only when the file is absent. |

## See also

- [README.md](README.md) for what EarthNow is and how to install it.
- [TESTING.md](TESTING.md) for the gate, the floors and the checks a person
  makes.
- [DEVELOPMENT.md](DEVELOPMENT.md) for the tools, the build and the release
  steps.

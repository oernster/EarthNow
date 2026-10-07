# Testing

How EarthNow is tested and what the gate checks. Every command is PowerShell,
run from the repository root.

## The gate

```powershell
./test.ps1
```

`build.ps1` runs it before building anything and has no switch to skip it. In
order:

1. `go list ./...` names the packages, leaving out any Go package an npm
   dependency ships under `node_modules`.
2. `gofmt -l internal tests installer` must list nothing.
3. `go vet` over those packages must pass.
4. staticcheck, at the version pinned in `test.ps1`, run through `go run`, must
   report nothing.
5. `go test` over those packages: the whole Go suite, structural tests
   included.
6. Coverage of `internal/domain` and `internal/application` together must reach
   100%. When it does not, every function short of it is named.
7. Each gated infrastructure package must reach its own floor (below).
8. In `frontend`: `npm run lint` (ESLint), `npx tsc --noEmit`, then `npm test`,
   which is `vitest run --coverage` held to the floors in
   `frontend/vite.config.ts`.
9. `python tools/notices.py --check`: `THIRD_PARTY_NOTICES` must be exactly
   what the tool would write from the dependencies shipped now.

It ends with `All green.`

## Reading the result

Trust the exit code, never the text. A failing step throws, which stops the
script with the reason. Run in the current session, a throw can leave
`$LASTEXITCODE` at `0` (measured), so run the gate as its own process:

```powershell
pwsh -NoProfile -File ./test.ps1
```

```powershell
$LASTEXITCODE
```

`0` means every step passed; `1` means one failed. `./test.ps1 -Floor 95`
changes the domain and application floor for a deliberate check, never for a
release.

## The coverage floors

The domain and the application are held at 100% because they are pure: no
network, no disk, no window. Every other floor is the figure that package
measured, never a target; it fails the moment cover is lost and is raised when
cover rises (NFR-MNT-001). The reasons are the ones written beside each floor in
`test.ps1` and `frontend/vite.config.ts`.

| Package | Floor | Why not 100 |
|---|---|---|
| `internal/domain`, `internal/application` | 100 | |
| `infrastructure/geo`, `pngcheck`, `settings` | 100 | |
| `infrastructure/providers/eonet`, `usgs`, `gvp` | 100 | |
| `infrastructure/httpfetch` | 98.0 | A request-building failure that no valid method and context can produce. |
| `infrastructure/clouds` | 97.8 | Encoding the drawn image into memory, which cannot fail. |
| `infrastructure/gwis` | 96.9 | Encoding the composed image into memory, which cannot fail. |
| `infrastructure/cache` | 92.6 | Five faults the operating system will not produce on demand: an open failing other than for absence, encoding a type that always encodes, then creating, writing or closing a temporary file in a folder just made. |
| `infrastructure/oslocale` | 85.7 | The region call failing (absent before Windows 10 1709, else answering nothing); neither happens on a current Windows. |
| `infrastructure/runlog` | 77.4 | Sending the error output to the log runs only in a crashing child process, where coverage is not collected; the crash tests prove the report lands. Beyond that, faults the operating system will not produce on demand. |
| `infrastructure/setup` | 59.9 | What acts on the machine: the uninstall entry's registry writes, the shortcut made through the Windows Script Host, then finding, ending, launching or scheduling the removal of a process. A test must not change the machine it runs on. The figure is for a machine without EarthNow installed; with it installed the package reads 61.4%. |

Not gated, deliberately: `internal/infrastructure/window` (Win32 focus
handling) and `installer` (the setup program's Wails facade). Neither has
anything a test can reach without the platform behind it. The root package is
the composition root and the Wails facade; it has no tests. `internal/product`
holds constants whose tests pin their values. `gofmt -l` reads `internal`,
`tests` and `installer` only, so the root package's formatting is not checked.

The page is measured with istanbul over `src/**`, leaving out the test files,
`test-setup.ts` and the composition root (`main.tsx` and `App.tsx`):

| Page measure | Floor |
|---|---|
| Statements | 87.68 |
| Branches | 81.65 |
| Functions | 85.54 |
| Lines | 89.46 |

These are measured figures too. The globe's drawing needs WebGL and real
layout, which jsdom lacks; the page's own rules are tested against a stand-in
for globe.gl.

## What the tests prove

### Domain and application

| Package | What it proves | Requirements |
|---|---|---|
| `domain/freshness` | Freshness wording at every boundary, day precision in days, staleness at three intervals | NFR-FRESH-001, NFR-FRESH-002, FR-SEL-004 |
| `domain/window`, `application/services` | Window membership; the newest observation as the event's time and place; the clock-skew bound. Ongoing volcanoes shown in every window and in a replay from their report week, gone once the report is past 14 days, with the status saying why; the week worded across months and New Year | FR-TW-002, FR-RPL-009, FR-PRV-016, FR-SEL-004, DATA-009 |
| `application/services` | The scheduler on a fake clock: intervals, backoff to its ceiling, recovery, the manual refresh and its cooldown; which providers count as refreshing | FR-PRV-010, FR-STS-007 |
| `application/services` | The globe and store over fake providers: one failure leaving the rest standing, a 304 keeping the set, the cache restored first, the notices, the dropped-item count, the minimum magnitude filtering the view while the cache keeps every quake; the https rule for source links | FR-SET-002 |
| `providers/usgs` | The offline promise end to end over a fake fetcher: an unusable answer keeps the held quakes on the globe and on disk with the reason shown | FR-PRV-012 |
| `domain/window`, `application/services`, `settings` | Storm trails clipped to the window and ending at the marker; trails on for a first run and an older settings file | FR-TRL-001, FR-TRL-004, FR-TRL-005 |
| `domain/event`, `application/services` | Earthquake depth wording at each band edge, rounding, depth above sea level, no negative zero, the fixed 10 km mark | FR-SEL-010 to FR-SEL-013 |
| `domain/cloud`, `application/services` | The cloud ramp, veil and status; on a fake clock: no request while hidden, an hourly check, fetching only new times, the held image kept through failure, the cached image drawn at start | FR-CLD-003 to FR-CLD-007, FR-CLD-009 to FR-CLD-011, FR-CLD-013, FR-CLD-014, FR-CLD-016 |
| `domain/sun`, `application/services`, `settings` | The subsolar point within 0.1 degrees of NOAA's calculator; the light ramp; the clouds' night floor; the sun moving on a fake clock; the layer's first-run default kept | FR-DAY-001, FR-DAY-002, FR-DAY-004, FR-DAY-007, FR-DAY-009 |
| `domain/region`, `application/services`, `geo` | Region codes read as a country or none; the label point answered, else no start view with the reason | FR-GLB-014, FR-GLB-015, FR-GLB-016 |
| `domain/burnt`, `domain/window`, `application/services` | Every UTC day a window touches; the union keeping the highest opacity; the status in UTC; on a fake source: one request per day, only missing days, nothing while hidden, failed days counted, held days drawn offline | FR-BA-001 to FR-BA-006, FR-BA-008, FR-BA-009, FR-BA-012 to FR-BA-014, FR-BA-017 |
| `domain/window`, `domain/freshness`, `domain/cloud`, `application/services` | Replay: the instant a position names; a frame showing what had happened by it; the count and status lines; a late event waiting; the replay's clouds fetched, chosen, named when missing, retried and released | FR-RPL-001, FR-RPL-009 to FR-RPL-011, FR-RPL-013 to FR-RPL-019, FR-RPL-022 |

### Adapters and structure

| Package | What it proves | Requirements |
|---|---|---|
| `providers/eonet`, `usgs`, `gvp` | Feeds captured from the real sources (in `testdata`) through a fake fetcher: request URLs, category mapping, malformed items dropped and counted, an all-malformed answer refused, withdrawn quakes left out, GDACS polygon order, the volcano report's Latin-1 mended, report weeks read from titles | FR-PRV-012, FR-PRV-015, DATA-013 |
| `clouds` | EUMETSAT's captured capabilities and images made in the test: the newest time, the GetMap query, the drawn pixels, refusal of an XML exception or a wrong PNG; a replay image at half size | FR-CLD-012, FR-RPL-015 |
| `gwis`, `pngcheck` | One day asked for at its size; a day drawn only when a pixel is burnt; only the PNG asked for accepted; the composed window in red | FR-BA-002, FR-BA-006, FR-BA-009, FR-BA-013, FR-CLD-012 |
| `httpfetch` | Against local TLS servers: the host allowlist, size cap, status check and a 304. Every redirect status followed within the host and refused elsewhere; plain http refused before the dial; lookalike hosts refused. It proves the client's decisions, not DNS or certificates | none |
| `cache`, `settings` | Round trips in temporary folders; an older settings file starting the day and night layer shown; another schema, a damaged or oversized file and a failed save | FR-DAY-007 |
| `oslocale`, `geo` | This machine's region read as a code or none; one label point per country in the embedded table | FR-GLB-017 |
| `runlog` | Rotation at start and while running; a child process panics and its log is read | none |
| `setup` | Extraction and its fence, paths, sizes, copies, versions, the step log, shortcut boxes over redirected folders, registry reads | none |
| `tests/structural` | The layers, a pure domain, one composition root, the 400-line limit, the single homes, the page's rules read from source, the wire between the Go DTOs and `frontend/src/types.ts`, the gate and build order, one host per `httpfetch.New` call; every Must verified by T is named by a test | NFR-MNT-001 |

[ARCHITECTURE.md](ARCHITECTURE.md) lists each structural rule beside the test
that enforces it.

### The page

Vitest under jsdom; `src/test-setup.ts` supplies an inert `ResizeObserver` and
nothing invents a measurement.

| Test file | What it proves | Requirements |
|---|---|---|
| `clusters`, `cursor`, `categories`, `autoScroll`, `settleKeyboard`, `ring` | Clustering and separation altitude, the keyboard cursor, the category table, the auto-scroll machine tick by tick, the keyboard repair, the ring (with layout stated through `testLayout.ts`) | none |
| `detail`, `help`, `helpdialogs` | The Help surfaces and rail order; the Depth row and the guide's depth note; a volcano's report week; the status popover's notices and dropped-item count | FR-SEL-004, FR-SEL-010, FR-SEL-013, FR-SEL-014, FR-PRV-013, FR-PRV-016 |
| `key` | Seven categories, an unknown one falling to Other | DATA-002 |
| `trails` | Storms handed to the globe as paths with their colour ramp, none while off, the Settings box | FR-TRL-002 to FR-TRL-004 |
| `refresh`, `clouds` | The status wording, the Refresh button's turn and last time; the cloud button, line and sphere | none |
| `daynight` | The button, the sun in three-globe's frame, when the sun is asked for, the night lights loaded once, the light while hidden, the shader injected over three's own sources | FR-DAY-003 to FR-DAY-006, FR-DAY-008, FR-DAY-009 |
| `burnt` | The Settings box, the line, the guide's limits, the sphere beneath the clouds, the shared image reader | FR-BA-005 to FR-BA-008, FR-BA-010, FR-BA-012, FR-BA-016 |
| `replay` | Play and Pause, the ring stops, Space and the scrubber, a pass on stubbed frames held at its end, Now, the speed button and its passes, pause and seek, the lines and the guide | FR-RPL-002 to FR-RPL-008, FR-RPL-013, FR-RPL-014, FR-RPL-019, FR-RPL-021, FR-RPL-023 to FR-RPL-025, NFR-KBD-009 |
| `clusterlist` | Two quakes 2.0 km apart stay one cluster; the list opens near the minimum altitude; choosing a member opens it | FR-MRK-011, FR-MRK-012 |
| `markers` | Markers drawn over the globe, shown to the horizon and hidden past it | none |
| `globeview` | Idle rotation: input stops it, the setting applies at once, it ignores reduced motion, it returns to the fit altitude first; the countdown bar over the idle delay, refilled by input | FR-GLB-003, FR-GLB-004, FR-GLB-011, FR-GLB-012, FR-GLB-018, FR-GLB-019 |
| `noborderfocus`, `help` | The noborderfocus rule's two guards, each proved by planting the defect back | NFR-KBD-007 |

## What the tests never do

- **Write to the registry.** Only the setup package's registry reads run.
- **Touch real shortcuts.** The shortcut tests redirect `USERPROFILE` and
  `APPDATA` into temporary folders and fail if the redirection did not take.
- **Launch, find or end EarthNow.** No test calls setup's process functions.
- **Reach a provider, EUMETSAT or GWIS.** Adapters read captured fixtures or
  built images; `httpfetch` talks to a server on this machine.
- **Read or write the real settings, cache or log.** Each such test works in
  `t.TempDir()`.
- **Open a browser.** That call lives in the untested facade.

The gate itself reaches the network once: the first `go run` of the pinned
staticcheck on a machine fetches it.

## Running part of the suite

```powershell
go test ./internal/domain/...
```

```powershell
go test -run TestFRPRV014_OnlyTheCompositionRootNamesAProvider -v ./tests/structural
```

```powershell
go test -cover ./internal/infrastructure/cache
```

```powershell
npm --prefix frontend test
```

```powershell
python tools/notices.py --check
```

## Before a first run on Windows

- **Install the page's packages.** The gate stops at its front-end step without
  `frontend/node_modules`; `tools/notices.py` also reads the production tree
  through `npm ls`:

  ```powershell
  npm --prefix frontend install
  ```

- **Put Python on the path as `python`.** `tools/notices.py` needs only the
  standard library.
- **Let the antivirus leave test binaries alone.** `go test` builds and runs
  each test binary in Go's scratch directory; the runlog tests also start it
  again as a child. Exclude the directory this prints; when it prints nothing,
  Go uses the default temporary folder:

  ```powershell
  go env GOTMPDIR
  ```

## Checked by a person

Some requirements need a real window, a real GPU, the network or a person
looking. They are verified by demonstration (D) or inspection (I) in
[REQUIREMENTS.md](REQUIREMENTS.md). The Phase 0 spike's measured results are
recorded there, in section 3.1.

| Check | How |
|---|---|
| The Phase 0 spike (FR-SPK-001, FR-SPK-002, FR-SPK-003, FR-SPK-004, FR-SPK-005, FR-SPK-006, FR-SPK-007) | Its measured results, section 3.1 of REQUIREMENTS.md. |
| The globe draws with its texture, turns from launch and again after 10 s idle, follows a drag and centres a selected event at unchanged altitude (FR-GLB-001, FR-GLB-002, FR-GLB-005, FR-GLB-007) | Leave it, drag it, select an event on the far side. |
| Reset view returns to the fit altitude; the whole globe fits the globe area; the focus animation looks like one second (FR-GLB-008, FR-GLB-013, NFR-UX-003) | Zoom, reset, resize the window. |
| Every category's emoji draws; hover shows the tooltip; the selection ring shows; no marker animates; a marker or cluster near the globe's edge draws whole until it passes behind (FR-MRK-002, FR-MRK-005, FR-MRK-006, FR-MRK-009) | Look; leave the globe turning and watch markers and cluster badges reach the edge. |
| A cluster zooms until its members separate (FR-MRK-008) | Activate a cluster. |
| Activating a marker opens its detail (FR-SEL-001) | Click one. |
| The key never overlaps the globe; rail and donate tooltips are not clipped (FR-KEY-003, FR-RAIL-003, FR-DON-007) | At the minimum window size, 960 by 700 (NFR-UX-002). |
| The keyboard works with no click at start (NFR-KBD-003) | Launch, press Tab; the log records each focus attempt. |
| The guide names every button and category (FR-HLP-004); categories differ by emoji alone (NFR-A11Y-001) | Read the guide. |
| Credits and notices (NFR-LEG-002, FR-GEO-008) | Read About and `THIRD_PARTY_NOTICES`. |
| One dark palette in the window; the setup program's theme toggle shows the mode it switches to (NFR-UX-005, NFR-UX-004) | Look at the window; press the setup program's toggle both ways. |
| Nothing is fetched from the donate address; no feature depends on a donation (FR-DON-009) | Read `internal/product` and `app.go`'s `Donate`. |
| A panic leaves a record; a panic on a goroutine the application starts is recovered, logged and shown (NFR-REL-002, NFR-REL-003) | A planted panic in a debug build, on the main path and on each goroutine; read the log and the status popover. |
| A failure found at startup reaches the window rather than ending the run (NFR-REL-001) | Launch with the data folder unwritable; the window opens and says so. |
| Start time, frame time, memory over a day, refreshes without a stall (NFR-PERF-001, NFR-PERF-002, NFR-PERF-003, NFR-PERF-004) | The log's timestamps for start and refreshes; frame time as the Phase 0 spike measured it (section 3.1 of REQUIREMENTS.md), since the application keeps no frame-time log. |
| The cloud layer draws over the texture, turns with it and stays beneath every marker; the veil reads as unseen rather than as cloud; frame time holds with the layer shown (FR-CLD-008, FR-CLD-015, NFR-PERF-005) | Show the clouds, leave the globe turning, look at the poles. With no frame-time log in the application, smoothness is judged by eye; the result and the spike's measured thresholds are in section 3.2.10 of REQUIREMENTS.md. |
| The lit side faces the sun and the terminator runs through dawn and dusk; the city lights show on the night side only; clouds over the night side dim and never glow white; frame time holds with both layers shown and 2,500 markers (FR-DAY-003, FR-DAY-009, NFR-PERF-006) | Show both layers and compare the terminator with NOAA's or any day and night map for the same minute; leave the globe turning. With no frame-time log in the application, smoothness is judged by eye, as for NFR-PERF-005. |
| The burnt areas draw in red over the texture, turn with it and lie beneath the clouds and every marker; the spike's measurements are taken; frame time holds with every layer shown (FR-BA-006, FR-BA-007, FR-BA-015, NFR-PERF-007) | Tick Settings' box with 7 days chosen, show the clouds too, leave the globe turning. The spike's results go in section 3.2.13 of REQUIREMENTS.md; smoothness is judged by eye, as for NFR-PERF-005. |
| Replay plays the chosen window in 30 s at normal speed with events appearing, tracks growing, the sun sweeping and the burnt days building; the replay clouds arrive while it plays; it holds at the end until Now; the speed button changes the pace and is kept after a restart; the controls sit on the top bar's one row at the minimum window; frame time holds through a 7 day pass with every layer shown (FR-RPL-020, FR-RPL-024, FR-RPL-025, NFR-PERF-008) | Choose 7 days with every layer shown and press Play; drag the slider; let it reach the end and check it holds; press Now; press the speed button to 2x, play, restart the app and check it still reads 2x; resize to 960 by 700. Smoothness is judged by eye, as for NFR-PERF-005. |
| A storm's track fades from faint to strong and ends in its marker without crowding the globe (FR-TRL-002) | Show the 7 day window with a storm in it; look, then clear Settings' box. |
| The globe opens facing your country, then turns from there, on Windows, a Mac and Linux (FR-GLB-014, FR-GLB-015) | Launch; your country faces you before rotation begins. The log's "Start view:" line names the region and the point. |
| Idle rotation first brings the whole globe back into view, smoothly, from zoomed in or out; a drag during that return stops it where it is (FR-GLB-018) | Zoom in on an event and leave the mouse alone for 10 s: the globe eases back to its usual size over a second, then turns. Zoom out and do the same. Zoom in, press Stop rotating then Start rotating: the same return, then it turns. Drag while it eases back: it stops at once and turns only after another 10 s. |
| Setup (DEL-002) | Install, update, go back, repair, reinstall and uninstall, each with EarthNow running; inspect the folders and the Apps list afterwards. |

## See also

- [README.md](README.md) for what EarthNow is and how to install it.
- [ARCHITECTURE.md](ARCHITECTURE.md) for the invariants the structural tests
  enforce and the design decisions.
- [DEVELOPMENT.md](DEVELOPMENT.md) for the tools, the build and the release
  steps.

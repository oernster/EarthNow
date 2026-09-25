# Testing

How EarthNow is tested, what the gate checks and what a person has to check.
Every command is PowerShell, run from the repository root.

## The gate

```powershell
./test.ps1
```

`build.ps1` runs it before building anything and cannot skip it. In order:

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

### Reading the result

Trust the exit code, never the text. A failing step throws, which stops the
script with the reason. Run in the current session, a throw leaves
`$LASTEXITCODE` holding whatever the last native command returned, which can be
`0` (measured). Run the gate as its own process to read a verdict:

```powershell
pwsh -NoProfile -File ./test.ps1
```

```powershell
$LASTEXITCODE
```

`0` means every step passed; `1` means one failed.

`-Floor` changes the domain and application floor for a deliberate check,
never for a release:

```powershell
./test.ps1 -Floor 95
```

### The coverage floors

The domain and the application are held at 100% because they are pure: no
network, no disk, no window, so nothing in them is out of a test's reach. Every
other floor is the figure that package measured, never a target: it fails the
moment cover is lost and is raised when cover rises (NFR-MNT-001). The reasons
below are the ones written beside each floor in `test.ps1` and
`frontend/vite.config.ts`.

| Package | Floor | Why not 100 |
|---|---|---|
| `internal/domain`, `internal/application` | 100 | |
| `infrastructure/geo` | 100 | |
| `infrastructure/providers/eonet` | 100 | |
| `infrastructure/providers/usgs` | 100 | |
| `infrastructure/providers/gvp` | 100 | |
| `infrastructure/settings` | 100 | |
| `infrastructure/pngcheck` | 100 | |
| `infrastructure/httpfetch` | 97.7 | A request-building failure that no valid method and context can produce. |
| `infrastructure/oslocale` | 85.7 | The region call failing: absent before Windows 10 1709, else answering nothing. Neither happens on a current Windows. |
| `infrastructure/clouds` | 97.8 | Encoding the drawn image into memory, which cannot fail. |
| `infrastructure/gwis` | 96.9 | Encoding the composed image into memory, which cannot fail. |
| `infrastructure/cache` | 92.6 | Five faults the operating system will not produce on demand (measured): an open failing other than for absence, encoding a type that always encodes, then creating, writing or closing a temporary file in a folder just made. |
| `infrastructure/runlog` | 77.4 | Sending the error output to the log is reached only in a crashing child process, where coverage is not collected; the crash tests prove the report lands. Beyond that, faults the operating system will not produce on demand: the log failing to open, to report its size or to close; the runtime refusing a crash file. |
| `infrastructure/setup` | 59.9 | What acts on the machine itself: the uninstall entry's registry writes, creating a shortcut through the Windows Script Host, then finding, ending, launching or scheduling the removal of a process. A test must not change the machine it runs on. |

`infrastructure/setup` reads 61.4% on a machine where EarthNow is installed:
the installed-version read then runs three statements past its early return,
121 of 197 against 118 (measured). The floor is the figure for a machine
without it; raising it to 61.4 would fail the gate on any machine that has never
installed the product.

Not gated, deliberately: `internal/infrastructure/window` (Win32 focus
handling) and `installer` (the setup program's Wails facade over acts that
change the machine). Neither has anything a test can reach without the platform
behind it, so a floor over either would be a floor at zero. The root package
(`main.go`, `app.go`) has no tests: it is the composition root and the Wails
facade. Running it would be running the application. `internal/product` holds
constants and its tests pin their values, so it carries no floor either;
`tests/structural` is tests and nothing else. `gofmt -l` reads `internal`,
`tests` and `installer`, so the root package's formatting is not checked.

The page is measured with istanbul over `src/**`, leaving out the test files
(the shared helper `testLayout.ts` is measured), `test-setup.ts` and the page's
composition root, `main.tsx` and `App.tsx`,
which wire the parts together and are checked by eye:

| Page measure | Floor |
|---|---|
| Statements | 87.08 |
| Branches | 80.37 |
| Functions | 85.06 |
| Lines | 88.9 |

These are measured figures too. `GlobeView.tsx`'s own rules (marker
placement, the cursor's tooltip, the focus animation) and the idle rotation it
takes from `useIdleRotation.ts` are tested against a stand-in for globe.gl; the drawing itself needs WebGL and real layout, which
jsdom does not have, so it is exercised by eye, in the checks below.

## What the tests prove

### The domain and the application

- **Freshness wording** at every boundary of NFR-FRESH-002, day precision
  worded in days (FR-SEL-004) and staleness at three intervals (NFR-FRESH-001).
- **The time window**: membership; the newest observation inside a window as
  the event time and position; the `ClockSkew` bound, both at it and one
  second past it.
- **The scheduler** on a fake clock: each provider on its own interval, the
  backoff doubling to its ceiling, recovery after a success, the manual refresh
  and its cooldown with the time of the last one (FR-PRV-010); which
  providers count as refreshing (FR-STS-007).
- **The globe and the store** over hand-written fake providers: a failed
  provider leaving the others' events standing, a not-modified answer keeping
  the stored set, the cache restored before any fetch, the notices when there
  is no cache or no settings file. The https rule for source links is tested
  there as well.
- **Storm trails**: a storm's fixes clipped to the window and ending at the
  marker, none for one fix or for any other category (FR-TRL-001); the trail
  reaching the wire as pairs, empty rather than absent; trails on for a first
  run and for a settings file from before them, the choice kept (FR-TRL-004,
  FR-TRL-005).
- **Earthquake depth**: the wording at each band edge and either side of it,
  a value rounding across an edge taking the band it shows, a negative depth
  above sea level, -0.04 reading 0.0 rather than -0.0 (proved by planting the
  sign back), exactly 10 km marked as often fixed while 10.04 is not, the
  wording reaching the wire and no depth leaving the row empty (FR-SEL-010
  to 013).
- **The cloud layer**: the brightness ramp at both thresholds and between them
  (FR-CLD-006), the veil for a pixel with no data (FR-CLD-007), the status
  wording and its staleness mark (FR-CLD-009, FR-CLD-010). On a fake clock and
  fake service: no request while hidden, one check per hour while shown, an
  image fetched only for a newly listed time, the held image kept through a
  failure with the backoff, the first failure said, the cached image drawn at
  start with its age (FR-CLD-003 to FR-CLD-005, FR-CLD-011, FR-CLD-013,
  FR-CLD-014, FR-CLD-016).
- **The day and night layer**: the subsolar point against NOAA's solar
  calculator at 2026's solstices and equinoxes, read off the calculator
  itself, within 0.1 degrees; the elevation at latitude and longitude 0
  against the calculator's; the longitude wrapping at the date line a
  quarter degree a minute (FR-DAY-001). The light ramp at -6, 0 and 6
  degrees (FR-DAY-002) and the clouds' night floor (FR-DAY-009). On a fake
  clock the sun moves between two readings a minute apart (FR-DAY-004); a
  first run shows the layer and hiding it is kept (FR-DAY-007).
- **The start view**: region codes and locale names read as a country or as
  none (FR-GLB-014); the region's label point answered, else no start view
  with the reason for the log (FR-GLB-015, FR-GLB-016).
- **The burnt areas**: every UTC day a window touches, whatever the clock's
  zone (FR-BA-001); the union keeping each pixel's highest opacity (FR-BA-006);
  the status wording in UTC with its age, else none mapped yet (FR-BA-008,
  FR-BA-009). On a fake clock and fake source: one request per day, every day
  again each interval, only the days a new window lacks, nothing while hidden,
  a failed day leaving the others drawn, a refused answer counted as a failed
  day, the maps said to be unavailable when nothing is held, held days drawn
  offline with old ones discarded and the image composed once per key
  (FR-BA-002 to FR-BA-006, FR-BA-012 to FR-BA-014, FR-BA-017).
- **Replay**: the instant a position names along the span (FR-RPL-001); a
  frame showing what had happened by its instant, a storm's trail growing
  with it and the burnt days building up (FR-RPL-009, FR-RPL-010,
  FR-RPL-013); the count and status lines naming the instant (FR-RPL-011,
  FR-RPL-019); an event from after the span's end waiting for the return
  (FR-RPL-022). The replay's clouds on a fake source: the span's three-hourly
  times fetched one a round at the replay's size, the image drawn the latest
  at or before the instant, play not waiting for them, a missing one named,
  a failed listing retried and the images released on the return
  (FR-RPL-014 to FR-RPL-018).

### The adapters

- **Each provider** parses a feed captured from the real source (in its
  `testdata`), through a fake fetcher: the request URL, the categories mapped,
  malformed items dropped and counted, withdrawn earthquakes left out, each
  GDACS polygon read in the order its own coordinates prove (else the order the
  feed's proven polygons show) and the volcano report's Latin-1 decoded. Which
  source link counts as a page is tested in the domain.
- **The cloud adapter** reads the layer's capabilities document captured from
  EUMETSAT (in its `testdata`) and draws images made in the test: the newest
  valid time, the GetMap query, the drawn pixels against the domain's ramp and
  the refusal of an XML exception served with 200, a PNG of another size and a
  PNG cut short (FR-CLD-012); a replay's image asked for at half the size
  (FR-RPL-015).
- **The GWIS adapter** against images made in the test: one day asked for at
  the size and the day, a day counted as drawn only when a pixel is burnt, the
  refusal of anything but the PNG asked for and the composed window keeping
  the highest opacity in red (FR-BA-002, FR-BA-006, FR-BA-009, FR-BA-013).
  `pngcheck`, which both map adapters share, decodes only a PNG of the size
  asked for (FR-CLD-012, FR-BA-013).
- **`httpfetch`** against a local test server: the host allowlist, a redirect
  held to the allowed hosts, the size cap, the status check, `If-Modified-Since` and a 304.
- **The cache and the settings** in temporary folders: round trips (the cloud
  image with its valid time and the burnt-area days among them), a settings
  file from before the day and night layer starting it shown (FR-DAY-007),
  another
  schema version read as absent, a damaged file, an oversized file and a save
  that cannot be written.
- **The region and the label table**: this machine's real region setting
  read as a code (GB on the reference machine) or as none; the embedded table
  holding one label point per country, the countries whose dependencies share
  their code among them (FR-GLB-017). The generator's refusal of a code left
  with two rows was proved by planting one.
- **The log**: rotation at start and while running with one previous file
  kept. A child process really panics after a rotation; its log is then read.
- **Setup**: the extraction and its fence against an entry that climbs out, the
  paths, sizes, copies, versions, the step log, the shortcut boxes over
  redirected folders and the registry reads.

### The structure

`tests/structural` reads the repository rather than running it. It holds the
architecture (layers, a pure domain, one composition root, providers named only
there, the 400-line limit and its danger band, documented exported types), the
single homes (the product's name, the donate address, the category emoji and
the EONET category ids), the page's rules read from its source (the CSP, the
palette's contrast, the ring states, the rail's glyph box, the globe area's
share of the minimum window, the heading), the wire between the Go DTOs and
`frontend/src/types.ts`, what the gate and the build scripts run and in what
order, plus the traceability test: every Must in REQUIREMENTS.md is named by a
test; failing that, it is listed in the table under "Checked by a person"
below, where each ID is spelt in full. [ARCHITECTURE.md](ARCHITECTURE.md) lists
each rule beside the test or test file that enforces it.

### The page

Vitest under jsdom, with `src/test-setup.ts` supplying an inert
`ResizeObserver`; nothing there invents a measurement. The suites cover the
clustering and the altitude at which a cluster's members separate, the keyboard
cursor's walk, the category table, the auto-scroll machine driven tick by tick,
the keyboard repair shared with the setup page, the keyboard ring (with the
page's shape stated through `testLayout.ts`, since jsdom lays nothing out) and
the Help surfaces with the rail's order, the detail panel's Depth row shown
or left out and the guide's note on the fixed depth (FR-SEL-010, FR-SEL-013,
FR-SEL-014), the storms handed to the globe as paths with their colour ramp,
none while switched off (proved by planting the switch away) and the Settings
box (FR-TRL-002 to FR-TRL-004). The refresh indicator is covered too:
the status line's wording, the turning Refresh button held for one turn and
the last refresh time. So is the cloud layer: the button's name and artwork in
each state, the press, the cloud line shown, marked or absent; also the sphere
sitting between the texture and the markers, drawn only while it has an image.
The day and night layer too: the button's name and artwork in each state, the
press, the sun placed in three-globe's own frame, the sun asked for on showing
and within a minute and never while hidden, the night lights loaded once on
the first show, a light of 1 everywhere while hidden (FR-DAY-003 to
FR-DAY-006, FR-DAY-008). The injected light is assembled over three's own
Phong and basic shader sources, so a chunk three renames fails a test rather
than drawing nothing (FR-DAY-003, FR-DAY-009).
The burnt-area layer as well: the Settings box, the line shown, marked or
absent, the guide's four limits, the sphere lying beneath the clouds and
following its image; also the shared reader that asks for a layer's image only
when its key changes (FR-BA-006 to FR-BA-010, FR-BA-016).
Replay too: the Play/Pause button's name and artwork, its two ring stops after
the time window with Space and the scrubber's step, a pass played on stubbed
animation frames to its end and back to now, pause and seek, another window
returning to now, each image asked for once its key names one, the lines and
the guide (FR-RPL-002 to FR-RPL-008, FR-RPL-013, FR-RPL-014, FR-RPL-019,
FR-RPL-021, FR-RPL-023, NFR-KBD-009).
So is the cluster list: two quakes 2.0 km apart staying one cluster at the
minimum altitude, the list opening within half a zoom step of it and
choosing a member opening that event (FR-MRK-011, FR-MRK-012).
So is idle rotation: input stopping it, the setting and its speed applied at
once, rotation whatever the reduced-motion setting says and the return to the
fit altitude from zoomed in or out before turning, stopped by input and taken
again when rotation is switched on (FR-GLB-003, FR-GLB-004, FR-GLB-011,
FR-GLB-012, FR-GLB-018).
The noborderfocus rule has two guards,
each proved by planting the defect back.

## What the tests never do

- **Write to the registry.** The registry writes in
  `internal/infrastructure/setup` are not exercised at all; only the reads are,
  since a read cannot damage anything.
- **Touch real shortcuts.** The shortcut tests redirect `USERPROFILE` and
  `APPDATA` into temporary folders and fail if the redirection did not take.
- **Launch, find or end EarthNow.** No test calls the process functions in
  setup; the root package that runs the application has no tests.
- **Reach a provider, EUMETSAT or GWIS.** The adapters read captured fixtures
  or built images through fakes; `httpfetch` talks to a server on this
  machine.
- **Read or write the real settings, cache or log.** Every such test works in
  `t.TempDir()`.
- **Open a browser.** The call that hands a link to the desktop is in the
  untested facade.

The gate itself can reach the network once: the first `go run` of the pinned
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
  `frontend/node_modules`; `tools/notices.py` also reads the page's production
  tree through `npm ls`:

  ```powershell
  npm --prefix frontend install
  ```

- **Put Python on the path as `python`.** The last step runs
  `tools/notices.py`, which needs only the standard library.
- **Let the antivirus leave test binaries alone.** `go test` builds each test
  binary in Go's scratch directory and runs it from there; the runlog tests
  also start that binary again as a child. Where an antivirus holds or removes
  those files, exclude the directory this prints. When it prints nothing, Go
  uses the default temporary folder instead:

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
| Every category's emoji draws; hover shows the tooltip; the selection ring shows; no marker animates (FR-MRK-002, FR-MRK-005, FR-MRK-006, FR-MRK-009) | Look. |
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
| Replay plays the chosen window in 30 s with events appearing, tracks growing, the sun sweeping and the burnt days building; the replay clouds arrive while it plays; the controls sit on the top bar's one row at the minimum window; frame time holds through a 7 day pass with every layer shown (FR-RPL-020, NFR-PERF-008) | Choose 7 days with every layer shown and press Play; drag the slider; press Play at the end; resize to 960 by 700. Smoothness is judged by eye, as for NFR-PERF-005. |
| A storm's track fades from faint to strong and ends in its marker without crowding the globe (FR-TRL-002) | Show the 7 day window with a storm in it; look, then clear Settings' box. |
| The globe opens facing your country, then turns from there, on Windows, a Mac and Linux (FR-GLB-014, FR-GLB-015) | Launch; your country faces you before rotation begins. The log's "Start view:" line names the region and the point. |
| Idle rotation first brings the whole globe back into view, smoothly, from zoomed in or out; a drag during that return stops it where it is (FR-GLB-018) | Zoom in on an event and leave the mouse alone for 10 s: the globe eases back to its usual size over a second, then turns. Zoom out and do the same. Zoom in, press Stop rotating then Start rotating: the same return, then it turns. Drag while it eases back: it stops at once and turns only after another 10 s. |
| Setup (DEL-002) | Install, update, go back, repair, reinstall and uninstall, each with EarthNow running; inspect the folders and the Apps list afterwards. |

## See also

- [ARCHITECTURE.md](ARCHITECTURE.md) for the invariants the structural tests
  enforce and the design decisions.
- [DEVELOPMENT.md](DEVELOPMENT.md) for the tools, the build and the release
  steps.

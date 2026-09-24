# Development

How to build EarthNow from source, what each script does and how a release is
cut. Every command is PowerShell, one command per block, run from the
repository root. [README.md](README.md) is for someone using the application;
this is for someone building it.

## Tools

| Tool | Needed for | Get it |
|---|---|---|
| Go 1.26.3, which `go.mod` declares | everything | [go.dev/dl](https://go.dev/dl/) |
| Node.js with npm; `package.json` pins no minimum | the page: lint, type check, tests and build | [nodejs.org](https://nodejs.org/) |
| Wails CLI v2.12.0, the version of the Wails module `go.mod` requires | building the application and the setup program | `go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0` |
| WebView2 runtime | running either program | [Microsoft's WebView2 page](https://developer.microsoft.com/microsoft-edge/webview2/) |
| Python 3, as `python` on the path | the notices check the gate ends with, standard library only | [python.org](https://www.python.org/downloads/) |
| Pillow | regenerating the icons, plus the Linux Flatpak's icon sizes | `python -m pip install pillow` |

staticcheck is not installed: `test.ps1` runs it at a pinned version through
`go run`, so a new release of it cannot change the result for unchanged code.
The first run on a machine fetches it, so that run needs the network.

cgo is not used on Windows; the Linux and macOS builds need it, since Wails
renders there through a C web view. `build.ps1` pins `CGO_ENABLED=0` for the gate and the build,
so no C compiler is needed. `./test.ps1` run on its own does not set it, so on a
machine with a C compiler the gate runs as you run it with the Go default.

`wails.exe` lands in `%USERPROFILE%\go\bin`. If this is not found, that folder
is not on the path:

```powershell
wails doctor
```

## First steps after cloning

```powershell
go mod download
```

```powershell
npm --prefix frontend install
```

`wails build` runs `npm install` itself through the `frontend:install` hook in
`wails.json`. The gate runs before that; it stops at its front-end step
without `frontend/node_modules`.

## Building

```powershell
./build.ps1
```

In order, it:

1. Reads `VERSION` and makes `-ldflags "-X main.appVersion=<version>"` from it.
2. Runs `stamp_version.py`, which writes that version into every
   `<!--VERSION-->` token of the site under `docs/`, so the site never offers an
   older number than the setup program. It touches nothing already current.
3. Sets `CGO_ENABLED=0` for everything that follows.
4. Runs [`test.ps1`](TESTING.md). A failure stops the build; there is no switch
   to skip it.
5. Refuses to go on without `assets/application-icon.png` and
   `assets/application-icon.ico`, then copies both into `build/` and
   `installer/build/` as `appicon.png` and `windows/icon.ico`.
6. Runs `wails build` for the application. That runs the page's
   `npm run build`, which is ESLint, `tsc` and `vite build`, then writes
   `build/bin/EarthNow.exe`.
7. With `-SkipInstaller`, stops here.
8. Zips `build/bin` into `installer/payload.zip`.
9. Copies `LICENSE` to `installer/frontend/dist/LICENSE.txt`, since the setup
   page has no build step to read it from the root.
10. Runs `wails build` in `installer/`, with the same `-ldflags`.
11. Copies `installer/build/bin/EarthNowSetup.exe` to
    `dist-installer/EarthNowSetup.exe`, the one file that ships.
12. Writes the 22-byte empty zip back over `installer/payload.zip`, so
    `go build ./...` and the tests keep working without a full build.

Step 12 runs only when the steps before it succeed. If the setup program fails
to build, `installer/payload.zip` still holds the full payload: run
`./build.ps1` again rather than committing it.

To build only the application, the faster loop when the setup program has not
changed:

```powershell
./build.ps1 -SkipInstaller
```

| Output | What it is |
|---|---|
| `build/bin/EarthNow.exe` | the application |
| `dist-installer/EarthNowSetup.exe` | the setup program, application included |

Everything generated is ignored by git: `build/`, `frontend/dist/`,
`frontend/wailsjs/`, `dist-installer/`, `installer/build/bin/`, the icon and
manifest copies under `installer/build/`, `installer/frontend/wailsjs/` and
`installer/frontend/dist/LICENSE.txt`, plus Wails' `frontend/package.json.md5`.
`installer/payload.zip` is tracked only as the empty archive.

## Running from source

`wails.json` names `npm run dev` as the page's dev watcher, so the development
loop serves the page from Vite and rebuilds the Go side on change:

```powershell
wails dev
```

To run a built copy instead:

```powershell
./build.ps1 -SkipInstaller
```

```powershell
./build/bin/EarthNow.exe
```

A binary built without `build.ps1` reports its version as `0.0.0-dev`.

Everything EarthNow does is written to `%LOCALAPPDATA%\EarthNow\Log.txt`: the
version at start, the settings file read, each fetch with its status, event
count and dropped count, each failure with its next retry and each keyboard
handover. The log rotates at 5 MB, keeping one `Log.previous.txt`. Read it
first when something looks wrong.

## Generated files

Both scripts are run by hand; their output is committed.

### The icons

`tools/genicons.py` reads the masters in `assets/` and writes:

- the rail icons into `frontend/src/assets/icons`, at 208 pixels, plus
  `rotate-stop.png` and `cloud-cover-hide.png`, made by laying `negative.png`
  over `rotate.png` and `cloud-cover.png`, so the two states of the rotation and
  cloud buttons cannot drift apart;
- `assets/application-icon.ico` at 16, 24, 32, 48, 64, 128 and 256 pixels,
  which `build.ps1` puts on both executables;
- the setup page's header mark (256 pixels) and its sun and moon (128 pixels)
  into `installer/frontend/dist`;
- the donate mark into `frontend/src/assets/donate.png` and `docs/donate.png`
  from one render, cropped to its artwork and scaled by height to four times
  the rail's 48-pixel glyph;
- the application icon, from one call, for the page's own mark and as the
  site's `docs/icon.png`;
- `docs/earth.jpg`, the page's NASA texture at 760 pixels high, for the
  site's turning globe (CON-009: the Earth is never drawn).

```powershell
python tools/genicons.py
```

Run it when a master changes, then commit the results. A clone builds without
Python or Pillow for this. The one exception is the Linux icon theme:
`build_flatpak.sh` runs `python3 tools/genicons.py --hicolor build/linux/icons`
on every build, which writes the application icon at each hicolor size into
that gitignored folder and nothing else.

### The third-party notices

`tools/notices.py` writes `THIRD_PARTY_NOTICES` from what actually ships: the
Go modules linked into the application and the setup program
(`go list -deps`) and the page's production tree (`npm ls --omit=dev --all`),
each with its licence, then every licence text in full.

```powershell
python tools/notices.py
```

Run it after any dependency change: a module added, removed or bumped in
`go.mod`; the same for a production package in `frontend/package.json`. The gate runs it
with `--check` and fails until the file matches.

### Not generated here

- `frontend/src/assets/earth.jpg` is the NASA Blue Marble Next Generation
  texture, downloaded and committed.
- `internal/infrastructure/geo/data` holds the Natural Earth places, borders
  and Antarctic ice shelves the application embeds. `tools/geodata.py`, first
  written for the Phase 0 spike, converts Natural Earth shapefiles. It is given the
  folder the layers are unpacked into plus an output folder, writes a file
  only for each layer it finds there, then the result is copied into `data`
  by hand. Pass only the layer being refreshed: the shapefiles are not kept,
  so a fresh download of the others might not match what is embedded.

## Versioning

`VERSION` holds the only version string (CON-005). Change it there and nowhere
else. `build.ps1` passes it into both programs through
`-ldflags "-X main.appVersion=<version>"`; `appVersion` is a `var` in
`main.go` and in `installer/main.go`, because `-X` does nothing to a `const`.
The About dialog shows it and setup records it in the Apps list entry, which is
how the next setup program decides between update, go back and repair.

## Cutting a release

1. Set `VERSION`.
2. Run `./build.ps1`. It will not build from a tree that fails the gate.
3. Commit, tag the commit with the version and push.
4. Create a GitHub release for the tag on `oernster/EarthNow` and attach
   `dist-installer/EarthNowSetup.exe`.

The README's install instructions point at the Releases page, so step 4 is
what users see.

## Linux and macOS builds (release 2, not yet verified)

Three bash scripts, ported from PigeonPost with SymDiary's hardening, build
EarthNow for the other two platforms (DEL-005 to DEL-007). None has been run on
its own platform yet. Each was syntax-checked; the Flatpak manifest was
generated and parsed with the flatpak tools stubbed out, all on Windows. Until
a build has run and the globe has drawn, neither platform is supported and the
README says so.

| Script | Runs on | What it makes |
|---|---|---|
| `build_flatpak.sh` | Linux (Ubuntu is the reference) | `earthnow.flatpak` and a user install of `uk.codecrafter.EarthNow`, on the GNOME 50 runtime with webkit2gtk-4.1 |
| `cleanup_flatpak.sh` | Linux | uninstalls it and removes only the flatpak artefacts; the user's settings and cache stay |
| `builddmg.sh` | an Apple Silicon Mac | `EarthNow.dmg`, signed and notarised; `ALLOW_UNNOTARIZED=1` for a local test build only |

```bash
bash build_flatpak.sh
```

- **The globe needs WebGL.** Wails v2 turns webkit2gtk's GPU acceleration off
  unless told otherwise, so `main.go` passes `options.Linux` with the policy
  set to Always (RSK-002). Whether a given machine then offers WebGL2 is
  measured on that machine.
- **The Flatpak keeps the network** for the three sources and the cloud
  image; it asks for no
  filesystem access: its settings, cache and log live in the sandbox's own
  cache folder.
- **The DMG** copies `assets/application-icon.png` to `build/appicon.png`, as
  `build.ps1` does; Wails makes the bundle's icon from it. Notarisation
  needs a keychain profile named `EarthNow` (the script prints how to make one)
  or `APPLE_ID` and `APPLE_APP_PASSWORD`.
- `.gitattributes` holds the three scripts at LF endings, since a Windows
  checkout would otherwise give them a shebang ending in a carriage return.

## Where things live

| Path | What it holds |
|---|---|
| `main.go` | the composition root |
| `app.go` | the Wails facade the page calls |
| `binding_pass.go`, `binding_pass_off.go` | the switch that keeps the build's bindings pass out of the user's log |
| `internal/domain` | `cloud`, `event`, `freshness`, `window`: no I/O |
| `internal/application` | `ports`, `services` and the `dto` wire shapes |
| `internal/infrastructure` | `cache`, `clouds`, `geo`, `httpfetch`, `providers/eonet`, `providers/gvp`, `providers/usgs`, `runlog`, `settings`, `setup`, `window` |
| `internal/product` | the name, slug, licence line, copyright, donate address and credits |
| `frontend/src` | the page |
| `installer/` | the setup program, a Wails application of its own |
| `tests/structural` | the tests that hold the architecture in place |
| `tools/` | `genicons.py`, `notices.py` and `geodata.py` |
| `stamp_version.py` | stamps `VERSION` into the site's version pill; `build.ps1` runs it first |
| `build_flatpak.sh`, `cleanup_flatpak.sh`, `builddmg.sh` | the Linux and macOS builds, release 2 |
| `assets/` | the master artwork |
| `docs/` | the GitHub Pages site: one hand-written page, no build step |

## Standing rules

- **No magic numbers.** A literal that needs a comment to say what it
  represents is a named constant or comes from data.
- **The product is named once,** in `internal/product/product.go`. A structural
  test fails when a Go string literal outside a test, a page source file,
  `frontend/index.html` or a setup page file spells it. The Wails configuration files and the build
  scripts sit outside that test and still carry it.
- **The donate address lives once,** beside the name; a structural test holds
  it there.
- **No file over 400 lines,** and none in the danger band of 381 to 400: one
  that lands there is reduced to 350 or fewer (CON-008). The structural test
  counts the Go files, the page's source and the setup page.
- **The layer direction is enforced.** The domain imports nothing of
  EarthNow's outside itself and reads no clock; the application imports no
  infrastructure; only `main.go` and `app.go` join the two.
- **A new provider** is a package under `internal/infrastructure/providers`
  plus a line in `main.go`; a test fails if anything else imports it.
- **A new wire shape** is a struct in `internal/application/dto`, an interface
  in `frontend/src/types.ts` and a pair in `wireShapes` in
  `tests/structural/wire_test.go`.
- **Domain and application stay at 100% coverage;** no infrastructure floor is
  lowered.
- **Every exported type has a doc comment.**
- **No version string outside `VERSION`.**
- **After a dependency change,** run `tools/notices.py`.

The reasons are in [ARCHITECTURE.md](ARCHITECTURE.md); the checks are in
[TESTING.md](TESTING.md).

## See also

- [ARCHITECTURE.md](ARCHITECTURE.md) for the invariants and the design
  decisions.
- [TESTING.md](TESTING.md) for the gate, the floors and the checks a person
  makes.

# Development

How to build EarthNow from source and cut a release. Every command is
PowerShell, run from the repository root, except the macOS and Linux build
scripts, which run in bash on those platforms.

## Tools

| Tool | What for | Where to get it |
|---|---|---|
| Go, at the version `go.mod` declares | everything | [go.dev/dl](https://go.dev/dl/) |
| Node.js with npm (`package.json` pins no minimum) | the page: lint, type check, tests and build | [nodejs.org](https://nodejs.org/) |
| Wails CLI, at the Wails version `go.mod` requires | building the application and the setup program | `go install github.com/wailsapp/wails/v2/cmd/wails@$(go list -m -f '{{.Version}}' github.com/wailsapp/wails/v2)` |
| WebView2 runtime | running either program | [Microsoft's WebView2 page](https://developer.microsoft.com/microsoft-edge/webview2/) |
| Python 3, as `python` | `stamp_version.py` and the gate's notices check; standard library only | [python.org](https://www.python.org/downloads/) |
| Pillow | regenerating the icons | `python -m pip install pillow` |

staticcheck needs no install: `test.ps1` runs it at a pinned version through
`go run`, so the first run on a machine needs the network. No C compiler is
needed on Windows; `build.ps1` sets `CGO_ENABLED=0`. `wails.exe` lands in
`%USERPROFILE%\go\bin`; if this fails, that folder is not on the path:

```powershell
wails doctor
```

After cloning:

```powershell
go mod download
```

```powershell
npm --prefix frontend install
```

## What build.ps1 does

```powershell
./build.ps1
```

1. Reads `VERSION` into `-ldflags "-X main.appVersion=<version>"`.
2. Runs `stamp_version.py` (see Versioning).
3. Sets `CGO_ENABLED=0` for everything that follows.
4. Runs `test.ps1`. A failure stops the build; there is no switch to skip it.
5. Stops unless `assets/application-icon.png` and `.ico` exist, then copies
   them into `build/` and `installer/build/`.
6. Runs `wails build`, whose `npm run build` is ESLint, `tsc` and
   `vite build`, writing `build/bin/EarthNow.exe`.
7. With `-SkipInstaller`, stops here.
8. Zips `build/bin` into `installer/payload.zip`.
9. Copies `LICENSE` to `installer/frontend/dist/LICENSE.txt`.
10. Runs `wails build` in `installer/` with the same `-ldflags`.
11. Copies `installer/build/bin/EarthNowSetup.exe` to
    `dist-installer/EarthNowSetup.exe`, the file that ships.
12. Writes the 22-byte empty zip back over `installer/payload.zip`.

If step 10 fails, `installer/payload.zip` still holds the full payload: run
the build again rather than committing it. The faster loop when the setup
program has not changed:

```powershell
./build.ps1 -SkipInstaller
```

Every output is ignored by git; `installer/payload.zip` is tracked only as the
empty archive.

## Running from source

`wails dev` serves the page from Vite and rebuilds the Go side on change:

```powershell
wails dev
```

To run a built copy:

```powershell
./build/bin/EarthNow.exe
```

A binary built without `build.ps1` reports a development placeholder in place
of a version. The log is `%LOCALAPPDATA%\EarthNow\Log.txt`: the version, each
fetch with its status and counts, each failure with its next retry. It rotates
at 5 MB, keeping one `Log.previous.txt`. Read it first when something looks
wrong.

## Generated assets

Both scripts are run by hand; their output is committed.

`tools/genicons.py` reads the masters in `assets/` and writes the page's icons
into `frontend/src/assets/icons`, `assets/application-icon.ico`, the setup
page's marks into `installer/frontend/dist`, the in-app donate mark, plus the
site's `docs/icon.png` and `docs/earth.jpg`. The site's `docs/donate.png` is
shared across projects and never generated. Run it when a master changes:

```powershell
python tools/genicons.py
```

`build_flatpak.sh` also runs it with `--hicolor build/linux/icons` on every
Linux build, writing only the icon theme into that ignored folder.

`tools/notices.py` writes `THIRD_PARTY_NOTICES` from the Go modules linked
into both programs and the page's production npm tree. Run it after any
dependency change; the gate runs it with `--check`:

```powershell
python tools/notices.py
```

Not generated here: the NASA textures `frontend/src/assets/earth.jpg` and
`earth-night.jpg` were downloaded, resampled and committed. The Natural Earth
data in `internal/infrastructure/geo/data` comes from `tools/geodata.py`, given
the folder of unpacked shapefiles and an output folder; copy the result into
`data` by hand and pass only the layer being refreshed. Country label points
alone:

```powershell
python tools/geodata.py --labels <unpacked layers> internal/infrastructure/geo/data
```

## Versioning

`VERSION` is the single source of the version (CON-005); change it there and
nowhere else. `build.ps1` passes it into both programs through `-ldflags`;
`appVersion` is a `var` in `main.go` and `installer/main.go` because `-X` does
nothing to a `const`. `stamp_version.py` writes it into every
`<!--VERSION-->` token of the site under `docs/` and versions the site's
stylesheet and script links by content hash. `build.ps1` runs it first; to run
it alone:

```powershell
python stamp_version.py
```

## Cutting a release

1. Set `VERSION`.
2. Run `./build.ps1`.
3. On an Apple Silicon Mac run `bash builddmg.sh`; on Linux run
   `bash build_flatpak.sh`.
4. Commit, tag the commit with the version and push.
5. Create a GitHub release for the tag on `oernster/EarthNow` and attach
   `dist-installer/EarthNowSetup.exe`, `EarthNow.dmg` and `earthnow.flatpak`.

| Script | Runs on | Makes |
|---|---|---|
| `builddmg.sh` | an Apple Silicon Mac | `EarthNow.dmg`, signed and notarised; `ALLOW_UNNOTARIZED=1` for a local test build only |
| `build_flatpak.sh` | Linux (Ubuntu is the reference) | `earthnow.flatpak` and a user install of `uk.codecrafter.EarthNow` |
| `cleanup_flatpak.sh` | Linux | uninstalls it and removes the build artefacts; the user's data stays |

Notarisation needs a keychain profile named `EarthNow` (the script prints how
to make one) or `APPLE_ID` with `APPLE_APP_PASSWORD`. `.gitattributes` holds
the scripts at LF endings.

## Standing rules

- **No magic numbers.** A literal that needs a comment is a named constant or
  comes from data.
- **The product name and the donate address live once,** in
  `internal/product/product.go`; structural tests hold them there.
- **No file over 400 lines;** one in the band 381 to 400 is cut to 350 or
  fewer (CON-008).
- **The layers point inward.** The domain does no I/O and reads no clock; the
  application imports no infrastructure; only `main.go` and `app.go` join them.
- **A new provider** is a package under `internal/infrastructure/providers`
  plus a line in `main.go`.
- **A new wire shape** is a struct in `internal/application/dto`, an interface
  in `frontend/src/types.ts` and a pair in `wireShapes` in
  `tests/structural/wire_test.go`.
- **Domain and application stay at 100% coverage;** no floor is lowered.
- **Every exported type has a doc comment.**
- **No version string outside `VERSION`.**

## See also

- [README.md](README.md) for what EarthNow is and how to install it.
- [ARCHITECTURE.md](ARCHITECTURE.md) for the layers and the reasons behind
  these rules.
- [TESTING.md](TESTING.md) for the gate, the floors and the manual checks.

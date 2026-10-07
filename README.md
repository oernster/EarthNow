# EarthNow

EarthNow is a desktop application for Windows, macOS and Linux that shows the
natural events of the past week on a slowly turning globe: earthquakes from the
USGS, wildfires, storms, floods and sea ice tracked by NASA EONET, plus the
volcanoes in the Smithsonian and USGS Weekly Volcanic Activity Report. Every
event names the source that reported it and how old that report is.

> **Commercial licences available.** EarthNow is free and open source under the
> GNU General Public License, version 3. If those terms do not suit what you are
> building, such as a closed-source product, a commercial licence can be bought
> from me separately. It covers my own code; third-party libraries keep their
> own licences. See
> [commercial licensing](https://ernster.dev/commercial-licensing.html).

## Who it is for

- Anyone curious about what the planet is doing this hour, day or week, who
  would rather see it on a globe than read it in a list.
- Anyone who wants each event traced to its public source with its age stated.
- Keyboard users: every control and every event can be reached without a
  pointer.

## Who it is not for

- Anyone who needs warnings. EarthNow sends no notifications and makes no
  forecast.
- Anyone who wants history. The widest window is seven days.
- Anyone after a GIS tool or a weather service. There is no rain, wind,
  temperature or forecast.

## What it does

The [project site](https://earthnow.world/) gives the full tour.

- Draws the globe from NASA's Blue Marble, opening over your own country.
- Turns slowly while idle; a bar counts down the pause after you touch the globe.
- Places every event with its category's emoji; overlapping markers cluster.
- Names the nearest place to each event, worked out on your machine.
- Opens an event's detail: times, position, any measurement and a source link.
- Optional layers: EUMETSAT clouds, day and night with city lights, GWIS burnt areas.
- Draws each storm's track through the window.
- Filters by category, provider, smallest earthquake and a time window of 1 h to 7 days.
- Replays the chosen window at half, normal or double speed.
- Walks the events from the keyboard.
- Refreshes each source on its own schedule, retrying a failed one with a growing delay.
- Keeps each source's last good set on disk, so it opens offline and says how old it is.
- Marks a source stale rather than calling anything live; a status panel says why.
- Remembers your choices between runs.
- Carries a guide to every control under Help.

## What it does not do

- **No account and no telemetry.** EarthNow's own code contacts
  `earthquake.usgs.gov`, `eonet.gsfc.nasa.gov` and `volcano.si.edu`; while the
  clouds show, `view.eumetsat.int`; while the burnt areas show,
  `maps.effis.emergency.copernicus.eu`. Each source's client allows its own
  host alone over https, refusing a redirect anywhere else. A proxy named in
  `HTTPS_PROXY` is used. What the platform's web view itself contacts is its
  maker's.
- **No request from the page.** Every fetch is made by the Go side; the page's
  Content-Security-Policy allows no other origin.
- **No update check, no tray icon, no start with Windows.**
- **No administrator rights.** Setup installs for your account only on
  Windows; the Flatpak is a user install on Linux.

## Built with

| Part | Choice |
|---|---|
| Backend | Go |
| Desktop shell | Wails v2 over WebView2 (Windows), WKWebView (macOS) and WebKitGTK (Linux) |
| Front end | React and TypeScript, built with Vite |
| Globe | globe.gl on three.js |
| Place names | Natural Earth data, embedded |
| Setup program | a second Wails application |
| Build and gate | PowerShell (`build.ps1`, `test.ps1`); bash for the DMG and the Flatpak |

## Install and run

Every build is on the
[Releases page](https://github.com/oernster/EarthNow/releases). Each needs a
graphics driver offering WebGL2.

| Platform | Download | Install |
|---|---|---|
| Windows 10 and 11, x64 | `EarthNowSetup.exe` | Run it and choose Install. It needs the WebView2 runtime. |
| macOS on Apple Silicon | `EarthNow.dmg` | Open it and drag EarthNow to Applications. Signed and notarised. |
| Linux (tested on the latest Ubuntu LTS) | `earthnow.flatpak` | `flatpak install --user earthnow.flatpak` |

On Windows setup installs to `%LOCALAPPDATA%\Programs\EarthNow`. Run again, it
offers Update, Repair or Uninstall as the installed version warrants; the Apps
list opens the same program. Ticking **Also forget my settings and cached
events** on uninstall removes the data folder too.

| Platform | Settings, cache and log (`Log.txt`) |
|---|---|
| Windows | `%LOCALAPPDATA%\EarthNow`; the setup log is `%TEMP%\EarthNowSetup.log` |
| macOS | `~/Library/Caches/EarthNow` |
| Linux | `~/.var/app/uk.codecrafter.EarthNow/cache/EarthNow` |

On Linux `flatpak uninstall --user uk.codecrafter.EarthNow` removes the
application; `--delete-data` removes that folder as well.

## Build and test

```powershell
./build.ps1
```

```powershell
./test.ps1
```

The first runs the gate, then builds `build/bin/EarthNow.exe` and
`dist-installer/EarthNowSetup.exe`; the second is the gate alone. See
[DEVELOPMENT.md](DEVELOPMENT.md) and [TESTING.md](TESTING.md).

## Documents

- [ARCHITECTURE.md](ARCHITECTURE.md): the layers and the rules the tests enforce.
- [DEVELOPMENT.md](DEVELOPMENT.md): the tools, the build and the release steps.
- [TESTING.md](TESTING.md): the gate, the coverage floors and the manual checks.
- [REQUIREMENTS.md](REQUIREMENTS.md): the specification.
- [TECH_DEBT.md](TECH_DEBT.md): what is open and what only looks like debt.
- [DECISIONS-TRADEOFFS.md](DECISIONS-TRADEOFFS.md): the decisions with their gains and costs.

## Data sources and credits

- Earth imagery: NASA Earth Observatory (Blue Marble Next Generation).
- Night lights: NASA Earth Observatory (Black Marble 2016).
- Natural events: NASA Earth Observatory Natural Event Tracker (EONET).
- Earthquakes: USGS Earthquake Hazards Program.
- Volcanoes: Global Volcanism Program, Smithsonian Institution
  (https://volcano.si.edu/) with the USGS Volcano Hazards Program, Weekly
  Volcanic Activity Report.
- Place names and borders: Natural Earth.
- Cloud images: EUMETSAT, world cloud map (EUMETView).
- Burnt areas: © European Union, Global Wildfire Information System (GWIS),
  Copernicus Emergency Management Service, CC BY 4.0; redrawn over the globe.

None of NASA, the USGS, the Smithsonian, EUMETSAT or GWIS endorses EarthNow.

## Supporting the project

EarthNow is free and stays free. There is no paid tier, no licence key and no
feature held back; a donation supports its maintenance.

<a href="https://www.paypal.com/ncp/payment/9LWU8TKV2MSRE"><img src="docs/donate.png" alt="Donate to EarthNow" width="120"></a>

## Licence

EarthNow's own code is under the GNU General Public License, version 3; see
[LICENSE](LICENSE). The components and data it ships keep their own licences,
listed in full in [THIRD_PARTY_NOTICES](THIRD_PARTY_NOTICES). Both show under
Help.

A commercial licence is available: see
[commercial licensing](https://ernster.dev/commercial-licensing.html).

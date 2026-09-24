# EarthNow

Open a globe and see what is happening on Earth right now.

EarthNow is a Windows desktop application. It shows a slowly turning globe
carrying the natural events of the past week at the places they happened:
earthquakes from the USGS, natural events tracked by NASA EONET (wildfires,
storms, floods, sea ice and more) and the volcanoes named in the Smithsonian and
USGS Weekly Volcanic Activity Report. Every marker says which source reported it
and how old that report is.

## Who it is for

- Anyone curious about what the planet is doing this hour, this day or this
  week, who wants to see it on a globe rather than read it in a list.
- Anyone who wants each event traced to the public source that reported it,
  with its age stated plainly.
- Keyboard users: every control and every event on the globe can be reached
  without a pointer.

## Who it is not for

- Anyone who needs warnings. EarthNow sends no notifications and makes no
  forecast; it shows what the sources have already published.
- Anyone who wants history. The widest time window is seven days; there is no
  archive and no playback.
- Anyone after a GIS tool. There are no layers, measurements or projections,
  no satellite imagery and no weather.
- Anyone on Linux or macOS. EarthNow is built for Windows.

## What it does

- **Draws the globe** with NASA's Blue Marble imagery on a black background. It
  starts turning after ten seconds without input; a button or the settings stop
  it. The wheel, the zoom buttons and the plus and minus keys zoom.
- **Places every event** with its category's emoji: earthquake, volcano,
  wildfire, severe storm, flood, landslide, drought, dust, ice or other.
  Earthquakes are sized by magnitude. Markers that overlap on screen draw as
  one cluster with its count; activating a cluster zooms in until its members
  separate.
- **Names where it is.** Hovering a marker shows the event with its nearest
  populated place, its country, the distance and the compass direction, worked
  out on your own machine from embedded Natural Earth data.
- **Opens the detail** of an event: title, category, provider, position, the
  event time and the retrieval time (each as an age, in UTC and in local time),
  any measurement and a link to the source page, which opens in your browser.
- **Filters** by category and by provider from the key down the right side,
  which also counts the events shown in each category. The time window picks
  1 h, 6 h, 24 h, 3 days or 7 days.
- **Keeps itself fresh.** USGS is asked every minute, EONET every ten minutes
  and the volcano report every hour. A source that fails is retried with a
  growing delay while the others carry on. Refresh asks again at once, no more
  than once every 30 seconds.
- **Survives being offline.** Each source's last good set is kept on disk, so
  the globe opens on it and says how old it is.
- **Is honest about age.** Nothing is labelled live. A source that has not
  answered for three of its intervals is marked stale.

## What it does not do

- **No account and no telemetry.** EarthNow's own code contacts three hosts
  and no others: `earthquake.usgs.gov`, `eonet.gsfc.nasa.gov` and
  `volcano.si.edu`. The HTTP client refuses any other host. What the WebView2
  runtime itself contacts is Microsoft's.
- **No request from the page.** Every fetch is made by the Go side; the page's
  Content-Security-Policy allows it no other origin.
- **No update check, no tray icon, no start with Windows.**
- **No administrator rights.** Setup installs for your account only.

## Built with

| Part | Choice |
|---|---|
| Backend | Go |
| Desktop shell | Wails v2 over WebView2 |
| Front end | React and TypeScript, built with Vite |
| Globe | globe.gl on three.js |
| Place names | Natural Earth data, embedded |
| Setup program | a second Wails application with a hand-written page |
| Build and gate | PowerShell: `build.ps1` and `test.ps1` |

## Install and run

EarthNow targets Windows 10 and 11 on x64. It needs the WebView2 runtime and a
graphics driver offering WebGL2.

1. Download `EarthNowSetup.exe` from the
   [Releases page](https://github.com/oernster/EarthNow/releases).
2. Run it and choose Install.

Setup puts EarthNow in `%LOCALAPPDATA%\Programs\EarthNow` and offers a Start
Menu entry and a Desktop shortcut. Running it again reads the version already
installed: over an older one it offers Update; over a newer one, Go back; over
the same one, Repair and Reinstall. Each of those screens offers Uninstall too.
EarthNow is also listed under Settings, then Apps, whose Modify and
Uninstall open the same program.

| What | Where |
|---|---|
| The program | `%LOCALAPPDATA%\Programs\EarthNow` |
| Settings | `%LOCALAPPDATA%\EarthNow\settings.json` |
| Each source's last good set | `%LOCALAPPDATA%\EarthNow\cache` |
| The log | `%LOCALAPPDATA%\EarthNow\Log.txt`, with one previous file kept |
| The window's web view data | `%LOCALAPPDATA%\EarthNow\webview` |
| The setup log | `%TEMP%\EarthNowSetup.log` |

Uninstall removes the program, the shortcuts and the Apps list entry. Ticking
**Also forget my settings and cached events** removes `%LOCALAPPDATA%\EarthNow`
as well.

## Build and test

```powershell
./build.ps1
```

That runs the whole gate, then writes `build/bin/EarthNow.exe` and
`dist-installer/EarthNowSetup.exe`. The gate alone is `./test.ps1`.

## Documents

- [ARCHITECTURE.md](ARCHITECTURE.md): the layers, the rules the tests enforce
  and why the design is the shape it is.
- [DEVELOPMENT.md](DEVELOPMENT.md): the tools, the build, the generated files
  and the release steps.
- [TESTING.md](TESTING.md): the gate, the coverage floors and the checks a
  person makes.
- [REQUIREMENTS.md](REQUIREMENTS.md): the specification.
- [TECH_DEBT.md](TECH_DEBT.md): what is open and what only looks like debt.

## Supporting the project

The donate button sits at the foot of the rail down the left side. It hands a
PayPal page to your browser; EarthNow itself fetches nothing from that address.
The address lives once, in `internal/product`; a structural test holds it
there.

EarthNow is free and stays free. A donation supports its maintenance and
continued development. Nothing is held back without one: there is no paid
tier, no licence key and no feature a donation unlocks.

## Data sources and credits

The About dialog carries these credits:

- Earth imagery: NASA Earth Observatory (Blue Marble Next Generation).
- Natural events: NASA Earth Observatory Natural Event Tracker (EONET).
- Earthquakes: USGS Earthquake Hazards Program.
- Volcanoes: Global Volcanism Program, Smithsonian Institution
  (https://volcano.si.edu/) with the USGS Volcano Hazards Program, Weekly
  Volcanic Activity Report.
- Place names and borders: Natural Earth.
- Neither NASA, the USGS nor the Smithsonian endorses EarthNow.

## Licence

EarthNow's own code is distributed under the GNU General Public License,
version 3; see [LICENSE](LICENSE). The components and data it ships, each
under its own licence or terms, are listed with their full licence texts in
[THIRD_PARTY_NOTICES](THIRD_PARTY_NOTICES). The application shows both under
Help.

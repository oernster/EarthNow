# EarthNow

Open a globe and see what is happening on Earth right now.

> **Commercial licences available.** EarthNow is free and open source under the
> GNU General Public License, version 3. If those terms do not suit what you are
> building, such as a closed-source product, a commercial licence can be bought
> from me separately. It covers my own code; third-party libraries and data keep
> their own licences. See
> [commercial licensing](https://ernster.dev/commercial-licensing.html).

EarthNow is a desktop application for Windows, macOS and Linux. It shows a
slowly turning globe carrying the natural events of the past week at the places
they happened: earthquakes from the USGS, natural events tracked by NASA EONET
(wildfires, storms, floods, sea ice and more) and the volcanoes named in the
Smithsonian and USGS Weekly Volcanic Activity Report. Every event names the
source that reported it and how old that report is.

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
- Anyone who wants history. The widest time window is seven days. Replay plays
  that window back; there is no archive beyond it.
- Anyone after a GIS tool or a weather service. There are no measurements or
  projections. There are three layers. The clouds show where cloud lay when the
  image was made; the burnt areas show ground mapped as burnt; day and night
  shows where the sun is up now. There is no rain, wind, temperature or
  forecast.

## What it does

- **Draws the globe** with NASA's Blue Marble imagery on a black background. It
  opens over your own country, as your computer's country or region setting
  names it (read on your machine and sent nowhere), then turns slowly from
  there. It pauses while you use it. After ten seconds without input it
  first brings the whole globe back into view, zooming out or in as needed,
  then turns again. Switching rotation back on does the same; input during
  that return stops it where it is. A button or the settings stop it; the
  settings also
  offer three speeds. The wheel, the zoom buttons and the plus and minus keys
  zoom; Reset view brings the whole globe back.
- **Shows the clouds** when asked, from EUMETSAT's world cloud map: a mosaic of
  weather satellites' infrared images made every three hours, laid over the
  globe beneath the markers. The line beneath the globe gives the image's time
  in UTC with its age. A grey veil marks where no satellite sees. The layer
  starts hidden and is fetched only while it is shown.
- **Shows day and night** as the real sun lights the Earth now. The night side
  darkens and NASA's Black Marble city lights show through it; dawn and dusk
  fade across a twilight band that moves with the sun. The lights are a 2016
  composite, so they show where cities are rather than which lights are on.
  The sun's place is worked out on your machine from the time, with no
  request. Clouds over the night side are drawn fainter. The layer shows from
  the first run; a button hides it.
- **Shows burnt areas** in red, worldwide, from the Global Wildfire Information
  System (GWIS), which maps burnt ground from satellites one UTC day at a time.
  Every day the time window touches is drawn, so 7 days shows eight days of
  maps; a short window may have none mapped yet, which the line beneath the
  globe says. The layer lies beneath the clouds and the markers. It shows from
  the first run; Settings hides it. The maps are fetched only while shown and
  kept for use offline.
- **Places every event** with its category's emoji: earthquake, volcano,
  wildfire, severe storm, flood, landslide, drought, dust, ice or other.
  Earthquakes are sized by magnitude. Markers that overlap on screen draw as
  one cluster with its count; activating a cluster zooms in until its members
  separate. One still together at the closest zoom lists its members instead;
  choosing one opens its detail. A storm draws a faint track through the positions its source gave
  inside the time window, fading with age; Settings can hide it.
- **Names where it is.** Hovering a marker shows the event with its nearest
  populated place, its country, the distance and the compass direction, worked
  out on your own machine from embedded Natural Earth data.
- **Opens the detail** of an event: title, category, provider, position, the
  event time and the retrieval time (each as an age, in UTC and in local time;
  an event its source dates by day alone has no local time), any measurement,
  an earthquake's depth with USGS's shallow, intermediate or deep band, then a
  link to the source page, which opens in your browser. An event its
  source has closed is marked as ended.
- **Walks the events from the keyboard:** Up and Down move between them, Enter
  opens one.
- **Filters** by category and by provider from the key down the right side,
  which also counts the events shown in each category. The time window picks
  1 h, 6 h, 24 h, 3 days or 7 days. An event dated more than 15 minutes ahead
  of your clock waits until its time arrives.
- **Replays the window.** Play, beside the time window, plays the chosen
  window from its start in 30 seconds at normal speed: events appear at their
  own times, storm tracks grow, the sun sweeps round, the burnt days build up
  and the clouds move. The slider beside it moves through the window by hand.
  The replay holds at the end of the window until Now returns to the
  present; the speed button beside Now plays at half, normal or double speed.
  Replay's clouds are smaller images fetched for it, held in memory and
  let go when it ends. It shows what the sources hold now about those days,
  which may since have been revised.
- **Remembers your choices:** the time window, the filters, rotation with its
  speed, the replay's speed, the smallest earthquake shown (every one, else 1.0, 2.5, 3.0 or
  4.5 and above) plus whether the clouds, day and night, storm tracks and
  burnt areas show.
- **Keeps itself fresh.** USGS is asked every minute, EONET every ten minutes
  and the volcano report every hour. A source that fails is retried with a
  growing delay while the others carry on. Refresh asks again at once, no more
  than once every 30 seconds; its icon turns and the status line says which
  sources are refreshing until they answer, then the time of the last refresh.
- **Survives being offline.** Each source's last good set is kept on disk, so
  the globe opens on it and says how old it is.
- **Is honest about age.** Nothing is labelled live. A source that has not
  answered for three of its intervals is marked stale. The provider status
  panel gives each source's state and the reason for any failed fetch.
- **Explains itself.** Help opens a guide to every control, the About details,
  the licence and the third-party notices.

## What it does not do

- **No account and no telemetry.** EarthNow's own code contacts three hosts:
  `earthquake.usgs.gov`, `eonet.gsfc.nasa.gov` and `volcano.si.edu`. While
  the clouds are shown or a replay fetches its clouds it also contacts
  `view.eumetsat.int`; while the burnt areas are shown,
  `maps.effis.emergency.copernicus.eu`. It contacts no others: the HTTP client
  refuses any other host. What the platform's web view itself contacts
  (WebView2, WKWebView or WebKitGTK) is its maker's.
- **No request from the page.** Every fetch is made by the Go side; the page's
  Content-Security-Policy allows it no other origin.
- **No update check, no tray icon, no start with Windows.**
- **No administrator rights.** On Windows, setup installs for your account
  only; on Linux the Flatpak is a user install.

## Built with

| Part | Choice |
|---|---|
| Backend | Go |
| Desktop shell | Wails v2 over WebView2 (Windows), WKWebView (macOS) and WebKitGTK (Linux) |
| Front end | React and TypeScript, built with Vite |
| Globe | globe.gl on three.js |
| Place names | Natural Earth data, embedded |
| Setup program | a second Wails application with a hand-written page |
| Build and gate | PowerShell: `build.ps1` and `test.ps1`; bash for the DMG and the Flatpak |

## Install and run

Every build is on the
[Releases page](https://github.com/oernster/EarthNow/releases). Each needs a
graphics driver offering WebGL2.

| Platform | Download | Install |
|---|---|---|
| Windows 10 and 11, x64 | `EarthNowSetup.exe` | Run it and choose Install. It needs the WebView2 runtime. |
| macOS on Apple Silicon | `EarthNow.dmg` | Open it and drag EarthNow to Applications. The DMG is signed and notarised. |
| Linux (tested on the latest Ubuntu LTS) | `earthnow.flatpak` | `flatpak install --user earthnow.flatpak`, which fetches the GNOME runtime from Flathub. |

### Windows

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
| Each source's last good set, the last cloud image and the burnt-area maps | `%LOCALAPPDATA%\EarthNow\cache` |
| The log | `%LOCALAPPDATA%\EarthNow\Log.txt`, with one previous file kept |
| The window's web view data | `%LOCALAPPDATA%\EarthNow\webview` |
| The setup log | `%TEMP%\EarthNowSetup.log` |

Uninstall removes the program, the shortcuts and the Apps list entry. Ticking
**Also forget my settings and cached events** removes `%LOCALAPPDATA%\EarthNow`
as well.

### macOS and Linux

On macOS the settings, cache and log live in `~/Library/Caches/EarthNow`;
removing EarthNow is dragging it out of Applications. On Linux they live in
`~/.var/app/uk.codecrafter.EarthNow/cache/EarthNow`;
`flatpak uninstall --user uk.codecrafter.EarthNow` removes the application and
leaves that folder, which `--delete-data` removes as well.

## Build and test

```powershell
./build.ps1
```

That runs the whole gate, then writes `build/bin/EarthNow.exe` and
`dist-installer/EarthNowSetup.exe`. The gate alone is `./test.ps1`. The macOS
and Linux builds are `builddmg.sh` and `build_flatpak.sh`, run on those
platforms; [DEVELOPMENT.md](DEVELOPMENT.md) covers them.

## Documents

- [ARCHITECTURE.md](ARCHITECTURE.md): the layers, the rules the tests enforce
  and why the design is the shape it is.
- [DEVELOPMENT.md](DEVELOPMENT.md): the tools, the build, the generated files
  and the release steps.
- [TESTING.md](TESTING.md): the gate, the coverage floors and the checks a
  person makes.
- [REQUIREMENTS.md](REQUIREMENTS.md): the specification.
- [TECH_DEBT.md](TECH_DEBT.md): what is open and what only looks like debt.

## Data sources and credits

The About dialog carries these credits:

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
- None of NASA, the USGS, the Smithsonian, EUMETSAT or GWIS endorses EarthNow.

## Supporting the project

The donate button sits at the foot of the rail down the left side. It hands a
PayPal page to your browser; EarthNow itself fetches nothing from that address.
The address lives once, in `internal/product`; a structural test holds it
there.

EarthNow is free and stays free. A donation supports its maintenance and
continued development. Nothing is held back without one: there is no paid
tier, no licence key and no feature a donation unlocks.

<a href="https://www.paypal.com/ncp/payment/9LWU8TKV2MSRE"><img src="docs/donate.png" alt="Donate to EarthNow" width="120"></a>

## Licence

EarthNow's own code is distributed under the GNU General Public License,
version 3; see [LICENSE](LICENSE). The components and data it ships, each
under its own licence or terms, are listed with their full licence texts in
[THIRD_PARTY_NOTICES](THIRD_PARTY_NOTICES). The application shows both under
Help.

A commercial licence for EarthNow's own code is available: see
[commercial licensing](https://ernster.dev/commercial-licensing.html).

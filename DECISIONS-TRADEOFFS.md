# Decisions and trade-offs

The deliberate choices EarthNow rests on: what was chosen, what was given up
for it and why. Each entry is the decision as the product makes it today.
The detail behind each one, with the tests that hold it, lives in
[ARCHITECTURE.md](ARCHITECTURE.md) and the specification
([REQUIREMENTS.md](REQUIREMENTS.md)), whose Won't list records what is
deliberately not planned; [TECH_DEBT.md](TECH_DEBT.md) holds what only looks
like debt.

## The product as a whole

### One sentence settles every unclear choice

EarthNow lets you open a globe and see what is happening on Earth right now.
Where a choice is unclear, that sentence decides it.

- **Rather than:** a feature list grown one request at a time.
- **Gains:** a calm globe stays the centre; a proposal that does not serve the
  sentence has an answer before it is argued.
- **Costs:** useful neighbouring features are turned away.

### What EarthNow deliberately is not

It shows what public sources have already published. It sends no warnings,
makes no forecast, keeps no archive beyond a week and offers no weather beyond
its three layers. It is not a GIS tool: no measurement, no projections, no
layer management.

- **Rather than:** an alerting service, a weather service or a GIS engine.
- **Gains:** a small surface that can be held to a high bar.
- **Costs:** those jobs need other tools.

### A week at most

The widest time window is seven days; Replay plays back the chosen window and
nothing older. The cache keeps only what some window can still show.

- **Rather than:** a history to browse; a replay beyond the window or saved as
  a video.
- **Gains:** each source's own weekly feed covers everything shown; the cache
  stays small.
- **Costs:** an event a week old is gone.

### Specification before code

Every feature area is written down as requirements before it is built. Each
change of course is a numbered amendment saying why. A structural test fails
when a Must is named by no test and is not listed for a person to check.

- **Rather than:** building first and describing afterwards.
- **Gains:** a ruled-out idea stays ruled out instead of being argued again;
  every Must has a named way of being verified.
- **Costs:** keeping the specification true to the code is work of its own,
  done at every documentation pass.

## Privacy and the network

### Every request is made by the Go side

The page makes no request of its own. Its Content-Security-Policy allows it
no origin but its own, which a structural test holds. Images the page draws
arrive from Go as data.

- **Rather than:** letting the page fetch, as the globe library's own examples
  do.
- **Gains:** one place to read to know what EarthNow contacts.
- **Costs:** every image crosses from Go to the page as encoded text.

### One client, a list of hosts and a size cap

All network traffic goes through one client. It reaches only the hosts named
in the composition root, follows a redirect only to one of them, gives up on a
request after 30 seconds and refuses a body over 16 MB (the largest feed
measured was 1.51 MB).

- **Rather than:** the standard client, which follows a redirect anywhere.
- **Gains:** the README's list of hosts is a property of the code; an allowed
  host answering with a redirect cannot carry the request elsewhere (a case
  that was reproduced before it was closed).
- **Costs:** a new source is an edit to that list as well as an adapter.

### Two layers ask only while shown

The cloud layer reaches EUMETSAT only while it is shown or while a replay
needs its clouds; the burnt-area layer reaches GWIS only while it is shown.
Hidden, each asks nothing at all. The clouds start hidden.

- **Rather than:** fetching every layer in the background whether shown or
  not.
- **Gains:** hiding a layer means nothing is sent to its source.
- **Costs:** the burnt areas and day and night start shown, so a first run
  does contact GWIS.

### No account, no telemetry, no update check

There is nothing to sign in to and nothing reports on use. There is no update
check, no tray icon and no start with Windows.

- **Rather than:** the conveniences other desktop applications offer.
- **Gains:** nothing about the viewer leaves the machine beyond the requests
  the sources need.
- **Costs:** a new release is found only by visiting the site; there are no
  usage figures to steer development.

### Places are named on the machine

The nearest place, its country, the distance and the direction come from
Natural Earth data built into the application. A structural test fails if the
geocoder imports any network package.

- **Rather than:** a geocoding service.
- **Gains:** hovering over a marker sends nothing; it works offline.
- **Costs:** the place data is fixed at build time. Refreshing it is a hand
  step with a generator; the source shapefiles are not kept.

### The globe opens over the viewer's own country

At launch the globe faces the country the operating system's country or
region setting names, at Natural Earth's label point for it, then turns from
there. The setting is read on the machine and sent nowhere.

- **Rather than:** the time zone, which knows only a band of longitude; the
  display language, which on a UK machine is often American English; a
  capital or an outline's centre, which can sit at an edge; a home location
  set in the application.
- **Gains:** the right view with no question asked.
- **Costs:** where the setting names no country the globe opens as it always
  did; the log says why.

### Donations go through the browser

The donate address is held once, on the Go side. It is opened in the default
browser through the same check as a source link: https, a host and a page
rather than a data file. EarthNow fetches nothing from it; nothing depends on
a donation.

- **Rather than:** the page holding the address.
- **Gains:** one home for it, which a structural test holds; no connection
  from inside the application.
- **Costs:** EarthNow never learns what happened next.

### One data folder

Settings, the cache and the log live in one folder under the user's local
application data. The web view's own data is placed there too.

- **Rather than:** the web view's default, which fell outside that folder
  (measured).
- **Gains:** removing that one folder removes everything EarthNow wrote,
  which is what setup's "forget" box does.
- **Costs:** none recorded.

## The sources

### Three public sources, none needing a key

Earthquakes come from the USGS, other natural events from NASA EONET and
volcanoes from the Smithsonian and USGS Weekly Volcanic Activity Report. The
volcano report was added when EONET was found to track none.

- **Rather than:** EONET alone, which tracked no volcano in the 30 days
  measured while that week's report listed 20.
- **Gains:** volcanoes are on the globe; nothing to sign up for or leak.
- **Costs:** a third schema to own, an XML feed in an old encoding.

### Every EONET event of the week, open and closed

EONET is asked for closed events as well as open ones; a closed event is
marked as ended in its detail.

- **Rather than:** open events only: 17 against 80 that week, dropping most
  wildfires and floods.
- **Gains:** the week as it happened.
- **Costs:** some markers are for events that have already ended.

### The earthquake floor starts at 2.5

The smallest earthquake shown starts at magnitude 2.5; Settings offers every
one, 1.0, 3.0 and 4.5 as well.

- **Rather than:** 3.0, the first default: 243 quakes in the week measured
  against 362 at 2.5. Every magnitude, which costs the most frame time.
- **Gains:** more of the world's activity at a count the globe draws smoothly.
- **Costs:** small local quakes are hidden until asked for.

### Each source on its own clock

USGS is asked every minute, matching its feed's measured cache lifetime;
EONET every ten minutes; the volcano report every hour. A failed source is
retried with a delay doubling up to 30 minutes while the others carry on.

- **Rather than:** one interval for all; retrying at a fixed pace.
- **Gains:** each source is asked about as often as it changes; a source that
  is down is not hammered.
- **Costs:** a source that comes back may wait up to half an hour to be asked
  again.

### Refresh is limited and says when it last ran

Refresh asks every source at once, no more than once in 30 seconds. After a
press the status line gives the time of the last refresh made. While a
source is fetching, the line says so and the button turns.

- **Rather than:** saying whether a refresh is available. Worded in whole
  minutes, every wait inside the cooldown read as "now" and never cleared.
- **Gains:** the line is always true; a refresh that worked no longer looks
  like one that did nothing.
- **Costs:** pressed at every chance, Refresh asks EONET up to 120 times an
  hour; the owner accepted that.

### Each adapter says what it accepts

Each adapter states the media types it asks for; no parser trusts the type
a source labels its answer with.

- **Rather than:** one request header for every source.
- **Gains:** every source answers. The Smithsonian's feed refuses a request
  asking only for JSON (measured); EONET labels its JSON as RSS.
- **Costs:** none recorded.

### The last good set kept as a file per source

Each source's last successful set is kept as one JSON file, written beside
the old one and then renamed over it, stamped with a schema version. At start
the globe opens on these sets, each marked with its age.

- **Rather than:** a database. The sets are small and read whole.
- **Gains:** an interrupted write leaves the previous set intact; the globe
  survives being offline; a file from another schema version is simply read
  as absent.
- **Costs:** a whole file is rewritten after every fetch.

### A failed source costs only itself

A source that fails keeps its last set on the globe while the others are
refreshed as normal. One not heard from for three of its intervals is marked
stale.

- **Rather than:** clearing a failed source's events; one failure stopping
  the round.
- **Gains:** a source that is down hides nothing the others report.
- **Costs:** old events can be on screen; the status says how old.

### Nothing from the future

An event dated more than 15 minutes ahead of the machine's clock waits until
its time arrives.

- **Rather than:** showing any date the source gives. A flood alert arrived
  dated 11 days ahead and showed as recent.
- **Gains:** only what has happened is shown; a slightly slow clock still
  hides nothing new.
- **Costs:** a source whose own clock runs well ahead loses those events for
  a while.

### Volcanoes are ongoing

A volcano in the current weekly report counts in every window from the first
day of the week the report covers, until the report is more than 14 days old.
The status says when it is too old to show.

- **Rather than:** dating each volcano by the report's issue day. Once that
  day left the window every window showed 0 volcanoes against the feed's 20.
- **Gains:** an erupting volcano is shown as erupting.
- **Costs:** a volcano stays on the globe for the report's whole life, however
  its week ended.

### Kinds no source publishes fold into Other

The key has seven categories. Landslides, drought and dust haze are filed
under Other, keeping the source's own kind on the event.

- **Rather than:** a key row of their own, each reading 0 for ever since
  EONET published none in a year; rows shown only when filled, which would
  change the key's shape.
- **Gains:** a key whose every row can hold something.
- **Costs:** a landslide, should one come, is filed under Other.

### No merging across sources

An event reported by two sources would show as two markers.

- **Rather than:** a rule for spotting duplicates. EONET holds 15 earthquakes
  in its whole history and none from the last year, so the case did not arise
  when measured.
- **Gains:** no matching rule to get wrong.
- **Costs:** if it ever arises, the same event shows twice.

### No common severity scale

A measurement is shown in its source's own unit. Nothing ranks a quake
against a storm. USGS's tsunami flag is kept but never worded, since it marks
large oceanic events rather than a tsunami; a structural test holds that.

- **Rather than:** one severity scale across categories.
- **Gains:** no comparison is invented; nothing claims what the source does
  not.
- **Costs:** the viewer compares events by reading them.

## Reading the sources faithfully

### Flood polygons read in the order the data proves

GDACS flood polygons arrive latitude first, against the GeoJSON standard. A
ring holding a value beyond 90 is read in the order that value proves; the
rest follow the order the feed's proving rings show, latitude first when none
proves anything.

- **Rather than:** GeoJSON order, which dropped five floods that week and drew
  nine in the wrong place (Honduras in Antarctica); a fixed latitude-first
  rule, which would draw every flood swapped once the source is corrected;
  guessing by which reading lands on a country, which put a coastal Kenya
  flood in Spain.
- **Gains:** floods are placed correctly now and follow a correction upstream
  as soon as it appears. Over 30 days, 15 of 57 rings proved their order.
- **Costs:** a feed whose rings prove nothing is read latitude first.

### A polygon across the date line stays in the Pacific

An event's position from a polygon is the mean of its vertices with
longitudes measured the short way from the first one. Every vertex is checked
before it counts.

- **Rather than:** a plain mean, which put a ring across the date line on the
  far side of the Earth.
- **Gains:** a ring that does not cross the line keeps its plain mean exactly.
- **Costs:** none recorded.

### An ice shelf is Antarctica

Natural Earth's Antarctic ice shelves are embedded beside the countries, so a
point on the Ross or Ronne shelf names Antarctica. A point at exactly 180
degrees east is looked up at 180 west, where the split polygon carries it.

- **Rather than:** the countries alone, which draw Antarctica only to its
  grounded coast, so a shelf read as open sea.
- **Gains:** a place line that matches where the event is.
- **Costs:** one more layer embedded.

### A source link is a page

The detail panel links the first source that names a page. A data file is
shown as text rather than as a link.

- **Rather than:** the first source given. For a storm that was a warning
  file, which downloaded instead of opening.
- **Gains:** a link opens something to read.
- **Costs:** a source that is only a data file cannot be opened from
  EarthNow.

### Two lost characters put back, nothing else

The volcano report's encoding holds neither the curly apostrophe nor the
subscript two, so it sends a question mark for each. A letter's "?s" is read
as an apostrophe and "SO?" as sulphur dioxide; every other question mark is
kept as sent.

- **Rather than:** showing the text as sent, which reads as a fault in
  EarthNow; replacing every question mark, which would lose real ones.
- **Gains:** the report reads as it was written.
- **Costs:** another character lost the same way would still show as a
  question mark.

### Earthquake depth with its meaning

An earthquake's detail gives its depth to one decimal with USGS's band:
shallow, intermediate or deep. A depth of exactly 10 km is marked as often a
fixed value, since USGS assigns it when it cannot compute one.

- **Rather than:** a bare number.
- **Gains:** the most common depth is not mistaken for a measurement.
- **Costs:** a quake genuinely at 10 km carries the note too.

## Honesty about age

### Nothing is labelled live

Every event, layer and source gives its age. No wording anywhere uses the
word "live"; a structural test holds that.

- **Rather than:** the "live" badge such displays usually carry.
- **Gains:** nothing claims a freshness the sources do not give.
- **Costs:** more words on screen.

## The layers

### Images are made in Go

The cloud image and the burnt areas are fetched, checked and drawn in Go,
then handed to the page as finished pictures laid on spheres over the globe.

- **Rather than:** drawing them on the page, which would have to fetch them.
- **Gains:** the page keeps its rule of no network origin; the drawing rules
  sit in the domain, where they are tested.
- **Costs:** each image crosses to the page as encoded text.

### The newest cloud image found cheaply

The cloud layer reads its own layer's capabilities document for the newest
image time about once an hour. It fetches an image only when that time is
new.

- **Rather than:** the whole service's document: 282 KB against 6.4 KB
  (measured).
- **Gains:** most checks cost a few kilobytes.
- **Costs:** a new image can wait up to an hour to appear.

### A map answer must be the picture asked for

Both map sources must answer with a PNG of exactly the size asked for, its
size read before the pixels are decoded. Anything else is a failed fetch.

- **Rather than:** trusting the status code. The cloud service reports
  errors as XML with a success status.
- **Gains:** an error page is never drawn as cloud.
- **Costs:** none recorded.

### Burnt areas a day at a time

GWIS is asked for one UTC day per request, for every day the window touches.
Each day is held apart; Go composes the days into one image laid beneath the
clouds. A day that fails keeps its held image while the others draw.

- **Rather than:** a date range in one request, which GWIS answers with an
  empty body (measured); one image per day on the page, up to eight textures
  decoded there.
- **Gains:** a widened window fetches only the days it lacks; one failed day
  costs one day.
- **Costs:** up to eight requests a round.

### The sun worked out from the time

The sun's position comes from NOAA's solar position equations, written in
the Go domain; the page asks for it once a minute while the layer shows.
Tests hold it within 0.1 degrees of NOAA's own calculator.

- **Rather than:** the npm package the globe library's day and night example
  uses, which would put astronomy on the page; that example also fetches its
  textures from a network address the page may not reach.
- **Gains:** no request; one home for the figures.
- **Costs:** the equations are EarthNow's own to maintain.

### Day and night lights the globe's own material

The globe keeps the library's own lit material; NASA's night lights are added
to it and one light factor per point dims the day and reveals the lights.
Hidden, the factor is 1 everywhere, so the globe draws exactly as before.
Clouds over the night side dim to a floor.

- **Rather than:** the example's own unlit shader, which would change the day
  side and leave hiding the layer unable to restore the globe.
- **Gains:** switching the layer off truly switches it off.
- **Costs:** the light is injected into the rendering library's shader
  source; a test fails if a part it relies on is renamed.

### The Earth is never drawn

Every picture of the Earth is NASA imagery, in the application and on the
website.

- **Rather than:** generated artwork. The site once showed a painted globe
  with invented markers.
- **Gains:** no fabricated record of the planet anywhere.
- **Costs:** the imagery ships with the application; its credits must be
  kept.

## The globe and its markers

### globe.gl rather than CesiumJS

The globe is globe.gl on three.js with the Blue Marble texture bundled.

- **Rather than:** CesiumJS: several times the shipped size, default imagery
  from a network service needing an evaluation token and a GIS engine where
  one calm globe is wanted.
- **Gains:** a small, offline globe that does what the product needs.
- **Costs:** no clustering built in; EarthNow writes its own.

### Emoji drawn as sprites, one table their home

Each category is drawn as its emoji on a sprite; categories differ by emoji,
never by colour alone. The emoji are written in one table that the key,
markers, clusters and tooltips all read.

- **Rather than:** thousands of page elements moved every frame.
- **Gains:** smooth: the spike measured 2,501 sprites at a median frame of
  10.00 ms.
- **Costs:** each picture is drawn and cached by the page itself.

### Markers keep their size on screen

Every marker keeps the size it has at launch whatever the zoom.

- **Rather than:** a fixed size on the globe, which grows on screen as the
  camera nears. Overlapping markers would then never separate.
- **Gains:** zooming in pulls neighbours apart.
- **Costs:** zooming in never makes a marker bigger.

### Overlaps become clusters; a cluster that cannot part becomes a list

Markers that overlap on screen draw as one cluster with its count, wearing
its leading category's emoji. Activating it zooms in until its members stand
apart. One still together at the closest zoom lists its members instead.

- **Rather than:** a library's clustering, of which globe.gl offers none;
  fanning the markers out, where the owner chose the list; zooming alone,
  which left two quakes 2.0 km apart unreachable by pointer.
- **Gains:** every event can be opened.
- **Costs:** clustering is EarthNow's own code to maintain.

### Markers drawn whole at the edge

Markers are drawn over the globe rather than tested against its depth; before
each frame those beyond the horizon are hidden.

- **Rather than:** the depth test. A sprite always faces the camera, so near
  the edge it stood upright and sank half into the sphere.
- **Gains:** a marker reaching the edge is drawn whole until it passes
  behind.
- **Costs:** a check over every marker on every frame.

### Rotation follows its own switch

The globe turns after ten seconds without input, first easing back to the
whole-globe view if zoomed. The operating system's reduced-motion setting
does not stop it; the application's own setting does.

- **Rather than:** following the system setting, which on Windows follows
  the animation effects switch people turn off for speed; turning at whatever
  zoom was left.
- **Gains:** the feature is not silently lost; the turning globe is always the
  whole globe.
- **Costs:** a viewer who wants stillness turns rotation off in EarthNow; a
  close look is undone after ten idle seconds.

### A missing WebGL2 is said, not shown blank

Without WebGL2 the globe area says so in plain words and the rest of the
window keeps working.

- **Rather than:** a blank window.
- **Gains:** the viewer knows what is missing.
- **Costs:** no fallback globe.

### Only storms leave a track

A severe storm draws a faint track through the positions its source gave
inside the window, oldest faintest, ending at its marker. Settings can hide
the tracks.

- **Rather than:** tracks for iceberg drift and earthquake swarms, for which
  no rule is agreed.
- **Gains:** a storm's path reads at a glance without crowding the globe.
- **Costs:** none recorded.

## Replay

### Replay plays the chosen window, as the sources hold it now

Play replays the chosen window in 30 seconds at normal speed: events appear,
tracks grow, the sun sweeps round, the burnt days build up and the clouds
move. Half and double speed are offered.

- **Rather than:** a history of its own, kept by EarthNow.
- **Gains:** nothing extra is stored; Replay needs no archive.
- **Costs:** it shows what the sources hold now about those days, which may
  since have been revised.

### The page keeps the clock

The page holds the replay's position and plays it on animation frames. It
asks Go for a frame at most every 100 ms and for an image only when the frame
names a new one.

- **Rather than:** Go driving the clock and pushing frames.
- **Gains:** no timer on the Go side for what is an animation; the page asks
  at its own pace.
- **Costs:** none recorded.

### Replay's clouds are smaller and kept in memory

A replay fetches its three-hourly cloud images at half the live image's width
and height, holds them in memory and lets them go when the replay ends. Play
does not wait for them.

- **Rather than:** full-size images; keeping them on disk.
- **Gains:** each image is a quarter of the pixels; nothing is left behind on
  disk.
- **Costs:** coarser clouds while replaying; the next replay fetches them
  again.

### Time travel is explicit

A replay holds at the end of its span. Only Now or another time window
returns to the present. Now is shown only while replaying and sits last in
the row.

- **Rather than:** drifting back to the present on its own; a Now button
  shown disabled, which wore a permanent red ring.
- **Gains:** the viewer always knows which time is on screen; Now coming and
  going moves no other control.
- **Costs:** one more press to get back.

## The interface

### An action rail down the left

The actions sit in a rail down the left side, 68 pixels wide; the key sits
on the right. The globe area keeps at least 70% of the window, which a test
checks.

- **Rather than:** full-width bars. At the minimum window when this was
  decided they would have had 55 pixels of height.
- **Gains:** the globe keeps the room it needs.
- **Costs:** the actions are icons, explained by tooltips and the guide.

### One dark palette

The main window has one dark palette, its contrast checked by test. The
setup program keeps its light and dark toggle.

- **Rather than:** a light theme for the main window.
- **Gains:** one palette to hold to its contrast; the Earth on black.
- **Costs:** no light option.

### Everything from the keyboard, the globe included

Every control is on the keyboard ring. The globe is a stop of its own: Up and
Down walk the events, Enter opens one. At launch the application focuses its
web view directly and the page asks again if it finds no keyboard.

- **Rather than:** mouse-first controls; trusting the window framework alone
  to hand over the keyboard, which lost a race at the first focus (seen in
  the log).
- **Gains:** the whole application works without a pointer from the moment it
  opens.
- **Costs:** every new control needs its place in the ring.

### Reading dialogs read themselves

Long help pages scroll gently on their own and stop the moment the reader
takes over. The machine is shared with the setup program. Like rotation, it
is not gated on reduced motion.

- **Rather than:** static pages.
- **Gains:** long text can be read hands free.
- **Costs:** none recorded.

### Source text is shown as text

The page never sets markup from a string, so whatever a feed sends is shown
as plain text. A structural test holds it.

- **Rather than:** rendering formatting a feed might carry.
- **Gains:** a feed cannot put markup into the window.
- **Costs:** any formatting in a feed is lost.

## Building and installing

### Installed for one user, without administrator rights

On Windows the setup program installs into the user's own folders and
registry. On Linux the Flatpak is a user install that asks for no file system
access.

- **Rather than:** a machine-wide install.
- **Gains:** no administrator prompt.
- **Costs:** each account on a machine installs separately.

### A setup program of its own

Install, update, going back, repair and removal are one bespoke program. It
reads the machine once to choose its route, refuses to touch a file while
EarthNow is running and refuses any archive entry that would land outside the
install folder. Its install rules live in a Go package; the program itself is
a facade, which a structural test holds.

- **Rather than:** a generic installer; designing one afresh rather than
  porting the house setup program.
- **Gains:** one identity throughout; a locked executable never leaves an
  install half done.
- **Costs:** the setup program is EarthNow's own to maintain.

### The gate cannot be skipped

The build runs the whole test gate first, with no switch to skip it. cgo is
pinned off for the gate and the Windows build.

- **Rather than:** a skip switch, which is used on the day it would have
  caught something; leaving cgo to the machine, where one with a C compiler
  would quietly build a different binary.
- **Gains:** every build comes from a tree that passed.
- **Costs:** every build waits for the whole suite.

### Each platform builds on itself

The Windows setup program, the macOS DMG and the Linux Flatpak are each built
on their own platform. The DMG is signed and notarised. On Linux the web
view's GPU use is switched on explicitly.

- **Rather than:** cross-compiling; leaving the Linux GPU setting at the
  framework's default, which turns acceleration off and leaves the globe no
  WebGL.
- **Gains:** each package is built by the tools that know that platform; the
  globe draws on Linux.
- **Costs:** a machine of each kind to build on; an Apple developer account
  for the signing.

### Notices generated from what ships

The third-party notices are written by a tool from the Go modules and page
packages actually shipped, each licence in full. The gate fails until the
file matches.

- **Rather than:** notices kept by hand.
- **Gains:** a dependency added or bumped cannot ship without its notice.
- **Costs:** the tool must be run after every dependency change.

### GPL plus a commercial licence

EarthNow's own code is GPL-3.0. A commercial licence for that code is offered
separately; third-party libraries and data keep their own terms.

- **Rather than:** one licence only.
- **Gains:** the source stays open; closed-source use has a route.
- **Costs:** two sets of terms to explain.

### A log that keeps every run

The log keeps earlier runs and rotates at 5 MB while running, keeping one
previous file. The application's first act points its error output at the
log, so a crash leaves a record. The build's own run of the program writes
nothing there.

- **Rather than:** rotating only at start; EarthNow runs for days.
- **Gains:** faults can be read after the fact; a build never adds stray
  lines to a user's log.
- **Costs:** none recorded.

### The site's stylesheet is named by its content

Each local stylesheet and script on the site is linked with a hash of its
content. The version is stamped into the site from the one version file.

- **Rather than:** plain links, which let a browser pair a new page with a
  cached old stylesheet.
- **Gains:** a change to the site is seen at once.
- **Costs:** the stamp must run before a release, which the build does.

## Engineering

### Layers with one place where they meet

The code is split into domain, application, infrastructure and interface,
each allowed to depend only inward. Only two root files join the application
to the infrastructure. A new source is a package plus a line in the
composition root; a test fails if anything else names it.

- **Rather than:** convention alone.
- **Gains:** the rules about events, time and the sun are tested with no disk,
  network, clock or screen.
- **Costs:** more packages and more explicit wiring.

### Complete coverage where it means something

The domain and the application together must reach 100% coverage. Every
other gated package holds the figure it measured, raised as cover rises. The
page is measured with istanbul.

- **Rather than:** one figure over the whole program; v8 coverage, which
  reported the globe component at 100% with no test importing it.
- **Gains:** anything short of complete in the pure layers is a decision
  nobody made.
- **Costs:** code that acts on the machine (the registry, shortcuts, window
  focus) relies on targeted tests and a person's checks.

### Small modules

No source file may exceed 400 lines; one between 381 and 400 is cut to 350
or fewer. The page's source and the setup page are counted too.

- **Rather than:** letting files grow. A stylesheet reached 402 lines unseen
  while the rule skipped the page.
- **Gains:** modules split at real seams.
- **Costs:** many small files.

### Every value has one home

The product's name, the donate address, the version and each category's
emoji are each written in one place; tests hold the name, the address and
the emoji there. Everything else reads them.

- **Rather than:** copies written where they are needed.
- **Gains:** a change is made once and cannot drift.
- **Costs:** static files such as the website have to be stamped or generated
  from the source.

### The wire is checked from both sides

Every shape that crosses from Go to the page is compared field by field with
its TypeScript twin. Every call the page makes to Go must name what happens on
refusal; a call that does not will not compile.

- **Rather than:** trusting the two sides to agree.
- **Gains:** a renamed field fails the suite rather than the screen; no failed
  call goes unhandled.
- **Costs:** a new shape is declared three times: in Go, in TypeScript and in
  the test's list.

### Use cases hold no timers

The scheduler, the layers and Replay's cloud fetcher hold no timer. The
facade asks what is due and when to wake.

- **Rather than:** timers inside the use cases.
- **Gains:** every timing rule runs on a fake clock in its tests.
- **Costs:** the facade carries the loop that does the waiting.

### Background work cannot take the application down

Every goroutine the application starts recovers from a panic, logs the stack
and tells the page.

- **Rather than:** fire-and-forget goroutines.
- **Gains:** one failing fetch cannot end the run.
- **Costs:** more ceremony around background work.

### Tests with real parts and no network

The suite never reaches a source: adapters read feeds captured from the real
services or images made in the test. Fakes are written by hand. Each guard is
proved by planting a violation and watching it fail.

- **Rather than:** live calls in tests; mocks and assumed guards.
- **Gains:** the suite gives the same answer offline; a guard is known to
  bite.
- **Costs:** captured feeds age; the fakes are EarthNow's own to maintain.

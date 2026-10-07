# Decisions and trade-offs

The deliberate choices EarthNow rests on: what was chosen, what was given up
for it, what it gains and what it costs, as the product stands today. The
detail lives in [ARCHITECTURE.md](ARCHITECTURE.md) and
[REQUIREMENTS.md](REQUIREMENTS.md), whose Won't list records what is
deliberately not planned; [TECH_DEBT.md](TECH_DEBT.md) holds what only looks
like debt.

## The product as a whole

### One sentence decides; what EarthNow is not

EarthNow lets you open a globe and see what is happening on Earth right now.
Where a choice is unclear, that sentence decides it: no warnings, no forecast,
no weather beyond its layers, no GIS tools. A feature list grown one request
at a time was rejected. The surface stays small enough to hold to a high bar;
neighbouring jobs need other tools.

### A week at most

The widest time window is seven days. Replay plays back the chosen window and
nothing older; the cache keeps only what some window can still show. A history
to browse was rejected. The sources' own weekly feeds cover everything shown
and the cache stays small; an event older than a week is gone.

### Specification before code

Every feature area is written down as requirements before it is built. Each
change of course is a recorded amendment saying why; a test fails when a Must
has no way of being verified. Building first and describing afterwards was
rejected. A ruled-out idea stays ruled out; keeping the specification true is
work of its own at every documentation pass.

## Privacy and the network

### Every request is made by the Go side

The page makes no request and its content policy allows it no origin but its
own. Images such as the clouds and the burnt areas are fetched, checked and
drawn in Go, then handed over as finished pictures. Letting the page fetch, as
the globe library's examples do, was rejected. One place says what EarthNow
contacts; every image crosses to the page as encoded text.

### One client per host, https only, with a size cap

Each source gets a network client for its own host alone. It sends only over
https, follows a redirect only to that host and refuses an oversized or hung
answer. The standard client, which follows a redirect anywhere, was rejected;
so was one client for every host. The README's list of hosts is a property of
the code; a source that moves host fails until the code changes.

### Two layers ask only while shown

The cloud layer reaches its service only while shown, a replay included; the
burnt-area layer likewise. The cloud layer reads its own small listing about
once an hour and fetches an image only when a newer one is listed. Fetching
every layer in the background was rejected. Hiding a layer sends nothing to
its source. The burnt areas start shown, so a first run contacts that service;
a new cloud image can wait up to an hour.

### No account, no telemetry, no update check

There is nothing to sign in to, nothing reports on use and there is no update
check, tray icon or start with Windows. Nothing about the viewer leaves the
machine beyond what the sources need. A new release is found only on the site;
there are no usage figures to steer development.

### Places and home are read on the machine

The nearest place and its country come from Natural Earth data built in, ice
shelves included so a shelf names Antarctica. At launch the globe faces the
country the system's region setting names. A geocoding service was rejected, as
were the time zone and the display language as guides to home. Nothing is sent;
the place data is fixed at build time.

### Donations go through the browser

The donate address is held once on the Go side and opened in the default
browser through the same check as a source link. The page holding the address
was rejected. Nothing inside EarthNow connects to it; EarthNow never learns
what happened next.

### One data folder

Settings, the cache, the log and the web view's own data live in one folder
under the user's local application data. The web view's default, outside that
folder, was rejected. Removing the folder removes everything EarthNow wrote,
which is what setup's "forget" option does.

## The sources

### Three public sources, none needing a key

Earthquakes come from the USGS, other events from NASA EONET and volcanoes from
the Smithsonian and USGS weekly report, added because EONET tracked none. EONET
is asked for closed events as well as open ones; a closed one is marked ended.
EONET alone and open events alone were rejected as leaving out most of the
week. There is nothing to sign up for; the cost is a third schema in an old
encoding.

### The earthquake floor starts at 2.5

The smallest earthquake shown starts at magnitude 2.5; Settings offers others.
Raising it hides quakes at once without dropping the held set. A higher default
was rejected as hiding too much; dropping the set was rejected as emptying the
globe offline. Small local quakes wait until asked for.

### Each source on its own clock, failing alone

Each source is asked about as often as it changes. A failed source keeps its
last set on the globe and is retried with a doubling delay while the others
carry on. One interval for all and clearing a failed source were rejected. A
source that is down hides nothing the others report; old events can show,
marked stale.

### Refresh is limited and says when it last ran

Refresh asks every source at once, no more than once in 30 seconds. The status
line then gives the time of the last refresh. Saying whether a refresh is
available was rejected: worded in whole minutes it read "now" and never cleared.
The line is always true; pressing at every chance still asks the sources often,
which the owner accepted.

### The last good set kept; a broken answer kept out

Each source's last good set is kept as a file, replaced whole after each fetch;
the globe opens on it at start. An answer listing items yet yielding none
usable is a failed fetch that keeps the set. A database was rejected as more
than small sets need; so was taking a broken answer as an empty source, which
once emptied a source and overwrote its file. The globe survives being offline
or a feed changing shape.

### Nothing from the future

An event dated more than 15 minutes ahead of the machine's clock waits until
its time arrives. Showing any date the source gives was rejected after a flood
alert arrived days ahead and showed as recent. A source whose clock runs well
ahead loses those events for a while.

### Volcanoes are ongoing

A volcano in the current weekly report counts in every window from the first
day of its week until the report is more than 14 days old. Dating each volcano
by the issue day was rejected: once that day left the window, none showed. An
erupting volcano shows as erupting; it stays for the report's whole life.

### Kinds no source publishes fold into Other

The key has seven categories; landslides, drought and dust haze are filed under
Other with the source's own kind kept on the event. Rows of their own, reading
zero for ever, were rejected. Every key row can hold something.

### Each source speaks for itself

An event reported by two sources shows as two markers; each measurement is in
its source's own unit and USGS's tsunami flag is never worded. A
duplicate-matching rule and a common severity scale were rejected. Nothing is
claimed that the sources do not; the viewer compares events by reading them.

## Reading the sources faithfully

### Polygons placed by what the data proves

Flood polygons arrive latitude first against the standard, so the feed is read
in the order its own impossible values prove, latitude first when none do. A
polygon across the date line takes its longitudes the short way and stays in
the Pacific. The standard order and a fixed swap were rejected for misplacing
floods. Floods sit where they are and follow a correction upstream.

### A source link is a page

The detail panel links the first source that names a page; a data file is shown
as text. Linking the first source given was rejected after a storm's link
downloaded a warning file. A data-only source cannot be opened from EarthNow.

### Two lost characters put back, nothing else

The volcano report's encoding cannot hold the curly apostrophe or the subscript
two. A letter's "?s" is read as an apostrophe and "SO?" as sulphur dioxide;
every other question mark is kept. Showing the text as sent and replacing every
question mark were rejected. Another character lost the same way still shows
as a question mark.

### Earthquake depth with its meaning

A depth is shown with its USGS band: shallow, intermediate or deep. A depth of
exactly 10 km is marked as often a fixed value, which USGS assigns when it
cannot compute one. A bare number was rejected. A quake genuinely at that depth
carries the note too.

## Honesty about age

### Nothing is labelled live

Every event, layer and source gives its age; no wording uses the word "live".
The badge such displays usually carry was rejected. Nothing claims a freshness
the sources do not give; the cost is more words on screen.

## The layers

### Burnt areas a day at a time

The burnt-area service is asked for one day per request; Go composes the days
into one image beneath the clouds. A date range in one request, which the
service answers empty, was rejected; so was one image per day on the page. A
widened window fetches only the days it lacks and one failed day costs one day;
a round can take several requests.

### Day and night from the time, on the globe's own material

The sun's position is worked out in Go from the time alone, checked against
NOAA's calculator. The globe keeps its lit material with night lights added.
The library's own example was rejected: it fetches from the network and cannot
restore the globe when hidden. The layer truly switches off; the light is
injected into the library's shader, so a test guards the parts it relies on.

### The Earth is never drawn

Every picture of the Earth, in the application and on the website, is NASA
imagery. Generated artwork was rejected after the site once showed a painted
globe. No fabricated record of the planet exists; the imagery's credits must be
kept.

## The globe and its markers

### globe.gl rather than CesiumJS

The globe is globe.gl on three.js with the Blue Marble texture bundled.
CesiumJS was rejected: several times the size, token-bound imagery and a GIS
engine where one calm globe is wanted. The globe is small and offline; it has
no clustering, so EarthNow writes its own.

### Emoji sprites from one table

Each category is drawn as its emoji on a sprite, read from one table, so
categories differ by more than colour. Markers draw over the globe with those
beyond the horizon hidden. Page elements per marker and the depth test, which
sank edge markers into the sphere, were rejected. Drawing stays smooth at
thousands of markers; a horizon check runs every frame.

### Markers keep their size; overlaps become clusters

Every marker keeps its launch size on screen whatever the zoom, so zooming in
pulls neighbours apart. Overlapping markers draw as one cluster with its count.
Activating it zooms until its members part; one still together at the closest
zoom lists them instead. A fixed size on the globe and zooming alone were
rejected, since some overlaps never separate. Every event can be opened; the
clustering is EarthNow's own to maintain.

### Rotation follows its own switch

The globe turns after ten seconds without input, first easing back to the whole
globe. Only the application's setting stops it; the system's reduced-motion
setting was rejected, since Windows ties it to a switch people turn off for
speed. While paused, a silent bar along the foot of the globe area empties over
those ten seconds; nothing on screen was rejected, since it read as a stuck
globe. A close look is undone after ten idle seconds; the bar is one more
element on the globe.

### A missing WebGL2 is said, not shown blank

Without WebGL2 the globe area says so in plain words and the rest of the window
keeps working. A blank window was rejected. There is no fallback globe.

### Only storms leave a track

A severe storm draws a faint track through its positions inside the window,
ending at its marker; Settings can hide the tracks. Tracks for iceberg drift
and earthquake swarms were rejected, since no rule for them is agreed. A
storm's path reads at a glance without crowding the globe.

## Replay

### Replay plays the window as the sources hold it now

Play replays the chosen window over 30 seconds at normal speed, with half and
double offered; its clouds are fetched at half size, held in memory and let go
when it ends. A history kept by EarthNow and full-size or stored replay clouds
were rejected. Nothing extra is stored; it shows what the sources hold now
about those days, with coarser clouds fetched again each time.

### Time travel is explicit

A replay holds at the end of its span; only Now or another window returns to
the present. Now is shown only while replaying, last in its row. Drifting back
on its own and a disabled Now button were rejected. The viewer always knows
which time is on screen; getting back takes one more press.

## The interface

### An action rail down the left

The actions sit in a narrow rail on the left with the key on the right, so the
globe keeps at least 70% of the window. Full-width bars, too short at the
minimum window, were rejected. The actions are icons, explained by tooltips
and the guide.

### One dark palette

The main window has one dark palette, its contrast checked by test; setup keeps
its light and dark toggle. A light theme was rejected. One palette is held to
contrast with the Earth on black; there is no light option.

### Everything from the keyboard, the globe included

Every control is on the keyboard ring. The globe is a stop of its own where
the arrow keys walk the events and Enter opens one. At launch the application
hands its web view the keyboard itself. Trusting the window framework alone,
which lost the first focus, was rejected. Every new control needs its place in
the ring.

### Reading dialogs read themselves

Long help pages scroll gently on their own and stop the moment the reader takes
over, by the mechanism setup uses. Static pages were rejected. Like rotation,
it does not follow the system reduced-motion setting.

### Source text is shown as text

The page never sets markup from a string, so whatever a feed sends shows as
plain text. Rendering a feed's formatting was rejected. A feed cannot put
markup into the window; any formatting it carries is lost.

## Building and installing

### Installed for one user, without administrator rights

On Windows setup installs into the user's own folders and registry; on Linux
the Flatpak is a user install with no file system access. A machine-wide
install was rejected. There is no administrator prompt; each account installs
separately.

### A setup program of its own

Install, update, going back, repair and removal are one bespoke program that
reads the machine once to choose its route. When EarthNow is running it offers
to close it first and never writes while it still runs; no archive entry may
land outside the install folder. A generic installer was rejected. One identity
runs throughout; the program is EarthNow's own to maintain.

### The gate cannot be skipped

The build runs the whole test gate first with no switch to skip it. It is
pinned to pure Go. A skip switch and leaving the compiler choice to the machine
were rejected. Every build comes from a tree that passed; every build waits for
the whole suite.

### Each platform builds on itself

The Windows setup program, the macOS DMG and the Linux Flatpak are each built on
their own platform; the DMG is signed and notarised. On Linux GPU use is
switched on explicitly, since the framework's default leaves the globe no
WebGL. Cross-compiling was rejected. The cost is a machine of each kind and an
Apple developer account.

### Notices generated from what ships

The third-party notices are generated from what actually ships, each licence in
full; the gate fails until the file matches. Notices kept by hand were rejected.
A dependency cannot ship without its notice; the tool must be run after every
dependency change.

### GPL plus a commercial licence

EarthNow's own code is GPL-3.0 with a commercial licence offered separately;
third-party parts keep their own terms. One licence only was rejected. The
source stays open while closed-source use has a route; two sets of terms need
explaining.

### A log that keeps every run

The log keeps earlier runs and rotates by size while running. The application's
first act points its error output at the log, so a crash leaves a record.
Rotating only at start was rejected, since EarthNow runs for days.

## Engineering

### Layers with one place where they meet

The code is split into domain, application, infrastructure and interface, each
depending only inward and joined in one composition root. Convention alone was
rejected. The rules about events, time and the sun are tested with no disk,
network, clock or screen; the cost is more packages and explicit wiring.

### Complete coverage where it means something

The domain and the application must reach complete coverage; every other gated
package holds its measured figure. One figure over the whole program was
rejected. Code that acts on the machine relies on targeted tests and a person's
checks.

### Small modules

No source file may pass a fixed line limit; one just below it must be cut back.
Letting files grow was rejected after a stylesheet passed the limit unseen.
Modules split at real seams; there are many small files.

### Every value has one home

The product's name, the donate address, the version and each category's emoji
are each written once, held there by tests. Copies written where needed were
rejected. A change is made once; static files such as the website must be
stamped from the source at build.

### The wire is checked from both sides

Every shape crossing from Go to the page is compared field by field with its
TypeScript twin. Every call the page makes must say what happens on refusal or
it will not compile. Trusting the two sides to agree was rejected. A renamed
field fails the suite rather than the screen; a new shape is declared three
times.

### Use cases hold no timers

The use cases hold no timer; the facade asks what is due and when to wake. A
replay's position is kept by the page, which plays it on animation frames.
Timers inside the use cases and Go driving the replay's clock were rejected.
Every timing rule runs on a fake clock; the facade carries the waiting loop.

### Background work cannot take the application down

Every goroutine the application starts recovers from a panic, logs the stack
and tells the page. Fire-and-forget goroutines were rejected. One failing fetch
cannot end the run.

### Tests with real parts and no network

The suite never reaches a source: adapters read feeds captured from the real
services and fakes are written by hand. Each guard is proved by planting a
violation. Live calls and mocks were rejected. The suite answers the same
offline; captured feeds age.

# Verifies EarthNow: formatting, vet, staticcheck, the test suite, the coverage floors,
# the frontend suite (Vitest) and the third-party notices.
# Ported from ED Voyage Companion's test.ps1.
#
#   ./test.ps1              run everything
#   ./test.ps1 -Floor 95    run with a different floor over domain and application, for a deliberate check
#
# build.ps1 will run this before it builds, so a release cannot be cut from a tree that
# fails it. Run it directly while working.
#
# Every floor is the measured number, not a target. A floor picked from an aspiration only
# teaches people to lower it; a floor at the measured number fails the moment cover is lost.
param(
    [double]$Floor = 100
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

# Domain and application hold every rule and need no network, filesystem or window, so
# anything short of 100% there is a decision nobody made (NFR-MNT-001).
$gated = './internal/domain/...', './internal/application/...'

# The Phase 0 spike is its own module under spike/, so go list leaves it out; the filter
# keeps any npm package that ships Go files out of the tools' reach as well.
$packages = go list ./... | Where-Object { $_ -notmatch '/node_modules/' }
if ($LASTEXITCODE -ne 0) { throw "go list failed with exit code $LASTEXITCODE" }

Write-Host 'Checking formatting...'
$unformatted = gofmt -l internal tests installer | Where-Object { $_ }
if ($unformatted) { throw "gofmt reports unformatted files:`n$($unformatted -join "`n")" }

Write-Host 'Vetting...'
go vet $packages
if ($LASTEXITCODE -ne 0) { throw "go vet failed with exit code $LASTEXITCODE" }

Write-Host 'Running staticcheck...'
go run honnef.co/go/tools/cmd/staticcheck@v0.8.1 $packages
if ($LASTEXITCODE -ne 0) { throw "staticcheck failed with exit code $LASTEXITCODE" }

Write-Host 'Running the whole suite...'
go test $packages
if ($LASTEXITCODE -ne 0) { throw "go test failed with exit code $LASTEXITCODE" }

Write-Host "Measuring coverage of $($gated -join ', ')..."
$profilePath = Join-Path ([System.IO.Path]::GetTempPath()) 'earthnow-coverage.out'
try {
    go test "-coverprofile=$profilePath" @gated
    if ($LASTEXITCODE -ne 0) { throw "the coverage run failed with exit code $LASTEXITCODE" }

    $summary = go tool cover "-func=$profilePath"
    if ($LASTEXITCODE -ne 0) { throw "go tool cover failed with exit code $LASTEXITCODE" }

    $total = ($summary | Select-Object -Last 1)
    if ($total -notmatch '([0-9]+(?:\.[0-9]+)?)%\s*$') { throw "could not read a total from: $total" }
    $percent = [double]$Matches[1]
    if ($percent -lt $Floor) {
        Write-Host 'Not covered:'
        $summary | Where-Object { $_ -notmatch '100\.0%\s*$' -and $_ -notmatch '^total:' } | ForEach-Object { Write-Host "  $_" }
        throw "coverage is $percent%, below the floor of $Floor%"
    }
    Write-Host "Coverage $percent%, floor $Floor%."
} finally {
    if (Test-Path $profilePath) { Remove-Item $profilePath -Force }
}

# Infrastructure, each package held at the number it actually reaches. httpfetch stops
# short only at a request-building failure no valid method and context can produce.
# cache (the providers' files and the cloud image share one reader and one writer)
# stops short at five faults the OS will not produce on demand (measured): an open
# failing other than for absence, encoding a type that always encodes, then creating,
# writing or closing a temporary file in a folder just made.
# setup stops short at what acts on the machine itself (measured): the registry writes
# (the uninstall entry), creating a shortcut through the Windows Script Host, then
# finding, ending, launching or scheduling the removal of a process. A test must not
# change the machine it runs on, so those are the deliberate gap; everything portable
# (extraction and its fence, paths, sizes, copies, versions, the step log) is covered.
# setup also reads 61.4% on a machine where EarthNow is installed: the installed-version
# read then runs three statements past its early return, 121 of 197 against 118
# (measured). The floor is the figure for a machine without it.
# runlog stops short at what only a crashing child process reaches, where coverage is
# not collected (sending the error output to the log; its crash tests prove it lands),
# then at faults the OS will not produce on demand: the log failing to open, to report
# its size or to close; the runtime refusing a crash file.
# clouds stops short only at encoding the drawn image into memory, which cannot fail;
# gwis likewise at encoding the composed burnt-area image.
# oslocale stops short at the region call failing: kernel32 lacking it (before Windows
# 10 1709) or answering nothing. Neither can be produced on a current Windows (measured
# on Windows only; the macOS and Linux readers are vetted, not run, here).
$measured = [ordered]@{
    './internal/infrastructure/cache'             = 92.6
    './internal/infrastructure/clouds'            = 97.8
    './internal/infrastructure/geo'               = 100
    './internal/infrastructure/gwis'              = 96.9
    './internal/infrastructure/httpfetch'         = 97.7
    './internal/infrastructure/oslocale'          = 85.7
    './internal/infrastructure/pngcheck'          = 100
    './internal/infrastructure/providers/eonet'   = 100
    './internal/infrastructure/providers/usgs'    = 100
    './internal/infrastructure/providers/gvp'     = 100
    './internal/infrastructure/runlog'            = 77.4
    './internal/infrastructure/settings'          = 100
    './internal/infrastructure/setup'             = 59.9
}

# Not gated at all, deliberately: internal/infrastructure/window is Win32 focus
# handling; installer is the setup program's Wails facade over acts that change the
# machine. Neither has anything a test can reach without the platform behind it, so a
# floor over either would be a floor at zero, which asserts nothing.

Write-Host 'Measuring infrastructure...'
foreach ($package in $measured.Keys) {
    $floor = $measured[$package]
    $reported = go test -cover $package
    if ($LASTEXITCODE -ne 0) { throw "$package failed with exit code $LASTEXITCODE" }
    $line = $reported | Where-Object { $_ -match 'coverage: ' } | Select-Object -First 1
    if ($line -notmatch 'coverage: ([0-9]+(?:\.[0-9]+)?)%') { throw "could not read a coverage figure for ${package}: $line" }
    $reached = [double]$Matches[1]
    if ($reached -lt $floor) { throw "$package is at $reached%, below its floor of $floor%" }
    Write-Host ("  {0,-44} {1,5}%  floor {2}%" -f $package, $reached, $floor)
}

Write-Host 'Checking the page: lint, types, then the suite and its coverage floor...'
Push-Location (Join-Path $root 'frontend')
try {
    npm run lint
    if ($LASTEXITCODE -ne 0) { throw "ESLint failed with exit code $LASTEXITCODE" }
    npx tsc --noEmit
    if ($LASTEXITCODE -ne 0) { throw "the type check failed with exit code $LASTEXITCODE" }
    # The floors live in frontend/vite.config.ts beside the reason for each.
    npm test
    if ($LASTEXITCODE -ne 0) { throw "the frontend suite failed with exit code $LASTEXITCODE" }
} finally {
    Pop-Location
}

# NFR-LEG-001: the notices must name every shipped component and carry its licence
# text, so a dependency added or bumped without them fails here.
Write-Host 'Checking the third-party notices...'
python (Join-Path $root 'tools/notices.py') --check
if ($LASTEXITCODE -ne 0) { throw "the notices check failed with exit code $LASTEXITCODE" }

Write-Host 'All green.'


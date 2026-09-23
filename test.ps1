# Verifies EarthNow: formatting, vet, staticcheck, the test suite and the coverage floors.
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
$unformatted = gofmt -l internal tests | Where-Object { $_ }
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
$measured = [ordered]@{
    './internal/infrastructure/geo'               = 100
    './internal/infrastructure/httpfetch'         = 96.6
    './internal/infrastructure/providers/eonet'   = 100
    './internal/infrastructure/providers/usgs'    = 100
}

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

Write-Host 'All green.'

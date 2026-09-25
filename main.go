// EarthNow: open a globe and see what is happening on Earth right now.
//
// main.go is the composition root (CON-002): the one place that builds the
// infrastructure and hands it to the application.
package main

import (
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/application/services"
	"github.com/oernster/EarthNow/internal/infrastructure/cache"
	"github.com/oernster/EarthNow/internal/infrastructure/clouds"
	"github.com/oernster/EarthNow/internal/infrastructure/geo"
	"github.com/oernster/EarthNow/internal/infrastructure/gwis"
	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch"
	"github.com/oernster/EarthNow/internal/infrastructure/oslocale"
	"github.com/oernster/EarthNow/internal/infrastructure/providers/eonet"
	"github.com/oernster/EarthNow/internal/infrastructure/providers/gvp"
	"github.com/oernster/EarthNow/internal/infrastructure/providers/usgs"
	"github.com/oernster/EarthNow/internal/infrastructure/runlog"
	"github.com/oernster/EarthNow/internal/infrastructure/settings"
	"github.com/oernster/EarthNow/internal/product"
)

//go:embed all:frontend/dist
var assets embed.FS

// The licence and the notices each have one home at the repository root; the
// help dialogs read the same files the repository ships (FR-HLP-002, FR-HLP-003).

//go:embed LICENSE
var licenceText string

//go:embed THIRD_PARTY_NOTICES
var noticesText string

// appVersion is overridden at build time with -ldflags "-X main.appVersion=x.y.z"
// (build.ps1 reads VERSION), so no version literal lives in the source. It is a var
// because -X silently does nothing to a const.
var appVersion = "0.0.0-dev"

// The window geometry (NFR-UX-002).
const (
	windowWidth  = 1280
	windowHeight = 800
	minWidth     = 960
	minHeight    = 700
)

// responseCap is FR-PRV-011's size cap; the largest measured feed was 1.51 MB.
const responseCap = 16 << 20

// requestTimeout bounds one fetch so a hung source cannot stall its refresh.
const requestTimeout = 30 * time.Second

// systemClock is the real clock, injected wherever the time is needed.
type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

// cacheFolder holds the per-provider cache files inside the data folder.
const cacheFolder = "cache"

// webviewFolder holds WebView2's own data inside the data folder.
const webviewFolder = "webview"

// dataDir is %LOCALAPPDATA%\EarthNow, created when absent (NFR-PRIV-002).
func dataDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, product.Slug)
	return dir, os.MkdirAll(dir, 0o755)
}

// keepLog opens %LOCALAPPDATA%\EarthNow\Log.txt, keeping what earlier runs wrote
// and rotating at its limit (NFR-OBS-002), then points the run's error output at
// it, so a panic leaves a record (NFR-REL-002). A log that cannot be opened falls
// back to standard error rather than ending the run (NFR-REL-001).
func keepLog() io.Writer {
	if bindingPass {
		return os.Stderr
	}
	dir, err := dataDir()
	if err != nil {
		return os.Stderr
	}
	runLog, err := runlog.Open(dir, appVersion, time.Now())
	if err != nil {
		return os.Stderr
	}
	if err := runLog.Keep(); err != nil {
		_, _ = fmt.Fprintf(runLog, "the error output could not be kept in the log: %v\n", err)
	}
	return runLog
}

func main() {
	// NFR-REL-002: the first act, before anything that could fail.
	log.SetOutput(keepLog())
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	client := httpfetch.New(&http.Client{Timeout: requestTimeout}, responseCap, eonet.Host, usgs.Host, gvp.Host, clouds.Host, gwis.Host)
	quakes := usgs.New(client, usgs.AllMagnitudes)
	providers := []ports.Provider{eonet.New(client), quakes, gvp.New(client)}
	clock := systemClock{}
	globe := services.NewGlobe(services.NewStore(clock), clock, providers)
	dir, err := dataDir()
	// With no data folder the image layers' caches have no home; each then
	// holds nothing and refuses to save, which its status says (FR-CLD-014,
	// FR-BA-014).
	layerDir := ""
	if err == nil {
		globe.UseCache(cache.New(filepath.Join(dir, cacheFolder), responseCap))
		layerDir = filepath.Join(dir, cacheFolder)
	} else {
		log.Printf("no data folder: %v", err)
		dir = ""
	}
	// The settings load before the first fetch, so it asks with the saved minimum.
	settingsStore := settings.New(dir)
	prefs := services.NewPreferences(settingsStore, quakes.SetMinimum)
	prefs.Load()
	log.Printf("settings: %s; notice %q", settingsStore.Path(), prefs.Notice())
	globe.RestoreCached()
	// FR-CLD-005: the cloud service is reached only while the layer is shown;
	// the use case asks nothing while hidden, so its host being allowed costs
	// nothing then (NFR-PRIV-001).
	cloudLayer := services.NewClouds(clock, clouds.New(client, clouds.BaseURL), cache.NewCloud(layerDir, responseCap))
	cloudLayer.Restore()
	cloudLayer.SetShown(prefs.Current().CloudsShown)
	// FR-BA-005: GWIS is reached only while the burnt-area layer is shown.
	burntLayer := services.NewBurntAreas(clock, gwis.New(client, gwis.BaseURL), cache.NewBurnt(layerDir, responseCap))
	burntLayer.Restore()
	burntLayer.SetWindow(prefs.Current().WindowKey)
	burntLayer.SetShown(prefs.Current().BurntShown)
	// Replay's cloud images come from the same service at a smaller size, held
	// in memory only (FR-RPL-015).
	sun := services.NewSun(clock)
	replayClouds := services.NewReplayClouds(clock, clouds.New(client, clouds.BaseURL))
	replay := services.NewReplay(globe, sun, cloudLayer, burntLayer, replayClouds)
	help := Help{
		About:   dto.About{Name: product.Name, Version: appVersion, Copyright: product.Copyright, Licence: product.Licence, Attributions: product.Attributions()},
		Licence: licenceText,
		Notices: noticesText,
	}
	// The label table is built into the binary; one that will not load leaves
	// an empty table, so the globe opens as before and the log says why.
	labels, err := geo.LoadLabels()
	if err != nil {
		log.Printf("Start view: %v", err)
		labels = &geo.Labels{}
	}
	start := services.NewStartView(oslocale.Setting{}, labels)
	app := NewApp(globe, services.NewScheduler(clock, providers), prefs, cloudLayer, burntLayer, replay, replayClouds, sun, start, help, providers)

	// NFR-PRIV-002: WebView2 keeps its data inside the data folder. Left unset it
	// falls back to %APPDATA%\EarthNow.exe (measured), outside it.
	windowsOptions := &windows.Options{}
	if dir != "" {
		windowsOptions.WebviewUserDataPath = filepath.Join(dir, webviewFolder)
	}

	// DEL-005, RSK-002: left nil, Wails v2.12.0 sets webkit2gtk's GPU policy to
	// Never on Linux (internal/frontend/desktop/linux/window.go, read), which
	// leaves the globe no WebGL. Stated here rather than left to the zero value.
	linuxOptions := &linux.Options{WebviewGpuPolicy: linux.WebviewGpuPolicyAlways}

	err = wails.Run(&options.App{
		Windows:          windowsOptions,
		Linux:            linuxOptions,
		Title:            product.Name,
		Width:            windowWidth,
		Height:           windowHeight,
		MinWidth:         minWidth,
		MinHeight:        minHeight,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 1},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		Bind:             []interface{}{app},
	})
	if err != nil {
		log.Printf("wails.Run failed: %v", err)
	}
}

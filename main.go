// EarthNow: open a globe and see what is happening on Earth right now.
//
// main.go is the composition root (CON-002): the one place that builds the
// infrastructure and hands it to the application.
package main

import (
	"embed"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/oernster/EarthNow/internal/application/dto"
	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/application/services"
	"github.com/oernster/EarthNow/internal/infrastructure/cache"
	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch"
	"github.com/oernster/EarthNow/internal/infrastructure/providers/eonet"
	"github.com/oernster/EarthNow/internal/infrastructure/providers/gvp"
	"github.com/oernster/EarthNow/internal/infrastructure/providers/usgs"
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

// The log file and window geometry (NFR-UX-002). The product's name has one home,
// internal/product, shared with the setup program.
const (
	logFileName  = "Log.txt"
	windowWidth  = 1280
	windowHeight = 800
	minWidth     = 960
	minHeight    = 600
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

// openLog points the standard logger at %LOCALAPPDATA%\EarthNow\Log.txt; a log
// that cannot be opened falls back to standard error rather than ending the run.
func openLog() io.Writer {
	dir, err := dataDir()
	if err != nil {
		return os.Stderr
	}
	f, err := os.Create(filepath.Join(dir, logFileName))
	if err != nil {
		return os.Stderr
	}
	return f
}

func main() {
	log.SetOutput(openLog())
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("%s %s", product.Name, appVersion)

	client := httpfetch.New(&http.Client{Timeout: requestTimeout}, responseCap, eonet.Host, usgs.Host, gvp.Host)
	quakes := usgs.New(client, usgs.AllMagnitudes)
	providers := []ports.Provider{eonet.New(client), quakes, gvp.New(client)}
	clock := systemClock{}
	globe := services.NewGlobe(services.NewStore(clock), clock, providers)
	dir, err := dataDir()
	if err == nil {
		globe.UseCache(cache.New(filepath.Join(dir, cacheFolder), responseCap))
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
	help := Help{
		About:   dto.About{Name: product.Name, Version: appVersion, Copyright: product.Copyright, Licence: product.Licence, Attributions: product.Attributions()},
		Licence: licenceText,
		Notices: noticesText,
	}
	app := NewApp(globe, services.NewScheduler(clock, providers), prefs, help, providers)

	// NFR-PRIV-002: WebView2 keeps its data inside the data folder. Left unset it
	// falls back to %APPDATA%\EarthNow.exe (measured), outside it.
	windowsOptions := &windows.Options{}
	if dir != "" {
		windowsOptions.WebviewUserDataPath = filepath.Join(dir, webviewFolder)
	}

	err = wails.Run(&options.App{
		Windows:          windowsOptions,
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

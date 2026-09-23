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

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/application/services"
	"github.com/oernster/EarthNow/internal/infrastructure/cache"
	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch"
	"github.com/oernster/EarthNow/internal/infrastructure/providers/eonet"
	"github.com/oernster/EarthNow/internal/infrastructure/providers/usgs"
)

//go:embed all:frontend/dist
var assets embed.FS

// Product identity and window geometry (NFR-UX-002).
const (
	productName  = "EarthNow"
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

// dataDir is %LOCALAPPDATA%\EarthNow, created when absent (NFR-PRIV-002).
func dataDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, productName)
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

	client := httpfetch.New(&http.Client{Timeout: requestTimeout}, responseCap, eonet.Host, usgs.Host)
	providers := []ports.Provider{eonet.New(client), usgs.New(client, usgs.DefaultMinimum)}
	clock := systemClock{}
	globe := services.NewGlobe(services.NewStore(clock), clock, providers)
	if dir, err := dataDir(); err == nil {
		globe.UseCache(cache.New(filepath.Join(dir, cacheFolder), responseCap))
	} else {
		log.Printf("no cache folder: %v", err)
	}
	globe.RestoreCached()
	app := NewApp(globe, services.NewScheduler(clock, providers), providers)

	err := wails.Run(&options.App{
		Title:            productName,
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

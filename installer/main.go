// Command installer is the bespoke EarthNow setup program (CON-006, DEL-002).
// Ported from ED Voyage Companion's installer/main.go.
//
// It is built as a Wails app so it wears the same WebView as the application it
// installs, with the house light theme and dark theme alike (NFR-UX-005). It carries
// the built application as an embedded zip and covers install, update, repair,
// reinstall, downgrade and uninstall, all per user with no administrator rights.
package main

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/oernster/EarthNow/internal/infrastructure/setup"
	"github.com/oernster/EarthNow/internal/product"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	windowsoptions "github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed payload.zip
var payload []byte

// appVersion is overridden at build time with -ldflags "-X main.appVersion=x.y.z",
// so the setup program never holds a version literal of its own. It is a var because
// -X silently does nothing to a const.
var appVersion = "0.0.0-dev"

const (
	windowTitle = product.Name + " Setup"
	// The window is fixed, so its height has to clear the tallest screen: the
	// install one, which carries three options under the path box. These are ED
	// Voyage Companion's measured figures, carried over with its type sizes; they
	// are rechecked whenever the type changes: text that grows without the window
	// growing with it turns a fixed dialog into a scrolling one.
	windowWidth  = 860
	windowHeight = 780
	// webviewFolder holds the setup window's own WebView2 cache. It is pinned under
	// TEMP rather than left to default into %APPDATA%, so running setup leaves no
	// folder behind next to the application's own.
	webviewFolder = product.Slug + "Setup"
)

// dark and light are the setup page's two surface colours (setup.css --surface).
// The dark one is the application's own --surface-solid. One of them paints the
// window before the page loads, so setup never flashes the wrong ground.
var (
	dark  = options.RGBA{R: 0x0e, G: 0x12, B: 0x1a, A: 1}
	light = options.RGBA{R: 0xf4, G: 0xf6, B: 0xfa, A: 1}
)

func main() {
	prefersDark := setup.SystemPrefersDark()
	background := light
	if prefersDark {
		background = dark
	}
	app := NewApp(payload, appVersion, prefersDark)
	_ = wails.Run(&options.App{
		Title:            windowTitle,
		Width:            windowWidth,
		Height:           windowHeight,
		DisableResize:    true,
		BackgroundColour: &background,
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		Bind:             []interface{}{app},
		Windows: &windowsoptions.Options{
			WebviewUserDataPath: filepath.Join(os.TempDir(), webviewFolder),
		},
	})
}

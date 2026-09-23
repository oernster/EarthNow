package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/oernster/EarthNow/internal/infrastructure/setup"
	"github.com/oernster/EarthNow/internal/infrastructure/window"
)

// uninstallFlag is the argument the Apps list passes back to this program, recorded
// in the registry as the UninstallString.
const uninstallFlag = "-uninstall"

// uninstallExeName is the copy of setup left inside the install directory, so the
// Apps list has something to call even after the downloaded setup file is gone.
const uninstallExeName = "uninstall.exe"

// Progress points for each step of the work. They are ED Voyage Companion's figures,
// not a measurement of this program's own timings.
const (
	pctExtract      = 10
	pctRegister     = 55
	pctChoices      = 80
	pctShortcutsOff = 20
	pctRegistryOff  = 50
	pctStateOff     = 70
	pctFilesOff     = 90
	pctDone         = 100
)

// App is the Wails facade for the setup program. Everything the front end can do
// goes through a method here; the install policy itself is in
// internal/infrastructure/setup (DEL-003).
type App struct {
	ctx           context.Context
	payload       []byte
	version       string
	uninstallMode bool
	prefersDark   bool
	log           *setup.StepLog
}

// NewApp builds the facade. Started with -uninstall, setup opens on the removal
// screen rather than on the manage one.
func NewApp(payload []byte, version string, prefersDark bool) *App {
	uninstall := len(os.Args) > 1 && os.Args[1] == uninstallFlag
	return &App{
		payload:       payload,
		version:       version,
		uninstallMode: uninstall,
		prefersDark:   prefersDark,
		log:           setup.NewStepLog(setup.StepLogPath(), time.Now),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.note("setup %s started", a.version)
}

// domReady takes the keyboard once the page exists.
//
// Wails hands the webview its keyboard from the main window's WM_SETFOCUS, which
// Windows raises only on a change of focus; the handler for it is bound inside an
// asynchronous callback. Whether the window's first focus arrives before there is a
// handler for it is a race, one that setup loses. Measured on a cold launch of ED
// Voyage Companion's setup: the primary button was focused by the page yet drew no
// ring and Enter did nothing at all, which reads as a dead keyboard.
func (a *App) domReady(context.Context) { a.show() }

// show gives the webview the keyboard, falling back to asking Wails for the window.
func (a *App) show() {
	if window.TakeFocus() {
		return
	}
	if a.ctx != nil {
		wailsruntime.WindowShow(a.ctx)
	}
}

// TakeKeyboard is called by the page when it finds it has no keyboard.
//
// The page is the only thing that can tell: from Go the window looks focused either
// way. It is the second half of the repair, because domReady runs before the webview
// is necessarily ready to keep what it is given.
func (a *App) TakeKeyboard() { a.show() }

// StateDTO describes what setup should offer, given what is already on the machine.
//
// AppName is sent because the page must not write the product's name down: a string
// in a page nobody compiles goes stale through a rename without a word from anything.
type StateDTO struct {
	AppName          string `json:"appName"`
	Mode             string `json:"mode"`
	Relation         string `json:"relation"`
	Installed        bool   `json:"installed"`
	InstalledVersion string `json:"installedVersion"`
	ThisVersion      string `json:"thisVersion"`
	InstallDir       string `json:"installDir"`
	StartMenu        bool   `json:"startMenu"`
	Desktop          bool   `json:"desktop"`
	PrefersDark      bool   `json:"prefersDark"`
	LogPath          string `json:"logPath"`
}

// OptionsDTO carries the choices made on the install or reinstall screen.
type OptionsDTO struct {
	StartMenu bool `json:"startMenu"`
	Desktop   bool `json:"desktop"`
}

// Progress is emitted on the "progress" event while a long operation runs.
type Progress struct {
	Pct int    `json:"pct"`
	Msg string `json:"msg"`
}

// relationNames turn the comparison into the word the front end routes on.
var relationNames = map[setup.Relation]string{
	setup.Newer: "newer",
	setup.Same:  "same",
	setup.Older: "older",
}

// DetectState inspects the machine once and returns the mode setup should open in.
func (a *App) DetectState() StateDTO {
	dir, _ := setup.InstallDir()
	installedVersion, installed := setup.InstalledVersion()
	shortcuts := setup.CurrentShortcuts()

	mode := "install"
	switch {
	case a.uninstallMode:
		mode = "uninstall"
	case installed:
		mode = "manage"
	}
	relation := setup.Same
	if installed {
		relation = setup.Compare(a.version, installedVersion)
	}
	a.note("state: mode %s, installed %q, carried %s", mode, installedVersion, a.version)
	return StateDTO{
		AppName:          setup.AppName,
		Mode:             mode,
		Relation:         relationNames[relation],
		Installed:        installed,
		InstalledVersion: installedVersion,
		ThisVersion:      a.version,
		InstallDir:       dir,
		StartMenu:        shortcuts.StartMenu,
		Desktop:          shortcuts.Desktop,
		PrefersDark:      a.prefersDark,
		LogPath:          a.log.Path(),
	}
}

// AppRunning reports whether the application is open, so the front end can offer to
// close it rather than failing later on a locked executable.
func (a *App) AppRunning() bool { return setup.IsAppRunning() }

// CloseRunningApp ends the running application so setup can proceed.
func (a *App) CloseRunningApp() error {
	a.note("closing the running application")
	return a.outcome(setup.CloseRunningApp())
}

// Install performs a fresh install, an update, a downgrade or a reinstall. All four
// are the same act: overwrite the files, then apply the options as given.
func (a *App) Install(choices OptionsDTO) error {
	a.note("install %s: %+v", a.version, choices)
	return a.outcome(a.write(choices))
}

// Repair re-extracts and re-registers the application, leaving every option exactly
// as it stands. It is the quick fix for a damaged install, as distinct from a
// reinstall, which puts the options back as a new install leaves them.
func (a *App) Repair() error {
	shortcuts := setup.CurrentShortcuts()
	a.note("repair %s", a.version)
	return a.outcome(a.write(OptionsDTO{
		StartMenu: shortcuts.StartMenu,
		Desktop:   shortcuts.Desktop,
	}))
}

// write is the single install path behind Install and Repair.
func (a *App) write(choices OptionsDTO) error {
	// Refuse to overwrite a running instance: extracting over a locked executable
	// fails part way and leaves a half-written install.
	if setup.IsAppRunning() {
		return setup.ErrAppRunning
	}
	dir, err := setup.InstallDir()
	if err != nil {
		return err
	}

	a.progress(pctExtract, "Extracting files...")
	if err := setup.ExtractZip(a.payload, dir); err != nil {
		return fmt.Errorf("extract files: %w", err)
	}
	exePath := filepath.Join(dir, setup.ExeName)

	a.progress(pctRegister, "Registering the application...")
	if err := a.register(dir, exePath); err != nil {
		return err
	}

	a.progress(pctChoices, "Applying your choices...")
	setup.ApplyShortcuts(exePath, dir, setup.Shortcuts{
		StartMenu: choices.StartMenu,
		Desktop:   choices.Desktop,
	})

	a.progress(pctDone, "Done.")
	return nil
}

// register leaves a copy of setup beside the application and writes the Apps list
// entry that points at it.
func (a *App) register(dir, exePath string) error {
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate setup: %w", err)
	}
	uninstallExe := filepath.Join(dir, uninstallExeName)
	if err := setup.CopyFile(self, uninstallExe); err != nil {
		return fmt.Errorf("write the uninstaller: %w", err)
	}
	sizeKB, _ := setup.DirSizeKB(dir)
	if err := setup.WriteUninstallEntry(setup.UninstallInfo{
		Version:      a.version,
		InstallDir:   dir,
		UninstallExe: uninstallExe,
		IconPath:     exePath,
		EstimatedKB:  sizeKB,
	}); err != nil {
		return fmt.Errorf("register the application: %w", err)
	}
	return nil
}

// Uninstall removes the shortcuts, the registry record and the installed files. The
// saved settings, event cache, log and window state go too when the user asks.
func (a *App) Uninstall(removeState bool) error {
	a.note("uninstall, removing saved state: %v", removeState)
	return a.outcome(a.remove(removeState))
}

// remove is the work behind Uninstall.
func (a *App) remove(removeState bool) error {
	// The scheduled deletion cannot remove a locked executable, so a running
	// application has to close first.
	if setup.IsAppRunning() {
		return setup.ErrAppRunning
	}
	dir, err := setup.InstallDir()
	if err != nil {
		return err
	}

	a.progress(pctShortcutsOff, "Removing shortcuts...")
	setup.RemoveShortcuts()

	a.progress(pctRegistryOff, "Removing registry entries...")
	_ = setup.RemoveUninstallEntry()

	if removeState {
		a.progress(pctStateOff, "Removing your saved settings...")
		if state, stateErr := setup.DataDir(); stateErr == nil {
			_ = setup.RemoveTree(state)
		}
	}

	a.progress(pctFilesOff, "Removing files...")
	setup.ScheduleDirDeletion(dir)

	a.progress(pctDone, "Done.")
	return nil
}

// LaunchApp starts the installed application, backing the "start it when setup
// finishes" option.
func (a *App) LaunchApp() error {
	a.note("launching the application")
	return a.outcome(setup.LaunchApp())
}

// SetShortcuts applies the shortcut boxes live from the manage screen, so unticking
// one takes the shortcut away there and then.
func (a *App) SetShortcuts(startMenu, desktop bool) error {
	dir, err := setup.InstallDir()
	if err != nil {
		return err
	}
	a.note("shortcuts: start menu %v, desktop %v", startMenu, desktop)
	setup.ApplyShortcuts(filepath.Join(dir, setup.ExeName), dir, setup.Shortcuts{
		StartMenu: startMenu,
		Desktop:   desktop,
	})
	return nil
}

// Quit closes the setup program.
func (a *App) Quit() {
	a.note("setup closed")
	wailsruntime.Quit(a.ctx)
}

// progress reports how far a long operation has got and records the step.
func (a *App) progress(pct int, msg string) {
	a.note("%d%% %s", pct, msg)
	wailsruntime.EventsEmit(a.ctx, "progress", Progress{Pct: pct, Msg: msg})
}

// note writes one line to the step log. A log that cannot be written must not stop
// the install it is recording, so its own failure is dropped here.
func (a *App) note(format string, args ...any) { _ = a.log.Step(format, args...) }

// outcome records how an act ended and hands its error back unchanged.
func (a *App) outcome(err error) error {
	if err != nil {
		a.note("failed: %v", err)
		return err
	}
	a.note("finished")
	return nil
}

//go:build bindings

package main

// bindingPass is true in the copy `wails build` compiles with the bindings tag
// and runs on the build machine to write the frontend's bindings (read in
// wails v2.12.0 pkg/commands/bindings). That copy passes through main, so it
// must not touch the user's log: each build otherwise left a stray
// "0.0.0-dev started" line there (seen 2026-09-23 and 2026-09-24).
const bindingPass = true

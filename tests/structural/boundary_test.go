// Package structural enforces the architecture with tests rather than convention
// (REQUIREMENTS.md CON-002, CON-008, FR-PRV-014). Ported from ED Voyage
// Companion's tests/structural/boundary_test.go.
//
// Every assertion here has been proved to bite by planting a violation and
// watching it fail. An assertion never seen to fail is not yet a guard.
package structural

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// lineLimit is the module-size cap. dangerBand is five per cent below it: a file
// landing between them is refactored down to safeLanding rather than left one
// edit away from breaching.
const (
	lineLimit   = 400
	dangerBand  = lineLimit - lineLimit/20
	safeLanding = 350
)

// modulePath prefixes every internal import.
const modulePath = "github.com/oernster/EarthNow/"

// maxWalkUp bounds the search for go.mod above the test directory.
const maxWalkUp = 6

// compositionRoot names the files allowed to import both application services
// and infrastructure. Nothing else may join them.
var compositionRoot = map[string]bool{"main.go": true, "app.go": true}

// forbiddenInDomain names the packages that would make the domain impure.
var forbiddenInDomain = []string{
	"net", "net/http", "os", "path/filepath", "math/rand", "math/rand/v2",
	"database/sql", "io/ioutil", "os/exec", "log", "log/slog", "sync",
}

// forbiddenCallsInDomain read the wall clock or the global random source.
var forbiddenCallsInDomain = []string{"time.Now", "time.Since", "time.Until", "rand.Intn", "rand.Float64"}

// skippedDirs are trees that are not this module's Go source.
var skippedDirs = map[string]bool{"frontend": true, "node_modules": true, ".git": true, "spike": true, "spike-assets": true}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("working directory: %v", err)
	}
	for range maxWalkUp {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("could not find go.mod above the test directory")
	return ""
}

func goFiles(t *testing.T) []string {
	t.Helper()
	var found []string
	err := filepath.WalkDir(repoRoot(t), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if skippedDirs[entry.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			found = append(found, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking repository: %v", err)
	}
	if len(found) == 0 {
		t.Fatal("no Go files found, the walk is wrong")
	}
	return found
}

func importsOf(t *testing.T, path string) []string {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	out := make([]string, 0, len(parsed.Imports))
	for _, item := range parsed.Imports {
		out = append(out, strings.Trim(item.Path.Value, `"`))
	}
	return out
}

// relative answers a file's slash-separated path from the module root.
func relative(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return ""
	}
	return filepath.ToSlash(rel)
}

func layerOf(root, path string) string {
	parts := strings.Split(relative(root, path), "/")
	if len(parts) >= 2 && parts[0] == "internal" {
		return parts[1]
	}
	return ""
}

func TestDomainHasNoOutwardImports(t *testing.T) {
	root := repoRoot(t)
	for _, path := range goFiles(t) {
		if layerOf(root, path) != "domain" {
			continue
		}
		for _, imported := range importsOf(t, path) {
			if strings.HasPrefix(imported, modulePath) && !strings.HasPrefix(strings.TrimPrefix(imported, modulePath), "internal/domain") {
				t.Errorf("%s imports %s: the domain depends on nothing", relative(root, path), imported)
			}
		}
	}
}

func TestDomainIsPure(t *testing.T) {
	root := repoRoot(t)
	for _, path := range goFiles(t) {
		if layerOf(root, path) != "domain" || strings.HasSuffix(path, "_test.go") {
			continue
		}
		for _, imported := range importsOf(t, path) {
			for _, banned := range forbiddenInDomain {
				if imported == banned {
					t.Errorf("%s imports %q: the domain performs no IO", relative(root, path), banned)
				}
			}
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		for _, call := range forbiddenCallsInDomain {
			if strings.Contains(string(raw), call+"(") {
				t.Errorf("%s calls %s: inject the clock instead", relative(root, path), call)
			}
		}
	}
}

func TestApplicationDoesNotImportInfrastructure(t *testing.T) {
	root := repoRoot(t)
	for _, path := range goFiles(t) {
		if layerOf(root, path) != "application" {
			continue
		}
		for _, imported := range importsOf(t, path) {
			if strings.Contains(imported, "internal/infrastructure") || strings.Contains(imported, "wails") || imported == "net/http" {
				t.Errorf("%s imports %s: the application depends on ports only", relative(root, path), imported)
			}
		}
	}
}

func TestCompositionRootIsWhitelisted(t *testing.T) {
	root := repoRoot(t)
	for _, path := range goFiles(t) {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		var application, infrastructure bool
		for _, imported := range importsOf(t, path) {
			application = application || strings.Contains(imported, "internal/application/services")
			infrastructure = infrastructure || strings.Contains(imported, "internal/infrastructure")
		}
		if application && infrastructure && !compositionRoot[relative(root, path)] {
			t.Errorf("%s wires application to infrastructure: only the composition root may", relative(root, path))
		}
	}
}

// providersDir holds one package per provider adapter.
const providersDir = "internal/infrastructure/providers/"

// TestFRPRV014_OnlyTheCompositionRootNamesAProvider keeps adding a provider a
// change to infrastructure and the composition root alone.
func TestFRPRV014_OnlyTheCompositionRootNamesAProvider(t *testing.T) {
	root := repoRoot(t)
	for _, path := range goFiles(t) {
		rel := relative(root, path)
		if compositionRoot[rel] {
			continue
		}
		for _, imported := range importsOf(t, path) {
			inner := strings.TrimPrefix(imported, modulePath)
			if !strings.HasPrefix(inner, providersDir) {
				continue
			}
			if !strings.HasPrefix(rel, inner+"/") {
				t.Errorf("%s imports the %s adapter: only its own package and the composition root may", rel, strings.TrimPrefix(inner, providersDir))
			}
		}
	}
}

func lineCount(t *testing.T, path string) int {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return strings.Count(string(raw), "\n") + 1
}

// setupFrontendDir is the setup program's front end. It has no build step, so the
// files there are the source rather than an output of one.
var setupFrontendDir = filepath.Join("installer", "frontend", "dist")

// setupFrontendExtensions are the files there that the size rule governs. The icons
// and the licence copy beside them are not source.
var setupFrontendExtensions = map[string]bool{".html": true, ".css": true, ".js": true}

// setupFrontendFiles returns the setup program's own source files.
//
// The size rule walks the Go. That walk skips every directory called frontend, so the
// setup program's page sat outside it. In ED Voyage Companion that page reached
// 880 lines in one file holding its markup, its whole stylesheet and its whole script,
// with no rule anywhere having anything to say about it.
func setupFrontendFiles(t *testing.T) []string {
	t.Helper()
	dir := filepath.Join(repoRoot(t), setupFrontendDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading %s: %v", filepath.ToSlash(setupFrontendDir), err)
	}
	var found []string
	for _, entry := range entries {
		if entry.IsDir() || !setupFrontendExtensions[strings.ToLower(filepath.Ext(entry.Name()))] {
			continue
		}
		found = append(found, filepath.Join(dir, entry.Name()))
	}
	if len(found) == 0 {
		t.Fatalf("no source found in %s, the walk is wrong", filepath.ToSlash(setupFrontendDir))
	}
	return found
}

// pageTree answers every source file of the application's page, tests included.
// The Go walk skips every directory called frontend, so without this the page sat
// outside the size rule: style.css reached 402 lines with nothing to say so.
func pageTree(t *testing.T) []string {
	t.Helper()
	root := repoRoot(t)
	var found []string
	err := filepath.WalkDir(filepath.Join(root, "frontend", "src"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && sourceExtensions[filepath.Ext(path)] {
			found = append(found, path)
		}
		return nil
	})
	if err != nil || len(found) == 0 {
		t.Fatalf("walking the page source found %d files: %v", len(found), err)
	}
	return found
}

// sizedFiles is every file the size rule governs: the Go, the page and the setup page.
func sizedFiles(t *testing.T) []string {
	t.Helper()
	return append(append(goFiles(t), pageTree(t)...), setupFrontendFiles(t)...)
}

func TestCON008_NoFileExceedsTheLineLimit(t *testing.T) {
	root := repoRoot(t)
	for _, path := range sizedFiles(t) {
		if count := lineCount(t, path); count > lineLimit {
			t.Errorf("%s has %d lines, over the %d limit", relative(root, path), count, lineLimit)
		}
	}
}

func TestCON008_NoFileInTheDangerBand(t *testing.T) {
	root := repoRoot(t)
	for _, path := range sizedFiles(t) {
		if count := lineCount(t, path); count > dangerBand && count <= lineLimit {
			t.Errorf("%s has %d lines, inside the danger band %d to %d: reduce it to %d or fewer",
				relative(root, path), count, dangerBand+1, lineLimit, safeLanding)
		}
	}
}

func TestEveryExportedTypeIsDocumented(t *testing.T) {
	root := repoRoot(t)
	for _, path := range goFiles(t) {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		for _, declaration := range parsed.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE || general.Doc != nil {
				continue
			}
			for _, spec := range general.Specs {
				if typed, ok := spec.(*ast.TypeSpec); ok && typed.Name.IsExported() && typed.Doc == nil {
					t.Errorf("%s: exported type %s has no doc comment", relative(root, path), typed.Name.Name)
				}
			}
		}
	}
}

package structural

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// inOrder reports the first step missing from text or out of order after the
// one before it; "" when every step is there in order.
func inOrder(text string, steps ...string) string {
	at := 0
	for _, step := range steps {
		i := strings.Index(text[at:], step)
		if i < 0 {
			return step
		}
		at += i + len(step)
	}
	return ""
}

// NFR-MNT-002, NFR-REL-004, NFR-LEG-001: the gate runs every check in order, each
// throwing on a non-zero exit, so none can pass unnoticed: the Go checks, the
// page's lint, its type check (which is what makes an unhandled refusal fail to
// compile) and suite, then the notices.
func TestNFRMNT002_TheGateRunsEveryCheck(t *testing.T) {
	t.Parallel()
	gate := readRepoFile(t, "test.ps1")
	steps := []string{
		"gofmt -l", "throw", "go vet", "throw", "staticcheck", "throw", "go test $packages", "throw",
		"npm run lint", "throw", "npx tsc --noEmit", "throw", "npm test", "throw",
		"tools/notices.py') --check", "throw",
	}
	if missing := inOrder(gate, steps...); missing != "" {
		t.Errorf("test.ps1 lacks %q where the gate order needs it", missing)
	}
}

// NFR-MNT-001: domain and application are held at 100% by default.
func TestNFRMNT001_DomainAndApplicationAreGatedAt100(t *testing.T) {
	t.Parallel()
	gate := readRepoFile(t, "test.ps1")
	for _, want := range []string{"[double]$Floor = 100", "'./internal/domain/...'", "'./internal/application/...'"} {
		if !strings.Contains(gate, want) {
			t.Errorf("test.ps1 lacks %q", want)
		}
	}
}

// DEL-001: build.ps1 reads VERSION, runs the gate and throws on its failure
// before building anything, passes the version to a var both mains declare, stops
// after the application on -SkipInstaller and writes the setup program.
func TestDEL001_TheBuildRunsTheGateFirst(t *testing.T) {
	t.Parallel()
	build := readRepoFile(t, "build.ps1")
	if missing := inOrder(build,
		"'VERSION'", "-X main.appVersion=$version", "'test.ps1'", "throw", "wails build",
		"if ($SkipInstaller)", "exit 0", "'dist-installer'", "EarthNowSetup.exe",
	); missing != "" {
		t.Errorf("build.ps1 lacks %q where the build order needs it", missing)
	}
	for _, main := range []string{"main.go", filepath.Join("installer", "main.go")} {
		if !strings.Contains(readRepoFile(t, main), "var appVersion = ") {
			t.Errorf("%s must declare appVersion as a var: -X does nothing to a const", main)
		}
	}
}

// DEL-004: one committed .ico is placed on both executables.
func TestDEL004_OneIconOnBothExecutables(t *testing.T) {
	t.Parallel()
	build := readRepoFile(t, "build.ps1")
	for _, want := range []string{"'assets/application-icon.ico'", "@('build', 'installer/build')", `"$dir/windows/icon.ico"`} {
		if !strings.Contains(build, want) {
			t.Errorf("build.ps1 lacks %q", want)
		}
	}
	if _, err := os.Stat(filepath.Join(repoRoot(t), "assets", "application-icon.ico")); err != nil {
		t.Errorf("the committed icon is missing: %v", err)
	}
}

// installPolicy names the imports that would put install logic in the setup
// program's facade rather than in internal/infrastructure/setup.
var installPolicy = []string{"archive/zip", "os/exec", "golang.org/x/sys/windows/registry", "golang.org/x/sys/windows"}

// DEL-003: the install policy lives in internal/infrastructure/setup; the setup
// program's facade uses it and imports none of the means to act alone.
func TestDEL003_TheSetupFacadeOwnsNoInstallLogic(t *testing.T) {
	t.Parallel()
	facade := filepath.Join(repoRoot(t), "installer", "app.go")
	imports := importsOf(t, facade)
	if !contains(imports, modulePath+"internal/infrastructure/setup") {
		t.Error("installer/app.go does not use internal/infrastructure/setup")
	}
	for _, forbidden := range installPolicy {
		if contains(imports, forbidden) {
			t.Errorf("installer/app.go imports %s: install logic belongs in internal/infrastructure/setup", forbidden)
		}
	}
}

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

// NFR-MNT-004: the four documents are present.
func TestNFRMNT004_TheFourDocumentsExist(t *testing.T) {
	t.Parallel()
	for _, doc := range []string{"README.md", "ARCHITECTURE.md", "TESTING.md", "DEVELOPMENT.md"} {
		if _, err := os.Stat(filepath.Join(repoRoot(t), doc)); err != nil {
			t.Errorf("%s: %v", doc, err)
		}
	}
}

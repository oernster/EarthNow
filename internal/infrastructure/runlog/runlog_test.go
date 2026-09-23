package runlog

// NFR-OBS-002 and NFR-REL-002: the log keeps what earlier runs wrote, rotates at
// its limit keeping one previous file and keeps writing when it cannot rotate.

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/product"
)

const testVersion = "9.8.7"

// started is the time every test's run starts at.
var started = time.Date(2026, time.September, 23, 18, 4, 5, 0, time.UTC)

func startLine() string {
	return product.Name + " " + testVersion + " started 2026-09-23 18:04:05\n"
}

func read(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(raw)
}

func mustOpen(t *testing.T, dir string, limit int64) *Log {
	t.Helper()
	l, err := open(dir, testVersion, started, limit)
	if err != nil {
		t.Fatalf("opening the log: %v", err)
	}
	t.Cleanup(func() { _ = l.file.Close() })
	return l
}

func TestAStartKeepsWhatEarlierRunsWrote(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	earlier := "panic: the run before this one\n"
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(earlier), filePerm); err != nil {
		t.Fatal(err)
	}

	l := mustOpen(t, dir, MaxBytes)

	if got := read(t, l.Path()); got != earlier+startLine() {
		t.Errorf("the log reads %q, want the earlier run's crash then the start line", got)
	}
}

func TestTheFolderIsMadeWhereThereIsNone(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "not", "yet")

	l := mustOpen(t, dir, MaxBytes)

	if got := read(t, l.Path()); got != startLine() {
		t.Errorf("the new log reads %q", got)
	}
}

func TestALogAtItsLimitRotatesAsTheRunStarts(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	full := strings.Repeat("x", 64)
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(full), filePerm); err != nil {
		t.Fatal(err)
	}

	l := mustOpen(t, dir, int64(len(full)))

	if got := read(t, filepath.Join(dir, PreviousName)); got != full {
		t.Errorf("the previous log reads %q", got)
	}
	if got := read(t, l.Path()); got != startLine() {
		t.Errorf("the new log reads %q", got)
	}
}

func TestTheLogRotatesWhileRunningKeepingOnePreviousFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	limit := int64(len(startLine()) + len("first\n"))
	l := mustOpen(t, dir, limit)
	// Each of these fills the log alone, so each one after the first rotates it.
	second := strings.Repeat("2", int(limit)-1) + "\n"
	third := strings.Repeat("3", int(limit)-1) + "\n"

	for _, line := range []string{"first\n", second, third} {
		if _, err := l.Write([]byte(line)); err != nil {
			t.Fatalf("writing %q: %v", line, err)
		}
	}

	if got := read(t, filepath.Join(dir, PreviousName)); got != second {
		t.Errorf("the previous log reads %q, want only the generation before this one", got)
	}
	if got := read(t, l.Path()); got != third {
		t.Errorf("the log reads %q", got)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Errorf("the folder holds %d files, want the log and one previous", len(entries))
	}
}

// Windows gives the file opened after a rotation the closed file's handle only
// sometimes (3 times in 5, measured 2026-09-23), so a crash test passes by luck
// without the re-pointing. This asserts the re-pointing itself.
func TestARotationPointsTheErrorOutputAtTheNewFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	l := mustOpen(t, dir, int64(len(startLine())))
	var kept []*os.File
	if err := l.keepWith(func(f *os.File) error { kept = append(kept, f); return nil }); err != nil {
		t.Fatal(err)
	}

	if _, err := l.Write([]byte("after the rotation\n")); err != nil {
		t.Fatalf("writing: %v", err)
	}

	if len(kept) != 2 || kept[1] != l.file || kept[0] == kept[1] {
		t.Errorf("the error output was pointed at %v; want the first file, then the new one %v", kept, l.file)
	}
}

func TestARotationThatCannotRepointSaysSoInTheLog(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	l := mustOpen(t, dir, int64(len(startLine())))
	calls := 0
	refuse := func(*os.File) error {
		calls++
		if calls > 1 {
			return os.ErrPermission
		}
		return nil
	}
	if err := l.keepWith(refuse); err != nil {
		t.Fatal(err)
	}

	if _, err := l.Write([]byte("after the rotation\n")); err != nil {
		t.Fatalf("writing: %v", err)
	}

	if got := read(t, l.Path()); !strings.Contains(got, "could not rotate") {
		t.Errorf("the log reads %q, want the refusal said", got)
	}
}

// A test binary has an error output, so Keep leaves it and copies crash reports
// to the log; the copy is withdrawn afterwards. Where there is none, the crash
// tests' child shows everything going to the log.
func TestKeepWithAnErrorOutputCopiesCrashesToTheLog(t *testing.T) {
	l := mustOpen(t, t.TempDir(), MaxBytes)
	t.Cleanup(func() { _ = debug.SetCrashOutput(nil, debug.CrashOptions{}) })

	if err := l.Keep(); err != nil || l.keep == nil {
		t.Errorf("keeping: %v; a keep set %v", err, l.keep != nil)
	}
}

func TestKeepRefusedLeavesNothingToRepoint(t *testing.T) {
	t.Parallel()
	l := mustOpen(t, t.TempDir(), MaxBytes)

	if err := l.keepWith(func(*os.File) error { return os.ErrPermission }); err == nil || l.keep != nil {
		t.Errorf("got %v with a keep set %v, want the refusal and nothing kept", err, l.keep != nil)
	}
}

func TestALogThatCannotRotateKeepsWriting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// A folder where the previous file belongs refuses the rename.
	if err := os.Mkdir(filepath.Join(dir, PreviousName), folderPerm); err != nil {
		t.Fatal(err)
	}
	l := mustOpen(t, dir, int64(len(startLine())))

	if _, err := l.Write([]byte("still here\n")); err != nil {
		t.Fatalf("writing: %v", err)
	}

	got := read(t, l.Path())
	if !strings.HasPrefix(got, startLine()) || !strings.Contains(got, "could not rotate") ||
		!strings.HasSuffix(got, "still here\n") {
		t.Errorf("the log reads %q, want the line kept and the refusal said", got)
	}
}

func TestALogThatCannotRotateAtTheStartIsStillOpened(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	full := strings.Repeat("x", 64)
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(full), filePerm); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, PreviousName), folderPerm); err != nil {
		t.Fatal(err)
	}

	l := mustOpen(t, dir, int64(len(full)))

	got := read(t, l.Path())
	if !strings.HasPrefix(got, full) || !strings.HasSuffix(got, startLine()) ||
		strings.Count(got, "could not rotate") != 1 {
		t.Errorf("the log reads %q, want the old log kept, the refusal said once and the start line", got)
	}
}

func TestAStuckRotationIsTriedAgainOnlyAfterAnotherLimit(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, PreviousName), folderPerm); err != nil {
		t.Fatal(err)
	}
	l := mustOpen(t, dir, int64(len(startLine())))

	for i := 0; i < 3; i++ {
		if _, err := l.Write([]byte("x\n")); err != nil {
			t.Fatalf("writing: %v", err)
		}
	}

	if got := strings.Count(read(t, l.Path()), "could not rotate"); got != 1 {
		t.Errorf("the refusal was said %d times over three short lines, want once", got)
	}
}

func TestAFolderThatCannotBeMadeIsRefusedWithItsReason(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "a file")
	if err := os.WriteFile(file, nil, filePerm); err != nil {
		t.Fatal(err)
	}

	if _, err := Open(filepath.Join(file, "log"), testVersion, started); err == nil ||
		!strings.Contains(err.Error(), "making the folder") {
		t.Errorf("got %v, want a refusal naming the folder", err)
	}
}

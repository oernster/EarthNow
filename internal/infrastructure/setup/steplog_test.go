package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fixedClock stands still, so a stamped line can be asserted whole.
func fixedClock() time.Time { return time.Date(2026, time.September, 23, 9, 30, 0, 0, time.UTC) }

// Each step is on disk the moment Step returns. A later step adds to the record
// rather than replacing it: a run that dies part way still says how far it got.
func TestEveryStepIsOnDiskAsSoonAsItIsWritten(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), stepLogName)
	log := NewStepLog(path, fixedClock)
	if log.Path() != path {
		t.Fatalf("path = %q, want %q", log.Path(), path)
	}

	steps := []string{"Extracting files...", "Registering the application..."}
	for count, step := range steps {
		if err := log.Step("%s", step); err != nil {
			t.Fatalf("step %q: %v", step, err)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading after %q: %v", step, err)
		}
		lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
		if len(lines) != count+1 {
			t.Fatalf("after %q the log holds %d lines, want %d", step, len(lines), count+1)
		}
		want := fixedClock().Format(stampLayout) + " " + step
		if lines[count] != want {
			t.Errorf("line %d = %q, want %q", count, lines[count], want)
		}
	}
}

// A log that cannot be written says so rather than dropping the record silently.
func TestAStepLogThatCannotBeWrittenIsReported(t *testing.T) {
	t.Parallel()
	blocked := filepath.Join(t.TempDir(), "a directory, not a file")
	if err := os.Mkdir(blocked, 0o755); err != nil {
		t.Fatalf("planting the blocker: %v", err)
	}
	if err := NewStepLog(blocked, fixedClock).Step("anything"); err == nil {
		t.Fatal("a step was reported written into a directory")
	}
}

// The log sits in the temporary folder, outside both the install directory and the
// data folder, so an uninstall that removes them keeps the record of doing so.
func TestTheStepLogLivesInTheTemporaryFolder(t *testing.T) {
	t.Parallel()
	if got, want := StepLogPath(), filepath.Join(os.TempDir(), stepLogName); got != want {
		t.Fatalf("step log path = %q, want %q", got, want)
	}
}

package runlog

// NFR-REL-002: a panic leaves a record. A crash ends the process that has it, so
// these tests start this test binary again as a child; TestMain turns the child
// into the crash asked for. Ported from Bridge Talk's runlog tests.

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	// childEnv names how the child keeps its log; dirEnv names the folder.
	childEnv = "EARTHNOW_RUNLOG_CHILD"
	dirEnv   = "EARTHNOW_RUNLOG_DIR"

	sendAllAct     = "send-all"
	copyCrashesAct = "copy-crashes"

	plantedPanic = "a planted panic on another goroutine"
	beforeCrash  = "the line written before the crash\n"

	// childLimit is small enough that beforeCrash rotates the child's log, so the
	// crash has to follow the error output into the new file.
	childLimit = 64

	goCrashExitCode       = 2
	childFailedExitCode   = 3
	childSurvivedExitCode = 4
	childWait             = 10 * time.Second
)

func TestMain(m *testing.M) {
	if act := os.Getenv(childEnv); act != "" {
		crash(act, os.Getenv(dirEnv))
	}
	os.Exit(m.Run())
}

// crash is the child: it opens a log in dir, keeps its error output there as act
// asks, writes past the limit so the log rotates, then panics on a goroutine.
func crash(act, dir string) {
	l, err := open(dir, testVersion, started, childLimit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "opening the log: %v\n", err)
		os.Exit(childFailedExitCode)
	}
	keep := sendAll
	if act == copyCrashesAct {
		keep = copyCrashes
	}
	if err := l.keepWith(keep); err != nil {
		fmt.Fprintf(os.Stderr, "keeping the error output: %v\n", err)
		os.Exit(childFailedExitCode)
	}
	if _, err := l.Write([]byte(beforeCrash + strings.Repeat("x", childLimit))); err != nil {
		fmt.Fprintf(os.Stderr, "writing: %v\n", err)
		os.Exit(childFailedExitCode)
	}
	go func() { panic(plantedPanic) }()
	time.Sleep(childWait)
	os.Exit(childSurvivedExitCode)
}

// runChild runs the child as act and answers the log it left and its own error
// output.
func runChild(t *testing.T, act string) (logged, errorOutput string) {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("finding the test binary: %v", err)
	}
	dir := t.TempDir()
	cmd := exec.Command(executable, "-test.run=^$")
	cmd.Env = append(os.Environ(), childEnv+"="+act, dirEnv+"="+dir)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err = cmd.Run()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != goCrashExitCode {
		t.Fatalf("the child ended with %v, want the runtime's crash exit %d; it said %q",
			err, goCrashExitCode, stderr.String())
	}
	return read(t, filepath.Join(dir, FileName)), stderr.String()
}

func TestAPanicAfterARotationLandsInTheNewLog(t *testing.T) {
	t.Parallel()
	logged, errorOutput := runChild(t, sendAllAct)

	if !strings.Contains(logged, "panic: "+plantedPanic) {
		t.Errorf("the log lacks the panic: %q", logged)
	}
	if strings.Contains(errorOutput, plantedPanic) {
		t.Errorf("the error output still took the panic: %q", errorOutput)
	}
}

func TestWithAnErrorOutputThePanicIsCopiedToTheLog(t *testing.T) {
	t.Parallel()
	logged, errorOutput := runChild(t, copyCrashesAct)

	if !strings.Contains(logged, "panic: "+plantedPanic) {
		t.Errorf("the log lacks the panic: %q", logged)
	}
	if !strings.Contains(errorOutput, plantedPanic) {
		t.Errorf("the error output lost the panic: %q", errorOutput)
	}
}

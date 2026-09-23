package setup

// The step log is not in ED Voyage Companion's setup program. It is added here
// because REQUIREMENTS.md DEL-002 asks for one, after the installer skill: the worst
// installer failures are the ones that never raise; a record written only at the
// end is lost with the run that needed it.

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// stepLogName is the file the setup program's record is written to, in the
	// temporary folder so an uninstall that removes the install directory and the
	// data folder does not remove the record of itself doing so.
	stepLogName = InstallFolder + "Setup.log"
	// logPerm is the step log's permission; it is the user's own record.
	logPerm = 0o644
	// stampLayout puts a time to each line, to the millisecond, so the gaps between
	// steps are a measurement rather than a guess.
	stampLayout = "2006-01-02 15:04:05.000"
)

// StepLogPath returns where the step log is written, which a failed run names.
func StepLogPath() string { return filepath.Join(os.TempDir(), stepLogName) }

// StepLog is the record of what one run of setup did, one line per step.
type StepLog struct {
	path string
	now  func() time.Time
}

// NewStepLog opens nothing yet: each step opens the file, writes one line and
// closes it again, so the record is on disk the moment a step starts.
func NewStepLog(path string, now func() time.Time) *StepLog {
	return &StepLog{path: path, now: now}
}

// Path is where this log writes.
func (l *StepLog) Path() string { return l.path }

// Step appends one stamped line and flushes it to disk before returning.
func (l *StepLog) Step(format string, args ...any) error {
	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, logPerm)
	if err != nil {
		return fmt.Errorf("open step log %q: %w", l.path, err)
	}
	defer file.Close()
	line := l.now().Format(stampLayout) + " " + fmt.Sprintf(format, args...) + "\n"
	if _, err := file.WriteString(line); err != nil {
		return fmt.Errorf("write step log %q: %w", l.path, err)
	}
	return file.Sync()
}

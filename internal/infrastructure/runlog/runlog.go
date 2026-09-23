// Package runlog keeps the log a run leaves: Log.txt in the data folder, with one
// previous file beside it (NFR-OBS-002). It carries what the run reports, plus
// the Go runtime's own report of a crash when the run fails (NFR-REL-002).
//
// A windowed Windows program started from a shortcut is given no error output, so
// everything written there is lost, the runtime's panic report included. Pointing
// the error output at the log means a run that dies leaves a record rather than a
// silence. Ported from Bridge Talk's runlog, which learned it from a run that
// vanished with nothing to read; the rotation while running is EarthNow's own,
// since EarthNow runs for days where Bridge Talk runs for an evening.
package runlog

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"github.com/oernster/EarthNow/internal/product"
)

const (
	// FileName names the log inside the data folder.
	FileName = "Log.txt"
	// PreviousName names the one earlier log kept beside it (NFR-OBS-002).
	PreviousName = "Log.previous.txt"
	// MaxBytes is the size at which the log rotates (NFR-OBS-002). Worked out
	// from the code rather than measured, a day of refreshes writes roughly
	// 350 KB (two lines of about 110 bytes per fetch: USGS every minute, EONET
	// every ten, GVP hourly), so a run left going rotates about fortnightly.
	MaxBytes = 5 << 20

	startedLayout = "2006-01-02 15:04:05"
	folderPerm    = 0o755
	filePerm      = 0o644
)

// Log is the open log. It is an io.Writer that rotates itself; every method is
// safe to call from any goroutine.
type Log struct {
	mu    sync.Mutex
	path  string
	limit int64
	// next is the size at which the log next rotates: limit, else a limit further
	// on after a rotation failed, so a stuck rotation is tried again only after
	// another limit's worth rather than on every line.
	next int64
	file *os.File
	size int64
	// keep re-points the run's error output after a rotation; nil until Keep.
	keep func(*os.File) error
}

// Open opens the log in dir for a run of version started at started, making the
// folder where there is none, then writes the line naming the build. What the log
// already holds is kept, so the record of a crash survives the next start.
func Open(dir, version string, started time.Time) (*Log, error) {
	return open(dir, version, started, MaxBytes)
}

func open(dir, version string, started time.Time, limit int64) (*Log, error) {
	if err := os.MkdirAll(dir, folderPerm); err != nil {
		return nil, fmt.Errorf("making the folder for the log: %w", err)
	}
	l := &Log{path: filepath.Join(dir, FileName), limit: limit, next: limit}
	if err := l.reopen(); err != nil {
		return nil, err
	}
	// The first question about any report is which binary wrote it. A log
	// already at its limit rotates on this line, as on any other.
	if _, err := fmt.Fprintf(l, "%s %s started %s\n", product.Name, version, started.Format(startedLayout)); err != nil {
		_ = l.file.Close()
		return nil, fmt.Errorf("writing to %s: %w", l.path, err)
	}
	return l, nil
}

// Path answers where the log is.
func (l *Log) Path() string { return l.path }

func (l *Log) previous() string { return filepath.Join(filepath.Dir(l.path), PreviousName) }

// reopen opens the log at its path for appending and reads its size.
func (l *Log) reopen() error {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePerm)
	if err != nil {
		return fmt.Errorf("opening %s: %w", l.path, err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return fmt.Errorf("reading the size of %s: %w", l.path, err)
	}
	l.file, l.size = f, info.Size()
	return nil
}

// Write appends p, rotating first when p would take the log past its limit. A
// rotation that fails leaves the writing where it was: a log kept too long costs
// disk space, a line lost costs the next diagnosis.
func (l *Log) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.size > 0 && l.size+int64(len(p)) > l.next {
		if err := l.rotate(); err != nil {
			line := fmt.Sprintf("the log could not rotate: %v\n", err)
			n, _ := l.file.WriteString(line)
			l.size += int64(n)
			l.next = l.size + l.limit
		} else {
			l.next = l.limit
		}
	}
	n, err := l.file.Write(p)
	l.size += int64(n)
	return n, err
}

// rotate moves the log to the previous file, replacing the one there, then starts
// it afresh and points the error output at the new file. Windows will not rename
// an open file, so the log is closed first; where the rename then fails, the old
// file is opened again.
func (l *Log) rotate() error {
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("closing %s: %w", l.path, err)
	}
	moved := os.Rename(l.path, l.previous())
	if err := l.reopen(); err != nil {
		return err
	}
	if moved != nil {
		return fmt.Errorf("keeping the previous log: %w", moved)
	}
	if l.keep != nil {
		if err := l.keep(l.file); err != nil {
			return err
		}
	}
	return nil
}

// Keep sends what the run reports as it fails to the log, for the rest of the
// run, rotations included. Where the run has no error output (a windowed build
// started from a shortcut), all of it goes to the log; where it has one, it stays
// there and the runtime's crash report is copied to the log as well.
func (l *Log) Keep() error {
	if hasErrorOutput() {
		return l.keepWith(copyCrashes)
	}
	return l.keepWith(sendAll)
}

func (l *Log) keepWith(keep func(*os.File) error) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := keep(l.file); err != nil {
		return err
	}
	l.keep = keep
	return nil
}

// copyCrashes copies the runtime's crash report to log. Go's crash file leaves
// out a fatal error's first line (Bridge Talk measured it against Go 1.26.3), so
// it is the second best, used only where the error output is not lost anyway.
func copyCrashes(log *os.File) error {
	if err := debug.SetCrashOutput(log, debug.CrashOptions{}); err != nil {
		return fmt.Errorf("copying crash reports to %s: %w", log.Name(), err)
	}
	return nil
}

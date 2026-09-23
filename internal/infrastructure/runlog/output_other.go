//go:build !windows

package runlog

import (
	"errors"
	"os"
)

// hasErrorOutput answers yes off Windows, so Keep copies the crash report to the
// log and leaves the error output where it is: EarthNow ships for Windows alone.
func hasErrorOutput() bool { return true }

// sendAll is not built for this platform, so it says so rather than doing nothing.
func sendAll(*os.File) error {
	return errors.New("sending error output to a file is not built for this platform")
}

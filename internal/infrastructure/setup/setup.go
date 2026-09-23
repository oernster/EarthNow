// Package setup holds the per-user install logic behind the bespoke setup program:
// the payload extraction and the paths, which are portable and unit tested, plus the
// registry, shortcut and process side effects in the Windows files beside this one.
// Ported from ED Voyage Companion's internal/infrastructure/setup (DEL-003).
//
// Everything is per-user. Nothing here needs administrator rights, so the whole flow
// runs without an elevation prompt.
package setup

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/oernster/EarthNow/internal/product"
)

const (
	// AppName is the product name shown to the user, used for the Start Menu entry
	// and for the Apps list.
	AppName = product.Name
	// InstallFolder names the install directory and the registry key. It carries no
	// spaces so a path never needs quoting for the sake of readability.
	InstallFolder = product.Slug
	// ExeName is the installed application executable.
	ExeName = product.Slug + ".exe"
	// Publisher is recorded in the uninstall registry entry.
	Publisher = "Oliver Ernster"

	installSubdir = "Programs"
	dirPerm       = 0o755
	bytesPerKB    = 1024
)

// InstallDir returns the per-user install directory,
// %LOCALAPPDATA%\Programs\EarthNow. Installing under LOCALAPPDATA is what keeps the
// whole flow free of an administrator prompt.
func InstallDir() (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return "", fmt.Errorf("LOCALAPPDATA is not set")
	}
	return filepath.Join(base, installSubdir, InstallFolder), nil
}

// DataDir returns %LOCALAPPDATA%\EarthNow, where the application keeps its settings,
// its event cache, its log and its WebView2 data (NFR-PRIV-002). It is everything the
// application writes outside its install directory, so it is what an uninstall has to
// consider keeping.
func DataDir() (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return "", fmt.Errorf("LOCALAPPDATA is not set")
	}
	return filepath.Join(base, product.Slug), nil
}

// ExtractZip extracts a zip archive into dest, creating directories as needed and
// refusing any entry whose path would escape dest.
func ExtractZip(data []byte, dest string) error {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("open payload: %w", err)
	}
	if err := os.MkdirAll(dest, dirPerm); err != nil {
		return fmt.Errorf("create install dir %q: %w", dest, err)
	}
	for _, file := range reader.File {
		if err := extractEntry(file, dest); err != nil {
			return err
		}
	}
	return nil
}

// extractEntry writes one archive entry, rejecting a name that climbs out of dest.
func extractEntry(file *zip.File, dest string) error {
	target := filepath.Join(dest, file.Name)
	fence := filepath.Clean(dest) + string(os.PathSeparator)
	if !strings.HasPrefix(filepath.Clean(target)+string(os.PathSeparator), fence) {
		return fmt.Errorf("unsafe path in payload: %q", file.Name)
	}
	if file.FileInfo().IsDir() {
		return os.MkdirAll(target, dirPerm)
	}
	if err := os.MkdirAll(filepath.Dir(target), dirPerm); err != nil {
		return fmt.Errorf("create dir for %q: %w", target, err)
	}
	source, err := file.Open()
	if err != nil {
		return fmt.Errorf("open entry %q: %w", file.Name, err)
	}
	defer source.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, dirPerm)
	if err != nil {
		return fmt.Errorf("create %q: %w", target, err)
	}
	defer out.Close()
	if _, err := io.Copy(out, source); err != nil {
		return fmt.Errorf("write %q: %w", target, err)
	}
	return nil
}

// DirSizeKB returns the total size of a directory tree in kilobytes, which is what
// the uninstall entry's EstimatedSize value wants.
func DirSizeKB(dir string) (uint32, error) {
	var total int64
	err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("size %q: %w", dir, err)
	}
	return uint32(total / bytesPerKB), nil
}

// RemoveTree deletes a directory tree. It is used for the saved state, which the
// uninstall screen offers to keep.
func RemoveTree(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("remove %q: %w", dir, err)
	}
	return nil
}

// CopyFile copies one file, used to leave a copy of the setup program inside the
// install directory so the Apps list has an uninstaller to call.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %q: %w", src, err)
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, dirPerm)
	if err != nil {
		return fmt.Errorf("create %q: %w", dst, err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("write %q: %w", dst, err)
	}
	return nil
}

// quotedPath wraps a path in plain double quotes, the form every command line in the
// registry holds.
//
// It exists because Go's %q is not that form: it escapes the separators inside the
// string, so a Windows path reaches the registry with every separator doubled. ED
// Voyage Companion shipped exactly that in its login entry and Windows spent every
// sign-in looking for a place that does not exist; the uninstall entry there was
// still written with %q. Here both of its command lines go through this instead.
func quotedPath(path string) string {
	return `"` + path + `"`
}

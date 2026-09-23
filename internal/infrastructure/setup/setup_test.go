package setup

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// zipOf builds an in-memory archive from a name-to-content map.
func zipOf(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range entries {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatalf("creating entry %q: %v", name, err)
		}
		if _, err := file.Write([]byte(content)); err != nil {
			t.Fatalf("writing entry %q: %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("closing archive: %v", err)
	}
	return buffer.Bytes()
}

func TestExtractZipWritesEveryEntry(t *testing.T) {
	t.Parallel()
	dest := t.TempDir()
	payload := zipOf(t, map[string]string{
		ExeName:             "binary",
		"assets/guide.html": "<p>guide</p>",
	})

	if err := ExtractZip(payload, dest); err != nil {
		t.Fatalf("extracting: %v", err)
	}
	for name, want := range map[string]string{
		ExeName:             "binary",
		"assets/guide.html": "<p>guide</p>",
	} {
		raw, err := os.ReadFile(filepath.Join(dest, filepath.FromSlash(name)))
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		if string(raw) != want {
			t.Errorf("%s holds %q, want %q", name, string(raw), want)
		}
	}
}

// TestExtractZipRejectsAPathThatEscapes proves the fence bites. An archive that
// climbs out of the install directory is the one thing extraction must refuse.
func TestExtractZipRejectsAPathThatEscapes(t *testing.T) {
	t.Parallel()
	dest := filepath.Join(t.TempDir(), "install")
	payload := zipOf(t, map[string]string{"../escaped.txt": "no"})

	if err := ExtractZip(payload, dest); err == nil {
		t.Fatal("extraction accepted a path that climbs out of the destination")
	}
}

func TestExtractZipRejectsAnArchiveItCannotRead(t *testing.T) {
	t.Parallel()
	if err := ExtractZip([]byte("not a zip"), t.TempDir()); err == nil {
		t.Fatal("extraction accepted something that is not an archive")
	}
}

func TestDirSizeKBTotalsTheTree(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	nested := filepath.Join(dir, "assets")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("creating %s: %v", nested, err)
	}
	// Two kilobytes plus a byte, so the division to whole kilobytes is exercised
	// rather than landing on a round number by luck.
	if err := os.WriteFile(filepath.Join(nested, "big"), make([]byte, 2*bytesPerKB+1), 0o644); err != nil {
		t.Fatalf("writing the file: %v", err)
	}

	size, err := DirSizeKB(dir)
	if err != nil {
		t.Fatalf("sizing: %v", err)
	}
	if size != 2 {
		t.Errorf("size is %d KB, want 2", size)
	}
}

func TestDirSizeKBFailsOnAMissingTree(t *testing.T) {
	t.Parallel()
	if _, err := DirSizeKB(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("sizing accepted a directory that is not there")
	}
}

func TestCopyFileReproducesTheContent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := filepath.Join(dir, "setup.exe")
	dst := filepath.Join(dir, "uninstall.exe")
	if err := os.WriteFile(src, []byte("payload"), 0o755); err != nil {
		t.Fatalf("writing the source: %v", err)
	}

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("copying: %v", err)
	}
	raw, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("reading the copy: %v", err)
	}
	if string(raw) != "payload" {
		t.Errorf("the copy holds %q, want %q", string(raw), "payload")
	}
}

func TestCopyFileFailsWhenTheSourceIsAbsent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := CopyFile(filepath.Join(dir, "absent"), filepath.Join(dir, "out")); err == nil {
		t.Fatal("the copy accepted a source that is not there")
	}
}

func TestRemoveTreeDeletesTheWholeTree(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "state")
	if err := os.MkdirAll(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatalf("creating the tree: %v", err)
	}

	if err := RemoveTree(dir); err != nil {
		t.Fatalf("removing: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("the tree is still there: %v", err)
	}
}

// TestTheUninstallCommandNamesTheRealPath guards the command lines written into the
// uninstall entry. Go's %q escapes the separators inside a string, so a path written
// with it reaches the registry with every separator doubled; ED Voyage Companion's
// login entry shipped that way and pointed at nowhere.
func TestTheUninstallCommandNamesTheRealPath(t *testing.T) {
	t.Parallel()
	// Built from the package's own names rather than written out, since the product is
	// named in one place and read from there.
	path := filepath.Join(`C:\Users\Someone\AppData\Local\Programs`, InstallFolder, "uninstall.exe")
	written := quotedPath(path)

	if strings.Contains(written, `\\`) {
		t.Errorf("quoted = %s, want single separators: a doubled path is not a place", written)
	}
	if written != `"`+path+`"` {
		t.Errorf("quoted = %s, want the path in plain quotes", written)
	}
}

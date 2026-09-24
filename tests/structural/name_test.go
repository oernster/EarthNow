package structural

// The product's name has one home, internal/product. Ported from ED Voyage
// Companion's tests/structural/identity_test.go.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oernster/EarthNow/internal/product"
)

// identityHome is the one file allowed to write the product's name down.
var identityHome = filepath.Join("internal", "product", "product.go")

// The name reaches a reader in the window title, About, the donate button, the
// heading mark, the setup program, the Start Menu and the Apps list. Written out
// at each of those it is a set of copies held together by memory, which is how a
// rename leaves one behind.
//
// In Go only string literals are examined, import paths aside: a comment naming
// the product is prose. A page file is read whole, since a name in a comment
// there ships to every user as much as one in a string.
func TestTheProductIsNamedOnce(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	for _, path := range goFiles(t) {
		if path == filepath.Join(root, identityHome) || isTest(filepath.Base(path)) {
			continue
		}
		checkGoLiterals(t, root, path)
	}
	pages := append(pageFiles(t, root), filepath.Join(root, "frontend", "index.html"))
	for _, path := range append(pages, setupFrontendFiles(t)...) {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		reportName(t, root, path, string(raw))
	}
}

// pageFiles answers the application page's shipped source.
func pageFiles(t *testing.T, root string) []string {
	t.Helper()
	var found []string
	err := filepath.WalkDir(filepath.Join(root, "frontend", "src"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && sourceExtensions[filepath.Ext(path)] && !isTest(d.Name()) {
			found = append(found, path)
		}
		return nil
	})
	if err != nil || len(found) == 0 {
		t.Fatalf("walking the page source found %d files: %v", len(found), err)
	}
	return found
}

// checkGoLiterals looks in a Go file's string literals, skipping import paths,
// which carry the module's name rather than the product's.
func checkGoLiterals(t *testing.T, root, path string) {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}
	ast.Inspect(parsed, func(node ast.Node) bool {
		if _, ok := node.(*ast.ImportSpec); ok {
			return false
		}
		if literal, ok := node.(*ast.BasicLit); ok && literal.Kind == token.STRING {
			reportName(t, root, path, literal.Value)
		}
		return true
	})
}

// reportName fails the test where a piece of text spells the product out.
func reportName(t *testing.T, root, path, text string) {
	t.Helper()
	for _, form := range []string{product.Name, product.Slug} {
		if strings.Contains(text, form) {
			t.Errorf("%s writes %q: the product is named in %s and read from there",
				relative(root, path), form, filepath.ToSlash(identityHome))
			return
		}
	}
}

package structural

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oernster/EarthNow/internal/product"
)

// shippedSources are the trees whose code ships in the application.
var shippedSources = []string{"internal", filepath.Join("frontend", "src")}

// sourceExtensions are the files that can carry a literal.
var sourceExtensions = map[string]bool{".go": true, ".ts": true, ".tsx": true, ".js": true, ".css": true}

// isTest reports a test file: tests pin the address on purpose, so they are
// not second homes of it.
func isTest(name string) bool {
	return strings.HasSuffix(name, "_test.go") || strings.Contains(name, ".test.")
}

// FR-DON-004: the donate address has exactly one home in shipped code, a named
// constant beside the product identity. The page asks the Go side to open it,
// so the page never needs a copy.
func TestFRDON004_TheDonateAddressHasOneHome(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	var homes []string
	for _, tree := range shippedSources {
		err := filepath.WalkDir(filepath.Join(root, tree), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !sourceExtensions[filepath.Ext(path)] || isTest(d.Name()) {
				return nil
			}
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for range strings.Count(string(body), product.DonateURL) {
				homes = append(homes, relative(root, path))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(homes) != 1 || homes[0] != "internal/product/product.go" {
		t.Fatalf("the donate address must live only in internal/product/product.go; found in %v", homes)
	}
}

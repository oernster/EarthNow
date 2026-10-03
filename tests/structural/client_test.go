package structural

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// clientArgs is httpfetch.New's argument count holding exactly one host: the
// standard client, the size cap and the host.
const clientArgs = 3

// NFR-PRIV-001: each adapter is given a client of its own host alone, so one
// source's redirect can reach neither another source's host nor a layer's while
// that layer is hidden (audit round 2, E-4). Every call building a client
// outside the tests names exactly one host, never a list.
func TestNFRPRIV001_EveryClientHoldsOneHost(t *testing.T) {
	root := repoRoot(t)
	built := 0
	for _, path := range goFiles(t) {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", relative(root, path), err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || !isCall(call, "httpfetch", "New") {
				return true
			}
			built++
			if len(call.Args) != clientArgs || call.Ellipsis.IsValid() {
				t.Errorf("%s builds a client holding %d hosts: give each adapter a client of its own host", relative(root, path), len(call.Args)-clientArgs+1)
			}
			return true
		})
	}
	if built == 0 {
		t.Fatal("no client is built anywhere, the walk is wrong")
	}
}

// isCall reports whether call is pkg.name(...).
func isCall(call *ast.CallExpr, pkg, name string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != name {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == pkg
}

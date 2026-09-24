package structural

// Rules REQUIREMENTS.md promises a structural test for: FR-GEO-004, FR-KEY-002,
// FR-STS-006 and NFR-SEC-001.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// readPage answers a page file's text; withoutComments strips its comments, so
// prose explaining a rule is never read as breaking it.
func readPage(t *testing.T, path string, withoutComments bool) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	text := string(raw)
	if withoutComments {
		text = tsLineComment.ReplaceAllString(tsBlockNote.ReplaceAllString(text, ""), "")
	}
	return text
}

// FR-GEO-004: places are resolved from the embedded Natural Earth data, never
// from a network service, so a hover never sends where the pointer is anywhere.
// The event providers use the network; the geocoder must not.
func TestFRGEO004_TheGeocoderReachesNoNetwork(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	geo := filepath.Join(root, "internal", "infrastructure", "geo")
	forbidden := []string{"net", "net/http", "net/url", "github.com/oernster/EarthNow/internal/infrastructure/httpfetch"}
	checked := 0
	for _, path := range goFiles(t) {
		if filepath.Dir(path) != geo || strings.HasSuffix(path, "_test.go") {
			continue
		}
		checked++
		for _, imported := range importsOf(t, path) {
			for _, bad := range forbidden {
				if imported == bad {
					t.Errorf("%s imports %s", relative(root, path), bad)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no geocoder source found, the walk is wrong")
	}
}

// emojiLiteral matches the pictographs the category table uses, with the
// variation selector some of them carry.
var emojiLiteral = regexp.MustCompile("[\U0001F300-\U0001FAFF☀-➿〰️]")

// FR-KEY-002: every category's emoji has one home, the category table, which the
// key, the markers, the clusters and the tooltips all read. The emoji themselves
// are the product; a second copy of one is what this refuses.
func TestFRKEY002_EmojiLiveOnlyInTheCategoryTable(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	home := filepath.Join(root, "frontend", "src", "categories.ts")
	for _, path := range pageFiles(t) {
		if path != home && emojiLiteral.MatchString(readPage(t, path, false)) {
			t.Errorf("%s writes an emoji: they are read from frontend/src/categories.ts", relative(root, path))
		}
	}
}

// liveWord matches "live" as a word, in any case.
var liveWord = regexp.MustCompile(`(?i)\blive\b`)

// FR-STS-006: no data is labelled "live"; every source publishes with a delay
// and is fetched on a schedule. Read in Go string literals and in page source
// outside comments, which is where every word a user sees is written.
func TestFRSTS006_NothingIsLabelledLive(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	for _, path := range goFiles(t) {
		if isTest(filepath.Base(path)) {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			if literal, ok := node.(*ast.BasicLit); ok && literal.Kind == token.STRING && liveWord.MatchString(literal.Value) {
				t.Errorf("%s says %s", relative(root, path), literal.Value)
			}
			return true
		})
	}
	for _, path := range pageFiles(t) {
		if liveWord.MatchString(readPage(t, path, true)) {
			t.Errorf("%s uses the word live outside a comment", relative(root, path))
		}
	}
}

// NFR-SEC-001: provider text is rendered as text only. React escapes what it
// renders, so the two ways round that are the ones refused.
func TestNFRSEC001_ProviderTextIsRenderedAsText(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	for _, path := range pageFiles(t) {
		text := readPage(t, path, true)
		for _, bad := range []string{"dangerouslySetInnerHTML", "innerHTML"} {
			if strings.Contains(text, bad) {
				t.Errorf("%s uses %s", relative(root, path), bad)
				break
			}
		}
	}
}

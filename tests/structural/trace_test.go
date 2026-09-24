package structural

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// requirementID matches an ID such as FR-GLB-013, NFR-A11Y-002 or DATA-005.
const requirementID = `(?:[A-Z][A-Z0-9]*-)+\d{3}`

// mustRow reads a Must row of a requirements table: its ID, then the cells after
// the priority. anyMustRow counts every row claiming Must, so a row the first
// pattern cannot read fails the test rather than going unchecked.
var (
	mustRow    = regexp.MustCompile(`(?m)^\| (` + requirementID + `) \| Must \|(.*)$`)
	anyMustRow = regexp.MustCompile(`(?m)^\|[^|\n]*\| Must \|`)
	anyID      = regexp.MustCompile(requirementID)
	testMethod = regexp.MustCompile(`\bT\b`)
)

// verifyCells is the width of an FR table after the priority: requirement,
// acceptance and the Verify cell naming the method (T, D or I).
const verifyCells = 3

// personHeading opens TESTING.md's table of the checks a person settles.
const personHeading = "## Checked by a person"

type must struct {
	id        string
	needsTest bool // its Verify cell includes T, so no person can stand in for a test
}

func readRepoFile(t *testing.T, name string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(repoRoot(t), name))
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	return strings.ReplaceAll(string(body), "\r\n", "\n")
}

func musts(t *testing.T) []must {
	t.Helper()
	body := readRepoFile(t, "REQUIREMENTS.md")
	rows := mustRow.FindAllStringSubmatch(body, -1)
	if claimed := len(anyMustRow.FindAllString(body, -1)); len(rows) != claimed || claimed == 0 {
		t.Fatalf("read %d Must rows of the %d in REQUIREMENTS.md; the pattern is wrong", len(rows), claimed)
	}
	var found []must
	for _, row := range rows {
		cells := strings.Split(strings.TrimSuffix(strings.TrimSpace(row[2]), "|"), "|")
		needsTest := len(cells) == verifyCells && testMethod.MatchString(cells[verifyCells-1])
		found = append(found, must{id: row[1], needsTest: needsTest})
	}
	return found
}

// thisFile names the IDs above as examples, so it is no test of them.
const thisFile = "trace_test.go"

// testCorpus is the text of every test: the Go tests and the page's.
func testCorpus(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	for _, path := range append(goFiles(t), pageTree(t)...) {
		if name := filepath.Base(path); !isTest(name) || name == thisFile {
			continue
		}
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		b.Write(body)
		b.WriteByte('\n')
	}
	return b.String()
}

// personTable is TESTING.md from its "Checked by a person" heading to the next.
func personTable(t *testing.T) string {
	t.Helper()
	body := readRepoFile(t, "TESTING.md")
	_, rest, ok := strings.Cut(body, personHeading)
	if !ok {
		t.Fatalf("TESTING.md has no %q section", personHeading)
	}
	table, _, _ := strings.Cut(rest, "\n## ")
	return table
}

// namedIn reports an ID spelt in full in text, with its dashes (an it() title,
// a comment) or without them (a Go test name such as TestFRGEO004_...).
func namedIn(text, id string) bool {
	for _, spelling := range []string{id, strings.ReplaceAll(id, "-", "")} {
		if regexp.MustCompile(regexp.QuoteMeta(spelling) + `(?:\D|$)`).MatchString(text) {
			return true
		}
	}
	return false
}

// Appendix C: each Must is named by the test that verifies it; failing that,
// it is listed with the checks a person settles. A Must verified by T is named by a test whatever
// the table says, so the table cannot stand in for a missing test.
func TestAppendixC_EveryMustIsNamedByATest(t *testing.T) {
	t.Parallel()
	corpus, people := testCorpus(t), personTable(t)
	var untested, parked []string
	for _, m := range musts(t) {
		switch tested := namedIn(corpus, m.id); {
		case tested:
		case m.needsTest:
			untested = append(untested, m.id)
		case !namedIn(people, m.id):
			untested = append(untested, m.id)
		}
		if m.needsTest && !namedIn(corpus, m.id) && namedIn(people, m.id) {
			parked = append(parked, m.id)
		}
	}
	if len(untested) > 0 {
		t.Errorf("%d Musts named by no test nor TESTING.md's person table: %v", len(untested), untested)
	}
	if len(parked) > 0 {
		t.Errorf("verified by T yet only in the person table: %v", parked)
	}
}

// Every ID the person table names is a requirement, so a typo there is not
// mistaken for a check.
func TestAppendixC_ThePersonTableNamesRealRequirements(t *testing.T) {
	t.Parallel()
	known := map[string]bool{}
	for _, id := range anyID.FindAllString(readRepoFile(t, "REQUIREMENTS.md"), -1) {
		known[id] = true
	}
	var unknown []string
	for _, id := range anyID.FindAllString(personTable(t), -1) {
		if !known[id] {
			unknown = append(unknown, id)
		}
	}
	sort.Strings(unknown)
	if len(unknown) > 0 {
		t.Errorf("TESTING.md's person table names IDs REQUIREMENTS.md does not hold: %v", unknown)
	}
}

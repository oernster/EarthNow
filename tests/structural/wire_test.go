package structural

// NFR-MNT-003: the wire is stated twice, as Go structs with json tags and as
// TypeScript interfaces; this compares the two statements. Ported from ED
// Voyage Companion's tests/structural/wire_test.go.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// wirePackage holds every shape the bound methods hand the page. Wails binds
// whatever a bound method returns, so the package is what says a struct crosses.
var wirePackage = filepath.Join("internal", "application", "dto")

// wireContract is the page's hand-written statement of those shapes. Wails
// generates the same shapes into frontend/wailsjs at build time; that file is
// gitignored and imported by nothing, so it is not the contract. The type checker
// sees the interface and the marshaller sees the struct; neither sees the other.
var wireContract = filepath.Join("frontend", "src", "types.ts")

// wireShapes pairs each struct with the interface that restates it, so adding a
// shape is a decision taken twice: once in the code and once here.
var wireShapes = map[string]string{
	"About":          "AboutDTO",
	"Choice":         "ChoiceDTO",
	"Event":          "EventDTO",
	"Provider":       "ProviderDTO",
	"SettingChoices": "SettingChoicesDTO",
	"Settings":       "SettingsDTO",
	"Speed":          "SpeedDTO",
	"View":           "ViewDTO",
}

// tsInterface captures one interface and its body; tsField matches one field in
// it. Comments are removed first so prose naming a field never reads as one.
var (
	tsInterface   = regexp.MustCompile(`(?s)export interface (\w+) \{(.*?)\n\}`)
	tsField       = regexp.MustCompile(`(?m)^\s+([A-Za-z_]\w*)\??:`)
	tsBlockNote   = regexp.MustCompile(`(?s)/\*.*?\*/`)
	tsLineComment = regexp.MustCompile(`(?m)//.*$`)
)

// wireFieldNames reads the wire names off one struct: the json tag, since the tag
// is what reaches the page, not the Go name.
func wireFieldNames(t *testing.T, name string, structure *ast.StructType) []string {
	t.Helper()
	var fields []string
	for _, field := range structure.Fields.List {
		if len(field.Names) == 0 || !field.Names[0].IsExported() {
			continue
		}
		if field.Tag == nil {
			t.Errorf("dto.%s.%s has no json tag, so it ships under its Go name", name, field.Names[0].Name)
			continue
		}
		raw, err := strconv.Unquote(field.Tag.Value)
		if err != nil {
			t.Fatalf("unreadable tag on dto.%s.%s", name, field.Names[0].Name)
		}
		tag := strings.Split(reflect.StructTag(raw).Get("json"), ",")[0]
		if tag != "" && tag != "-" {
			fields = append(fields, tag)
		}
	}
	sort.Strings(fields)
	return fields
}

// wireStructs answers every struct the dto package declares, by json field name.
func wireStructs(t *testing.T, root string) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	for _, path := range goFiles(t) {
		if filepath.Dir(path) != filepath.Join(root, wirePackage) || strings.HasSuffix(path, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		for _, declaration := range parsed.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE {
				continue
			}
			for _, spec := range general.Specs {
				typed := spec.(*ast.TypeSpec)
				if structure, ok := typed.Type.(*ast.StructType); ok && typed.Name.IsExported() {
					out[typed.Name.Name] = wireFieldNames(t, typed.Name.Name, structure)
				}
			}
		}
	}
	if len(out) == 0 {
		t.Fatalf("no structs found in %s, the walk is wrong", filepath.ToSlash(wirePackage))
	}
	return out
}

// wireInterfaces answers every interface the contract declares, by field name.
func wireInterfaces(t *testing.T, root string) map[string][]string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, wireContract))
	if err != nil {
		t.Fatalf("reading %s: %v", filepath.ToSlash(wireContract), err)
	}
	source := tsLineComment.ReplaceAll(tsBlockNote.ReplaceAll(raw, nil), nil)
	out := map[string][]string{}
	for _, match := range tsInterface.FindAllSubmatch(source, -1) {
		var fields []string
		for _, field := range tsField.FindAllSubmatch(match[2], -1) {
			fields = append(fields, string(field[1]))
		}
		sort.Strings(fields)
		out[string(match[1])] = fields
	}
	if len(out) == 0 {
		t.Fatalf("no interfaces found in %s, the pattern is wrong", filepath.ToSlash(wireContract))
	}
	return out
}

// missing answers the entries of want that got does not hold.
func missing(want, got []string) []string {
	present := map[string]bool{}
	for _, item := range got {
		present[item] = true
	}
	var out []string
	for _, item := range want {
		if !present[item] {
			out = append(out, item)
		}
	}
	return out
}

// A field renamed or removed on the Go side alone leaves the interface declaring
// something that never arrives, which the page reads as undefined with nothing
// to say so; the reverse sends a value on every call that nothing collects.
func TestNFRMNT003_TheWireMatchesOnBothSides(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	structs := wireStructs(t, root)
	interfaces := wireInterfaces(t, root)

	var unpaired []string
	for name := range structs {
		if _, ok := wireShapes[name]; !ok {
			unpaired = append(unpaired, name)
		}
	}
	sort.Strings(unpaired)
	for _, name := range unpaired {
		t.Errorf("dto.%s crosses the wire but is not paired in wireShapes", name)
	}

	names := make([]string, 0, len(wireShapes))
	for name := range wireShapes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		checkShape(t, name, wireShapes[name], structs, interfaces)
	}
}

// checkShape compares one pair, so each failure reads as a statement about it.
func checkShape(t *testing.T, name, shape string, structs, interfaces map[string][]string) {
	t.Helper()
	fields, ok := structs[name]
	if !ok {
		t.Errorf("wireShapes pairs dto.%s, which no longer exists", name)
		return
	}
	declared, ok := interfaces[shape]
	if !ok {
		t.Errorf("dto.%s has no interface %s in %s", name, shape, filepath.ToSlash(wireContract))
		return
	}
	for _, field := range missing(fields, declared) {
		t.Errorf("dto.%s sends %s but interface %s does not declare it", name, field, shape)
	}
	for _, field := range missing(declared, fields) {
		t.Errorf("interface %s declares %s but dto.%s does not send it", shape, field, name)
	}
}

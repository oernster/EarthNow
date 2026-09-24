package structural

import (
	"math"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The page's stylesheet, read as the browser would take it: the rules each
// requirement below rests on, read from the one place they are written.
const pageStyle = "frontend/src/style.css"

func cssToken(t *testing.T, css, name string) string {
	t.Helper()
	m := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(name) + `:\s*([^;]+);`).FindStringSubmatch(css)
	if m == nil {
		t.Fatalf("%s holds no %s", pageStyle, name)
	}
	return strings.TrimSpace(m[1])
}

func pixels(t *testing.T, value string) float64 {
	t.Helper()
	m := regexp.MustCompile(`(\d+)px`).FindStringSubmatch(value)
	if m == nil {
		t.Fatalf("%q states no pixels", value)
	}
	n, _ := strconv.ParseFloat(m[1], 64)
	return n
}

// rule answers the declarations of the first rule whose selector list is exactly
// selector, "" when there is none. A rule starts after the one before it or
// after the comment that introduces it.
func rule(css, selector string) string {
	m := regexp.MustCompile(`(?s)(?:^|\}|\*/)\s*` + regexp.QuoteMeta(selector) + `\s*\{([^}]*)\}`).FindStringSubmatch(css)
	if m == nil {
		return ""
	}
	return m[1]
}

// railGlyphPx is the rail's glyph size FR-RAIL-001 names.
const railGlyphPx = 48

// FR-RAIL-001, FR-DON-002: the rail is a column of buttons, each drawing its
// artwork in the glyph box; the donate button is one of them (FR-DON-001's test
// holds its class), so its artwork sits in the same box.
func TestFRRAIL001_TheRailDrawsEveryButtonInTheGlyphBox(t *testing.T) {
	t.Parallel()
	css := readRepoFile(t, pageStyle)
	if got := pixels(t, cssToken(t, css, "--rail-glyph-size")); got != railGlyphPx {
		t.Errorf("--rail-glyph-size is %vpx, want %dpx", got, railGlyphPx)
	}
	if !strings.Contains(rule(css, ".rail-actions"), "flex-direction: column") {
		t.Error(".rail-actions does not stack its buttons one above another")
	}
	img := rule(css, ".rail-btn img")
	for _, want := range []string{"width: var(--rail-glyph-size)", "height: var(--rail-glyph-size)"} {
		if !strings.Contains(img, want) {
			t.Errorf(".rail-btn img lacks %q, so a button's artwork (the donate mark included, FR-DON-002) leaves the glyph box", want)
		}
	}
}

// globeShareFloor is NFR-UX-001's least share of the window the globe area holds.
const globeShareFloor = 0.70

// NFR-UX-001: at the minimum window the globe area, the window less the rail and
// the key at full height, is at least 70% of it; a wider window only raises the
// share, since the rail and the key keep their widths.
func TestNFRUX001_TheGlobeAreaHoldsSeventyPercentAtTheMinimum(t *testing.T) {
	t.Parallel()
	css := readRepoFile(t, pageStyle)
	rail := pixels(t, cssToken(t, css, "--rail-glyph-size")) + pixels(t, strings.TrimPrefix(cssToken(t, css, "--rail-width"), "calc(var(--rail-glyph-size) + "))
	key := pixels(t, cssToken(t, css, "--key-width"))
	m := regexp.MustCompile(`minWidth\s*=\s*(\d+)`).FindStringSubmatch(readRepoFile(t, "main.go"))
	if m == nil {
		t.Fatal("main.go states no minWidth")
	}
	width, _ := strconv.ParseFloat(m[1], 64)
	if share := (width - rail - key) / width; share < globeShareFloor {
		t.Errorf("the globe area is %.3f of the minimum window, below %.2f", share, globeShareFloor)
	}
}

// NFR-KBD-008: rings follow the three states: green on an enabled control
// hovered or focused, a permanent danger ring while disabled.
func TestNFRKBD008_RingsFollowTheThreeStates(t *testing.T) {
	t.Parallel()
	css := readRepoFile(t, pageStyle)
	if !strings.Contains(rule(css, "button:enabled:hover,\nbutton:enabled:focus-visible"), "border-color: var(--ring)") {
		t.Error("an enabled button hovered or focused does not take the ring colour")
	}
	if !strings.Contains(rule(css, "button:disabled"), "border-color: var(--danger)") {
		t.Error("a disabled button does not keep the danger ring")
	}
}

func luminance(t *testing.T, hex string) float64 {
	t.Helper()
	v, err := strconv.ParseUint(strings.TrimPrefix(hex, "#"), 16, 32)
	if err != nil || len(hex) != len("#rrggbb") {
		t.Fatalf("%q is not a #rrggbb colour", hex)
	}
	channel := func(shift uint) float64 {
		c := float64(v>>shift&0xff) / 0xff
		if c <= 0.04045 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(16) + 0.7152*channel(8) + 0.0722*channel(0)
}

// contrast is the WCAG contrast ratio of two colours.
func contrast(t *testing.T, a, b string) float64 {
	la, lb := luminance(t, a), luminance(t, b)
	return (math.Max(la, lb) + 0.05) / (math.Min(la, lb) + 0.05)
}

// NFR-A11Y-002: text tokens reach 4.5:1 and glyph tokens 3:1 on both solid
// surfaces the page draws on. --border draws decoration only, never text.
func TestNFRA11Y002_ThePaletteMeetsItsContrast(t *testing.T) {
	t.Parallel()
	css := readRepoFile(t, pageStyle)
	floors := map[string]float64{"--text": 4.5, "--muted": 4.5, "--warn": 4.5, "--ring": 3, "--danger": 3}
	for _, surface := range []string{"--space", "--surface-solid"} {
		for name, floor := range floors {
			if got := contrast(t, cssToken(t, css, name), cssToken(t, css, surface)); got < floor {
				t.Errorf("%s on %s is %.2f:1, below %.1f:1", name, surface, got, floor)
			}
		}
	}
}

// networkSource is anything in a fetch directive that names an origin beyond the
// page's own: a scheme reaching the network or a wildcard.
var networkSource = regexp.MustCompile(`https?:|wss?:|\*`)

// NFR-SEC-002: the page's Content-Security-Policy lets no fetch directive reach a
// network origin: connect-src is 'self' alone and no directive names a network
// scheme or a wildcard (images may also be data: and blob:, which fetch nothing).
func TestNFRSEC002_ThePageReachesNoNetworkOrigin(t *testing.T) {
	t.Parallel()
	m := regexp.MustCompile(`http-equiv="Content-Security-Policy"\s+content="([^"]+)"`).FindStringSubmatch(readRepoFile(t, "frontend/index.html"))
	if m == nil {
		t.Fatal("frontend/index.html carries no Content-Security-Policy")
	}
	policy := m[1]
	if !strings.Contains(policy, "connect-src 'self';") || !strings.Contains(policy, "default-src 'self';") {
		t.Errorf("the policy does not hold default-src and connect-src to 'self': %s", policy)
	}
	if found := networkSource.FindString(policy); found != "" {
		t.Errorf("the policy names a network source %q: %s", found, policy)
	}
}

// eonetMap is the EONET adapter's category table, the one home of the mapping.
const eonetMap = "internal/infrastructure/providers/eonet/eonet.go"

// DATA-007: the category mapping is data in the EONET adapter. No EONET category
// id the table maps appears as a literal in the page, the domain or the
// application, which is where a conditional on one would have to be written.
func TestDATA007_TheCategoryMappingIsDataInTheAdapter(t *testing.T) {
	t.Parallel()
	table := regexp.MustCompile(`(?s)var categories = map\[string\]event\.Category\{(.*?)\n\}`).FindStringSubmatch(readRepoFile(t, eonetMap))
	if table == nil {
		t.Fatalf("%s holds no categories table", eonetMap)
	}
	ids := regexp.MustCompile(`"(\w+)":`).FindAllStringSubmatch(table[1], -1)
	if len(ids) == 0 {
		t.Fatal("the categories table maps nothing; the pattern is wrong")
	}
	root := repoRoot(t)
	for _, path := range append(pageTree(t), goFiles(t)...) {
		rel := relative(root, path)
		inScope := strings.HasPrefix(rel, "frontend/") || strings.HasPrefix(rel, "internal/domain/") || strings.HasPrefix(rel, "internal/application/")
		if isTest(filepath.Base(path)) || !inScope {
			continue
		}
		text := readRepoFile(t, rel)
		for _, id := range ids {
			if strings.Contains(text, `"`+id[1]+`"`) || strings.Contains(text, `'`+id[1]+`'`) {
				t.Errorf("%s names the EONET category %q: the mapping belongs in %s", rel, id[1], eonetMap)
			}
		}
	}
}

// DATA-010: USGS's tsunami flag marks large oceanic events only (R4), so nothing
// the user reads may word it: no shipped page or application source names it.
func TestDATA010_TheTsunamiFlagIsNeverWorded(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	var worded []string
	for _, path := range append(pageTree(t), goFiles(t)...) {
		rel := relative(root, path)
		if isTest(filepath.Base(path)) || !(strings.HasPrefix(rel, "frontend/") || strings.HasPrefix(rel, "internal/application/")) {
			continue
		}
		if strings.Contains(strings.ToLower(readRepoFile(t, rel)), "tsunami") {
			worded = append(worded, rel)
		}
	}
	if len(worded) > 0 {
		t.Errorf("the tsunami flag is worded in %v", worded)
	}
}

// NFR-UX-006: the main window's heading is the application mark, read as the
// product's name, which has its one home in internal/product.
func TestNFRUX006_TheHeadingReadsTheProductName(t *testing.T) {
	t.Parallel()
	if !strings.Contains(readRepoFile(t, "frontend/src/App.tsx"), `<h1><img className="brand-mark" src={icons.appMark} alt={product}`) {
		t.Error("App.tsx's heading is not the application mark named by the product")
	}
}

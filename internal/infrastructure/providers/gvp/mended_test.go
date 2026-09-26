package gvp

import (
	"strings"
	"testing"
)

// DATA-013: the two characters the feed's encoding loses are put back where the
// meaning is plain; every other question mark is left as the source sent it.
func TestDATA013_LostCharactersAreMendedAndRealQuestionMarksKept(t *testing.T) {
	t.Parallel()
	cases := []struct{ in, want string }{
		{"the eruption at Karangetang?s North Crater", "the eruption at Karangetang’s North Crater"},
		{"Instituto Geofísico del Perú?s (IGP)", "Instituto Geofísico del Perú’s (IGP)"},
		{"Kilauea?s summit and Great Sitkin?s crater", "Kilauea’s summit and Great Sitkin’s crater"},
		{"sulfur dioxide (SO?) emissions", "sulfur dioxide (SO₂) emissions"},
		{"Has it erupted? Scientists are unsure.", "Has it erupted? Scientists are unsure."},
		{"a plume? seen at dawn", "a plume? seen at dawn"},
		{"Kilauea?s", "Kilauea’s"},
		{"what?so", "what?so"},
		{"ALSO?", "ALSO?"},
	}
	for _, c := range cases {
		if got := mended(c.in); got != c.want {
			t.Errorf("mended(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// DATA-013 over the captured feed: no lost apostrophe or subscript survives.
func TestDATA013_TheCapturedFeedReadsWithoutLostCharacters(t *testing.T) {
	t.Parallel()
	events, _, err := Parse(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	mendedOne := false
	for _, e := range events {
		if lostApostrophe.MatchString(e.Description) || strings.Contains(e.Description, "SO?") {
			t.Errorf("%s still reads a lost character: %q", e.Title, e.Description)
		}
		if strings.Contains(e.Description, "Karangetang’s North Crater") {
			mendedOne = true
		}
	}
	if !mendedOne {
		t.Error("Karangetang's description was not mended")
	}
}

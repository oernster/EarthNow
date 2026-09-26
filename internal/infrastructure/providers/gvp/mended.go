package gvp

import "regexp"

// The feed is written in ISO-8859-1, which holds neither the curly apostrophe
// nor the subscript two, so the source writes a question mark for each
// (measured on the live feed on 2026-09-26: "Karangetang?s", "Kilauea?s",
// "Perú?s", "(SO?)"). mended puts back those two where the meaning is plain and
// leaves every other question mark as sent (REQUIREMENTS.md DATA-013).
var (
	lostApostrophe = regexp.MustCompile(`(\p{L})\?s\b`)
	lostSubscript  = regexp.MustCompile(`\bSO\?`)
)

// Their restored forms: the apostrophe the source wrote, then sulfur dioxide.
const (
	apostropheS   = "${1}’s"
	sulfurDioxide = "SO₂"
)

// mended answers text with a letter followed by "?s" read as a possessive or a
// contraction and "SO?" read as sulfur dioxide.
func mended(text string) string {
	text = lostApostrophe.ReplaceAllString(text, apostropheS)
	return lostSubscript.ReplaceAllString(text, sulfurDioxide)
}

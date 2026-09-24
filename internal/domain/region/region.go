// Package region reads the reader's country from what the operating system
// says about it (REQUIREMENTS.md FR-GLB-014): a region code as Windows answers
// one; the territory inside a locale name as macOS and Linux hold one. The
// answer is an ISO 3166-1 alpha-2 code; anything else is no answer, since a UN
// M.49 area such as 001 (the world) is not a country to face.
package region

import "strings"

// codeLength is an ISO 3166-1 alpha-2 code's length.
const codeLength = 2

// The separators of a locale name: POSIX's language_TERRITORY.codeset@modifier
// and BCP 47's language-Script-REGION.
const (
	posixTerritory = "_"
	bcp47Subtag    = "-"
	codesetMark    = "."
	modifierMark   = "@"
	// regionKeyword is the Unicode locale keyword macOS writes when the Region
	// setting differs from the language's own (en_US@rg=gbzzzz, ASM-010): its
	// first two letters are the region.
	regionKeyword = "rg="
	keywordEnd    = ";"
)

// Code answers a bare region code as an ISO alpha-2 code, upper case; false
// when it is not two letters.
func Code(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if len(s) != codeLength {
		return "", false
	}
	for _, r := range s {
		if !isLetter(r) {
			return "", false
		}
	}
	return strings.ToUpper(s), true
}

// FromLocale answers the territory a locale name carries: GB for en_GB.UTF-8,
// en-GB or en_US@rg=gbzzzz; false for C, POSIX, a bare language or nothing.
func FromLocale(locale string) (string, bool) {
	locale = strings.TrimSpace(locale)
	if at := strings.Index(locale, regionKeyword); at >= 0 {
		value := locale[at+len(regionKeyword):]
		if end := strings.Index(value, keywordEnd); end >= 0 {
			value = value[:end]
		}
		if len(value) >= codeLength {
			if code, ok := Code(value[:codeLength]); ok {
				return code, true
			}
		}
	}
	if cut := strings.IndexAny(locale, codesetMark+modifierMark); cut >= 0 {
		locale = locale[:cut]
	}
	parts := strings.FieldsFunc(locale, func(r rune) bool {
		return strings.ContainsRune(posixTerritory+bcp47Subtag, r)
	})
	// The first part is the language; the territory is the first two-letter
	// part after it (a BCP 47 script subtag is four letters).
	for i, part := range parts {
		if i == 0 {
			continue
		}
		if code, ok := Code(part); ok {
			return code, true
		}
	}
	return "", false
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

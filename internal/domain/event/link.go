package event

import (
	"path"
	"strings"
)

// pageExtensions are the endings a web page's address may carry. An address
// whose last segment has no ending at all is a page too. Anything else is a
// file a browser downloads rather than shows: measured in the EONET feed on
// 2026-09-23, JTWC storm sources are .tcw warning files and one US Ice Center
// source is a .csv table (FR-SEL-009). The Smithsonian's pages are ColdFusion .cfm.
var pageExtensions = map[string]bool{
	".html": true, ".htm": true, ".shtml": true, ".php": true, ".asp": true, ".aspx": true, ".jsp": true,
	".cfm": true,
}

// IsPage reports whether an address names a page a browser shows, judged by the
// ending of its last path segment. The query and fragment are not part of that
// judgement.
func IsPage(address string) bool {
	rest := address
	if i := strings.IndexAny(rest, "?#"); i >= 0 {
		rest = rest[:i]
	}
	if i := strings.Index(rest, "://"); i >= 0 {
		rest = rest[i+len("://"):]
	}
	slash := strings.Index(rest, "/")
	if slash < 0 {
		return true
	}
	ext := strings.ToLower(path.Ext(rest[slash:]))
	return ext == "" || pageExtensions[ext]
}

// FirstPage answers the first address that is a page; when none is, the first
// address at all, so a source is never lost, only shown differently. It answers
// "" for none.
func FirstPage(addresses []string) string {
	for _, a := range addresses {
		if IsPage(a) {
			return a
		}
	}
	if len(addresses) > 0 {
		return addresses[0]
	}
	return ""
}

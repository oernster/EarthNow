// Package product is the one place this application's identity is written down.
// Ported from ED Voyage Companion's internal/product.
//
// The name reaches a reader in the window title, the setup program, the Start Menu
// and the Apps list. The same identity also names the install directory, the
// executable, the uninstall key and the data folder. Written out at each of those it
// becomes several files kept in step by whoever remembers, which is how a rename
// leaves one of them behind.
//
// Two forms, because two are genuinely needed. Name is what a reader sees and may hold
// punctuation. Slug is what a file system and a registry hold, so it carries no space
// and no character a path cannot. For EarthNow the two spell the same today; they
// stay separate so a display name with a space or a colon in it can never reach a
// shortcut file name or a registry key by accident.
//
// It sits under internal rather than in a layer because it belongs to none of them: it
// is a leaf that the composition root and the installer may both read without either
// depending on the other.
package product

const (
	// Name is the product as a reader meets it.
	Name = "EarthNow"

	// Slug is the same identity where only a file name will do. It must stay free of
	// the characters a path refuses.
	Slug = "EarthNow"

	// Licence names the licence of EarthNow's own code (CON-004); the full text is
	// LICENSE at the repository root.
	Licence = "GNU General Public License, version 3"

	// Copyright is the notice the About dialog shows (FR-HLP-001), with the
	// symbol rather than the word (owner).
	Copyright = "© Oliver Ernster 2026"

	// DonateURL is where the donate button sends a browser (FR-DON-003, FR-DON-004):
	// its one home. The application never fetches it; the address is handed to the
	// desktop, which opens it in the system browser (FR-DON-009).
	DonateURL = "https://www.paypal.com/ncp/payment/9LWU8TKV2MSRE"
)

// Attributions answers the credits the About dialog shows (NFR-LEG-002): the
// imagery and the event data, credited without implying endorsement. A fresh
// slice each call, so no caller can change what the next one reads.
func Attributions() []string {
	return []string{
		"Earth imagery: NASA Earth Observatory (Blue Marble Next Generation).",
		"Natural events: NASA Earth Observatory Natural Event Tracker (EONET).",
		"Earthquakes: USGS Earthquake Hazards Program.",
		// The Smithsonian's minimum citation is this name with its home page
		// (volcano.si.edu terms of use, read 2026-09-23).
		"Volcanoes: Global Volcanism Program, Smithsonian Institution (https://volcano.si.edu/) " +
			"with the USGS Volcano Hazards Program, Weekly Volcanic Activity Report.",
		"Place names and borders: Natural Earth.",
		// NFR-LEG-002: the wording the owner confirmed (ASM-007).
		"Cloud images: EUMETSAT, world cloud map (EUMETView).",
		"Neither NASA, the USGS, the Smithsonian nor EUMETSAT endorses " + Name + ".",
	}
}

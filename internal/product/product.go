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
)

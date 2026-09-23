package product

import (
	"strings"
	"testing"
)

// The address is asserted literally, so a typo fails here rather than sending a
// supporter to a page that is not the author's (FR-DON-003).
func TestDonateURLIsTheAuthorsPage(t *testing.T) {
	t.Parallel()
	const want = "https://www.paypal.com/ncp/payment/9LWU8TKV2MSRE"
	if DonateURL != want {
		t.Fatalf("DonateURL = %q, want %q", DonateURL, want)
	}
	if !strings.HasPrefix(DonateURL, "https://") {
		t.Fatalf("DonateURL %q is not https", DonateURL)
	}
}

// NFR-LEG-002: the credits name the imagery and both event sources and say
// that neither body endorses the product.
func TestAttributionsCreditTheSources(t *testing.T) {
	t.Parallel()
	all := strings.Join(Attributions(), " ")
	for _, want := range []string{"NASA Earth Observatory", "EONET", "USGS Earthquake Hazards Program", "endorses"} {
		if !strings.Contains(all, want) {
			t.Errorf("attributions lack %q: %s", want, all)
		}
	}
	Attributions()[0] = "changed"
	if Attributions()[0] == "changed" {
		t.Error("Attributions shares its slice between callers")
	}
}

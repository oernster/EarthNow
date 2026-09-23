package event

import "testing"

// FR-SEL-009: a data file is not a page. The addresses are the ones measured in
// the live EONET feed on 2026-09-23.
func TestIsPage(t *testing.T) {
	t.Parallel()
	for address, want := range map[string]bool{
		"https://www.metoc.navy.mil/jtwc/products/ep1726.tcw":                        false,
		"https://usicecenter.gov/pub/Iceberg_Tabular.csv":                            false,
		"https://www.nhc.noaa.gov/archive/2026/POLO.shtml":                           true,
		"https://usicecenter.gov/Products/AntarcIcebergs":                            true,
		"https://irwin.doi.gov/observer/incidents/2026-LALAS-000390":                 true,
		"https://earthquake.usgs.gov/earthquakes/eventpage/us7000abcd":               true,
		"https://example.org/report.PDF":                                             false,
		"https://example.org/page.html?download=file.csv#part.tcw":                   true,
		"https://example.org":                                                        true,
		"https://example.org/":                                                       true,
		"https://earthobservatory.nasa.gov/images/152848/antarctic-ice-shelf-spawns": true,
	} {
		if got := IsPage(address); got != want {
			t.Errorf("IsPage(%q) = %v, want %v", address, got, want)
		}
	}
}

func TestFirstPagePrefersAPageAndKeepsAFileWhenThereIsNone(t *testing.T) {
	t.Parallel()
	tcw := "https://www.metoc.navy.mil/jtwc/products/ep1726.tcw"
	nhc := "https://www.nhc.noaa.gov/archive/2026/POLO.shtml"
	if got := FirstPage([]string{tcw, nhc}); got != nhc {
		t.Errorf("Polo's sources answered %q, want the NHC page", got)
	}
	if got := FirstPage([]string{tcw}); got != tcw {
		t.Errorf("a file alone answered %q, want it kept", got)
	}
	if got := FirstPage(nil); got != "" {
		t.Errorf("no sources answered %q", got)
	}
}

// FR-PRV-001: a closed EONET event still shows, marked as ended.
func TestEndedReadsTheClosedStatus(t *testing.T) {
	t.Parallel()
	if !(Event{Status: StatusClosed}).Ended() || (Event{Status: StatusOpen}).Ended() || (Event{}).Ended() {
		t.Error("Ended misreads the status")
	}
}

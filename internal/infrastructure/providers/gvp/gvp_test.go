package gvp

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch"
)

// fixture is the live feed captured on 2026-09-23: twenty volcanoes.
func fixture(t *testing.T) []byte {
	t.Helper()
	body, err := os.ReadFile("testdata/weekly.xml")
	if err != nil {
		t.Fatal(err)
	}
	return body
}

type fakeFetcher struct {
	body   []byte
	err    error
	url    string
	accept string
}

func (f *fakeFetcher) GetAccepting(_ context.Context, rawURL, _, accept string) (httpfetch.Response, error) {
	f.url = rawURL
	f.accept = accept
	return httpfetch.Response{Body: f.body}, f.err
}

func TestFRPRV015_ParseCapturedFeed(t *testing.T) {
	t.Parallel()
	events, dropped, err := Parse(fixture(t))
	if err != nil || dropped != 0 || len(events) != 20 {
		t.Fatalf("events %d, dropped %d, err %v", len(events), dropped, err)
	}
	krakatau := events[0]
	if krakatau.Provider != event.GVP || krakatau.Category != event.Volcano || krakatau.ProviderEventID != "262000" {
		t.Errorf("Krakatau = %+v", krakatau)
	}
	if krakatau.Title != "Krakatau (Indonesia), new eruptive activity" {
		t.Errorf("title = %q", krakatau.Title)
	}
	o := krakatau.Observations[0]
	if o.Where.Lat != -6.1009 || o.Where.Lng != 105.4233 || o.Precision != event.Day {
		t.Errorf("observation = %+v", o)
	}
	// Published Thu, 17 Sep 2026 01:20:04 -0400: 05:20 UTC on the 17th, kept as the date.
	if !o.At.Equal(time.Date(2026, 9, 17, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("date = %v", o.At)
	}
	if krakatau.SourceURL != "https://volcano.si.edu/reports_weekly.cfm#vn_262000" || !event.IsPage(krakatau.SourceURL) {
		t.Errorf("source = %q", krakatau.SourceURL)
	}
	if strings.Contains(krakatau.Description, "<") || !strings.HasPrefix(krakatau.Description, "The Pusat Vulkanologi") {
		t.Errorf("description = %q", krakatau.Description)
	}
}

func TestUnusableItemsAreDroppedAndCounted(t *testing.T) {
	t.Parallel()
	item := func(title, guid, date, point string) string {
		return "<item><title>" + title + "</title><guid>" + guid + "</guid><pubDate>" + date +
			"</pubDate><georss:point>" + point + "</georss:point></item>"
	}
	good := item("A (B) - Report for x - Continuing Eruptive Activity", "u#vn_1", "Thu, 17 Sep 2026 01:20:04 -0400", "1 2")
	body := `<?xml version="1.0" encoding="ISO-8859-1"?><rss xmlns:georss="http://www.georss.org/georss"><channel>` +
		good +
		item("no parts", "u#vn_2", "Thu, 17 Sep 2026 01:20:04 -0400", "1 2") +
		item("A (B) - w - c", "no number", "Thu, 17 Sep 2026 01:20:04 -0400", "1 2") +
		item("A (B) - w - c", "u#vn_3", "not a date", "1 2") +
		item("A (B) - w - c", "u#vn_4", "Thu, 17 Sep 2026 01:20:04 -0400", "1") +
		item("A (B) - w - c", "u#vn_5", "Thu, 17 Sep 2026 01:20:04 -0400", "x 2") +
		item("A (B) - w - c", "u#vn_6", "Thu, 17 Sep 2026 01:20:04 -0400", "91 2") +
		"</channel></rss>"
	events, dropped, err := Parse([]byte(body))
	if err != nil || len(events) != 1 || dropped != 6 {
		t.Fatalf("events %d, dropped %d, err %v", len(events), dropped, err)
	}
}

func TestADocumentThatIsNotTheFeedIsAnError(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"not xml",
		`<?xml version="1.0" encoding="UTF-16"?><rss></rss>`,
		`<other></other>`,
	} {
		if _, _, err := Parse([]byte(body)); err == nil {
			t.Errorf("%q parsed", body)
		}
	}
}

func TestLatin1DecodesEveryByteAsItsCodePoint(t *testing.T) {
	t.Parallel()
	body := []byte("<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?><rss xmlns:georss=\"http://www.georss.org/georss\"><channel>" +
		"<item><title>Nevado del Ruiz (Colombia) - Report for w - Continuing Eruptive Activity</title><description>Se\xf1al</description>" +
		"<guid>u#vn_351020</guid><pubDate>Thu, 17 Sep 2026 01:20:04 -0400</pubDate><georss:point>4.892 -75.324</georss:point></item>" +
		"</channel></rss>")
	events, _, err := Parse(body)
	if err != nil || len(events) != 1 || events[0].Description != "Señal" {
		t.Fatalf("events %+v, err %v", events, err)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("broken") }

func TestLatin1PassesAReadFailureOn(t *testing.T) {
	t.Parallel()
	if _, err := latin1("ISO-8859-1", failingReader{}); err == nil {
		t.Error("a failed read was swallowed")
	}
}

func TestFetchAsksForTheFeedAndNamesItselfInErrors(t *testing.T) {
	t.Parallel()
	f := &fakeFetcher{body: fixture(t)}
	a := New(f)
	got, err := a.Fetch(context.Background(), "")
	if err != nil || len(got.Events) != 20 || f.url != "https://volcano.si.edu/news/WeeklyVolcanoRSS.xml" {
		t.Fatalf("fetch = %d events, %v, %q", len(got.Events), err, f.url)
	}
	if f.accept != httpfetch.AcceptXML {
		t.Errorf("accept = %q; the feed refuses a JSON-only request", f.accept)
	}
	if a.Name() != event.GVP || a.Interval() != Interval {
		t.Error("identity")
	}
	if _, err := New(&fakeFetcher{err: errors.New("down")}).Fetch(context.Background(), ""); err == nil || !strings.HasPrefix(err.Error(), "GVP: ") {
		t.Errorf("fetch error = %v", err)
	}
	if _, err := New(&fakeFetcher{body: []byte("junk")}).Fetch(context.Background(), ""); err == nil || !strings.HasPrefix(err.Error(), "GVP: ") {
		t.Errorf("parse error = %v", err)
	}
}

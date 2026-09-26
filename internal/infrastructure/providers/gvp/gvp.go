// Package gvp adapts the Smithsonian / USGS Weekly Volcanic Activity Report to
// the event model (REQUIREMENTS.md FR-PRV-015). Nothing outside this package
// knows the report's schema.
//
// The feed is RSS 2.0 declared as ISO-8859-1, measured on 2026-09-23: each item
// carries a title "Name (Country) - Report for <week> - <activity>", an HTML
// description, a guid permalink ending in the volcano number, a publish date and
// a georss:point "lat lng". It is issued once a week, so each report is dated
// by its publish date at day precision rather than claiming an instant.
package gvp

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/oernster/EarthNow/internal/application/ports"
	"github.com/oernster/EarthNow/internal/domain/event"
	"github.com/oernster/EarthNow/internal/infrastructure/httpfetch"
)

// Host is the only host this adapter reaches.
const Host = "volcano.si.edu"

// feedPath is the weekly report's RSS feed.
const feedPath = "/news/WeeklyVolcanoRSS.xml"

// Interval is the report's refresh interval (FR-PRV-005). The report is issued
// weekly (its own channel description: by 2300 UTC every Thursday), so an hourly
// look finds a new one within the hour and costs one 26 KB request.
const Interval = time.Hour

// coordinateFields is a georss point's two numbers, latitude then longitude.
const coordinateFields = 2

// titleParts is the number of " - " separated parts of an item title: the
// volcano, the report week and the activity.
const titleParts = 3

// Fetcher is what the adapter needs from the HTTP client.
type Fetcher interface {
	GetAccepting(ctx context.Context, rawURL, validator, accept string) (httpfetch.Response, error)
}

// Adapter is the weekly volcanic activity report provider.
type Adapter struct {
	fetcher Fetcher
}

// New builds the adapter over fetcher.
func New(fetcher Fetcher) *Adapter { return &Adapter{fetcher: fetcher} }

// Name implements ports.Provider.
func (a *Adapter) Name() event.Provider { return event.GVP }

// Interval implements ports.Provider.
func (a *Adapter) Interval() time.Duration { return Interval }

// URL is the request of FR-PRV-015.
func URL() string { return "https://" + Host + feedPath }

// Fetch implements ports.Provider. Every fetch is a full one, asking for XML:
// the feed answers 403 to a request asking only for JSON (measured 2026-09-23).
func (a *Adapter) Fetch(ctx context.Context, _ string) (ports.Fetched, error) {
	resp, err := a.fetcher.GetAccepting(ctx, URL(), "", httpfetch.AcceptXML)
	if err != nil {
		return ports.Fetched{}, fmt.Errorf("GVP: %w", err)
	}
	events, dropped, err := Parse(resp.Body)
	if err != nil {
		return ports.Fetched{}, fmt.Errorf("GVP: %w", err)
	}
	return ports.Fetched{Events: events, Dropped: dropped}, nil
}

type wireItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Point       string `xml:"http://www.georss.org/georss point"`
}

// Parse maps a weekly report feed. A body that is not the expected document is
// an error; a single unusable item is dropped and counted (FR-PRV-012,
// FR-PRV-013).
func Parse(body []byte) ([]event.Event, int, error) {
	var doc struct {
		XMLName xml.Name   `xml:"rss"`
		Items   []wireItem `xml:"channel>item"`
	}
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.CharsetReader = latin1
	if err := decoder.Decode(&doc); err != nil {
		return nil, 0, fmt.Errorf("parsing the report: %w", err)
	}
	var out []event.Event
	dropped := 0
	for _, w := range doc.Items {
		e, ok := mapItem(w)
		if !ok {
			dropped++
			continue
		}
		out = append(out, e)
	}
	return out, dropped, nil
}

// latin1 decodes the feed's declared ISO-8859-1, in which every byte is the
// code point of the same number; anything else is refused rather than guessed.
func latin1(label string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(label) {
	case "iso-8859-1", "latin1", "latin-1":
	default:
		return nil, fmt.Errorf("unsupported encoding %q", label)
	}
	raw, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	for _, c := range raw {
		b.WriteRune(rune(c))
	}
	return strings.NewReader(b.String()), nil
}

func mapItem(w wireItem) (event.Event, bool) {
	parts := strings.Split(w.Title, " - ")
	id := volcanoNumber(w.GUID)
	if len(parts) != titleParts || id == "" {
		return event.Event{}, false
	}
	where, ok := point(w.Point)
	if !ok {
		return event.Event{}, false
	}
	issued, err := time.Parse(time.RFC1123Z, strings.TrimSpace(w.PubDate))
	if err != nil {
		return event.Event{}, false
	}
	y, m, d := issued.UTC().Date()
	issueDay := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	from, to, ok := reportWeek(parts[1])
	if !ok {
		return event.Event{}, false
	}
	activity := strings.TrimSpace(parts[2])
	return event.Event{
		Provider:        event.GVP,
		ProviderEventID: id,
		Category:        event.Volcano,
		Title:           strings.TrimSpace(parts[0]) + ", " + strings.ToLower(activity),
		Description:     plainText(w.Description),
		Observations:    []event.Observation{{At: issueDay, Precision: event.Day, Where: where}},
		Status:          event.StatusOpen,
		SourceURL:       strings.TrimSpace(w.GUID),
		Extras:          event.Extras{SourceCategory: activity},
		Report:          event.Report{WeekFrom: from, WeekTo: to, Issued: issueDay},
	}, true
}

// weekPrefix opens a title's week part: "Report for 10 September-16 September 2026".
const weekPrefix = "Report for "

// Layouts of the week's two days. Measured on 2026-09-23 and 2026-09-26: the
// last day carries its year and the first carries none. A first day with a
// year is accepted too, since how a week crossing New Year is written has not
// been seen.
const (
	dayWithYear    = "2 January 2006"
	dayWithoutYear = "2 January"
)

// reportWeek reads a title's week part as its first and last UTC days
// (FR-PRV-015). A first day without a year takes the last day's, less one when
// its month falls after the last day's month; false when either day is unreadable.
func reportWeek(part string) (time.Time, time.Time, bool) {
	rest, found := strings.CutPrefix(strings.TrimSpace(part), weekPrefix)
	first, last, split := strings.Cut(rest, "-")
	if !found || !split {
		return time.Time{}, time.Time{}, false
	}
	to, err := time.Parse(dayWithYear, strings.TrimSpace(last))
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	if from, err := time.Parse(dayWithYear, strings.TrimSpace(first)); err == nil {
		return from, to, true
	}
	from, err := time.Parse(dayWithoutYear, strings.TrimSpace(first))
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	year := to.Year()
	if from.Month() > to.Month() {
		year--
	}
	return time.Date(year, from.Month(), from.Day(), 0, 0, 0, 0, time.UTC), to, true
}

// volcanoNumber is the id after "#vn_" in a guid; "" when there is none.
func volcanoNumber(guid string) string {
	_, number, found := strings.Cut(guid, "#vn_")
	if !found {
		return ""
	}
	return strings.TrimSpace(number)
}

// point reads a georss point, "lat lng", through the domain's validation.
func point(raw string) (event.Point, bool) {
	fields := strings.Fields(raw)
	if len(fields) != coordinateFields {
		return event.Point{}, false
	}
	lat, errLat := strconv.ParseFloat(fields[0], 64)
	lng, errLng := strconv.ParseFloat(fields[1], 64)
	if errLat != nil || errLng != nil {
		return event.Point{}, false
	}
	p, err := event.NewPoint(lat, lng)
	return p, err == nil
}

// plainText turns the description's HTML into the plain text the detail panel
// shows: tags dropped, entities decoded, paragraphs kept apart by a blank line.
func plainText(markup string) string {
	var b strings.Builder
	inTag := false
	for _, r := range strings.ReplaceAll(markup, "</p>", "\n\n") {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(html.UnescapeString(b.String()))
}

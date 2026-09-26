package event

import (
	"reflect"
	"slices"
	"testing"
)

func fieldNames(v any) []string {
	t := reflect.TypeOf(v)
	names := make([]string, t.NumField())
	for i := range names {
		names[i] = t.Field(i).Name
	}
	return names
}

// DATA-001: the event holds what the requirement lists and nothing else, each
// sighting as an observation. A field a source may leave out (depth, a
// measurement) is a pointer, so an absent value stays absent rather than
// becoming a plausible zero.
func TestDATA001_TheEventHoldsWhatItLists(t *testing.T) {
	t.Parallel()
	want := map[string][]string{
		"Event":       {"Provider", "ProviderEventID", "Category", "Title", "Description", "Observations", "UpdatedAt", "Status", "SourceURL", "Extras", "Report"},
		"Observation": {"At", "Precision", "Where", "Measurement"},
		"Report":      {"WeekFrom", "WeekTo", "Issued"},
		"Extras":      {"SourceCategory", "DepthKm", "TsunamiFlag", "MagnitudeType"},
		"Measurement": {"Value", "Unit"},
	}
	got := map[string][]string{
		"Event": fieldNames(Event{}), "Observation": fieldNames(Observation{}),
		"Extras": fieldNames(Extras{}), "Measurement": fieldNames(Measurement{}),
		"Report": fieldNames(Report{}),
	}
	for name, fields := range want {
		if !slices.Equal(got[name], fields) {
			t.Errorf("%s holds %v, want %v", name, got[name], fields)
		}
	}
	optional := map[string]reflect.Type{
		"Extras.DepthKm":          reflect.TypeOf(Extras{}.DepthKm),
		"Observation.Measurement": reflect.TypeOf(Observation{}.Measurement),
	}
	for name, typ := range optional {
		if typ.Kind() != reflect.Pointer {
			t.Errorf("%s is %v: an absent value could not stay absent", name, typ)
		}
	}
}

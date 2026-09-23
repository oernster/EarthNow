package event

// The earthquake size bands of FR-MRK-003: below 3.0, 3.0 to 4.5, 4.5 to 6,
// 6 and above. Only a source-supplied magnitude sizes a marker.
const (
	bandSmall  = 3.0
	bandMedium = 4.5
	bandLarge  = 6.0
)

// Band answers the marker size band for an event: 0 for any event without a
// magnitude to size by (every non-earthquake, FR-MRK-004), else 1 to 4.
func Band(e Event, o Observation) int {
	if e.Category != Earthquake || o.Measurement == nil {
		return 0
	}
	switch m := o.Measurement.Value; {
	case m < bandSmall:
		return 1
	case m < bandMedium:
		return 2
	case m < bandLarge:
		return 3
	default:
		return 4
	}
}

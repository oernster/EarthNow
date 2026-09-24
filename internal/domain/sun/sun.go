// Package sun holds the rules of the day and night layer (REQUIREMENTS.md
// 3.2.11): where the sun stands overhead at an instant (FR-DAY-001), how high it
// stands at a point (FR-DAY-002) and how that becomes the light the globe is
// drawn in, clouds included (FR-DAY-002, FR-DAY-009).
//
// The position follows NOAA's solar position equations, as published with its
// solar calculator (gml.noaa.gov/grad/solcalc/calcdetails.html); the
// coefficients below are that published algorithm's, held as named tables.
package sun

import (
	"math"
	"time"

	"github.com/oernster/EarthNow/internal/domain/event"
)

// TwilightDegrees is FR-DAY-002's limit either side of the horizon: the light
// is none at this far below it, full at this far above, linear between.
const TwilightDegrees = 6

// CloudNightFloor is FR-DAY-009's share of a cloud's opacity kept at night, so
// clouds dim over the dark side without vanishing.
const CloudNightFloor = 0.25

// The Julian day count: the Unix epoch's Julian day, the J2000.0 epoch and the
// days in a Julian century.
const (
	unixEpochJulianDay = 2440587.5
	j2000JulianDay     = 2451545
	daysPerCentury     = 36525
)

// Angular constants: a full turn, the degrees the sun's hour angle moves a
// minute (a turn a day) and the minutes of a day.
const (
	fullTurn         = 360
	halfTurn         = fullTurn / 2
	minutesPerDay    = 24 * 60
	degreesPerMinute = float64(fullTurn) / minutesPerDay
)

// NOAA's polynomials in Julian centuries since J2000.0, lowest power first.
var (
	geomMeanLongitude = []float64{280.46646, 36000.76983, 0.0003032}
	geomMeanAnomaly   = []float64{357.52911, 35999.05029, -0.0001537}
	orbitEccentricity = []float64{0.016708634, -0.000042037, -0.0000001267}
	centreFirst       = []float64{1.914602, -0.004817, -0.000014}
	centreSecond      = []float64{0.019993, -0.000101}
	centreThird       = []float64{0.000289}
	nutationNode      = []float64{125.04, -1934.136}
	// Mean obliquity in arcseconds past 23 degrees 26 minutes.
	obliquitySeconds = []float64{21.448, -46.815, -0.00059, 0.001813}
)

// The remaining constants of NOAA's equations: the aberration and nutation
// corrections to the sun's longitude, the nutation correction to the
// obliquity and the whole degrees and minutes the mean obliquity starts from.
const (
	aberration          = 0.00569
	nutationInLongitude = 0.00478
	nutationInObliquity = 0.00256
	obliquityDegrees    = 23
	obliquityMinutes    = 26
	arcPerUnit          = 60
)

// Coefficients of NOAA's equation of time series, in radians before the
// conversion to minutes.
const (
	eotEccentricityTerm = 2
	eotCrossTerm        = 4
	eotSquareTerm       = 0.5
	eotEccentricSquare  = 1.25
)

// Subsolar is FR-DAY-001: the point where the sun stands overhead at t, its
// latitude the solar declination and its longitude the hour angle at
// Greenwich corrected by the equation of time.
func Subsolar(t time.Time) event.Point {
	c := centuries(t)
	declination, eot := position(c)
	utc := t.UTC()
	midnight := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
	minutes := utc.Sub(midnight).Minutes()
	noon := float64(minutesPerDay) / 2
	return event.Point{Lat: declination, Lng: normalise((noon - minutes - eot) * degreesPerMinute)}
}

// Elevation is the sun's height above the horizon at p, in degrees, with the
// sun overhead at subsolar (geometric: no refraction).
func Elevation(p, subsolar event.Point) float64 {
	lat, sunLat := radians(p.Lat), radians(subsolar.Lat)
	cosZenith := math.Sin(lat)*math.Sin(sunLat) +
		math.Cos(lat)*math.Cos(sunLat)*math.Cos(radians(p.Lng-subsolar.Lng))
	return degrees(math.Asin(math.Max(-1, math.Min(1, cosZenith))))
}

// Light is FR-DAY-002: 0 at or below the twilight limit under the horizon, 1
// at or above it over the horizon, linear between.
func Light(elevation float64) float64 {
	return math.Max(0, math.Min(1, (elevation+TwilightDegrees)/(2*TwilightDegrees)))
}

// CloudOpacity is FR-DAY-009: the share of a cloud's opacity kept at a point
// with the given light, from the night floor at none to all of it at full.
func CloudOpacity(light float64) float64 {
	return CloudNightFloor + (1-CloudNightFloor)*light
}

// centuries is the Julian centuries from J2000.0 to t.
func centuries(t time.Time) float64 {
	seconds := float64(t.UnixNano()) / float64(time.Second)
	julianDay := seconds/(24*time.Hour).Seconds() + unixEpochJulianDay
	return (julianDay - j2000JulianDay) / daysPerCentury
}

// position is NOAA's solar declination in degrees and equation of time in
// minutes at c Julian centuries from J2000.0.
func position(c float64) (declination, eot float64) {
	meanLong := radians(math.Mod(poly(c, geomMeanLongitude), fullTurn))
	anomaly := radians(poly(c, geomMeanAnomaly))
	eccentricity := poly(c, orbitEccentricity)
	centre := math.Sin(anomaly)*poly(c, centreFirst) +
		math.Sin(2*anomaly)*poly(c, centreSecond) +
		math.Sin(3*anomaly)*poly(c, centreThird)
	node := radians(poly(c, nutationNode))
	apparent := radians(degrees(meanLong) + centre - aberration - nutationInLongitude*math.Sin(node))
	meanObliquity := obliquityDegrees +
		(obliquityMinutes+poly(c, obliquitySeconds)/arcPerUnit)/arcPerUnit
	obliquity := radians(meanObliquity + nutationInObliquity*math.Cos(node))

	declination = degrees(math.Asin(math.Sin(obliquity) * math.Sin(apparent)))
	y := math.Pow(math.Tan(obliquity/2), 2)
	e := eccentricity
	radiansOfTime := y*math.Sin(2*meanLong) -
		eotEccentricityTerm*e*math.Sin(anomaly) +
		eotCrossTerm*e*y*math.Sin(anomaly)*math.Cos(2*meanLong) -
		eotSquareTerm*y*y*math.Sin(4*meanLong) -
		eotEccentricSquare*e*e*math.Sin(2*anomaly)
	return declination, degrees(radiansOfTime) / degreesPerMinute
}

// poly evaluates a polynomial in x, coefficients lowest power first.
func poly(x float64, coefficients []float64) float64 {
	sum := 0.0
	for i := len(coefficients) - 1; i >= 0; i-- {
		sum = sum*x + coefficients[i]
	}
	return sum
}

// normalise brings a longitude into -180 to 180.
func normalise(lng float64) float64 {
	lng = math.Mod(lng+halfTurn, fullTurn)
	if lng < 0 {
		lng += fullTurn
	}
	return lng - halfTurn
}

func radians(deg float64) float64 { return deg * math.Pi / halfTurn }
func degrees(rad float64) float64 { return rad * halfTurn / math.Pi }

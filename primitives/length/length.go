// Package length holds the lengths physics and astronomy are measured
// in, each in metres and each named once. Multiples come from the si's
// prefixes: a kiloparsec is si.Kilo(Parsec), a micron is si.Micro(Metre).
//
// Every length here but the Planck length is exact, by a definition
// cited beside it.
package length

import (
	"math"

	"github.com/janearc/libtheme-css/primitives/si"
)

// The lengths, in metres.
const (
	// Metre is the si's unit of length, fixed since 1983 by the speed of
	// light: the distance light travels in 1/299792458 of a second.
	Metre = 1.0

	// Angstrom is 1e-10 metres, a unit outside the si that x-ray papers
	// still use. SI Brochure, 9th edition, section 4, and ISO 80000-3.
	Angstrom = 1e-10

	// AU is the astronomical unit, fixed in metres by the IAU in 2012
	// (resolution B2). It began as the earth's mean distance from the
	// sun and is now a definition.
	AU = 149597870700.0

	// Parsec is the distance at which one AU subtends one arcsecond:
	// 648000/pi AU, by the IAU's 2015 resolution B2. 648000 is the
	// number of arcseconds in half a circle.
	Parsec = arcsecondsInHalfACircle / math.Pi * AU

	// LightYear is the distance light travels in a Julian year, by the
	// IAU's definition: c times 365.25 days of 86400 seconds.
	LightYear = si.Speed * julianYear
)

// arcsecondsInHalfACircle is 180 degrees of 3600 arcseconds each.
const arcsecondsInHalfACircle = 180 * 3600

// julianYear is the IAU's year for the light-year, in seconds: 365.25
// days of 86400 seconds, both fixed by definition.
const julianYear = 365.25 * 86400

// PlanckLength is sqrt(hbar G / c^3), about 1.616e-35 metres. It is the
// one length here that is measured, not defined: hbar and c are exact,
// but G carries the uncertainty si.Gravitation states, and the Planck
// length inherits half of it.
var PlanckLength = math.Sqrt(hbar * si.Gravitation / math.Pow(si.Speed, 3))

// hbar is the reduced Planck constant, h over two pi, exact since h is.
const hbar = si.Planck / (2 * math.Pi)

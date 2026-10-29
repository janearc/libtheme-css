package swatch_test

import (
	"math"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
)

// The published ranges of the two tables, in nanometres. They are defined by
// the CIE's tables, which data/README.md in this package gives the source of.
const (
	// observerFrom is the shortest wavelength of the 1931 observer table.
	// It is defined by the table.
	observerFrom = 360
	// observerTo is the longest wavelength of the 1931 observer table. It
	// is defined by the table.
	observerTo = 830
	// daylightFrom is the shortest wavelength of the D65 table. It is
	// defined by the table.
	daylightFrom = 300
	// daylightTo is the longest wavelength of the D65 table. It is defined
	// by the table.
	daylightTo = 830
	// perNanometre is the step between rows. It is defined by the tables.
	perNanometre = 1
)

// The embedded tables are whole: the observer runs 360 to 830 nm and
// daylight 300 to 830, a row at every nanometre. The reader refuses a table
// with a gap, so a table that loads is a table that is all there.
func TestTheEmbeddedTablesAreWhole(t *testing.T) {
	low, high := swatch.Observed()
	if low != observerFrom || high != observerTo {
		t.Errorf("the observer runs %d to %d, not %d to %d", low, high,
			observerFrom, observerTo)
	}
	daylight := swatch.D65()
	want := daylightTo - daylightFrom + perNanometre
	if len(daylight) != want {
		t.Errorf("daylight has %d rows, want %d", len(daylight), want)
	}
	for _, edge := range []int{daylightFrom, daylightTo} {
		if _, has := daylight[edge]; !has {
			t.Errorf("daylight has no row at %d nm", edge)
		}
	}
}

// The white and the dark the library computes with are its own. Programs can
// assign to the exported White or Black. That changes what they read there
// and not what WhitePoint and NoLight answer.
func TestAnAssignmentToWhiteLeavesTheLibrarysWhiteAlone(t *testing.T) {
	white, dark := swatch.White, swatch.Black
	swatch.White, swatch.Black = dark, white
	whitePoint, noLight := swatch.WhitePoint(), swatch.NoLight()
	swatch.White, swatch.Black = white, dark
	if whitePoint != white {
		t.Errorf("the white point followed an assignment: %v", whitePoint)
	}
	if noLight != dark {
		t.Errorf("no light followed an assignment: %v", noLight)
	}
}

// Temperatures that give no light, and the other numbers of the tests below.
const (
	// absoluteZero is the lowest temperature there is. It is defined by
	// the kelvin scale.
	absoluteZero = 0
	// belowAbsoluteZero is a temperature there is not. It is defined by
	// the kelvin scale.
	belowAbsoluteZero = -1
	// noRows is an empty spectrum. It is defined by counting.
	noRows = 0
	// aLuminance is any brightness; no light has none of it. It is defined
	// by this test: the answer must not depend on it.
	aLuminance = 1
)

// A black body at zero kelvin or below gives no light and no spectrum.
// No light has the white's chromaticity.
func TestNoTemperatureIsNoLight(t *testing.T) {
	for _, kelvin := range []float64{absoluteZero, belowAbsoluteZero,
		math.Inf(belowAbsoluteZero)} {
		if swatch.Planckian(kelvin) != swatch.NoLight() {
			t.Errorf("%vK gave light", kelvin)
		}
		if len(swatch.Blackbody(kelvin)) != noRows {
			t.Errorf("%vK gave a spectrum", kelvin)
		}
	}
	if swatch.NoLight().XY() != swatch.WhitePoint().XY() {
		t.Error("no light does not answer the white's chromaticity")
	}
	if swatch.FromXY(swatch.XY{}, aLuminance) != swatch.NoLight() {
		t.Error("a chromaticity with no Y gave light")
	}
}

// Corners of a gamut: a line that runs up from 0.3, and a point away from it.
const (
	// nearEnd is where the line starts, in both x and y. It is defined by
	// this test.
	nearEnd = 0.3
	// farEnd is where the line ends in y. It is defined by this test.
	farEnd = 0.5
	// outside is a place beyond the far end, in x and y. It is defined by
	// this test.
	outside = 0.9
)

// Two corners on one point make an edge with no length. The fit still
// lands on the other edge, at its nearer end.
func TestAGamutWithAnEdgeOfNoLengthStillFits(t *testing.T) {
	start := swatch.XY{X: nearEnd, Y: nearEnd}
	end := swatch.XY{X: nearEnd, Y: farEnd}
	got := swatch.Gamut{Red: start, Green: start, Blue: end}.Fit(
		swatch.XY{X: outside, Y: outside})
	if got != end {
		t.Errorf("fit to %v, not %v", got, end)
	}
}

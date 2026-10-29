package ok_test

import (
	"math"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
)

// Numbers the tests use that the arithmetic or the scale defines.
const (
	// noChroma is no chroma. It is defined by the scale.
	noChroma = 0
	// oneStep is one step of a count. It is defined by counting.
	oneStep = 1
	// bisection is the factor a range shrinks by at each halving. It is
	// defined by bisection.
	bisection = 2
)

// box lets in the colours up to a chroma.
type box struct{ edge float64 }

// holds says whether a swatch is in the box.
func (limit box) holds(colour swatch.Swatch) bool {
	return ok.FromSwatch(colour).Polar().C <= limit.edge
}

// everything lets every colour in.
func everything(swatch.Swatch) bool { return true }

// halvings is the lower end after some halvings of a range from nothing
// to top, hunting for a limit. It is the largest multiple of top over
// two to the count that is not past the limit.
func halvings(top, limit float64, count int) float64 {
	step := top / math.Pow(bisection, float64(count))
	return math.Floor(limit/step) * step
}

// Fit halves its range 24 times. Its answer is the 24th halving of the range.
func TestFitHalvesTwentyFourTimes(t *testing.T) {
	fit := pickedNumbers(t).Fit
	limit := box{fit.BoxEdge}
	got, lost := ok.Fit(fit.Colour, limit.holds)
	want := halvings(fit.Colour.C, fit.BoxEdge, fit.Halvings)
	if math.Abs(got.C-want) > fit.Tolerance || math.Abs(lost-(fit.Colour.C-want)) > fit.Tolerance {
		t.Errorf("chroma %v lost %v, the halvings give %v", got.C, lost, want)
	}
}

// Other counts give other answers. The pick of start chroma tells them apart.
func TestFitHalvingsOtherCountsGiveOtherAnswers(t *testing.T) {
	fit := pickedNumbers(t).Fit
	right := halvings(fit.Colour.C, fit.BoxEdge, fit.Halvings)
	for _, count := range []int{fit.Halvings - oneStep, fit.Halvings + oneStep} {
		other := halvings(fit.Colour.C, fit.BoxEdge, count)
		if math.Abs(other-right) <= fit.Tolerance {
			t.Errorf("%d halvings give the same answer as %d", count, fit.Halvings)
		}
	}
}

// Colours with a hair of chroma are grey. They have no chroma and no hue.
func TestPolarOfAHairOfChromaIsGrey(t *testing.T) {
	polar := pickedNumbers(t).Polar
	if got := polar.AHair.Polar(); got != polar.Grey {
		t.Errorf("a hair of chroma is %+v, not %+v", got, polar.Grey)
	}
}

// Colours with exactly Exact of chroma are colours. Grey is below Exact.
func TestPolarAtExactChromaIsAColour(t *testing.T) {
	polar := pickedNumbers(t).Polar
	got := polar.ExactlyExact.Polar()
	if got.C != ok.Exact {
		t.Errorf("chroma Exact is %+v", got)
	}
}

// Hues below the red axis are counted round to 360.
func TestPolarHueGoesRoundTheWheel(t *testing.T) {
	polar := pickedNumbers(t).Polar
	for _, want := range polar.Wheel {
		got := want.Lab.Polar().H
		if math.Abs(got-want.Hue) > polar.DegreeSlack {
			t.Errorf("%+v is at %v degrees, not %v", want.Lab, got, want.Hue)
		}
	}
}

// The space's white is the library's white to the three places of the reference values.
func TestWhiteIsWhereTheLibrarysWhiteIs(t *testing.T) {
	slack := pickedNumbers(t).White.Tolerance
	derived := ok.FromSwatch(swatch.WhitePoint())
	if gap := ok.Distance(derived, ok.White); gap > slack {
		t.Errorf("the library's white is %v from White", gap)
	}
}

// White and black are a whole lightness apart.
func TestWhiteIsNotBlack(t *testing.T) {
	slack := pickedNumbers(t).White.Tolerance
	if gap := ok.Distance(ok.White, ok.Black); gap < ok.White.L-slack {
		t.Errorf("white and black are %v apart", gap)
	}
}

// Colours exactly a tolerance apart are the same. Same includes the tolerance.
func TestSameIncludesTheToleranceItself(t *testing.T) {
	var origin ok.OKLab
	away := pickedNumbers(t).Same.Apart
	tolerance := ok.Distance(origin, away)
	if !ok.Same(origin, away, tolerance) {
		t.Errorf("colours %v apart are not the same within %v", tolerance, tolerance)
	}
}

// Tolerances a hair under the distance are not enough.
func TestSameRefusesUnderTheTolerance(t *testing.T) {
	var origin ok.OKLab
	away := pickedNumbers(t).Same.Apart
	under := math.Nextafter(ok.Distance(origin, away), noChroma)
	if ok.Same(origin, away, under) {
		t.Errorf("colours %v apart are the same within %v", ok.Distance(origin, away), under)
	}
}

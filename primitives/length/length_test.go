package length

import (
	"math"
	"testing"

	"github.com/janearc/libtheme-css/primitives/si"
)

// near is two numbers the same to a part in a hundred million.
func near(a, b float64) bool {
	return math.Abs(a-b) <= 1e-8*math.Max(math.Abs(a), math.Abs(b))
}

// The lengths are the published figures: the IAU's for the AU, parsec
// and light-year, the SI's for the angstrom, and CODATA 2018's
// 1.616255e-35 m for the Planck length, which is good to about one
// part in 100000.
func TestTheLengthsAreThePublishedFigures(t *testing.T) {
	exact := map[string][2]float64{
		"metre":      {Metre, 1},
		"angstrom":   {Angstrom, 1e-10},
		"au":         {AU, 149597870700},
		"light-year": {LightYear, 9460730472580800},
	}
	for name, pair := range exact {
		if pair[0] != pair[1] {
			t.Errorf("the %s is %v m, not %v", name, pair[0], pair[1])
		}
	}
	if !near(Parsec, 3.0856775814913673e16) {
		t.Errorf("the parsec is %v m", Parsec)
	}
	if math.Abs(PlanckLength-1.616255e-35) > 1e-5*1.616255e-35 {
		t.Errorf("the Planck length is %v m", PlanckLength)
	}
}

// A parsec is the distance at which an AU subtends one arcsecond: an
// AU over a parsec is that arcsecond, in radians.
func TestAParsecIsWhereAnAUIsAnArcsecond(t *testing.T) {
	arcsecond := math.Pi / arcsecondsInHalfACircle
	if !near(math.Atan(AU/Parsec), arcsecond) {
		t.Errorf("an AU at a parsec is %v rad", math.Atan(AU/Parsec))
	}
}

// Multiples come from the si's prefixes, not from names here: a micron
// is a millionth of a metre and a kiloparsec a thousand parsecs.
func TestMultiplesComeFromThePrefixes(t *testing.T) {
	if got := si.Micro(Metre); got != 1e-6 {
		t.Errorf("a micron is %v m", got)
	}
	if got := si.Kilo(Parsec); !near(got, 1000*Parsec) {
		t.Errorf("a kiloparsec is %v m", got)
	}
}

// The anti-test: the light-year is not the parsec, and neither is the
// AU; a slip that set one to another shows.
func TestTheLengthsAreDistinct(t *testing.T) {
	if LightYear == Parsec || AU == LightYear || AU == Parsec {
		t.Error("two astronomical lengths are the same")
	}
	if !(AU < LightYear && LightYear < Parsec) {
		t.Errorf("AU %v, light-year %v and parsec %v are out of order",
			AU, LightYear, Parsec)
	}
}

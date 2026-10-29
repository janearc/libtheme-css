package srgb_test

import (
	"math"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

const (
	// positive is the sign math.Inf is asked for to get plus infinity. It is
	// defined by math.Inf.
	positive = +1
	// negative is the sign math.Inf is asked for to get minus infinity. It
	// is defined by math.Inf.
	negative = -1
)

// nonFinite are the values that are not numbers on a line: not a number and
// the two infinities.
func nonFinite() []float64 {
	return []float64{math.NaN(), math.Inf(positive), math.Inf(negative)}
}

// bent is white's XYZ with one part replaced, in each of the three places.
func bent(part float64) []swatch.Swatch {
	x, y, z := srgb.White.Swatch().XYZ()
	return []swatch.Swatch{swatch.FromXYZ(part, y, z),
		swatch.FromXYZ(x, part, z), swatch.FromXYZ(x, y, part)}
}

// either asks math.IsInf for an infinity of either sign. It is defined by
// math.IsInf.
const either = 0

// isFinite is whether a number is neither infinite nor not a number.
func isFinite(number float64) bool {
	return !math.IsNaN(number) && !math.IsInf(number, either)
}

// The anti-test: a swatch with a part that is not finite is never in gamut,
// and its lamps are finite, so nothing meaningless is written to a device.
func TestANonFiniteSwatchIsNotInGamut(t *testing.T) {
	for _, part := range nonFinite() {
		for _, odd := range bent(part) {
			lamps, inGamut := srgb.FromSwatch(odd)
			if inGamut {
				t.Errorf("a swatch with a part %v is in gamut", part)
			}
			if !isFinite(lamps.R + lamps.G + lamps.B) {
				t.Errorf("lamps %+v are not finite", lamps)
			}
		}
	}
}

// Lamps that are not numbers are written as the dark byte. They are not left
// to whatever the platform makes of a conversion. Some platforms make the dark
// byte of it anyway (measured on arm64, read for amd64). There this test
// cannot fail without the guard. It holds what is written, not the guard.
func TestANaNLampIsWrittenDark(t *testing.T) {
	odd := srgb.White
	odd.R = math.NaN()
	red, green, blue := odd.Bytes()
	none, _, _ := srgb.Black.Bytes()
	if red != none || green == none || blue == none {
		t.Errorf("bytes %d %d %d", red, green, blue)
	}
}

// The property: whatever the XYZ, the lamps are finite and between dark and
// full; and a colour is in gamut only if all its parts are finite.
func TestLampsAreFiniteWhateverTheLight(t *testing.T) {
	rapid.Check(t, lampsAreFinite)
}

// lampsAreFinite is one draw of the property above.
func lampsAreFinite(t *rapid.T) {
	odd := rapid.SampledFrom(nonFinite())
	part := rapid.OneOf(rapid.Float64(), odd)
	x, y, z := part.Draw(t, "x"), part.Draw(t, "y"), part.Draw(t, "z")
	lamps, inGamut := srgb.FromSwatch(swatch.FromXYZ(x, y, z))
	for _, lamp := range []float64{lamps.R, lamps.G, lamps.B} {
		if !isFinite(lamp) || lamp < srgb.Black.R || lamp > srgb.White.R {
			t.Fatalf("XYZ %v %v %v gave lamp %v", x, y, z, lamp)
		}
	}
	if inGamut && !(isFinite(x) && isFinite(y) && isFinite(z)) {
		t.Fatalf("XYZ %v %v %v is in gamut", x, y, z)
	}
}

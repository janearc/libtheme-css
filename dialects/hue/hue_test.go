package hue

import (
	"math"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/primitives/swatch"
)

// near is whether two values agree within a tolerance.
func near(a, b swatch.XY, tol float64) bool {
	return math.Abs(a.X-b.X) <= tol && math.Abs(a.Y-b.Y) <= tol
}

// what a lamp says comes back as what it said: white's place through a
// swatch and back is white's place.
func TestPlaceRoundTrip(t *testing.T) {
	w := Of(swatch.White)
	if !near(Of(Swatch(w)), w, 1e-12) {
		t.Errorf("white did not round-trip: %v", Of(Swatch(w)))
	}
}

// a ramp made from a lamp's five places sampled back to five is those
// places, to the precision of going through oklab and back.
func TestRampToPointsIsIdentityAtTheStops(t *testing.T) {
	in := []swatch.XY{{X: 0.64, Y: 0.33}, {X: 0.5, Y: 0.4}, {X: 0.3127,
		Y: 0.329}, {X: 0.2, Y: 0.5}, {X: 0.15, Y: 0.06}}
	out := Points(Ramp(in...), 5)
	for i := range in {
		if !near(out[i], in[i], 1e-6) {
			t.Errorf("point %d: in %v out %v", i, in[i], out[i])
		}
	}
}

// 370 mirek is 2703 K, a warm bulb; it must sit on the locus between
// illuminant A (2856 K, x 0.4476) and warmer, so x above 0.44.
func TestMirekIsWarm(t *testing.T) {
	p := Of(Mirek(370))
	if p.X < 0.44 || p.X > 0.48 {
		t.Errorf("370 mirek gave x %.4f", p.X)
	}
}

// The numbers the mirek properties pick. Each is defined by this test.
const (
	// coolest is the lowest mirek of a hue lamp, which is its coolest white.
	// It is defined by hue's lamps: the doc of Mirek gives 153.
	coolest = 153
	// warmest is the highest mirek of a hue lamp, which is its warmest white.
	// It is defined by hue's lamps: the doc of Mirek gives 500.
	warmest = 500
	// perMegakelvin is what a mirek divides into to give kelvin. It is
	// defined by the unit: reciprocal megakelvin.
	perMegakelvin = 1e6
	// oneMirek is the step to the next mirek. It is defined by the unit.
	oneMirek = 1
	// atZero is a mirek that is none. It is defined by counting.
	atZero = 0
)

// The property: a mirek in a lamp's range is the black body at a million
// over it, and the whites get warmer, redder in x, as the mirek goes up.
func TestAMirekIsABlackBodyAndWarmsAsItRises(t *testing.T) {
	rapid.Check(t, isABlackBody)
}

// isABlackBody is one draw of the property above.
func isABlackBody(t *rapid.T) {
	mirek := rapid.IntRange(coolest, warmest-oneMirek).Draw(t, "mirek")
	got := Mirek(mirek)
	if want := swatch.Planckian(perMegakelvin / float64(mirek)); got != want {
		t.Fatalf("mirek %d: %v, want %v", mirek, got, want)
	}
	warmer := Mirek(mirek + oneMirek)
	if Of(warmer).X < Of(got).X {
		t.Fatalf("mirek %d is not warmer than %d", mirek+oneMirek, mirek)
	}
}

// The anti-test: a mirek of nothing or less is not a temperature, and is no
// light, at any size.
func TestAMirekOfNothingOrLessIsNoLight(t *testing.T) {
	rapid.Check(t, nothingBelowZero)
}

// nothingBelowZero is one draw of the property above.
func nothingBelowZero(t *rapid.T) {
	mirek := rapid.IntRange(math.MinInt32, atZero).Draw(t, "mirek")
	if got := Mirek(mirek); got != swatch.NoLight() {
		t.Fatalf("mirek %d gave light: %v", mirek, got)
	}
}

// gamuts a and c hold white; gamut b, the old bulbs, misses d65 by six
// ten-thousandths past its green-blue edge, which is a known fact about
// those lamps, so Fit moves white less than a thousandth. spectral red
// fits to gamut c's red corner, which lies short of it.
func TestPublishedGamuts(t *testing.T) {
	w := swatch.White.XY()
	for name, g := range map[string]swatch.Gamut{"A": GamutA, "C": GamutC} {
		if !g.Contains(w) {
			t.Errorf("white is outside gamut %s", name)
		}
	}
	if f := GamutB.Fit(w); math.Hypot(f.X-w.X, f.Y-w.Y) > 0.001 {
		t.Errorf("gamut b should miss white by a hair, not %v", f)
	}
	red := swatch.XY{X: 0.7347, Y: 0.2653}
	if f := GamutC.Fit(red); !near(f, GamutC.Red, 1e-9) {
		t.Errorf("spectral red should fit to gamut c's red corner, "+
			"got %v", f)
	}
}

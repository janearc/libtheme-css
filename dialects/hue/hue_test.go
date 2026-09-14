package hue

import (
	"math"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
)

// gamut C, as philips publishes it for the lamps that have it. typed
// here because a test may type what the library derives.
var gamutC = Gamut{Red: Point{0.6915, 0.3083}, Green: Point{0.17, 0.7}, Blue: Point{0.1532, 0.0475}}

func near(a, b Point, tol float64) bool {
	return math.Abs(a.X-b.X) <= tol && math.Abs(a.Y-b.Y) <= tol
}

// what a lamp says comes back as what it said: white's xy through a
// swatch and back is white's xy.
func TestPointRoundTrip(t *testing.T) {
	w := Of(swatch.White)
	if !near(Of(Swatch(w)), w, 1e-12) {
		t.Errorf("white did not round-trip: %v", Of(Swatch(w)))
	}
}

// a ramp made from a lamp's five points sampled back to five is those
// points, to the precision of going through oklab and back.
func TestRampToPointsIsIdentityAtTheStops(t *testing.T) {
	in := []Point{{0.64, 0.33}, {0.5, 0.4}, {0.3127, 0.329}, {0.2, 0.5}, {0.15, 0.06}}
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

// the white point is inside gamut C; spectral red is not, and lies past
// the red corner, so the nearest reachable point is that corner.
func TestGamutFit(t *testing.T) {
	if !gamutC.Contains(Of(swatch.White)) {
		t.Error("white is outside gamut C")
	}
	red := Point{0.7347, 0.2653} // the 700 nm end of the horseshoe
	if gamutC.Contains(red) {
		t.Error("spectral red is inside gamut C")
	}
	f := gamutC.Fit(red)
	if !gamutC.Contains(f) {
		t.Errorf("fit landed outside: %v", f)
	}
	if !near(f, gamutC.Red, 1e-9) {
		t.Errorf("spectral red should fit to the red corner, got %v", f)
	}
	if gamutC.Fit(Of(swatch.White)) != Of(swatch.White) {
		t.Error("fit moved a point that was inside")
	}
}

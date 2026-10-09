package srgb

import (
	"math"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
)

// The white is all three lamps at full, in light as in levels.
func TestWhiteIsOneOneOne(t *testing.T) {
	r, g, b := Light(swatch.White)
	for _, v := range []float64{r, g, b} {
		if math.Abs(v-1) > 1e-9 {
			t.Errorf("white is %v %v %v in light", r, g, b)
		}
	}
}

// Light and FromLight undo each other, bright light and dark alike, and
// light brighter than white stays brighter than white.
func TestLightRoundTrip(t *testing.T) {
	for _, c := range [][3]float64{{0.7, 0.3, 0.2}, {1, 0, 0},
		{0, 1, 0}, {3.5, 2, 0.1}, {0, 0, 0}} {
		r, g, b := Light(FromLight(c[0], c[1], c[2]))
		if math.Abs(r-c[0]) > 1e-12 || math.Abs(g-c[1]) > 1e-12 ||
			math.Abs(b-c[2]) > 1e-12 {
			t.Errorf("%v came back as %v %v %v", c, r, g, b)
		}
	}
}

// Red and green light together are yellow: the lamps add in light, and
// the sum is the swatch the lamps say yellow is.
func TestRedAndGreenLightAreYellow(t *testing.T) {
	x1, y1, z1 := FromLight(1, 0, 0).XYZ()
	x2, y2, z2 := FromLight(0, 1, 0).XYZ()
	sum := swatch.FromXYZ(x1+x2, y1+y2, z1+z2)
	got, _ := FromSwatch(sum)
	if got.Hex() != "#ffff00" {
		t.Errorf("red light and green light make %s", got.Hex())
	}
}

// The curve and its inverse undo each other across the whole range.
func TestTheCurveRoundTrips(t *testing.T) {
	for v := 0.0; v <= 1; v += 1.0 / 64 {
		if d := FromLinear(ToLinear(v)) - v; math.Abs(d) > 1e-12 {
			t.Errorf("%v came back %v off", v, d)
		}
	}
}

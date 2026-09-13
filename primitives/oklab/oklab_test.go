package oklab

import (
	"math"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
)

func near(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

// The reference values from Ottosson's description of the space, XYZ in
// and L, a, b out, to the three places he printed them.
func TestReferenceValues(t *testing.T) {
	cases := []struct {
		x, y, z float64
		l, a, b float64
	}{
		{0.950, 1.000, 1.089, 1.000, 0.000, 0.000},
		{1.000, 0.000, 0.000, 0.450, 1.236, -0.019},
		{0.000, 1.000, 0.000, 0.922, -0.671, 0.263},
		{0.000, 0.000, 1.000, 0.153, -1.415, -0.449},
	}
	for _, c := range cases {
		got := FromSwatch(swatch.FromXYZ(c.x, c.y, c.z))
		if !near(got.L, c.l, 1e-3) || !near(got.A, c.a, 1e-3) || !near(got.B, c.b, 1e-3) {
			t.Errorf("xyz %v %v %v gave %+v, want L %v a %v b %v", c.x, c.y, c.z, got, c.l, c.a, c.b)
		}
	}
}

// The library's own white is L 1 with no lean either way, and black is 0.
func TestWhiteAndBlack(t *testing.T) {
	w := FromSwatch(swatch.White)
	if !near(w.L, 1, 1e-3) || !near(w.A, 0, 1e-3) || !near(w.B, 0, 1e-3) {
		t.Errorf("white = %+v", w)
	}
	if k := FromSwatch(swatch.Black); k.L != 0 || k.A != 0 || k.B != 0 {
		t.Errorf("black = %+v", k)
	}
}

// Going to OKLab and back lands on the swatch you started from, to the
// float: the inverses are computed, not typed, and the cube and cube root
// undo each other.
func TestRoundTrip(t *testing.T) {
	for _, in := range []swatch.Swatch{swatch.White, swatch.FromXYZ(0.2, 0.1, 0.05), swatch.FromXYZ(0.4, 0.6, 0.9)} {
		out := FromSwatch(in).Swatch()
		ix, iy, iz := in.XYZ()
		ox, oy, oz := out.XYZ()
		if !near(ix, ox, 1e-12) || !near(iy, oy, 1e-12) || !near(iz, oz, 1e-12) {
			t.Errorf("%v came back as %v", in, out)
		}
	}
}

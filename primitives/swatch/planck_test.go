package swatch

import (
	"math"
	"testing"
)

// CIE illuminant A is defined as a black body at 2856 K, and its
// chromaticity is published to four places: x 0.4476, y 0.4074. The
// derived locus must land on it; this is the check that Planck's law,
// the observer tables and Illuminant agree.
func TestPlanckianIsIlluminantA(t *testing.T) {
	x, y := Planckian(2856).XY()
	if math.Abs(x-0.4476) > 0.001 || math.Abs(y-0.4074) > 0.001 {
		t.Errorf("2856 K gave xy %.4f %.4f, illuminant A is 0.4476 0.4074", x, y)
	}
}

// Daylight is not a black body, so 6504 K does not land exactly on D65;
// it lands close, on the warm side of it, and the gap is a known fact
// about daylight rather than an error here.
func TestPlanckianNearD65(t *testing.T) {
	x, y := Planckian(6504).XY()
	wx, wy := White.XY()
	if math.Hypot(x-wx, y-wy) > 0.01 {
		t.Errorf("6504 K gave xy %.4f %.4f, d65 is %.4f %.4f", x, y, wx, wy)
	}
}

// A chromaticity put back at unit luminance is the swatch it came from.
func TestXYRoundTrip(t *testing.T) {
	x, y := White.XY()
	s := FromXY(x, y, 1)
	sx, sy, sz := s.XYZ()
	wx, wy, wz := White.XYZ()
	if math.Abs(sx-wx) > 1e-12 || math.Abs(sy-wy) > 1e-12 || math.Abs(sz-wz) > 1e-12 {
		t.Errorf("white through xy came back %v %v %v", sx, sy, sz)
	}
	if FromXY(0.3, 0, 1) != Black {
		t.Error("y of zero is not black")
	}
}

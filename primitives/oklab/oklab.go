// Package oklab is the swatch in the coordinates an eye agrees with. The
// same colour as the swatch, not a different one: a second set of axes
// on the same point, chosen so that distance means something.
package oklab

import (
	"math"

	"github.com/janearc/libtheme-css/primitives/swatch"
)

// OKLab is a colour as L, a, b.
type OKLab struct {
	// lightness: 0 is black, 1 is the white. steps look even, so 0.5 looks
	//   halfway
	L float64

	// green to red: negative leans green, positive leans red, 0 is neither
	A float64

	// blue to yellow: negative leans blue, positive leans yellow, 0 is
	//   neither. A and B together say which way round the wheel and how
	//   far out; 0, 0 is a grey at whatever L
	B float64
}

// The conversion is two steps, and the two steps are the whole idea.
//
// First, XYZ is turned into how strongly each of the eye's three cone
// types would respond, called LMS for long, medium and short wavelength.
// XYZ was built in 1931 to be non-negative, not to be cones; this matrix
// undoes that choice. Then each response has its cube root taken: the
// eye reports ratios, not amounts, so doubling the light does not look
// like twice as much, and a cube root is the compression that matches
// what people report. Last, the three compressed responses are combined
// into one lightness and two opponent axes, because an eye cannot see a
// reddish green or a bluish yellow, and those pairs are what it actually
// compares.
//
// The twelve numbers in each matrix are fitted, not derived: Björn
// Ottosson chose them in 2020 by searching for the values under which
// hue stays put when lightness changes and equal steps look equal,
// against published measurements of what people see. They are the ones
// from his description of the space, https://bottosson.github.io/posts/oklab/,
// and CSS Color Level 4 adopted them unchanged. There is no derivation
// to show; there is a fit to check, and the tests below check it against
// the reference values in the same post.

// mat3 is a 3x3 matrix, rows first.
type mat3 [3][3]float64

// apply is the matrix times a column of three.
func (m mat3) apply(a, b, c float64) (float64, float64, float64) {
	return m[0][0]*a + m[0][1]*b + m[0][2]*c,
		m[1][0]*a + m[1][1]*b + m[1][2]*c,
		m[2][0]*a + m[2][1]*b + m[2][2]*c
}

// inverse is the matrix that undoes this one, by the adjugate over the
// determinant, which for a 3x3 is short enough to write out. The
// published inverses are rounded to ten places and do not quite undo the
// forward matrices; computing the inverse here makes a round trip exact
// to the float, so the only numbers typed in are the forward fit.
func (m mat3) inverse() mat3 {
	det := m[0][0]*(m[1][1]*m[2][2]-m[1][2]*m[2][1]) -
		m[0][1]*(m[1][0]*m[2][2]-m[1][2]*m[2][0]) +
		m[0][2]*(m[1][0]*m[2][1]-m[1][1]*m[2][0])
	return mat3{
		{(m[1][1]*m[2][2] - m[1][2]*m[2][1]) / det, (m[0][2]*m[2][1] - m[0][1]*m[2][2]) / det, (m[0][1]*m[1][2] - m[0][2]*m[1][1]) / det},
		{(m[1][2]*m[2][0] - m[1][0]*m[2][2]) / det, (m[0][0]*m[2][2] - m[0][2]*m[2][0]) / det, (m[0][2]*m[1][0] - m[0][0]*m[1][2]) / det},
		{(m[1][0]*m[2][1] - m[1][1]*m[2][0]) / det, (m[0][1]*m[2][0] - m[0][0]*m[2][1]) / det, (m[0][0]*m[1][1] - m[0][1]*m[1][0]) / det},
	}
}

// xyzToLMS is the first step: XYZ to the three cone responses.
var xyzToLMS = mat3{
	{0.8189330101, 0.3618667424, -0.1288597137},
	{0.0329845436, 0.9293118715, 0.0361456387},
	{0.0482003018, 0.2643662691, 0.6338517070},
}

// lmsToLab is the last step: compressed cone responses to L, a, b.
var lmsToLab = mat3{
	{0.2104542553, 0.7936177850, -0.0040720468},
	{1.9779984951, -2.4285922050, 0.4505937099},
	{0.0259040371, 0.7827717662, -0.8086757660},
}

var lmsToXYZ, labToLMS = xyzToLMS.inverse(), lmsToLab.inverse()

// FromSwatch is the swatch in OKLab.
func FromSwatch(s swatch.Swatch) OKLab {
	l, m, n := xyzToLMS.apply(s.XYZ())
	l, m, n = math.Cbrt(l), math.Cbrt(m), math.Cbrt(n)
	L, a, b := lmsToLab.apply(l, m, n)
	return OKLab{L, a, b}
}

// Swatch is the OKLab colour stored back as a swatch: the two steps run
// in reverse, cubes for cube roots and the inverse of each matrix.
func (c OKLab) Swatch() swatch.Swatch {
	l, m, n := labToLMS.apply(c.L, c.A, c.B)
	l, m, n = l*l*l, m*m*m, n*n*n
	return swatch.FromXYZ(lmsToXYZ.apply(l, m, n))
}

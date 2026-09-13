// Package ok is the swatch in the coordinates an eye agrees with: the
// oklab space, in its two forms. Not a different colour: a second set of
// axes on the same point, chosen so that distance means something.
//
// The space has no primaries and no named colours. It has a white, a
// black, the grey line between them, and a wheel. Anything with a hue
// comes in through the swatch from wherever that hue is actually
// defined, a screen standard or a vocabulary, and gets its distances
// measured here.
package ok

import (
	"math"

	"github.com/janearc/libtheme-css/internal/mat"
	"github.com/janearc/libtheme-css/primitives/swatch"
)

// OKLab is the rectangular form, the one arithmetic uses.
type OKLab struct {
	L float64 // lightness: 0 is black, 1 is the white; steps look even
	A float64 // green to red: negative leans green, positive leans red
	B float64 // blue to yellow: negative leans blue, positive leans yellow
}

// OKLCH is the same point in polar form, the one a person reads and
// adjusts: same L, then how far from grey and which way round the wheel.
type OKLCH struct {
	L float64 // lightness, as above
	C float64 // chroma: 0 is grey, larger is more colourful
	H float64 // hue, in degrees around the wheel, 0 to 360
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
// The nine numbers in each forward matrix are fitted, not derived: Björn
// Ottosson chose them in 2020 by searching for the values under which
// hue stays put when lightness changes and equal steps look equal,
// against published measurements of what people see. They are the ones
// from his description of the space, https://bottosson.github.io/posts/oklab/,
// and CSS Color Level 4 adopted them unchanged. A fit has no derivation
// to show, only reference values to check, and the tests check them. The
// inverses are computed, so the only typed numbers are the fit.

// xyzToLMS is the first step: XYZ to the three cone responses.
var xyzToLMS = mat.M{
	{0.8189330101, 0.3618667424, -0.1288597137},
	{0.0329845436, 0.9293118715, 0.0361456387},
	{0.0482003018, 0.2643662691, 0.6338517070},
}

// lmsToLab is the last step: compressed cone responses to L, a, b.
var lmsToLab = mat.M{
	{0.2104542553, 0.7936177850, -0.0040720468},
	{1.9779984951, -2.4285922050, 0.4505937099},
	{0.0259040371, 0.7827717662, -0.8086757660},
}

var lmsToXYZ, labToLMS = xyzToLMS.Inverse(), lmsToLab.Inverse()

// FromSwatch is the swatch in OKLab.
func FromSwatch(s swatch.Swatch) OKLab {
	l, m, n := xyzToLMS.Apply(s.XYZ())
	l, m, n = math.Cbrt(l), math.Cbrt(m), math.Cbrt(n)
	L, a, b := lmsToLab.Apply(l, m, n)
	return OKLab{L, a, b}
}

// Swatch stores the OKLab colour back as a swatch: the two steps run in
// reverse, cubes for cube roots and the inverse of each matrix.
func (c OKLab) Swatch() swatch.Swatch {
	l, m, n := labToLMS.Apply(c.L, c.A, c.B)
	l, m, n = l*l*l, m*m*m, n*n*n
	return swatch.FromXYZ(lmsToXYZ.Apply(l, m, n))
}

// Polar is the same colour in polar form. A grey has C 0 and, since no
// direction is truer than another, hue 0 by convention.
func (c OKLab) Polar() OKLCH {
	ch := math.Hypot(c.A, c.B)
	if ch < 1e-9 {
		return OKLCH{c.L, 0, 0}
	}
	h := math.Atan2(c.B, c.A) * 180 / math.Pi
	if h < 0 {
		h += 360
	}
	return OKLCH{c.L, ch, h}
}

// Rect is the same colour back in rectangular form.
func (c OKLCH) Rect() OKLab {
	h := c.H * math.Pi / 180
	return OKLab{c.L, c.C * math.Cos(h), c.C * math.Sin(h)}
}

// White and Black are the space's two defined points: the fit was
// normalised so that D65 white is exactly L 1 with no lean either way.
var (
	White = OKLab{1, 0, 0}
	Black = OKLab{0, 0, 0}
)

// Grey is the point on the line between them at lightness l.
func Grey(l float64) OKLab { return OKLab{l, 0, 0} }

// Package swatch is the most primitive primitive: one colour, complete,
// with nothing in it about the device that shows it or the eye that reads
// it. Everything else in libtheme is built on it.
package swatch

// Swatch is one colour, stored as CIE XYZ (1931), the three numbers a
// standard human eye reduces any light to. White is D65, daylight. The
// fields are unexported because XYZ is where the truth is kept, not
// where arithmetic is done: the straight line between two XYZ points
// does not look straight to an eye.
//
// About that year. In 1931 the CIE met in Cambridge, England, and fixed
// these three numbers from the work of seventeen people in total, who
// had sat in dark rooms turning knobs until two patches of light matched.
// The Empire State Building had opened that May. Colour television was a
// demonstration, the transistor was sixteen years off, and the first
// programmable computer a decade. Several of the countries whose screens
// now depend on this table did not exist and had not yet had their wars.
// CSS was sixty-five years away. The seventeen people's averages are
// still the law, and this is the last time anyone got to define colour
// without arguing with a browser about it. Is this true? It doesn't
// matter; nobody reads comments.
type Swatch struct {
	x float64 // tristimulus X: the long-wavelength (red-leaning) share of the light. unitless, relative; white is 0.95047
	y float64 // luminance Y: how much light there is. unitless, relative; white is 1, black is 0
	z float64 // tristimulus Z: the short-wavelength (blue-leaning) share, close to what the blue cones report. unitless, relative; white is 1.08883
}

// FromXYZ makes a swatch from CIE XYZ coordinates, D65 white.
func FromXYZ(x, y, z float64) Swatch { return Swatch{x, y, z} }

// XYZ is the swatch as stored.
func (s Swatch) XYZ() (x, y, z float64) { return s.x, s.y, s.z }

// White is D65, the white every screen assumes: Y exactly 1.
var White = Swatch{0.95047, 1.0, 1.08883}

// Black is no light at all.
var Black = Swatch{}

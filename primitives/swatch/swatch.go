// Package swatch is the most primitive primitive: one colour, complete,
// with nothing in it about the device that shows it or the eye that reads
// it. Everything else in libtheme is built on it and translates out of it.
package swatch

// Swatch is one colour, stored as CIE XYZ: the three numbers the 1931
// standard observer reduces any spectrum to. Y is luminance, how much
// light, with white at 1 and black at 0. X and Z carry the rest of what
// the eye's three cone types report; on their own they are not "red" or
// "blue", they are the other two axes of a space chosen in 1931 so that
// every number is non-negative. Every device space a screen or a lamp
// uses is defined by its relationship to these three, which is why they
// are the root and not sRGB.
//
// The white is D65, daylight, the one sRGB and every screen assume.
//
// The fields are unexported on purpose. XYZ is where the truth is kept,
// not where the arithmetic is done: the straight line between two XYZ
// points does not look straight to an eye, so averaging two of these
// gives mud. Operations that need a straight line live in a space built
// for that, and come back here to store the answer.
type Swatch struct {
	x, y, z float64
}

// FromXYZ makes a swatch from CIE XYZ coordinates, D65 white.
func FromXYZ(x, y, z float64) Swatch { return Swatch{x, y, z} }

// XYZ is the swatch as stored.
func (s Swatch) XYZ() (x, y, z float64) { return s.x, s.y, s.z }

// White is D65, the white every screen assumes: Y exactly 1.
var White = Swatch{0.95047, 1.0, 1.08883}

// Black is no light at all.
var Black = Swatch{}

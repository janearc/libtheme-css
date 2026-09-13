// Package srgb is the swatch as a screen is driven: three lamp levels,
// red, green and blue, on the curve every screen since 1996 assumes. It
// is the door to hex, and to the two cylinders, hsl and hsv, which are
// the same lamps described by angle. A swatch may fall outside what the
// lamps can make; every exit here says whether it did.
package srgb

import "github.com/janearc/libtheme-css/primitives/swatch"

// RGB is the three lamp levels as fractions, 0 off to 1 full.
type RGB struct {
	R, G, B float64
}

// HSL is the lamps described by angle: hue around the wheel, saturation,
// and lightness as the midpoint of the brightest and dimmest lamp.
type HSL struct {
	H, S, L float64
}

// HSV is the other cylinder: hue, saturation, and value, the brightest
// lamp on its own, so full red and white both have V 1.
type HSV struct {
	H, S, V float64
}

// FromSwatch is the swatch as lamp levels, and whether the lamps can
// make it. Not yet.
func FromSwatch(s swatch.Swatch) (c RGB, inGamut bool) { return RGB{}, false }

// Swatch stores the lamp levels back as a swatch. Not yet.
func (c RGB) Swatch() swatch.Swatch { return swatch.Black }

// Hex is the lamps as "#rrggbb". Not yet.
func (c RGB) Hex() string { return "" }

// FromHex reads "#rrggbb". Not yet.
func FromHex(h string) (RGB, error) { return RGB{}, nil }

// Package ok is the swatch in the coordinates an eye agrees with: the
// oklab space, in its two forms. Not a different colour: a second set of
// axes on the same point, chosen so that distance means something.
package ok

import "github.com/janearc/libtheme-css/primitives/swatch"

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

// FromSwatch is the swatch in OKLab. Not yet.
func FromSwatch(s swatch.Swatch) OKLab { return OKLab{} }

// Swatch stores the OKLab colour back as a swatch. Not yet.
func (c OKLab) Swatch() swatch.Swatch { return swatch.Black }

// Polar is the same colour in polar form. Not yet.
func (c OKLab) Polar() OKLCH { return OKLCH{} }

// Rect is the same colour back in rectangular form. Not yet.
func (c OKLCH) Rect() OKLab { return OKLab{} }

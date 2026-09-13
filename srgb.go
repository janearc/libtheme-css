package libtheme

import (
	"fmt"
	"math"
	"strings"
)

// sRGB is what a screen is driven with: three lamp levels, 0 to 255, on a
// curve that spends more numbers on the dark end. It is a device space, so
// a swatch may fall outside it; every exit here says whether it did.

// srgbToLinear undoes the screen's curve: a byte becomes light.
func srgbToLinear(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// linearToSRGB puts the curve back: light becomes a byte's fraction.
func linearToSRGB(c float64) float64 {
	if c <= 0.0031308 {
		return 12.92 * c
	}
	return 1.055*math.Pow(c, 1/2.4) - 0.055
}

// FromSRGB makes a swatch from sRGB fractions, 0 to 1 each.
func FromSRGB(r, g, b float64) Swatch {
	r, g, b = srgbToLinear(r), srgbToLinear(g), srgbToLinear(b)
	return Swatch{
		0.4123907993*r + 0.3575843394*g + 0.1804807884*b,
		0.2126390059*r + 0.7151686788*g + 0.0721923154*b,
		0.0193308187*r + 0.1191947798*g + 0.9505321522*b,
	}
}

// FromHex makes a swatch from "#rrggbb" or "rrggbb", read as sRGB, which
// is what every hex code you have ever typed silently was.
func FromHex(h string) (Swatch, error) {
	h = strings.TrimPrefix(strings.TrimSpace(h), "#")
	if len(h) != 6 {
		return Swatch{}, fmt.Errorf("hex colour wants six digits, not %q", h)
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(h, "%02x%02x%02x", &r, &g, &b); err != nil {
		return Swatch{}, fmt.Errorf("hex colour %q: %v", h, err)
	}
	return FromSRGB(float64(r)/255, float64(g)/255, float64(b)/255), nil
}

// MustHex is FromHex for a literal you know is well formed.
func MustHex(h string) Swatch {
	s, err := FromHex(h)
	if err != nil {
		panic(err)
	}
	return s
}

// SRGB is the swatch as sRGB fractions, and whether the screen can make it.
// Out of gamut, the fractions are clipped to 0..1 and ok is false: the
// nearest thing the lamps can do, and an honest word that it is not the
// same colour.
func (s Swatch) SRGB() (r, g, b float64, ok bool) {
	r = 3.2409699419*s.x - 1.5373831776*s.y - 0.4986107603*s.z
	g = -0.9692436363*s.x + 1.8759675015*s.y + 0.0415550574*s.z
	b = 0.0556300797*s.x - 0.2039769589*s.y + 1.0569715142*s.z
	ok = true
	clip := func(v float64) float64 {
		if v < -1e-6 || v > 1+1e-6 {
			ok = false
		}
		return math.Max(0, math.Min(1, v))
	}
	return linearToSRGB(clip(r)), linearToSRGB(clip(g)), linearToSRGB(clip(b)), ok
}

// SRGB8 is the swatch as the three bytes a terminal or a CSS hex wants.
func (s Swatch) SRGB8() (r, g, b uint8, ok bool) {
	fr, fg, fb, ok := s.SRGB()
	round := func(v float64) uint8 { return uint8(math.Round(v * 255)) }
	return round(fr), round(fg), round(fb), ok
}

// Hex is the swatch as "#rrggbb", clipped if the screen cannot make it.
func (s Swatch) Hex() string {
	r, g, b, _ := s.SRGB8()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

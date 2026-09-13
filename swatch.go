// Package libtheme normalises colour between the places one person keeps a
// theme: a terminal, an editor, a web page, a lamp. It thinks in one
// primitive, the Swatch, and translates out to whatever each device speaks.
package libtheme

import (
	"fmt"
	"math"
)

// Swatch is one colour, complete: everything about it that does not depend
// on who is showing it or who is looking. It is stored as CIE XYZ (1931,
// D65 white), the root every device space is defined against, so the
// working space can change without touching a stored value. The fields are
// unexported on purpose: averaging two XYZ triples gives mud, and the
// methods know where straight lines look straight (OKLab).
type Swatch struct {
	x, y, z float64
}

// OKLab is a swatch in the coordinates an eye agrees with: L lightness (0
// black, 1 white), A green-to-red, B blue-to-yellow. Equal distances here
// look about equally different, wherever they sit.
type OKLab struct {
	L, A, B float64
}

// OKLCH is OKLab in polar form, the one you say aloud: L lightness, C
// chroma (0 is grey), H hue in degrees around the wheel.
type OKLCH struct {
	L, C, H float64
}

// FromXYZ makes a swatch from CIE XYZ, D65.
func FromXYZ(x, y, z float64) Swatch { return Swatch{x, y, z} }

// XYZ is the swatch as stored.
func (s Swatch) XYZ() (x, y, z float64) { return s.x, s.y, s.z }

// FromOKLab makes a swatch from OKLab coordinates.
func FromOKLab(c OKLab) Swatch {
	l := c.L + 0.3963377774*c.A + 0.2158037573*c.B
	m := c.L - 0.1055613458*c.A - 0.0638541728*c.B
	n := c.L - 0.0894841775*c.A - 1.2914855480*c.B
	l, m, n = l*l*l, m*m*m, n*n*n
	return Swatch{
		1.2270138511*l - 0.5577999807*m + 0.2812561490*n,
		-0.0405801784*l + 1.1122568696*m - 0.0716766787*n,
		-0.0763812845*l - 0.4214819784*m + 1.5861632204*n,
	}
}

// OKLab is the swatch in OKLab.
func (s Swatch) OKLab() OKLab {
	l := 0.8189330101*s.x + 0.3618667424*s.y - 0.1288597137*s.z
	m := 0.0329845436*s.x + 0.9293118715*s.y + 0.0361456387*s.z
	n := 0.0482003018*s.x + 0.2643662691*s.y + 0.6338517070*s.z
	l, m, n = math.Cbrt(l), math.Cbrt(m), math.Cbrt(n)
	return OKLab{
		0.2104542553*l + 0.7936177850*m - 0.0040720468*n,
		1.9779984951*l - 2.4285922050*m + 0.4505937099*n,
		0.0259040371*l + 0.7827717662*m - 0.8086757660*n,
	}
}

// FromOKLCH makes a swatch from the polar form.
func FromOKLCH(c OKLCH) Swatch {
	h := c.H * math.Pi / 180
	return FromOKLab(OKLab{c.L, c.C * math.Cos(h), c.C * math.Sin(h)})
}

// OKLCH is the swatch in polar form. Hue is 0 to 360; a grey has C 0 and
// hue 0 by convention, since no direction is truer than another.
func (s Swatch) OKLCH() OKLCH {
	c := s.OKLab()
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

// Mix is the swatch t of the way from a to b, 0 giving a and 1 giving b,
// along a straight line in OKLab, which is the line that looks straight.
func Mix(a, b Swatch, t float64) Swatch {
	p, q := a.OKLab(), b.OKLab()
	return FromOKLab(OKLab{
		p.L + (q.L-p.L)*t,
		p.A + (q.A-p.A)*t,
		p.B + (q.B-p.B)*t,
	})
}

// Distance is how different two swatches look: Euclidean distance in
// OKLab. Around 0.02 is the least difference most eyes notice.
func Distance(a, b Swatch) float64 {
	p, q := a.OKLab(), b.OKLab()
	return math.Sqrt((p.L-q.L)*(p.L-q.L) + (p.A-q.A)*(p.A-q.A) + (p.B-q.B)*(p.B-q.B))
}

// String is the swatch as CSS says it: oklch(74% 0.20 345).
func (s Swatch) String() string {
	c := s.OKLCH()
	return fmt.Sprintf("oklch(%.0f%% %.2f %.0f)", c.L*100, c.C, c.H)
}

package main

import (
	"fmt"
	"strings"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/libtheme-css/primitives/swatch"
)

// bandsPageProse is what the page says.
const bandsPageProse = `
light before it is a colour: power in slices of the spectrum, 24 to
an octave. the first bar is the eye's own, one cell per five
nanometres; the second is the same light carried in bands, so each
slice is one colour and you can count them. below, black bodies as
the observer sees them and the same lamps through bands: the pairs
should look the same. then daylight redshifted, one step a cell,
each step 2.9 percent redder. last, light adds: a red lamp and a
brighter green one on the same place are yellow.`

// at is a light's colour, put at a brightness a screen can show.
func at(s swatch.Swatch, y float64) swatch.Swatch {
	x, ly, z := s.XYZ()
	if ly == 0 {
		return s
	}
	return swatch.FromXYZ(x/ly*y, y, z/ly*y)
}

// bandOf is the light of one whole band at unit power.
func bandOf(nm int) bands.Light {
	return bands.Split(map[int]float64{nm: 1})
}

// bandsPage paints the page: the spectrum, sliced.
func bandsPage() {
	title("bands: light before it is a colour")
	var eye, sliced strings.Builder
	for nm := 380; nm <= 700; nm += 5 {
		eye.WriteString(paint(lifted(swatch.Monochrome(nm)), 1))
		b := bands.Of(float64(nm))
		lo, hi := b.Edges()
		var mid bands.Light
		for w := int(lo) + 1; float64(w) < hi; w++ {
			mid = bands.Add(mid, bandOf(w))
		}
		sliced.WriteString(paint(lifted(mid.Swatch()), 1))
	}
	fmt.Printf("   %s  the eye\n   %s  in bands\n\n", eye.String(),
		sliced.String())

	var exact, banded strings.Builder
	for k := 1000.0; k <= 12000; k += 1000 {
		exact.WriteString(paint(at(swatch.Planckian(k), 0.8), 3))
		through := bands.Split(swatch.Blackbody(k)).Swatch()
		banded.WriteString(paint(at(through, 0.8), 3))
	}
	fmt.Printf("   %s  1000 K to 12000 K\n   %s  the same, in bands\n\n",
		exact.String(), banded.String())

	var shifted strings.Builder
	day := bands.Split(swatch.D65())
	for s := 0; s <= 16; s++ {
		shifted.WriteString(paint(at(day.Shift(s).Swatch(), 0.8), 3))
	}
	fmt.Printf("   %s  daylight, redshifted\n\n", shifted.String())

	// yellow wants more green than red, as it does from any screen's lamps
	red, green := lamp(620, 660, 0.2), lamp(520, 560, 0.6)
	both := bands.Add(red, green)
	fmt.Printf("   %s + %s = %s  light adds\n\n",
		paint(red.Swatch(), 6), paint(green.Swatch(), 6),
		paint(both.Swatch(), 6))
	say(lines(bandsPageProse)...)
}

// lamp is flat light from lo to hi nanometres, at luminance y, where white
// is 1.
func lamp(lo, hi int, y float64) bands.Light {
	spectrum := map[int]float64{}
	for nm := lo; nm <= hi; nm++ {
		spectrum[nm] = 1
	}
	l := bands.Split(spectrum)
	_, has, _ := l.Swatch().XYZ()
	return l.Scale(y / has)
}

// bandsCSS is the page as a css sheet. It writes the visible bands as hard
// stops, one colour a band, the way the second bar paints them.
func bandsCSS() string {
	var stops []string
	first, last := bands.Of(380), bands.Of(700)
	for b := first; b <= last; b++ {
		lo, hi := b.Edges()
		var l bands.Light
		for w := int(lo) + 1; float64(w) < hi; w++ {
			l = bands.Add(l, bandOf(w))
		}
		from := (max(lo, 380) - 380) / 320 * 100
		to := (min(hi, 700) - 380) / 320 * 100
		stops = append(stops, fmt.Sprintf("    %s %.1f%% %.1f%%",
			hexOf(lifted(l.Swatch())), from, to))
	}
	return ":root {\n  --bands: linear-gradient(in srgb,\n" +
		strings.Join(stops, ",\n") + ");\n}\n"
}

// Package bands is light carried as amounts in slices of the spectrum: the
// form a renderer needs when light has to be added, filtered, redshifted or
// bent by wavelength, before an eye reduces it to three numbers.
//
// A swatch is what light looks like. Light in bands is what it is, coarsely.
// Swatch folds it into a colour through the same observer as everything else
// in libtheme, so the two never disagree about what a light looks like.
package bands

import (
	"math"
	"sort"

	"github.com/janearc/libtheme-css/primitives/swatch"
)

// PerOctave is how many bands there are for every doubling of wavelength,
// and the one number this package types in. It is chosen, not derived; the
// measurements it was chosen from are in FUDGE.md under "bands".
//
// The bands are even in the logarithm of wavelength, not in nanometres.
// Redshift multiplies every wavelength by the same factor, so it is a slide
// along the bands, not a resampling.
const PerOctave = 24

// Band is one slice of the spectrum. Band i runs from 2^(i/PerOctave)
// nanometres up to, but not including, 2^((i+1)/PerOctave). Band 0 starts
// at one nanometre, and the wavelength doubles every PerOctave bands.
// Visible light is about bands 205 to 230.
type Band int

// Of is the band a wavelength in nanometres falls in.
func Of(nm float64) Band { return grid(PerOctave).of(nm) }

// Edges is where a band starts and where the next one does, in nanometres.
func (b Band) Edges() (lo, hi float64) { return grid(PerOctave).edges(b) }

// grid is a step, in bands per octave. The package uses PerOctave; the
// tests use others, to show why it is that one.
type grid int

// of is Of at a step.
func (g grid) of(nm float64) Band {
	return Band(math.Floor(float64(g) * math.Log2(nm)))
}

// edges is Edges at a step.
func (g grid) edges(b Band) (lo, hi float64) {
	return math.Exp2(float64(b) / float64(g)),
		math.Exp2(float64(b+1) / float64(g))
}

// Light is the power in each of a run of neighbouring bands, starting at
// band First. Its units are whatever the spectrum it came from used, times
// a nanometre. The zero value is no light.
type Light struct {
	g     grid
	first Band
	power []float64
}

// Split is a spectrum, given as power by wavelength in nanometres, as light
// in bands: each wavelength's power goes into the band it falls in.
//
// Every wavelength counts, including those an eye never sees. Light in
// infrared bands is still light, and it adds nothing to Swatch.
func Split(spectrum map[int]float64) Light {
	return grid(PerOctave).split(spectrum)
}

// split is Split at a step.
func (g grid) split(spectrum map[int]float64) Light {
	lines := make(map[float64]float64, len(spectrum))
	for nm, p := range spectrum {
		lines[float64(nm)] = p
	}
	return g.lines(lines)
}

// Span is the first and last band this light has any entry for. Empty light
// reports first greater than last.
func (l Light) Span() (first, last Band) {
	return l.first, l.first + Band(len(l.power)) - 1
}

// Power is the light in one band, zero outside the span.
func (l Light) Power(b Band) float64 {
	i := int(b - l.first)
	if i < 0 || i >= len(l.power) {
		return 0
	}
	return l.power[i]
}

// Shift slides the light along the bands: up by steps is redder. Redshift
// by a factor of 1 + z is PerOctave times log2(1 + z) steps, exact when that
// is whole. The power in each band travels with it unchanged; dimming with
// distance is the scene's business, not the spectrum's.
func (l Light) Shift(steps int) Light {
	out := Light{g: l.g, first: l.first + Band(steps)}
	out.power = append([]float64(nil), l.power...)
	return out
}

// Scale is the light made k times as strong in every band.
func (l Light) Scale(k float64) Light {
	out := Light{g: l.g, first: l.first}
	out.power = make([]float64, len(l.power))
	for i, p := range l.power {
		out.power[i] = p * k
	}
	return out
}

// Add is two lights shining on the same place: in every band, the sum.
func Add(a, b Light) Light {
	if len(a.power) == 0 {
		return b.Scale(1)
	}
	if len(b.power) == 0 {
		return a.Scale(1)
	}
	af, al := a.Span()
	bf, bl := b.Span()
	first, last := min(af, bf), max(al, bl)
	out := Light{g: a.g, first: first}
	out.power = make([]float64, last-first+1)
	for band := first; band <= last; band++ {
		out.power[band-first] = a.Power(band) + b.Power(band)
	}
	return out
}

// Swatch is the colour this light is, to the 1931 observer.
//
// Each band's power is weighed by the observer's average response across the
// whole wavelengths inside that band, and the three weighted sums are scaled
// so that D65, at the power its table gives, has Y of 1.
//
// So a light Split from D65 comes back as White, as nearly as bands can
// carry it. Twice that light has Y of 2.
func (l Light) Swatch() swatch.Swatch {
	var x, y, z float64
	for i, p := range l.power {
		if p == 0 {
			continue
		}
		wx, wy, wz := l.g.weight(l.first + Band(i))
		x += p * wx
		y += p * wy
		z += p * wz
	}
	return swatch.FromXYZ(x*unitY, y*unitY, z*unitY)
}

// weight is the observer's mean response across the whole wavelengths in a
// band, the same wavelengths Split puts there. A band with none, or with
// none the observer lists, weighs nothing.
func (g grid) weight(b Band) (x, y, z float64) {
	lo, hi := g.edges(b)
	n := 0
	for nm := int(math.Ceil(lo)); float64(nm) < hi+1; nm++ {
		if g.of(float64(nm)) != b {
			continue
		}
		mx, my, mz := swatch.Monochrome(nm).XYZ()
		x, y, z = x+mx, y+my, z+mz
		n++
	}
	if n == 0 {
		return 0, 0, 0
	}
	return x / float64(n), y / float64(n), z / float64(n)
}

// unitY is one over D65's luminance, worked out once.
var unitY = unit()

// unit is one over D65's luminance, summed at every wavelength with no
// bands involved, so that the unit of light does not depend on the step.
func unit() float64 {
	d := swatch.D65()
	order := make([]int, 0, len(d))
	for nm := range d {
		order = append(order, nm)
	}
	sort.Ints(order)
	var y float64
	for _, nm := range order {
		_, my, _ := swatch.Monochrome(nm).XYZ()
		y += d[nm] * my
	}
	return 1 / y
}

// light given at any wavelength is grouped into bands by nanometer wavelength,
// and for smaller bands, we have lines. gamma rays are a few thousandths of a
// nanometer, so we need to make room for that.
func Lines(spectrum map[float64]float64) Light {
	return grid(PerOctave).lines(spectrum)
}

// lines is Lines at a step, which Split and the tests share.
func (g grid) lines(spectrum map[float64]float64) Light {
	order := make([]float64, 0, len(spectrum))
	for nm := range spectrum {
		if nm > 0 {
			order = append(order, nm)
		}
	}
	if len(order) == 0 {
		return Light{g: g}
	}
	sort.Float64s(order)
	first, last := g.of(order[0]), g.of(order[len(order)-1])
	l := Light{g: g, first: first, power: make([]float64, last-first+1)}

	// order by wavelength shortest to longest
	for _, nm := range order {
		l.power[g.of(nm)-first] += spectrum[nm]
	}

	// send it back.
	return l
}

// adds support for radiation not ordinarily visible to the human eye
//
// in order to render this in your terminal or a browser, you'd need a
// profile, typically called  a false color profile.
//
// ranges are bounded by iso 21348.
package radiation

import (
	"fmt"
	"math"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/libtheme-css/primitives/swatch"
)

// we define the range in nm of a band of radiation
type Range struct{ Lo, Hi float64 }

// create a named object from a range of radiation and what it is mapped  to
type Profile struct {
	Name   string
	Source Range // this is probably a range in nm
	Render Range // this is probably a range of visible light
}

// define what "visible" means here, which is sometimes tricky and contentious:
// it is the 1931 observer's, from the first row of its table to the last, as
// every range in libtheme is.
var Visible = observed()

// observed is the observer's rows as a range.
func observed() Range {
	lo, hi := swatch.Observed()
	return Range{float64(lo), float64(hi)}
}

// iso 21348:2007 named ranges
var (
	Gamma              = Profile{"gamma ray", Range{1e-5, 1e-3}, Visible}
	XRay               = Profile{"x-ray", Range{1e-3, 10}, Visible}
	ExtremeUltraviolet = Profile{
		"extreme ultraviolet",
		Range{10, 121},
		Visible,
	}
	Ultraviolet = Profile{"ultraviolet", Range{100, 400}, Visible}
	Infrared    = Profile{"infrared", Range{760, 1e6}, Visible}
	Microwave   = Profile{"microwave", Range{1e6, 1.5e7}, Visible}
	Radio       = Profile{"radio", Range{1e5, 1e11}, Visible}
)

// the fermi mission has a little bit diffrent range than iso, and when you
// render gamma ray emitting objects in css or the console, you sometimes need
// to use a different range, depending on where that observation was
// eollected.
var Fermi = Profile{
	"fermi",
	Range{FromEnergy(300 * GeV), FromEnergy(10 * KeV)},
	Visible,
}

// enumerated list of available profiles in order, plus fermi (thx guys)
var Profiles = []Profile{
	Gamma, XRay, ExtremeUltraviolet, Ultraviolet,
	Infrared, Radio, Microwave, Fermi,
}

// convenience function for humans with eyeballs
func (p Profile) String() string {
	return fmt.Sprintf("%s, %g to %g nm, shown as %g to %g nm", p.Name,
		p.Source.Lo, p.Source.Hi, p.Render.Lo, p.Render.Hi)
}

// we have to be careful with our swatch primitive because things get weird
// if we are asked for radiation which is outside the range we're expecting.
// so the most cautious thing we can do is just return black if that happens,
// rather than decide for the user what kind of radiation they would like.
func (p Profile) Swatch(nm float64) swatch.Swatch {
	if !p.holds(nm) {
		return swatch.Black
	}
	return swatch.Monochrome(int(math.Round(p.onto(nm))))
}

// Of is the radiation a colour stands for: the wavelength in Source whose
// swatch has this colour's hue, as seen from white.
//
// ok is false for white, and for a colour no single wavelength makes, like a
// purple. Past about 690 nm of Render every wavelength looks the same red,
// so a colour there names the shortest of them.
func (p Profile) Of(s swatch.Swatch) (nm float64, ok bool) {
	w, c := swatch.White.XY(), s.XY()
	d := swatch.XY{X: c.X - w.X, Y: c.Y - w.Y}
	if math.Hypot(d.X, d.Y) < 1e-9 || !p.spectral(d) {
		return 0, false
	}
	best, closest := 0, math.Inf(1)
	for n := int(math.Ceil(p.Render.Lo)); float64(n) < p.Render.Hi; n++ {
		if swatch.Monochrome(n) == swatch.Black {
			continue
		}
		if a := angle(d, towards(n)); a < closest {
			best, closest = n, a
		}
	}
	return p.back(float64(best)), best != 0
}

// return a swatch of what the mapped (typically to visible) light looks like
// from a given radiation.
func (p Profile) Light(l bands.Light) swatch.Swatch {
	seen := map[int]float64{}
	first, last := l.Span()
	for b := first; b <= last; b++ {
		lo, hi := b.Edges()
		a, z := max(lo, p.Source.Lo), min(hi, p.Source.Hi)
		if power := l.Power(b); power != 0 && z > a {
			spread(seen, p.onto(a), p.onto(z), power*(z-a)/(hi-lo))
		}
	}
	return bands.Split(seen).Swatch()
}

// spread shares power evenly among the whole nanometres from a up to b, or
// gives it to the nearest one when the stretch holds none.
func spread(seen map[int]float64, a, b, power float64) {
	lo, hi := int(math.Ceil(a)), int(math.Ceil(b))
	if hi <= lo {
		seen[int(math.Round((a+b)/2))] += power
		return
	}
	for n := lo; n < hi; n++ {
		seen[n] += power / float64(hi-lo)
	}
}

// holds is whether a wavelength is inside the profile's range, and the
// profile's two ranges are ranges at all.
func (p Profile) holds(nm float64) bool {
	ok := func(r Range) bool { return r.Lo > 0 && r.Hi > r.Lo }
	return ok(p.Source) && ok(p.Render) && nm >= p.Source.Lo &&
		nm < p.Source.Hi
}

// onto is where a wavelength in Source lands in Render, evenly in the
// logarithm of wavelength
func (p Profile) onto(nm float64) float64 {
	t := math.Log(nm/p.Source.Lo) / math.Log(p.Source.Hi/p.Source.Lo)
	return p.Render.Lo * math.Pow(p.Render.Hi/p.Render.Lo, t)
}

// reciprocal function for onto
func (p Profile) back(nm float64) float64 {
	t := math.Log(nm/p.Render.Lo) / math.Log(p.Render.Hi/p.Render.Lo)
	return p.Source.Lo * math.Pow(p.Source.Hi/p.Source.Lo, t)
}

// towards is the direction from white to the colour of one wavelength.
func towards(nm int) swatch.XY {
	w, c := swatch.White.XY(), swatch.Monochrome(nm).XY()
	return swatch.XY{X: c.X - w.X, Y: c.Y - w.Y}
}

// angle is the angle between two directions, in radians.
func angle(a, b swatch.XY) float64 {
	return math.Abs(math.Atan2(a.X*b.Y-a.Y*b.X, a.X*b.X+a.Y*b.Y))
}

// spectral is whether a direction from white is one some wavelength in Source
// points: the directions of the Source wavelengths sweep most of the way
// to white. this creates uncomfortable notions of what "purple" is.
func (p Profile) spectral(d swatch.XY) bool {
	lo, hi := math.Inf(1), math.Inf(-1)
	var prev float64
	for n := int(math.Ceil(p.Render.Lo)); float64(n) < p.Render.Hi; n++ {
		if swatch.Monochrome(n) == swatch.Black {
			continue
		}
		v := towards(n)
		a := math.Atan2(v.Y, v.X)
		if !math.IsInf(lo, 1) {
			// unwrap, so the sweep is one run of angle
			a += math.Round((prev-a)/(2*math.Pi)) * 2 * math.Pi
		}
		prev = a
		lo, hi = min(lo, a), max(hi, a)
	}
	phi := math.Atan2(d.Y, d.X)
	for k := -1.0; k <= 1; k++ {
		if at := phi + k*2*math.Pi; at >= lo && at <= hi {
			return true
		}
	}
	return false
}

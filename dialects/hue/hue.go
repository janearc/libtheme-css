// Package hue is the philips hue dialect: how a hue lamp names a colour,
// and how that becomes a swatch and comes back. A hue lamp never says a
// colour. It says a place on the horseshoe, xy, or a colour temperature
// in mirek, and separately how bright, as a percent that is the lamp's
// own scale and not a luminance. A gradient lamp says five places in
// order along itself. The dialect turns each of those into swatches at
// unit luminance and a list of them into a ramp; the reverse samples a
// ramp back to as many places as a lamp has and fits each to the
// triangle the lamp can reach. Brightness is not translated here: it is
// a law between a leader and a follower, and it lives with them.
package hue

import (
	"math"

	"github.com/janearc/libtheme-css/primitives/functions"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
)

// Point is a chromaticity as hue sends it.
type Point struct {
	X, Y float64
}

// Swatch is the point at unit luminance.
func Swatch(p Point) swatch.Swatch { return swatch.FromXY(p.X, p.Y, 1) }

// Of is the swatch's place on the horseshoe, as hue would want it.
func Of(s swatch.Swatch) Point {
	x, y := s.XY()
	return Point{x, y}
}

// Mirek is a colour temperature as hue sends it, reciprocal megakelvin,
// as a swatch on the planckian locus at unit luminance. Hue's lamps
// speak 153 to 500, which is 6536 K down to 2000 K.
func Mirek(m int) swatch.Swatch {
	if m <= 0 {
		return swatch.Black
	}
	return swatch.Planckian(1e6 / float64(m))
}

// Ramp is a lamp's points in order as a ramp: even stops, mixed in
// oklab, so the colour between two of a gradient's points is the one
// the eye would put there. One point is a ramp that is that colour
// everywhere.
func Ramp(points ...Point) functions.Ramp {
	swatches := make([]swatch.Swatch, len(points))
	for i, p := range points {
		swatches[i] = Swatch(p)
	}
	return functions.Even(ok.Mix, swatches...)
}

// Points is a ramp sampled back to n places, for a lamp with n of them:
// five for a gradient signe, one for a bulb. The samples are taken
// evenly from 0 to 1, so a ramp made from a lamp's own points comes back
// as those points.
func Points(r functions.Ramp, n int) []Point {
	out := make([]Point, 0, n)
	for _, s := range r.Samples(n) {
		out = append(out, Of(s))
	}
	return out
}

// Gamut is the triangle a lamp can reach, as the lamp reports it; two
// lamps in one room need not share one. A point outside is shown by the
// lamp as the nearest point inside, and Fit says which, so a caller can
// know what the lamp will do before asking.
type Gamut struct {
	Red, Green, Blue Point
}

// Contains is whether the point is inside the triangle, edges included.
func (g Gamut) Contains(p Point) bool {
	d1 := side(p, g.Red, g.Green)
	d2 := side(p, g.Green, g.Blue)
	d3 := side(p, g.Blue, g.Red)
	neg := d1 < 0 || d2 < 0 || d3 < 0
	pos := d1 > 0 || d2 > 0 || d3 > 0
	return !(neg && pos)
}

// Fit is the point itself when the lamp can reach it, and otherwise the
// nearest point on the triangle's edge.
func (g Gamut) Fit(p Point) Point {
	if g.Contains(p) {
		return p
	}
	best, dist := p, math.Inf(1)
	for _, e := range [][2]Point{{g.Red, g.Green}, {g.Green, g.Blue}, {g.Blue, g.Red}} {
		q := nearest(p, e[0], e[1])
		if d := math.Hypot(p.X-q.X, p.Y-q.Y); d < dist {
			best, dist = q, d
		}
	}
	return best
}

// side is which side of the line a→b the point is on, by sign.
func side(p, a, b Point) float64 {
	return (p.X-b.X)*(a.Y-b.Y) - (a.X-b.X)*(p.Y-b.Y)
}

// nearest is the closest point to p on the segment a→b.
func nearest(p, a, b Point) Point {
	dx, dy := b.X-a.X, b.Y-a.Y
	l2 := dx*dx + dy*dy
	if l2 == 0 {
		return a
	}
	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / l2
	t = math.Max(0, math.Min(1, t))
	return Point{a.X + t*dx, a.Y + t*dy}
}

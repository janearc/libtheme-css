package functions

import "math"

// Field is the second resident: a function from a position on a surface,
// u and v in 0..1, to a t in 0..1. A gradient is a field in front of a
// ramp, ramp.At(field.T(u, v)), and is not a type of its own. Change the
// field and the same ramp is a different picture; the ramp never learns
// that geometry exists. A surface with time in it is a field that takes
// the time as a parameter when it is built, so an animation is a field
// per frame in front of one ramp.
type Field interface {
	T(u, v float64) float64
}

// Point is a position on the surface, 0..1 each way, v downward like a
// screen.
type Point struct{ U, V float64 }

// Linear projects the point onto the axis from A to B: 0 at A, 1 at B,
// held at the ends beyond them. A dot product, clamped. Daffy's linear
// gradient and css's linear-gradient are this.
type Linear struct{ A, B Point }

// T is the position along the axis.
func (l Linear) T(u, v float64) float64 {
	dx, dy := l.B.U-l.A.U, l.B.V-l.A.V
	l2 := dx*dx + dy*dy
	if l2 == 0 {
		return 0
	}
	return clamp(((u-l.A.U)*dx + (v-l.A.V)*dy) / l2)
}

// Radial is distance from a centre over a radius: 0 at the centre, 1 at
// the radius and beyond. A sun is this with the centre low on the
// horizon. Aspect is the caller's: on a grid of cells twice as tall as
// wide, pass the position already corrected or the circle is an oval.
type Radial struct {
	Centre Point
	Radius float64
}

// T is the distance from the centre, in radii.
func (r Radial) T(u, v float64) float64 {
	if r.Radius <= 0 {
		return 0
	}
	return clamp(math.Hypot(u-r.Centre.U, v-r.Centre.V) / r.Radius)
}

// Conic is the angle around a centre over a full turn: 0 straight up,
// increasing clockwise, back to 1 at the top. css's conic-gradient.
type Conic struct{ Centre Point }

// T is the fraction of a turn.
func (c Conic) T(u, v float64) float64 {
	a := math.Atan2(u-c.Centre.U, -(v - c.Centre.V))
	if a < 0 {
		a += 2 * math.Pi
	}
	return a / (2 * math.Pi)
}

// Through is a field seen through a matrix: the position is transformed
// before the field reads it, so a circle becomes an ellipse and an axis
// turns. M is row-major, applied to (u - 0.5, v - 0.5) about the centre
// of the surface, then moved back.
type Through struct {
	M     [2][2]float64
	Field Field
}

// T is the inner field at the transformed position.
func (t Through) T(u, v float64) float64 {
	x, y := u-0.5, v-0.5
	return t.Field.T(t.M[0][0]*x+t.M[0][1]*y+0.5, t.M[1][0]*x+t.M[1][1]*y+0.5)
}

// clamp holds a value to 0..1.
func clamp(t float64) float64 { return math.Max(0, math.Min(1, t)) }

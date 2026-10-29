package swatch_test

import (
	"math"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
)

// edge is one side of a triangle and the corner opposite it.
type edge struct{ from, to, away swatch.XY }

// edgesOf lists the three edges of a triangle.
func edgesOf(gamut swatch.Gamut) []edge {
	red, green, blue := gamut.Red, gamut.Green, gamut.Blue
	return []edge{{red, green, blue}, {green, blue, red}, {blue, red, green}}
}

// outward is the unit vector perpendicular to an edge. It points away from the opposite corner.
func outward(side edge) swatch.XY {
	runX, runY := side.to.X-side.from.X, side.to.Y-side.from.Y
	normal := vec(runY, -runX)
	towards := normal.X*(side.away.X-side.from.X) + normal.Y*(side.away.Y-side.from.Y)
	if towards > none {
		normal = vec(-normal.X, -normal.Y)
	}
	length := math.Hypot(normal.X, normal.Y)
	return vec(normal.X/length, normal.Y/length)
}

// fitToTheFoot puts a point outside an edge and fits it. It returns the
// fit and the foot of the perpendicular the point was put over.
func fitToTheFoot(gamut swatch.Gamut, side edge, along, outBy float64) (got, foot swatch.XY) {
	foot = vec(side.from.X+along*(side.to.X-side.from.X), side.from.Y+along*(side.to.Y-side.from.Y))
	away := outward(side)
	outside := vec(foot.X+outBy*away.X, foot.Y+outBy*away.Y)
	return gamut.Fit(outside), foot
}

// apart says how far two places are from each other.
func apart(left, right swatch.XY) float64 { return math.Hypot(left.X-right.X, left.Y-right.Y) }

// Points outside an edge fit to the foot of the perpendicular from them.
// This holds for each of the three edges.
func TestFitGoesToTheFootOnEachEdge(t *testing.T) {
	numbers := pickedNumbers(t)
	for place, side := range edgesOf(numbers.Gamut) {
		got, foot := fitToTheFoot(numbers.Gamut, side, numbers.Foot.Along, numbers.Foot.OutBy)
		if apart(got, foot) > numbers.Foot.Tolerance {
			t.Errorf("edge %d fits to %v, the foot is %v", place, got, foot)
		}
	}
}

// The same holds on a triangle with an edge of length exactly one.
func TestFitOnATriangleWithAnEdgeOfLengthOne(t *testing.T) {
	numbers := pickedNumbers(t)
	for place, side := range edgesOf(numbers.UnitEdge) {
		got, foot := fitToTheFoot(numbers.UnitEdge, side, numbers.Foot.Along, numbers.Foot.OutBy)
		if apart(got, foot) > numbers.Foot.Tolerance {
			t.Errorf("edge %d fits to %v, the foot is %v", place, got, foot)
		}
	}
}

// Points past a corner fit to the corner. They do not fit to the line of an edge.
func TestFitPastACornerIsTheCorner(t *testing.T) {
	numbers := pickedNumbers(t)
	red := numbers.Gamut.Red
	past := vec(red.X+numbers.Past.East, red.Y-numbers.Past.South)
	if got := numbers.Gamut.Fit(past); apart(got, red) > numbers.Foot.Tolerance {
		t.Errorf("past the red corner fits to %v, not %v", got, red)
	}
}

// Past the blue corner two edges are at the same distance. The first of them
// names the point, in the order red-green then green-blue then blue-red.
// Green to blue works the corner out as green plus the whole run. That is a
// last bit away from the blue the other edge gives. The first edge wins.
func TestFitPastTheBlueCornerIsTheFirstEdgesCorner(t *testing.T) {
	numbers := pickedNumbers(t)
	green, blue := numbers.Gamut.Green, numbers.Gamut.Blue
	first := vec(green.X+(blue.X-green.X), green.Y+(blue.Y-green.Y))
	if first == blue {
		t.Fatalf("the corner is the same from both edges: %v", blue)
	}
	for _, past := range numbers.Blue.Points {
		if got := numbers.Gamut.Fit(past); got != first {
			t.Errorf("past the blue corner %v fits to %v, not %v", past, got, first)
		}
	}
}

package swatch_test

import (
	"math"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
)

// vec is a chromaticity at a place.
func vec(x, y float64) swatch.XY {
	var place swatch.XY
	place.X, place.Y = x, y
	return place
}

// weights is a point in the triangle's own coordinates. All three lie
// between nothing and all of a scale inside the triangle.
type weights struct{ red, green, blue float64 }

// barycentric works out a point's weights in a triangle.
func barycentric(gamut swatch.Gamut, point swatch.XY) weights {
	corner := gamut.Blue
	spanX, spanY := corner.X-gamut.Green.X, gamut.Green.Y-corner.Y
	det := spanY*(gamut.Red.X-corner.X) + spanX*(gamut.Red.Y-corner.Y)
	red := (spanY*(point.X-corner.X) + spanX*(point.Y-corner.Y)) / det
	green := ((corner.Y-gamut.Red.Y)*(point.X-corner.X) + (gamut.Red.X-corner.X)*(point.Y-corner.Y)) / det
	return weights{red, green, whole - red - green}
}

// leastOf is the smallest of a point's weights, ignoring sign.
func (found weights) leastOf() float64 {
	return min(math.Abs(found.red), math.Abs(found.green), math.Abs(found.blue))
}

// inside says whether all three weights are above nothing.
func (found weights) inside() bool {
	return found.red > none && found.green > none && found.blue > none
}

// Contains agrees with barycentric coordinates at every point of a grid
// that is not on an edge. It does so on every side of every edge.
func TestContainsAgreesWithBarycentricCoordinates(t *testing.T) {
	numbers := pickedNumbers(t)
	grid := numbers.Grid
	for x := grid.From; x < grid.To; x += grid.Step {
		for y := grid.From; y < grid.To; y += grid.Step {
			point := vec(x, y)
			found := barycentric(numbers.Gamut, point)
			if found.leastOf() < grid.Margin {
				continue
			}
			want := found.inside()
			if got := numbers.Gamut.Contains(point); got != want {
				t.Fatalf("Contains(%v) is %v, barycentric coordinates say %v", point, got, want)
			}
		}
	}
}

// turned is the triangle with its corners named each of three ways round.
// Each edge is in turn the first, the second and the third.
func turned(gamut swatch.Gamut) []swatch.Gamut {
	second, third := gamut, gamut
	second.Red, second.Green, second.Blue = gamut.Green, gamut.Blue, gamut.Red
	third.Red, third.Green, third.Blue = gamut.Blue, gamut.Red, gamut.Green
	return []swatch.Gamut{gamut, second, third}
}

// Points exactly a hair out of one edge are in. The slack is allowed.
func TestContainsAllowsExactlyAHairOutsideEachEdge(t *testing.T) {
	for _, triangle := range pickedNumbers(t).Hair.Triangles {
		for turn, gamut := range turned(triangle.Gamut) {
			if !gamut.Contains(triangle.HairOut) {
				t.Errorf("%s, turn %d: a hair outside one edge is outside", triangle.Turn, turn)
			}
		}
	}
}

// Points two hairs out of one edge are out.
func TestContainsRefusesPastTheHair(t *testing.T) {
	for _, triangle := range pickedNumbers(t).Hair.Triangles {
		for turn, gamut := range turned(triangle.Gamut) {
			if gamut.Contains(triangle.TwoHairsOut) {
				t.Errorf("%s, turn %d: two hairs outside one edge is inside", triangle.Turn, turn)
			}
		}
	}
}

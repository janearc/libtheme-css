package ok

import (
	"math"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/primitives/swatch"
)

// a gamut that is a box on chroma, as in the fit test.
func inBox(s swatch.Swatch) bool { return FromSwatch(s).Polar().C <= 0.1 }

// every cell of the grid is inside the gamut, chroma zero is grey, and
// a cell's own colour is nearest to itself.
func TestReach(t *testing.T) {
	g := Grid(12, 6, inBox)
	for y := 0; y < g.Rows; y++ {
		for x := 0; x < g.Cols; x++ {
			if !inBox(g.At(x, y).Rect().Swatch()) {
				t.Fatalf("cell %d,%d is outside the gamut", x,
					y)
			}
		}
	}
	g.Chroma = 0
	if c := g.At(3, 3); c.C != 0 {
		t.Errorf("chroma zero gave %v", c)
	}
	g.Chroma = 1
	if x, y := g.Nearest(g.At(7, 2).Rect()); x != 7 || y != 2 {
		t.Errorf("nearest to a cell's own colour was %d,%d", x, y)
	}
}

// The numbers the property below picks. Each is defined by this test.
const (
	// gridColumns is the width of the grid the cells are drawn from. It is
	// defined by this test.
	gridColumns = 12
	// gridRows is the height of the grid the cells are drawn from. It is
	// defined by this test.
	gridRows = 7
	// firstCell is the index of the first cell. It is defined by indexing.
	firstCell = 0
	// lastCell is how far from a count the index of the last cell is. It is
	// defined by indexing.
	lastCell = 1
	// overshoot is how far past 0 to 1 a drawn chroma goes. It is defined by
	// this test.
	overshoot = 4.0
	// noChroma is grey. It is defined by chroma.
	noChroma = 0.0
	// fullChroma is all the way to the edge. It is defined by chroma.
	fullChroma = 1.0
	// minus is the sign Copysign gives a number to make it negative. It is
	// defined by that function.
	minus = -1.0
)

// boxGrid is the test grid, at a chroma.
func boxGrid(chroma float64) Reach {
	reach := Grid(gridColumns, gridRows, inBox)
	reach.Chroma = chroma
	return reach
}

// The property: a chroma past either end of 0 to 1 is held at that end.
// Below nothing a cell is grey; past all of it a cell is no further out than
// the edge of the gamut, the same as all of it.
func TestAReachChromaPastItsEndsIsHeldAtThem(t *testing.T) {
	rapid.Check(t, heldAtTheEnds)
}

// heldAtTheEnds is one draw of the property above.
func heldAtTheEnds(t *rapid.T) {
	chroma := rapid.Float64Range(-overshoot, overshoot).Draw(t, "chroma")
	column := rapid.IntRange(firstCell, gridColumns-lastCell).Draw(t, "column")
	row := rapid.IntRange(firstCell, gridRows-lastCell).Draw(t, "row")
	got := boxGrid(chroma).At(column, row)
	held := math.Max(noChroma, math.Min(fullChroma, chroma))
	if want := boxGrid(held).At(column, row); got != want {
		t.Fatalf("chroma %v: %v, not %v", chroma, got, want)
	}
	if chroma <= noChroma && got.C != noChroma {
		t.Fatalf("chroma %v is not grey: %v", chroma, got)
	}
	if edge := boxGrid(fullChroma).At(column, row); got.C > edge.C {
		t.Fatalf("chroma %v reaches past the edge: %v, edge %v", chroma, got,
			edge)
	}
}

// Negative zero chroma is grey. The cell's chroma is the zero, since the
// negative one would be written as -0. The anti-test is the plain zero,
// which has always been held.
func TestANegativeZeroChromaIsTheZero(t *testing.T) {
	negative := math.Copysign(noChroma, minus)
	for _, chroma := range []float64{negative, noChroma} {
		got := boxGrid(chroma).At(firstCell, firstCell)
		if got.C != noChroma || math.Signbit(got.C) {
			t.Errorf("chroma %v gives %v, not the zero", chroma, got.C)
		}
	}
}

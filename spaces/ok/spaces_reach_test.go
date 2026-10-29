package ok_test

import (
	"math"
	"testing"

	"github.com/janearc/libtheme-css/spaces/ok"
)

// firstCell is the index of the first cell. It is defined by indexing.
const firstCell = 0

// near says whether two numbers agree to an absolute tolerance.
func near(got, want, tolerance float64) bool { return math.Abs(got-want) <= tolerance }

// Grid starts as the whole hue circle and the lightness range its comment names.
func TestGridStartsAsItsDocSays(t *testing.T) {
	reach := pickedNumbers(t).Reach
	grid := ok.Grid(reach.Cols, reach.Rows, everything)
	defaults := reach.Defaults
	if grid.Cols != reach.Cols || grid.Rows != reach.Rows || grid.In == nil {
		t.Errorf("size or gamut lost: %+v", grid)
	}
	if grid.HueFrom != defaults.HueFrom || grid.HueTo != defaults.HueTo {
		t.Errorf("hue runs %v to %v", grid.HueFrom, grid.HueTo)
	}
	if grid.LightFrom != defaults.LightFrom || grid.LightTo != defaults.LightTo {
		t.Errorf("lightness runs %v to %v", grid.LightFrom, grid.LightTo)
	}
	if grid.Chroma != defaults.Chroma {
		t.Errorf("chroma is %v", grid.Chroma)
	}
}

// Across the grid the hue steps evenly. Down it the lightness does.
func TestReachStepsEvenly(t *testing.T) {
	reach := pickedNumbers(t).Reach
	grid := ok.Grid(reach.Cols, reach.Rows, everything)
	defaults := reach.Defaults
	for col := firstCell; col < reach.Cols; col++ {
		share := float64(col) / float64(reach.Cols-oneStep)
		want := defaults.HueFrom + (defaults.HueTo-defaults.HueFrom)*share
		if got := grid.At(col, firstCell).H; !near(got, want, reach.Tolerance) {
			t.Errorf("column %d is hue %v, not %v", col, got, want)
		}
	}
	for row := firstCell; row < reach.Rows; row++ {
		share := float64(row) / float64(reach.Rows-oneStep)
		want := defaults.LightFrom + (defaults.LightTo-defaults.LightFrom)*share
		if got := grid.At(firstCell, row).L; !near(got, want, reach.Tolerance) {
			t.Errorf("row %d is lightness %v, not %v", row, got, want)
		}
	}
}

// Grids of two reach both ends of both axes.
func TestReachOfTwoReachesBothEnds(t *testing.T) {
	reach := pickedNumbers(t).Reach
	two := reach.Small.Two
	grid := ok.Grid(two, two, everything)
	last := two - oneStep
	if got := grid.At(last, firstCell).H; got != reach.Defaults.HueTo {
		t.Errorf("the last of two columns is hue %v", got)
	}
	if got := grid.At(firstCell, last).L; !near(got, reach.Defaults.LightTo, reach.Tolerance) {
		t.Errorf("the last of two rows is lightness %v", got)
	}
}

// Grids of one stay at the start. They do not divide by nothing.
func TestReachOfOneStaysAtTheStart(t *testing.T) {
	reach := pickedNumbers(t).Reach
	one := reach.Small.One
	cell := ok.Grid(one, one, everything).At(firstCell, firstCell)
	if cell.H != reach.Defaults.HueFrom || cell.L != reach.Defaults.LightFrom || math.IsNaN(cell.C) {
		t.Errorf("one by one is %+v", cell)
	}
}

// The ends of two are a step on from the start.
func TestReachOfTwoMoves(t *testing.T) {
	reach := pickedNumbers(t).Reach
	two := reach.Small.Two
	grid := ok.Grid(two, two, everything)
	last := two - oneStep
	start := grid.At(firstCell, firstCell)
	if grid.At(last, firstCell).H == start.H || grid.At(firstCell, last).L == start.L {
		t.Errorf("a two by two grid stands still")
	}
}

// Grids that begin at another hue step from it to their end.
func TestReachStepsFromAHueThatIsNotZero(t *testing.T) {
	reach := pickedNumbers(t).Reach
	grid := ok.Grid(reach.Cols, reach.Rows, everything)
	grid.HueFrom, grid.HueTo = reach.Hues.From, reach.Hues.To
	for col := firstCell; col < reach.Cols; col++ {
		share := float64(col) / float64(reach.Cols-oneStep)
		want := reach.Hues.From + (reach.Hues.To-reach.Hues.From)*share
		if got := grid.At(col, firstCell).H; !near(got, want, reach.Tolerance) {
			t.Errorf("column %d is hue %v, not %v", col, got, want)
		}
	}
}

// Where the gamut lets in more than the probe the cell is the probe.
// The chroma setting scales it.
func TestReachCellIsTheProbeWhereNothingStopsIt(t *testing.T) {
	reach := pickedNumbers(t).Reach
	grid := ok.Grid(reach.Cols, reach.Rows, everything)
	cell := reach.Twin
	if got := grid.At(cell.Col, cell.Row).C; !near(got, reach.Probe, reach.Tolerance) {
		t.Errorf("an unstopped cell has chroma %v, not %v", got, reach.Probe)
	}
	grid.Chroma = reach.Half
	if got := grid.At(cell.Col, cell.Row).C; !near(got, reach.Probe*reach.Half, reach.Tolerance) {
		t.Errorf("at half, an unstopped cell has chroma %v", got)
	}
}

// Nearest finds every cell that has no twin where it is.
// The last column is the first one come round again.
func TestNearestFindsEveryCell(t *testing.T) {
	reach := pickedNumbers(t).Reach
	limit := box{reach.BoxEdge}
	grid := ok.Grid(reach.Cols, reach.Rows, limit.holds)
	for row := firstCell; row < reach.Rows; row++ {
		for col := firstCell; col < reach.Cols-oneStep; col++ {
			colour := grid.At(col, row).Rect()
			if foundCol, foundRow := grid.Nearest(colour); foundCol != col || foundRow != row {
				t.Errorf("cell %d,%d came back as %d,%d", col, row, foundCol, foundRow)
			}
		}
	}
}

// Colours far from every cell still have a nearest one.
func TestNearestOfAColourFarAway(t *testing.T) {
	reach := pickedNumbers(t).Reach
	limit := box{reach.BoxEdge}
	grid := ok.Grid(reach.Cols, reach.Rows, limit.holds)
	_, lightest := grid.Nearest(reach.Far.Light)
	_, darkest := grid.Nearest(reach.Far.Dark)
	if lightest != firstCell || darkest != reach.Rows-oneStep {
		t.Errorf("the far light is in row %d and the far dark in row %d", lightest, darkest)
	}
}

// Cells that are the same colour give the first of them.
func TestNearestOfTwinsIsTheFirst(t *testing.T) {
	reach := pickedNumbers(t).Reach
	limit := box{reach.BoxEdge}
	grid := ok.Grid(reach.Cols, reach.Rows, limit.holds)
	grid.Chroma = noChroma
	twin := reach.Twin
	grey := ok.Grey(grid.At(twin.Col, twin.Row).L)
	if col, row := grid.Nearest(grey); col != twin.Col || row != twin.Row {
		t.Errorf("the nearest of the twins is %d,%d", col, row)
	}
}

// Nothing off the grid is a cell. The colour of the cell past the end
// is answered by a cell that is there.
func TestNearestStaysOnTheGrid(t *testing.T) {
	reach := pickedNumbers(t).Reach
	limit := box{reach.BoxEdge}
	grid := ok.Grid(reach.Cols, reach.Rows, limit.holds)
	off := reach.OffTop
	pastCols := grid.At(grid.Cols, off.Row).Rect()
	pastRows := grid.At(off.Col, grid.Rows).Rect()
	col, _ := grid.Nearest(pastCols)
	_, row := grid.Nearest(pastRows)
	if col >= grid.Cols || row >= grid.Rows {
		t.Errorf("nearest to past the end is column %d and row %d", col, row)
	}
}

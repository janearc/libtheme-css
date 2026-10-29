package functions_test

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/janearc/libtheme-css/primitives/functions"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
)

// picked is the numbers these tests pick, from testdata/spaces.json.
type picked struct {
	HardEdge float64 `json:"hard edge at"`
	PastEdge float64 `json:"past the edge at"`
	Halfway  float64 `json:"halfway"`
	Two      int     `json:"two samples"`
}

// pickedNumbers reads the test numbers, or the test stops.
func pickedNumbers(t *testing.T) picked {
	t.Helper()
	raw, err := os.ReadFile("testdata/spaces.json")
	if err != nil {
		t.Fatal(err)
	}
	var numbers picked
	if err := json.Unmarshal(raw, &numbers); err != nil {
		t.Fatal(err)
	}
	return numbers
}

// The ends of the ramps below are the dark and the light.
var (
	start = swatch.Black
	end   = swatch.White
)

// The ends of the scale a stop sits on. They are defined by the scale.
const (
	// atStart is the first position. It is defined by the scale.
	atStart = 0
	// atEnd is the last position. It is defined by the scale.
	atEnd = 1
)

// hardRamp has a hard edge. It runs dark to light and then does it again.
func hardRamp(edge float64) functions.Ramp {
	first := functions.Stop{At: atStart, Swatch: start}
	lit := functions.Stop{At: edge, Swatch: end}
	dim := functions.Stop{At: edge, Swatch: start}
	last := functions.Stop{At: atEnd, Swatch: end}
	return functions.New(ok.Mix, first, lit, dim, last)
}

// Stops at one position keep the order they were given in.
func TestStopsAtOnePositionKeepTheirOrder(t *testing.T) {
	edge := pickedNumbers(t).HardEdge
	last := functions.Stop{At: atEnd, Swatch: end}
	lit := functions.Stop{At: edge, Swatch: end}
	dim := functions.Stop{At: edge, Swatch: start}
	first := functions.Stop{At: atStart, Swatch: start}
	ramp := functions.New(ok.Mix, last, lit, dim, first)
	want := []functions.Stop{first, lit, dim, last}
	if !slices.Equal(ramp.Stops, want) {
		t.Errorf("the stops are %v, not %v", ramp.Stops, want)
	}
}

// Stops at different positions are put in order of position.
func TestStopsAtDifferentPositionsAreOrdered(t *testing.T) {
	last := functions.Stop{At: atEnd, Swatch: end}
	first := functions.Stop{At: atStart, Swatch: start}
	ramp := functions.New(ok.Mix, last, first)
	want := []functions.Stop{first, last}
	if !slices.Equal(ramp.Stops, want) {
		t.Errorf("the stops are %v, not %v", ramp.Stops, want)
	}
}

// At a hard edge the ramp is the colour the first half ends on.
func TestAtAHardEdgeIsTheFirstHalfsEnd(t *testing.T) {
	edge := pickedNumbers(t).HardEdge
	got := ok.FromSwatch(hardRamp(edge).At(edge))
	distance := ok.Distance(got, ok.FromSwatch(end))
	if !(distance <= ok.Exact) {
		t.Errorf("the ramp at %v is %v from the first half's end", edge, distance)
	}
}

// Past the hard edge the ramp is on the second half.
func TestPastAHardEdgeIsTheSecondHalf(t *testing.T) {
	numbers := pickedNumbers(t)
	got := hardRamp(numbers.HardEdge).At(numbers.PastEdge)
	want := ok.Mix.Mix(start, end, numbers.Halfway)
	if got != want {
		t.Errorf("the ramp at %v is %v, not %v", numbers.PastEdge, got, want)
	}
}

// Two samples are the two ends.
func TestTwoSamplesAreTheEnds(t *testing.T) {
	two := pickedNumbers(t).Two
	ramp := functions.Even(ok.Mix, start, end)
	got := ramp.Samples(two)
	want := []swatch.Swatch{start, end}
	if !slices.Equal(got, want) {
		t.Errorf("two samples are %v, not %v", got, want)
	}
}

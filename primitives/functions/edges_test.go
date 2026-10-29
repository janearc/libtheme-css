package functions_test

import (
	"math"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/primitives/functions"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// The numbers the tests pick. Each is defined by this test.
const (
	// beyond is how far past the ends of 0 to 1 a drawn position goes. It
	// is defined by this test.
	beyond = 5.0
	// fewest is the fewest stops of a drawn ramp. It is defined by this test.
	fewest = 2
	// most is the most stops of a drawn ramp. It is defined by this test.
	most = 8
	// notLater is the least a position may be past its neighbour: a drawn
	// ramp's stops are strictly later each time. It is defined by this test.
	notLater = 0.01
	// widest is the most a drawn gap between two stops is. It is defined by
	// this test.
	widest = 1.0
	// first is the position of a ramp's first stop. It is defined by this
	// test: the ramp starts there.
	first = 0
	// within is how near the ramp is to a stop colour at its position, in
	// oklab. It is defined by this test: far below what an eye can tell.
	within = 1e-6
	// lastOffset is how far from the end of a list its last item is. It is
	// defined by indexing.
	lastOffset = 1
	// halfway is how far along a position is. It is defined by this test.
	halfway = 0.5
)

// colourOf is a drawn colour, from three bytes.
func colourOf(t *rapid.T, name string) swatch.Swatch {
	byteOf := rapid.Uint8()
	return srgb.RGB8(byteOf.Draw(t, name+" red"), byteOf.Draw(t, name+" green"),
		byteOf.Draw(t, name+" blue")).Swatch()
}

// drawn is a ramp of stops later each time, in oklab.
func drawn(t *rapid.T) functions.Ramp {
	count := rapid.IntRange(fewest, most).Draw(t, "stops")
	gap := rapid.Float64Range(notLater, widest)
	at := rapid.Float64Range(-widest, widest).Draw(t, "start")
	var stops []functions.Stop
	for range count {
		stops = append(stops, functions.Stop{At: at,
			Swatch: colourOf(t, "stop")})
		at += gap.Draw(t, "gap")
	}
	return functions.New(ok.Mix, stops...)
}

// The property: at the position of a stop the ramp is that stop's colour, as
// near as the mixer's arithmetic gets; before the first stop it is the
// first, and after the last it is the last, exactly, wherever the position
// is.
func TestARampHoldsItsStopsAndItsEnds(t *testing.T) {
	rapid.Check(t, holdsItsStops)
}

// holdsItsStops is one draw of the property above.
func holdsItsStops(t *rapid.T) {
	ramp := drawn(t)
	stops := ramp.Stops
	for _, stop := range stops {
		got := ok.FromSwatch(ramp.At(stop.At))
		want := ok.FromSwatch(stop.Swatch)
		if !ok.Same(got, want, within) {
			t.Fatalf("at %v the ramp is %v, not %v", stop.At, got, want)
		}
	}
	over := rapid.Float64Range(first, beyond).Draw(t, "beyond")
	if ramp.At(stops[first].At-over) != stops[first].Swatch {
		t.Fatalf("before the first stop the ramp is not the first colour")
	}
	end := stops[len(stops)-lastOffset]
	if ramp.At(end.At+over) != end.Swatch {
		t.Fatalf("after the last stop the ramp is not the last colour")
	}
}

// The anti-tests: a ramp with nothing in it is no light, wherever it is
// asked; a position that is not a number is not a place on the ramp, and the
// ramp answers its last colour rather than mixing with nothing; and a last
// stop that has no position is the last colour wherever it is asked after
// the first.
func TestARampWithNothingToShowShowsNothing(t *testing.T) {
	none := functions.Ramp{In: ok.Mix}
	for _, at := range []float64{-beyond, first, halfway, beyond} {
		if got := none.At(at); got != swatch.NoLight() {
			t.Errorf("an empty ramp at %v is %v", at, got)
		}
	}
	black, white := srgb.Black.Swatch(), srgb.White.Swatch()
	ramp := functions.Even(ok.Mix, black, white)
	if got := ramp.At(math.NaN()); got != white {
		t.Errorf("a ramp asked for no number is %v, not its last colour", got)
	}
	lost := functions.Ramp{In: ok.Mix, Stops: []functions.Stop{{At: first,
		Swatch: black}, {At: math.NaN(), Swatch: white}}}
	if got := lost.At(halfway); got != white {
		t.Errorf("a ramp whose last stop has no place is %v", got)
	}
}

package ok_test

import (
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// The numbers the tests pick. Each is defined by this test.
const (
	// tolerance is how far two oklab colours may be apart. It is defined by
	// this test: far below what an eye can tell, above float error.
	tolerance = 1e-9
	// none is no distance along a mix. It is defined by the mix.
	none = 0.0
	// whole is all the way along a mix. It is defined by the mix.
	whole = 1.0
	// overshoot is how far past 0 to 1 a drawn position goes. It is defined
	// by this test.
	overshoot = 4.0
)

// colourOf is a drawn colour from three bytes.
func colourOf(t *rapid.T, name string) swatch.Swatch {
	byteOf := rapid.Uint8()
	return srgb.RGB8(byteOf.Draw(t, name+" red"), byteOf.Draw(t, name+" green"),
		byteOf.Draw(t, name+" blue")).Swatch()
}

// mixed is two colours mixed, as oklab.
func mixed(left, right swatch.Swatch, along float64) ok.OKLab {
	return ok.FromSwatch(ok.Mix.Mix(left, right, along))
}

// The property: Mix is the straight line in oklab. At nothing it is the
// first colour, at all of it the second, and in between each coordinate is
// that far along.
func TestMixIsAStraightLineInOklab(t *testing.T) {
	rapid.Check(t, mixesStraight)
}

// mixesStraight is one draw of the property above.
func mixesStraight(t *rapid.T) {
	left, right := colourOf(t, "left"), colourOf(t, "right")
	along := rapid.Float64Range(none, whole).Draw(t, "along")
	from, to := ok.FromSwatch(left), ok.FromSwatch(right)
	want := ok.OKLab{L: from.L + (to.L-from.L)*along,
		A: from.A + (to.A-from.A)*along, B: from.B + (to.B-from.B)*along}
	if got := mixed(left, right, along); !ok.Same(got, want, tolerance) {
		t.Fatalf("%v of the way: %v, not %v", along, got, want)
	}
	if !ok.Same(mixed(left, right, none), from, tolerance) {
		t.Fatalf("a mix at nothing is not the first colour")
	}
	if !ok.Same(mixed(left, right, whole), to, tolerance) {
		t.Fatalf("a mix at all of it is not the second colour")
	}
}

// The anti-test: mixing a colour with itself does not move it, however far
// along: the line has no length, and the mix does not invent one.
func TestMixingAColourWithItselfStaysPut(t *testing.T) {
	rapid.Check(t, staysPut)
}

// staysPut is one draw of the property above.
func staysPut(t *rapid.T) {
	colour := colourOf(t, "colour")
	along := rapid.Float64Range(-overshoot, overshoot).Draw(t, "along")
	if got := mixed(colour, colour, along); !ok.Same(got,
		ok.FromSwatch(colour), tolerance) {
		t.Fatalf("a colour mixed with itself at %v moved to %v", along, got)
	}
}

package css_test

import (
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/css"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// brightest is the most light a drawn swatch has, in the units where white
// is 1. It is defined by this test: well past white, so some are out of
// gamut.
const brightest = 2

// lamps are the three bytes a sheet holds for a colour.
type lamps struct{ red, green, blue uint8 }

// bytesOf is the bytes a sheet would hold for a swatch.
func bytesOf(colour swatch.Swatch) lamps {
	rgb, _ := srgb.FromSwatch(colour)
	red, green, blue := rgb.Bytes()
	return lamps{red, green, blue}
}

// roundTrip is a swatch set in a sheet, written as css and read back.
func roundTrip(colour swatch.Swatch) (swatch.Swatch, string) {
	sheet := css.New()
	sheet.Set("c", colour)
	text := sheet.String()
	back, _ := css.Read(text).Roles.Get("c")
	return back, text
}

// A source holds sRGB bytes. A byte colour is read back as exactly the
// swatch it was written from.
func TestAByteColourSurvivesTheSheetExactly(t *testing.T) {
	rapid.Check(t, survivesExactly)
}

// survivesExactly is one draw of the property above.
func survivesExactly(t *rapid.T) {
	byteOf := rapid.Uint8()
	colour := srgb.RGB8(byteOf.Draw(t, "red"), byteOf.Draw(t, "green"),
		byteOf.Draw(t, "blue")).Swatch()
	if back, text := roundTrip(colour); back != colour {
		t.Fatalf("read back %v, not %v:\n%s", back, colour, text)
	}
}

// Any other swatch is held as the nearest byte colour. That covers one
// between bytes and one past what the lamps can make. Reading gives that
// colour, not the swatch. Writing it again keeps it. Only the note beside
// the hex changes, once, from the swatch to the byte colour.
func TestAnyOtherColourIsHeldAsItsNearestBytes(t *testing.T) {
	rapid.Check(t, heldAsBytes)
}

// heldAsBytes is one draw of the property above.
func heldAsBytes(t *rapid.T) {
	light := rapid.Float64Range(nothing, brightest)
	colour := swatch.FromXYZ(light.Draw(t, "x"), light.Draw(t, "y"),
		light.Draw(t, "z"))
	back, _ := roundTrip(colour)
	if bytesOf(back) != bytesOf(colour) {
		t.Fatalf("read back bytes %v, not %v", bytesOf(back), bytesOf(colour))
	}
	if again, _ := roundTrip(back); again != back {
		t.Fatalf("writing again changed the colour: %v, then %v", back, again)
	}
}

// The oklch beside a hex says what the colour is to a reader. It is not
// read: the hex is the colour, whatever the comment says.
func TestTheCommentBesideAColourIsNotRead(t *testing.T) {
	source := ":root {\n  --a: #ff0000; /* oklch(10% 0.001 100) */\n}\n"
	got, _ := css.Read(source).Roles.Get("a")
	if got != srgb.MustHex("#ff0000").Swatch() {
		t.Errorf("--a read as %v", got)
	}
}

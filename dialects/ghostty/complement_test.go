package ghostty_test

import (
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/dialects/ghostty"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// complete is what a byte and its complement add up to. It is defined by the
// byte.
const complete = 255

// noLight is the least light a drawn swatch has. It is defined by light: none
// is the least there is.
const noLight = 0

// brightest is the most light a drawn swatch has, in the units where white
// is 1. It is defined by this test: well past white, so some are out of
// gamut.
const brightest = 2

// lamps are the three bytes a file holds for a colour.
type lamps struct{ red, green, blue uint8 }

// bytesOf is a swatch as the bytes a file holds.
func bytesOf(value swatch.Swatch) lamps {
	rgb, _ := srgb.FromSwatch(value)
	red, green, blue := rgb.Bytes()
	return lamps{red, green, blue}
}

// fullPair is whether each byte of one and the same byte of the other add up
// to a full byte.
func fullPair(was, now lamps) bool {
	return int(was.red)+int(now.red) == complete &&
		int(was.green)+int(now.green) == complete &&
		int(was.blue)+int(now.blue) == complete
}

// The opposite of no light is white and of white is no light, exactly.
func TestComplementOfBlackAndWhite(t *testing.T) {
	if ghostty.Complement(srgb.Black.Swatch()) != srgb.White.Swatch() {
		t.Error("the complement of black is not white")
	}
	if ghostty.Complement(srgb.White.Swatch()) != srgb.Black.Swatch() {
		t.Error("the complement of white is not black")
	}
}

// The property: for any colour the lamps can hold, each byte of the
// complement and the byte it came from add up to 255, and the complement of
// the complement is the colour.
func TestComplementAddsToAFullByte(t *testing.T) {
	rapid.Check(t, addsToAFullByte)
}

// addsToAFullByte is one draw of the property above.
func addsToAFullByte(t *rapid.T) {
	byteOf := rapid.Uint8()
	start := srgb.RGB8(byteOf.Draw(t, "red"), byteOf.Draw(t, "green"),
		byteOf.Draw(t, "blue")).Swatch()
	turned := ghostty.Complement(start)
	if !fullPair(bytesOf(start), bytesOf(turned)) {
		t.Fatalf("%v and %v do not add to a full byte", bytesOf(start),
			bytesOf(turned))
	}
	if ghostty.Complement(turned) != start {
		t.Fatalf("the complement of the complement is %v, not %v",
			ghostty.Complement(turned), start)
	}
}

// The property: a swatch the lamps cannot make, or one off the byte grid,
// is clipped and quantised first. The complement is of the bytes the file
// would hold, whatever the swatch.
func TestComplementIsOfTheBytesNotTheSwatch(t *testing.T) {
	rapid.Check(t, complementsTheBytes)
}

// complementsTheBytes is one draw of the property above.
func complementsTheBytes(t *rapid.T) {
	light := rapid.Float64Range(noLight, brightest)
	start := swatch.FromXYZ(light.Draw(t, "x"), light.Draw(t, "y"),
		light.Draw(t, "z"))
	turned := ghostty.Complement(start)
	if !fullPair(bytesOf(start), bytesOf(turned)) {
		t.Fatalf("%v and its complement do not add to a full byte",
			bytesOf(start))
	}
}

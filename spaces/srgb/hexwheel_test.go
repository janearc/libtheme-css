package srgb_test

import (
	"math"
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/spaces/srgb"
)

// The numbers the tests pick. Each is defined by this test.
const (
	// tolerance is how far two lamp levels may differ. It is defined by this
	// test: far below one byte, above float error.
	tolerance = 1e-9
	// none is no light, and the start of a mix. It is defined by the mix.
	none = 0.0
	// whole is full light, and the end of a mix. It is defined by the mix.
	whole = 1.0
	// shortDigits is the digits of the short form of a hex colour. It is
	// defined by css: #rgb.
	shortDigits = 3
	// fullDigits is the digits of the full form of a hex colour. It is
	// defined by css: #rrggbb.
	fullDigits = 6
	// fewest is the least characters a drawn string has. It is defined by
	// this test: the empty string.
	fewest = 0
	// longest is the most characters a drawn string has. It is defined by
	// this test.
	longest = 10
)

// The characters a drawn string is made of.
const (
	// hexDigits are the characters of a hex colour. It is defined by hex.
	hexDigits = "0123456789abcdefABCDEF"
	// strays are characters that are no digit of hex. This test defines them:
	// signs, a prefix letter and punctuation.
	strays = "+-xg.,_"
)

// isHex is whether every character of a string is a digit of hex.
func isHex(text string) bool {
	return strings.Trim(text, hexDigits) == ""
}

// verdict says whether css takes a string, for a message.
func verdict(takes bool) string {
	if takes {
		return "takes"
	}
	return "refuses"
}

// The property: FromHex takes the strings css takes and no others. Css takes
// three or six digits, with or without the hash. A sign or a stray character
// is refused. What it takes, Hex writes back in the full lower-case form.
func TestFromHexTakesExactlyWhatCSSTakes(t *testing.T) {
	rapid.Check(t, takesWhatCSSTakes)
}

// takesWhatCSSTakes is one draw of the property above.
func takesWhatCSSTakes(t *rapid.T) {
	letters := rapid.SampledFrom([]rune(hexDigits + strays))
	text := string(rapid.SliceOfN(letters, fewest, longest).Draw(t, "text"))
	hash := rapid.SampledFrom([]string{"", "#"}).Draw(t, "hash")
	colour, err := srgb.FromHex(hash + text)
	length := len(text)
	takes := isHex(text) && (length == shortDigits || length == fullDigits)
	if (err == nil) != takes {
		t.Fatalf("%q: error %v, but css %s it", hash+text, err, verdict(takes))
	}
	if err == nil {
		back, again := srgb.FromHex(colour.Hex())
		if again != nil || back != colour {
			t.Fatalf("%q wrote %q, which reads as %v", text, colour.Hex(), back)
		}
	}
}

// refusedBy is whether the function panics.
func refusedBy(run func()) (panicked bool) {
	defer note(&panicked)
	run()
	return panicked
}

// note is deferred to say that a panic happened.
func note(panicked *bool) {
	if recover() != nil {
		*panicked = true
	}
}

// notAColour is a call that gives MustHex a string no colour is.
func notAColour() { srgb.MustHex("not a colour") }

// The anti-test: MustHex stops the program on a string that is no colour,
// and gives the colour for one that is.
func TestMustHexStopsOnANonColour(t *testing.T) {
	if srgb.MustHex("#ff0000") != srgb.Red {
		t.Error("#ff0000 is not red")
	}
	if !refusedBy(notAColour) {
		t.Error("MustHex took a string that is no colour")
	}
}

// The property: Mix is the straight line through the lamps. At nothing it
// is the first colour. At all of it, the second. Between, each lamp is
// that far along.
func TestMixIsAStraightLineThroughTheLamps(t *testing.T) {
	rapid.Check(t, mixesThroughTheLamps)
}

// mixesThroughTheLamps is one draw of the property above.
func mixesThroughTheLamps(t *rapid.T) {
	byteOf := rapid.Uint8()
	left := srgb.RGB8(byteOf.Draw(t, "r1"), byteOf.Draw(t, "g1"),
		byteOf.Draw(t, "b1"))
	right := srgb.RGB8(byteOf.Draw(t, "r2"), byteOf.Draw(t, "g2"),
		byteOf.Draw(t, "b2"))
	along := rapid.Float64Range(none, whole).Draw(t, "along")
	line := srgb.Mix.Mix(left.Swatch(), right.Swatch(), along)
	mixed, _ := srgb.FromSwatch(line)
	want := srgb.RGB{R: left.R + (right.R-left.R)*along,
		G: left.G + (right.G-left.G)*along,
		B: left.B + (right.B-left.B)*along}
	if math.Abs(mixed.R-want.R) > tolerance ||
		math.Abs(mixed.G-want.G) > tolerance ||
		math.Abs(mixed.B-want.B) > tolerance {
		t.Fatalf("%v of the way from %v to %v: %v, not %v", along, left, right,
			mixed, want)
	}
}

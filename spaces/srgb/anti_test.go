package srgb_test

import (
	"testing"

	"github.com/janearc/libtheme-css/spaces/srgb"
)

// The numbers the anti-tests pick. Each is defined by the comment on it.
const (
	// lessThanNone is an amount of light below none. It is defined by this
	// test: no lamp can make it.
	lessThanNone = -0.5
	// brighterThanWhite is an amount of light above the white. It is defined
	// by this test.
	brighterThanWhite = 4.0
	// greyByte is the byte of a mid grey in all three lamps. It is defined
	// by this test.
	greyByte = uint8(0x80)
)

// The anti-test: strings that look a little like hex colours and are not are
// refused. "00000+" and "+f0000" were once read as colours, because the
// reader below FromHex took a sign as part of a digit.
func TestFromHexRefusesAStrayCharacter(t *testing.T) {
	for _, text := range []string{"00000+", "+f0000", "0x0000", "-1-1-1",
		"ggg", "#12345", "", "#", "1234567"} {
		if got, err := srgb.FromHex(text); err == nil {
			t.Errorf("%q was read as %v", text, got)
		}
	}
}

// The anti-test: a grey has no hue and no saturation in either cylinder, so
// nothing reads an angle into three equal lamps.
func TestAGreyHasNoHueToSpeakOf(t *testing.T) {
	for _, grey := range []srgb.RGB{srgb.Black, srgb.White,
		srgb.RGB8(greyByte, greyByte, greyByte)} {
		hsl, hsv := grey.HSL(), grey.HSV()
		if hsl.S != none || hsv.S != none || hsl.H != none ||
			hsv.H != none {
			t.Errorf("%v: hsl %+v hsv %+v", grey, hsl, hsv)
		}
	}
}

// The anti-test: a colour the lamps cannot make is clipped into range. It
// says it is not in gamut. In says the same.
func TestAnImpossibleColourIsClippedAndSaid(t *testing.T) {
	impossible := srgb.FromLight(lessThanNone, whole, none)
	got, in := srgb.FromSwatch(impossible)
	if in || srgb.In(impossible) {
		t.Fatal("light below none claimed to be in gamut")
	}
	for _, lamp := range []float64{got.R, got.G, got.B} {
		if lamp < none || lamp > whole {
			t.Errorf("the clipped colour has a lamp at %v", lamp)
		}
	}
}

// The anti-test: the curve does not clip. Light above white stays above
// full and light below none stays below none. FromSwatch clips.
func TestTheCurveDoesNotClip(t *testing.T) {
	if above := srgb.FromLinear(brighterThanWhite); above <= whole {
		t.Errorf("light of %v came back as a level of %v",
			brighterThanWhite, above)
	}
	if below := srgb.FromLinear(lessThanNone); below >= none {
		t.Errorf("light of %v came back as a level of %v", lessThanNone, below)
	}
}

// The anti-test: a lamp past either end is written as the end. Hex never
// writes more than a byte, whatever the lamps hold.
func TestHexWritesNoMoreThanAByte(t *testing.T) {
	wild := srgb.RGB{R: lessThanNone, G: brighterThanWhite, B: none}
	if got := wild.Hex(); got != "#00ff00" {
		t.Errorf("%v was written as %s", wild, got)
	}
}

package ok_test

import (
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/spaces/ok"
)

// quarter divides the eye's distance to the chroma where a hue is dropped,
// and halfTurn is the hue opposite the red axis, where dropping a hue costs
// the most. Both are defined: by the rule under test, and by the circle.
const (
	// quarter is the part of the eye's distance a hue is dropped below. It
	// is defined by the rule under test.
	quarter = 4
	// halfTurn is the hue half way round the wheel. It is defined by the
	// circle.
	halfTurn = 180
)

// colour is a nearly white colour of the chroma and hue given.
func colour(chroma, hue float64) ok.OKLCH {
	var made ok.OKLCH
	made.L = ok.White.L - ok.Eye
	made.C, made.H = chroma, hue
	return made
}

// A colour with chroma below an eye's distance but well above nothing keeps
// its hue when written: dropping it would move the colour by about twice
// the chroma, which is more than an eye can tell.
func TestALowChromaColourKeepsItsHue(t *testing.T) {
	faint := colour(ok.Eye-ok.Exact, halfTurn)
	if text := faint.String(); strings.Contains(text, "none") {
		t.Errorf("%s dropped a hue that mattered", text)
	}
}

// The anti-test: a grey, and a colour with almost no chroma, still write
// none, since its hue is noise.
func TestAGreyHasNoHue(t *testing.T) {
	for _, almost := range []ok.OKLCH{colour(ok.Black.L, ok.Black.L),
		colour(ok.Eye/quarter/quarter, halfTurn)} {
		if text := almost.String(); !strings.HasSuffix(text, "none)") {
			t.Errorf("%s kept a hue that is noise", text)
		}
	}
}

// The chroma where the hue is dropped is a quarter of Eye, and not before:
// at it the hue is written, just under it is not.
func TestTheHueIsDroppedBelowAQuarterOfTheEye(t *testing.T) {
	edge := ok.Eye / quarter
	at := colour(edge, halfTurn).String()
	under := colour(edge-ok.Exact, halfTurn).String()
	if strings.Contains(at, "none") || !strings.Contains(under, "none") {
		t.Errorf("at the edge %s, under it %s", at, under)
	}
}

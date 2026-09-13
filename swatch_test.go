package libtheme

import (
	"math"
	"testing"
)

func near(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

// Every hex code survives a trip through the swatch and back unchanged.
func TestHexRoundTrip(t *testing.T) {
	for _, h := range []string{"#000000", "#ffffff", "#ff0000", "#00ff00", "#0000ff", "#ff6ec7", "#ffa2ff", "#160d2b", "#7f7f7f", "#01fe80"} {
		s := MustHex(h)
		if got := s.Hex(); got != h {
			t.Errorf("%s came back as %s", h, got)
		}
		if _, _, _, ok := s.SRGB8(); !ok {
			t.Errorf("%s reported out of gamut", h)
		}
	}
}

// The ends of L are black and white, and a grey has no chroma.
func TestLightnessEnds(t *testing.T) {
	if l := MustHex("#000000").OKLab().L; !near(l, 0, 1e-6) {
		t.Errorf("black L = %v", l)
	}
	if l := MustHex("#ffffff").OKLab().L; !near(l, 1, 1e-3) {
		t.Errorf("white L = %v", l)
	}
	if c := MustHex("#7f7f7f").OKLCH().C; !near(c, 0, 1e-3) {
		t.Errorf("grey C = %v", c)
	}
}

// The vaporwave pinks land where the colour tools say they do: the web
// pink at about oklch(74% 0.20 345), the terminal pink lighter and
// twenty degrees toward magenta.
func TestKnownPinks(t *testing.T) {
	web := MustHex("#ff6ec7").OKLCH()
	if !near(web.L, 0.74, 0.01) || !near(web.C, 0.20, 0.01) || !near(web.H, 345, 1.5) {
		t.Errorf("web pink = %+v", web)
	}
	term := MustHex("#ffa2ff").OKLCH()
	if !near(term.L, 0.84, 0.01) || !near(term.H, 327, 1.5) {
		t.Errorf("terminal pink = %+v", term)
	}
	if term.L <= web.L {
		t.Errorf("terminal pink should be the lighter one")
	}
}

// Mix holds its ends, and halfway between blue and yellow is not mud:
// its lightness sits between the two and it keeps some chroma.
func TestMix(t *testing.T) {
	blue, yellow := MustHex("#0000ff"), MustHex("#ffff00")
	if d := Distance(Mix(blue, yellow, 0), blue); d > 1e-6 {
		t.Errorf("mix at 0 moved by %v", d)
	}
	if d := Distance(Mix(blue, yellow, 1), yellow); d > 1e-6 {
		t.Errorf("mix at 1 moved by %v", d)
	}
	mid := Mix(blue, yellow, 0.5).OKLCH()
	lo, hi := blue.OKLab().L, yellow.OKLab().L
	if mid.L < lo || mid.L > hi {
		t.Errorf("midpoint L %v outside %v..%v", mid.L, lo, hi)
	}
	if mid.C < 0.02 {
		t.Errorf("midpoint went grey: C = %v", mid.C)
	}
}

// A swatch outside the screen's reach says so instead of pretending.
func TestGamutIsHonest(t *testing.T) {
	loud := FromOKLCH(OKLCH{0.7, 0.4, 145}) // greener than any sRGB lamp
	if _, _, _, ok := loud.SRGB(); ok {
		t.Errorf("an impossible green claimed to be in gamut")
	}
}

// String prints what CSS accepts.
func TestString(t *testing.T) {
	if got := MustHex("#ff6ec7").String(); got != "oklch(74% 0.20 345)" {
		t.Errorf("String = %q", got)
	}
}

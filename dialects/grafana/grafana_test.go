package grafana

import (
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"

	"github.com/janearc/libtheme-css/spaces/ok"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// pink is the vaporwave accent, and a hue with somewhere to go in both
// directions.
var pink = srgb.MustHex("#ff6ec7").Swatch()

// the five shades run from dark to super-light, and each is lighter than
// the one before it: what grafana's names say is what they do.
func TestShadesAscend(t *testing.T) {
	last := -1.0
	for _, s := range Shades {
		l := ok.FromSwatch(Shade(pink, s)).Polar().L
		if l <= last {
			t.Errorf("%q is not lighter than the one before: %.3f",
				s, l)
		}
		last = l
	}
}

// every shade is a colour sRGB can name, because a hex is all grafana can
// be told.
func TestShadesAreInGamut(t *testing.T) {
	for _, s := range Shades {
		if !InGamut(Shade(pink, s)) {
			t.Errorf("shade %q left the gamut", s)
		}
	}
}

// shaded is each shade of some colours, by hue and shade, as the hex a
// dashboard would hold. They were written by the code as it stood before the
// shade was mixed directly: the two ends of the gamut, a primary at its edge
// and two hues with room to move, so a change to the mix, the fit or the
// steps shows.
var shaded = map[string]string{
	"#ff6ec7 dark": "#63264c", "#ff6ec7 semi-dark": "#ae4987",
	"#ff6ec7 ": "#ff6ec7", "#ff6ec7 light": "#ff99d4",
	"#ff6ec7 super-light": "#ffbee1",
	"#1f60c4 dark":        "#06204a", "#1f60c4 semi-dark": "#113f85",
	"#1f60c4 ": "#1f60c4", "#1f60c4 light": "#5989d6",
	"#1f60c4 super-light": "#8fb1e5",
	"#00ff00 dark":        "#006300", "#00ff00 semi-dark": "#00ae00",
	"#00ff00 ": "#00ff00", "#00ff00 light": "#84ff7c",
	"#00ff00 super-light": "#b7ffb0",
	"#000000 dark":        "#000000", "#000000 semi-dark": "#000000",
	"#000000 ": "#000000", "#000000 light": "#222222",
	"#000000 super-light": "#636363",
	"#ffffff dark":        "#636363", "#ffffff semi-dark": "#aeaeae",
	"#ffffff ": "#ffffff", "#ffffff light": "#ffffff",
	"#ffffff super-light": "#ffffff",
}

// Every shade of every one of those colours is the hex it always was, and a
// shade nobody named leaves the colour as it was.
func TestShadesAreWhatTheyWere(t *testing.T) {
	for key, want := range shaded {
		hue, shade, _ := strings.Cut(key, " ")
		moved := Shade(srgb.MustHex(hue).Swatch(), shade)
		lamps, _ := srgb.FromSwatch(moved)
		if got := lamps.Hex(); got != want {
			t.Errorf("%s: %s, want %s", key, got, want)
		}
	}
	if Shade(pink, "nonesuch") != pink {
		t.Error("an unknown shade moved the colour")
	}
}

// the bare hue is the colour it was given, unmoved.
func TestBareHueIsUnchanged(t *testing.T) {
	if got := Shade(pink, ""); !ok.Same(ok.FromSwatch(got),
		ok.FromSwatch(pink), 1e-12) {
		t.Error("the bare hue moved")
	}
}

// a name goes to a shade and a hue and back to the same name, bare hues
// included.
func TestNamesRoundTrip(t *testing.T) {
	for _, hue := range []string{"blue", "green", "super"} {
		for _, shade := range Shades {
			name := Name(shade, hue)
			gotShade, gotHue := Split(name)
			if gotShade != shade || gotHue != hue {
				t.Errorf("%q read back as %q %q", name,
					gotShade, gotHue)
			}
		}
	}
}

// a palette holds five names per hue, all of them hex, and a name it was
// not given is refused rather than guessed.
func TestPalette(t *testing.T) {
	p := Palette(map[string]swatch.Swatch{"pink": pink})
	if len(p) != len(Shades) {
		t.Fatalf("palette has %d names, want %d", len(p), len(Shades))
	}
	for name, hex := range p {
		if len(hex) != 7 || hex[0] != '#' {
			t.Errorf("%s is %q, which is not a hex", name, hex)
		}
	}
	bound := map[string]swatch.Swatch{"pink": pink}
	if _, found := Lookup("dark-pink", bound); !found {
		t.Error("a bound hue was not found")
	}
	if _, found := Lookup("dark-teal", bound); found {
		t.Error("an unbound hue was answered")
	}
}

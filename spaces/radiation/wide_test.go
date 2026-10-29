package radiation_test

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/radiation"
)

// The numbers the tests pick. Each is defined by this test.
const (
	// wideLo is the shortest wavelength of a profile wider than the
	// observer, in nanometres. It is defined by this test: below the
	// observer's 360, where it has no row.
	wideLo = 300.0
	// wideHi is the longest wavelength of that profile, in nanometres. It is
	// defined by this test: above the observer's 830.
	wideHi = 900.0
	// shownLo is the shortest wavelength a drawn colour is made at. It is
	// defined by this test: clear of the violet and red ends, where several
	// wavelengths look the same.
	shownLo = 450.0
	// shownHi is the longest wavelength a drawn colour is made at. It is
	// defined by this test.
	shownHi = 650.0
	// within is how far, in nanometres, a colour's named wavelength is from
	// the one it was made at. It is defined by this test: the mapping is good
	// to a nanometre.
	within = 1.5
)

// wide is a profile whose range runs past both ends of what the observer
// has rows for, shown as itself.
func wide() radiation.Profile {
	both := radiation.Range{Lo: wideLo, Hi: wideHi}
	return radiation.Profile{Name: "wide", Source: both, Render: both}
}

// The property: in a profile wider than the observer, a colour made at a
// wavelength names that wavelength. The wavelengths the observer has no row
// for are passed over, not mistaken for the nearest colour.
func TestAWiderProfileStillNamesTheWavelength(t *testing.T) {
	rapid.Check(t, namesItsWavelength)
}

// namesItsWavelength is one draw of the property above.
func namesItsWavelength(t *rapid.T) {
	shown := rapid.Float64Range(shownLo, shownHi).Draw(t, "shown")
	profile := wide()
	named, ok := profile.Of(profile.Swatch(shown))
	if !ok || math.Abs(named-shown) > within {
		t.Fatalf("%g nm named %g, ok %v", shown, named, ok)
	}
}

// The anti-test: a wavelength outside the source range has no colour in the
// profile, whichever side it is off.
func TestAWavelengthOutsideTheProfileIsNoLight(t *testing.T) {
	rapid.Check(t, outsideIsNoLight)
}

// outsideIsNoLight is one draw of the property above.
func outsideIsNoLight(t *rapid.T) {
	below := rapid.Float64Range(math.SmallestNonzeroFloat64, wideLo).Draw(t, "below")
	above := rapid.Float64Range(wideHi, math.MaxFloat64).Draw(t, "above")
	profile := wide()
	if profile.Swatch(below) != swatch.NoLight() ||
		profile.Swatch(above) != swatch.NoLight() {
		t.Fatalf("%g nm or %g nm was a colour", below, above)
	}
}

// A profile says its name, its range and what it is shown as, for a person
// to read.
func TestAProfileSaysWhatItIs(t *testing.T) {
	for _, profile := range append(radiation.Profiles, wide()) {
		said := profile.String()
		for _, part := range []string{profile.Name,
			fmt.Sprintf("%g to %g nm", profile.Source.Lo, profile.Source.Hi),
			fmt.Sprintf("shown as %g to %g nm", profile.Render.Lo,
				profile.Render.Hi)} {
			if !strings.Contains(said, part) {
				t.Errorf("%q does not say %q", said, part)
			}
		}
	}
}

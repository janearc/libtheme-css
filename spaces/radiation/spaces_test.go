package radiation_test

import (
	"math"
	"slices"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/radiation"
)

// rangePair is a range the package has and the range the JSON says it is.
type rangePair struct{ got, want radiation.Range }

// The ranges are as radiation.go types them. The JSON says which of the
// figures I recall from the standard and which I could not check.
func TestTheNamedRanges(t *testing.T) {
	named := pickedNumbers(t).Named
	for name, pair := range map[string]rangePair{
		"gamma ray":           {radiation.Gamma.Source, named.Gamma},
		"x-ray":               {radiation.XRay.Source, named.XRay},
		"extreme ultraviolet": {radiation.ExtremeUltraviolet.Source, named.ExtremeUV},
		"ultraviolet":         {radiation.Ultraviolet.Source, named.Ultraviolet},
		"infrared":            {radiation.Infrared.Source, named.Infrared},
		"microwave":           {radiation.Microwave.Source, named.Microwave},
		"radio":               {radiation.Radio.Source, named.Radio},
	} {
		if pair.got != pair.want {
			t.Errorf("%s runs %v, not %v", name, pair.got, pair.want)
		}
	}
}

// Fermi's far end carries 10 keV and its near end 300 GeV.
func TestFermiEdgesCarryTheirEnergies(t *testing.T) {
	fermi := pickedNumbers(t).Fermi
	long := radiation.Energy(radiation.Fermi.Source.Hi)
	short := radiation.Energy(radiation.Fermi.Source.Lo)
	if math.Abs(long-fermi.Lowest) > fermi.Tolerance*fermi.Lowest {
		t.Errorf("the long edge carries %v eV, not %v", long, fermi.Lowest)
	}
	if math.Abs(short-fermi.Highest) > fermi.Tolerance*fermi.Highest {
		t.Errorf("the short edge carries %v eV, not %v", short, fermi.Highest)
	}
}

// The units are the SI's powers of ten.
func TestTheUnitsArePowersOfTen(t *testing.T) {
	ladders := pickedNumbers(t).Ladders
	energy := []float64{
		radiation.Electronvolt, radiation.KeV, radiation.MeV, radiation.GeV, radiation.TeV,
	}
	frequency := []float64{
		radiation.Hz, radiation.KHz, radiation.MHz, radiation.GHz, radiation.THz,
	}
	if !slices.Equal(energy, ladders.Energy) {
		t.Errorf("the energy units are %v, not %v", energy, ladders.Energy)
	}
	if !slices.Equal(frequency, ladders.Frequency) {
		t.Errorf("the frequency units are %v, not %v", frequency, ladders.Frequency)
	}
}

// The top of the source range is outside it. The profile shows it as no light.
func TestSwatchAtTheTopOfTheSourceIsNoLight(t *testing.T) {
	top := radiation.Ultraviolet.Source.Hi
	if got := radiation.Ultraviolet.Swatch(top); got != swatch.NoLight() {
		t.Errorf("the top of the source is %v, not no light", got)
	}
}

// The bottom of the source range is in it. It is the bottom of the rendered range.
func TestSwatchAtTheBottomOfTheSourceIsTheBottomOfTheRender(t *testing.T) {
	bottom := radiation.Ultraviolet.Source.Lo
	want := swatch.Monochrome(int(radiation.Ultraviolet.Render.Lo))
	if got := radiation.Ultraviolet.Swatch(bottom); got != want {
		t.Errorf("the bottom of the source is %v, not %v", got, want)
	}
}

// Profiles with a rendered range that is no range show nothing.
func TestProfilesWithNoRenderRangeShowNothing(t *testing.T) {
	profiles := pickedNumbers(t).Profile
	for _, broken := range []radiation.Profile{profiles.Flat, profiles.BackRend} {
		if got := broken.Swatch(profiles.InsideNM); got != swatch.NoLight() {
			t.Errorf("%s shows %v", broken.Name, got)
		}
	}
}

// Profiles with a source that runs backwards show nothing.
func TestProfilesWithNoSourceRangeShowNothing(t *testing.T) {
	profiles := pickedNumbers(t).Profile
	backwards := profiles.BackSrc
	for _, nm := range []float64{profiles.InsideNM, profiles.AboveNM} {
		if got := backwards.Swatch(nm); got != swatch.NoLight() {
			t.Errorf("%s shows %v at %v nm", backwards.Name, got, nm)
		}
	}
}

// Sound profiles show what they hold.
func TestSoundProfilesShowWhatTheyHold(t *testing.T) {
	profiles := pickedNumbers(t).Profile
	if got := profiles.Sound.Swatch(profiles.InsideNM); got == swatch.NoLight() {
		t.Errorf("%s shows nothing at %v nm", profiles.Sound.Name, profiles.InsideNM)
	}
}

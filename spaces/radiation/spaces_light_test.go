package radiation_test

import (
	"math"
	"testing"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/radiation"
)

// halving is the divisor that finds a half or a middle. It is defined by the arithmetic.
const halving = 2

// sameLight says whether two swatches agree in X, Y and Z to a relative tolerance.
func sameLight(left, right swatch.Swatch, tolerance float64) bool {
	leftX, leftY, leftZ := left.XYZ()
	rightX, rightY, rightZ := right.XYZ()
	return math.Abs(leftX-rightX) <= tolerance*math.Abs(rightX) &&
		math.Abs(leftY-rightY) <= tolerance*math.Abs(rightY) &&
		math.Abs(leftZ-rightZ) <= tolerance*math.Abs(rightZ)
}

// lineOf is a single line of light at a wavelength.
func lineOf(nm, power float64) bands.Light {
	spectrum := map[float64]float64{}
	spectrum[nm] = power
	return bands.Lines(spectrum)
}

// profileOver is a profile that renders a source range into a render range.
func profileOver(source, render radiation.Range) radiation.Profile {
	var profile radiation.Profile
	profile.Name = "over"
	profile.Source, profile.Render = source, render
	return profile
}

// squeezedOver renders from the start to the end of a source into one nanometre.
func squeezedOver(from, to float64, render radiation.Range) radiation.Profile {
	var source radiation.Range
	source.Lo, source.Hi = from, to
	return profileOver(source, render)
}

// cellLight is the light of a power at the one nanometre the squeezed range holds.
func cellLight(cell int, power float64) swatch.Swatch {
	seen := map[int]float64{}
	seen[cell] = power
	return bands.Split(seen).Swatch()
}

// Bands cut in half by the end of the source give half their power.
func TestLightGivesThePartOfTheBandInsideTheSource(t *testing.T) {
	light := pickedNumbers(t).Light
	lo, hi := bands.Of(light.LineNM).Edges()
	profile := squeezedOver(lo-light.Lead, (lo+hi)/halving, light.Squeezed)
	got := profile.Light(lineOf(light.LineNM, light.Power))
	want := cellLight(light.CellNM, light.Power/halving)
	if !sameLight(got, want, light.Tolerance) {
		t.Errorf("half a band shows %v, not %v", got, want)
	}
}

// Bands wholly inside the source give all their power.
func TestLightGivesAWholeBandInsideTheSource(t *testing.T) {
	light := pickedNumbers(t).Light
	lo, hi := bands.Of(light.LineNM).Edges()
	profile := squeezedOver(lo-light.Lead, hi+light.Lead, light.Squeezed)
	got := profile.Light(lineOf(light.LineNM, light.Power))
	want := cellLight(light.CellNM, light.Power)
	if !sameLight(got, want, light.Tolerance) {
		t.Errorf("a whole band shows %v, not %v", got, want)
	}
}

// Light outside the source is no light at all. It does not subtract.
func TestLightOutsideTheSourceIsNoLight(t *testing.T) {
	numbers := pickedNumbers(t)
	outside := numbers.Profile.Outside
	for _, nm := range []float64{numbers.Profile.BelowNM, numbers.Profile.AboveNM} {
		got := outside.Light(lineOf(nm, numbers.Light.Power))
		if got != swatch.NoLight() {
			t.Errorf("light at %v nm, outside %v, shows %v", nm, outside.Source, got)
		}
	}
}

// Bands spread their power evenly over the whole nanometres in them.
// No power goes to the first nanometre past the band.
func TestLightSpreadsAcrossTheWholeNanometres(t *testing.T) {
	light := pickedNumbers(t).Light
	lo, hi := bands.Of(light.LineNM).Edges()
	var source radiation.Range
	source.Lo, source.Hi = lo-light.Lead, hi+light.Lead
	profile := profileOver(source, source)
	first, past := int(math.Ceil(lo)), int(math.Ceil(hi))
	share := light.Power / float64(past-first)
	seen := map[int]float64{}
	for nm := first; nm < past; nm++ {
		seen[nm] = share
	}
	want := bands.Split(seen).Swatch()
	got := profile.Light(lineOf(light.LineNM, light.Power))
	if !sameLight(got, want, light.Tolerance) {
		t.Errorf("the band shows %v, not %v", got, want)
	}
}

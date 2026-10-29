package bands_test

import (
	"math"
	"slices"
	"testing"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/libtheme-css/primitives/swatch"
)

// Numbers the tests use that the scale or the table defines.
const (
	// none is no wavelength and no power. It is defined by the scale.
	none = 0
	// unitPower is a power of exactly one. It is defined by the scale.
	unitPower = 1.0
	// oneNM is the step of the observer's table. It is defined by the table.
	oneNM = 1
	// halving is the divisor that finds a middle. It is defined by the arithmetic.
	halving = 2
)

// blackBody is a black body at a temperature. It is split into bands and slid.
func blackBody(kelvin float64, slide int) bands.Light {
	spectrum := swatch.Blackbody(kelvin)
	light := bands.Split(spectrum)
	return light.Shift(slide)
}

// close says whether two numbers agree to a relative tolerance.
func close(got, want, tolerance float64) bool {
	gap := math.Abs(got - want)
	return gap <= tolerance*math.Abs(want)
}

// Visible light is the bands the Band comment names at 24 to an octave.
func TestVisibleLightIsTheBandsTheCommentNames(t *testing.T) {
	visible := pickedNumbers(t).Visible
	short, long := bands.Of(visible.Short), bands.Of(visible.Long)
	if int(short) != visible.First || int(long) != visible.Last {
		t.Errorf("visible light is bands %d to %d, not %d to %d",
			short, long, visible.First, visible.Last)
	}
}

// Visible light is twenty-six bands wide, as FUDGE.md says.
func TestVisibleLightIsTwentySixBandsWide(t *testing.T) {
	visible := pickedNumbers(t).Visible
	short, long := bands.Of(visible.Short), bands.Of(visible.Long)
	if width := int(long-short) + oneNM; width != visible.Width {
		t.Errorf("visible light is %d bands wide, not %d", width, visible.Width)
	}
}

// Adding keeps the span of the lights. It runs from the earlier first
// band to the later last band and no wider.
func TestAddSpansBothLightsExactly(t *testing.T) {
	numbers := pickedNumbers(t)
	left := blackBody(numbers.CoolLamp, numbers.Shift)
	right := blackBody(numbers.WarmLamp, -numbers.Shift)
	leftFirst, leftLast := left.Span()
	rightFirst, rightLast := right.Span()
	first, last := bands.Add(left, right).Span()
	if first != min(leftFirst, rightFirst) || last != max(leftLast, rightLast) {
		t.Errorf("the sum spans %d to %d", first, last)
	}
}

// Light added to no light spans what the light does. This holds from either side.
func TestAddingNoLightDoesNotWidenTheSpan(t *testing.T) {
	numbers := pickedNumbers(t)
	light := blackBody(numbers.MiddleLamp, numbers.Shift)
	var nothing bands.Light
	wantFirst, wantLast := light.Span()
	for name, sum := range map[string]bands.Light{
		"no light first": bands.Add(nothing, light),
		"no light last":  bands.Add(light, nothing),
	} {
		first, last := sum.Span()
		if first != wantFirst || last != wantLast {
			t.Errorf("%s spans %d to %d, not %d to %d", name, first, last, wantFirst, wantLast)
		}
	}
}

// wholeNanometres lists the whole nanometres in a band.
func wholeNanometres(band bands.Band) []int {
	lo, hi := band.Edges()
	var inside []int
	for nm := int(math.Ceil(lo)); float64(nm) < hi; nm += oneNM {
		if bands.Of(float64(nm)) == band {
			inside = append(inside, nm)
		}
	}
	return inside
}

// meanResponse is the observer's mean response over some nanometres.
func meanResponse(nms []int) (x, y, z float64) {
	for _, nm := range nms {
		mx, my, mz := swatch.Monochrome(nm).XYZ()
		x, y, z = x+mx, y+my, z+mz
	}
	count := float64(len(nms))
	return x / count, y / count, z / count
}

// daylightUnit is one over the luminance of D65 at the power its table gives.
func daylightUnit() float64 {
	var luminance float64
	for nm, power := range swatch.D65() {
		_, y, _ := swatch.Monochrome(nm).XYZ()
		luminance += power * y
	}
	return oneNM / luminance
}

// lineOf is a single line of light at a wavelength.
func lineOf(nm, power float64) bands.Light {
	spectrum := map[float64]float64{}
	spectrum[nm] = power
	return bands.Lines(spectrum)
}

// centreOf is the middle of a band in nanometres.
func centreOf(band bands.Band) float64 {
	lo, hi := band.Edges()
	return (lo + hi) / halving
}

// Each line is its band's mean response times its power. The scale is
// the one that gives D65 a Y of one.
func TestALineIsTheMeanResponseOfItsBand(t *testing.T) {
	numbers := pickedNumbers(t)
	band := bands.Band(numbers.WeightBand)
	mx, my, mz := meanResponse(wholeNanometres(band))
	scale := numbers.LinePower * daylightUnit()
	got := lineOf(centreOf(band), numbers.LinePower).Swatch()
	gx, gy, gz := got.XYZ()
	if !close(gx, mx*scale, numbers.Tolerance) || !close(gy, my*scale, numbers.Tolerance) ||
		!close(gz, mz*scale, numbers.Tolerance) {
		t.Errorf("the line is %v %v %v, not %v %v %v", gx, gy, gz, mx*scale, my*scale, mz*scale)
	}
}

// The next nanometre up is in the next band. The mean with it in is
// another number.
func TestTheMeanLeavesOutTheNextBand(t *testing.T) {
	numbers := pickedNumbers(t)
	band := bands.Band(numbers.WeightBand)
	inside := wholeNanometres(band)
	withNext := append(slices.Clone(inside), slices.Max(inside)+oneNM)
	mean, _, _ := meanResponse(inside)
	longer, _, _ := meanResponse(withNext)
	if close(mean, longer, numbers.Tolerance) {
		t.Errorf("the mean does not change with the next nanometre: %v", mean)
	}
}

// Power of exactly one is light and counts.
func TestAPowerOfOneCounts(t *testing.T) {
	numbers := pickedNumbers(t)
	got := lineOf(numbers.LineNM, unitPower).Swatch()
	_, y, _ := got.XYZ()
	if got == swatch.NoLight() || y <= none {
		t.Errorf("a power of one at %v nm is %v", numbers.LineNM, got)
	}
}

// samePowers says whether two lights have one power in every band of the first.
func samePowers(left, right bands.Light) bool {
	first, last := left.Span()
	for band := first; band <= last; band++ {
		if left.Power(band) != right.Power(band) {
			return false
		}
	}
	return true
}

// Light at no wavelength is no light. The result is the same without it.
func TestLinesIgnoreAWavelengthOfNothing(t *testing.T) {
	numbers := pickedNumbers(t)
	spectrum := map[float64]float64{}
	spectrum[numbers.LineNM] = unitPower
	without := bands.Lines(spectrum)
	spectrum[none] = unitPower
	with := bands.Lines(spectrum)
	withFirst, withLast := with.Span()
	first, last := without.Span()
	if withFirst != first || withLast != last || !samePowers(with, without) {
		t.Errorf("a line at nothing changed the light: %d to %d", withFirst, withLast)
	}
}

package srgb_test

import (
	"math"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// Numbers the tests use that the scale or the arithmetic defines.
const (
	// nothing is no light. It is defined by the scale.
	nothing = 0
	// unit is all of a scale. It is defined by the scale.
	unit = 1
)

// The gamut is the standard's primaries. The figures are in the JSON with their source.
func TestGamutIsTheStandardsPrimaries(t *testing.T) {
	want := pickedNumbers(t).Gamut
	if srgb.Gamut != want {
		t.Errorf("the gamut is %+v, not %+v", srgb.Gamut, want)
	}
}

// At the break the curve is still the straight piece.
func TestToLinearAtTheBreakIsStraight(t *testing.T) {
	curve := pickedNumbers(t).Curve
	want := curve.Break / curve.Slope
	got := srgb.ToLinear(curve.Break)
	if math.Abs(got-want) > curve.Tolerance*want {
		t.Errorf("ToLinear(%v) is %v, the straight piece gives %v", curve.Break, got, want)
	}
}

// Just past the break the curve is the power piece. It is not the straight one.
func TestToLinearPastTheBreakIsPower(t *testing.T) {
	curve := pickedNumbers(t).Curve
	past := math.Nextafter(curve.Break, unit)
	want := math.Pow((past+curve.Offset)/curve.Scale, curve.Power)
	straight := past / curve.Slope
	got := srgb.ToLinear(past)
	if math.Abs(got-want) > curve.Tolerance*want || math.Abs(got-straight) <= curve.Tolerance*straight {
		t.Errorf("ToLinear(%v) is %v, power gives %v and straight %v", past, got, want, straight)
	}
}

// The curve back is the straight piece at its own break.
func TestFromLinearAtTheBreakIsStraight(t *testing.T) {
	curve := pickedNumbers(t).Curve
	want := curve.Slope * curve.LinearBreak
	got := srgb.FromLinear(curve.LinearBreak)
	if math.Abs(got-want) > curve.Tolerance*want {
		t.Errorf("FromLinear(%v) is %v, the straight piece gives %v", curve.LinearBreak, got, want)
	}
}

// Just past that break the curve back is the power piece.
func TestFromLinearPastTheBreakIsPower(t *testing.T) {
	curve := pickedNumbers(t).Curve
	past := math.Nextafter(curve.LinearBreak, unit)
	want := curve.Scale*math.Pow(past, unit/curve.Power) - curve.Offset
	straight := curve.Slope * past
	got := srgb.FromLinear(past)
	if math.Abs(got-want) > curve.Tolerance*want || math.Abs(got-straight) <= curve.Tolerance*straight {
		t.Errorf("FromLinear(%v) is %v, power gives %v and straight %v", past, got, want, straight)
	}
}

// exactly finds a swatch whose red lamp in linear light is exactly the bound.
// It tries the floats next to the swatch the bound should make.
// Floating point does not promise a hit, so the search says if it missed.
func exactly(t *testing.T, bound float64) swatch.Swatch {
	t.Helper()
	reach := pickedNumbers(t).Slack.Reach
	x, y, z := srgb.FromLight(bound, nothing, nothing).XYZ()
	for step := -reach; step <= reach; step++ {
		bits := int64(math.Float64bits(x)) + int64(step)
		nudged := math.Float64frombits(uint64(bits))
		candidate := swatch.FromXYZ(nudged, y, z)
		if red, _, _ := srgb.Light(candidate); red == bound {
			return candidate
		}
	}
	t.Fatalf("no swatch within %d floats has a red lamp of exactly %v", reach, bound)
	return swatch.NoLight()
}

// Channels exactly on the slack are in gamut at both ends. The slack is allowed.
func TestAChannelExactlyOnTheSlackIsInGamut(t *testing.T) {
	slack := pickedNumbers(t).Slack
	for name, bound := range map[string]float64{"below": slack.Below, "above": slack.Above} {
		if _, in := srgb.FromSwatch(exactly(t, bound)); !in {
			t.Errorf("a red lamp exactly on the slack %s is out of gamut", name)
		}
	}
}

// justPast finds the swatch whose red lamp is the first one past the bound,
// moving x one float at a time away from the swatch the bound should make.
// Red rises with x, so towards minus infinity walks red down and towards
// plus infinity walks it up. It does not ask for any one exact lamp: the
// float just past a bound is often one no swatch reaches, and which one is
// reached depends on how the machine rounds. It asks for the nearest that
// is past, which every machine has.
func justPast(t *testing.T, bound, towards float64) swatch.Swatch {
	t.Helper()
	reach := pickedNumbers(t).Slack.Reach
	x, y, z := srgb.FromLight(bound, nothing, nothing).XYZ()
	for step := 0; step <= reach; step++ {
		candidate := swatch.FromXYZ(x, y, z)
		red, _, _ := srgb.Light(candidate)
		if (towards < 0 && red < bound) || (towards > 0 && red > bound) {
			return candidate
		}
		x = math.Nextafter(x, towards)
	}
	t.Fatalf("no swatch within %d floats has a red lamp past %v", reach, bound)
	return swatch.NoLight()
}

// Channels a hair past the slack are out at both ends: the first red lamp
// past each edge of the slack, a few floats away. So are channels far past
// it.
func TestAChannelPastTheSlackIsOutOfGamut(t *testing.T) {
	slack := pickedNumbers(t).Slack
	pasts := map[string]swatch.Swatch{
		"below": justPast(t, slack.Below, math.Inf(-1)),
		"above": justPast(t, slack.Above, math.Inf(1)),
	}
	for name, past := range pasts {
		if _, in := srgb.FromSwatch(past); in {
			t.Errorf("a red lamp a hair past the slack %s is in gamut", name)
		}
	}
	far := srgb.FromLight(slack.WayOver, nothing, nothing)
	if _, in := srgb.FromSwatch(far); in {
		t.Errorf("a red lamp at %v is in gamut", slack.WayOver)
	}
}

// Hues more than a turn round the wheel are the same hue wrapped.
func TestAHueMoreThanATurnIsTheSameHueWrapped(t *testing.T) {
	numbers := pickedNumbers(t)
	for _, pair := range numbers.Wrap {
		got, want := pair.HSV.RGB(), pair.SameAs.RGB()
		gap := math.Max(math.Abs(got.R-want.R), math.Abs(got.G-want.G))
		if gap = math.Max(gap, math.Abs(got.B-want.B)); gap > numbers.WheelTolerance {
			t.Errorf("hue %v is %+v, hue %v is %+v", pair.HSV.H, got, pair.SameAs.H, want)
		}
	}
}

// Lamps at exactly the dark threshold have no saturation in either cylinder.
func TestALampAtTheDarkThresholdHasNoSaturation(t *testing.T) {
	lamp := pickedNumbers(t).Dark.AtThreshold
	if saturation := lamp.HSL().S; saturation != nothing {
		t.Errorf("the hsl saturation is %v", saturation)
	}
	if saturation := lamp.HSV().S; saturation != nothing {
		t.Errorf("the hsv saturation is %v", saturation)
	}
}

// Lamps past the dark threshold have a saturation.
func TestALampPastTheDarkThresholdHasSaturation(t *testing.T) {
	lamp := pickedNumbers(t).Dark.PastThreshold
	if lamp.HSL().S == nothing || lamp.HSV().S == nothing {
		t.Errorf("a lamp at %v has no saturation", lamp.R)
	}
}

// Spreads at the dark threshold have the hue of their brightest lamp.
func TestASpreadAtTheDarkThresholdHasItsHue(t *testing.T) {
	dark := pickedNumbers(t).Dark
	if hue := dark.Green.HSV().H; hue != dark.GreenHue {
		t.Errorf("the hue of a green lamp at the threshold is %v, not %v", hue, dark.GreenHue)
	}
}

// RGBA gives image/color 16-bit channels and holds the levels in range.
func TestRGBAHoldsTheLevelsInRange(t *testing.T) {
	rgba := pickedNumbers(t).RGBA
	red, green, blue, alpha := rgba.Levels.RGBA()
	want := rgba.Want
	if red != want.Red || green != want.Green || blue != want.Blue || alpha != want.Alpha {
		t.Errorf("RGBA is %d %d %d %d, not %v", red, green, blue, alpha, want)
	}
}

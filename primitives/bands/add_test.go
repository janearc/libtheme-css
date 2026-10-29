package bands_test

import (
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/libtheme-css/primitives/swatch"
)

// The numbers the tests pick. Each is defined by this test.
const (
	// coolest is the lowest temperature of a drawn lamp, in kelvin. It is
	// defined by this test.
	coolest = 1500.0
	// hottest is the highest temperature of a drawn lamp, in kelvin. It is
	// defined by this test.
	hottest = 9000.0
	// weakest is the least a drawn lamp is scaled by. It is defined by this
	// test.
	weakest = 0.1
	// strongest is the most a drawn lamp is scaled by. It is defined by this
	// test.
	strongest = 10.0
	// slide is how many bands a drawn lamp's light is slid, at most. It is
	// defined by this test: enough to make the spans differ.
	slide = 30
)

// lamp is a drawn light: a black body of some temperature, made some strength
// and slid some bands, so two of them have spans that differ.
func lamp(t *rapid.T, name string) bands.Light {
	kelvin := rapid.Float64Range(coolest, hottest).Draw(t, name+" kelvin")
	strength := rapid.Float64Range(weakest, strongest).Draw(t, name+" strength")
	steps := rapid.IntRange(-slide, slide).Draw(t, name+" slide")
	return bands.Split(swatch.Blackbody(kelvin)).Scale(strength).Shift(steps)
}

// powers is every band's power in a light across a span.
func powers(light bands.Light, first, last bands.Band) []float64 {
	var out []float64
	for band := first; band <= last; band++ {
		out = append(out, light.Power(band))
	}
	return out
}

// sameLight is whether two lights have the same power in every band either
// has an entry for.
func sameLight(left, right bands.Light) bool {
	leftFirst, leftLast := left.Span()
	rightFirst, rightLast := right.Span()
	first, last := min(leftFirst, rightFirst), max(leftLast, rightLast)
	got, want := powers(left, first, last), powers(right, first, last)
	for place := range got {
		if got[place] != want[place] {
			return false
		}
	}
	return true
}

// The property: two lights on one place are, in every band, the sum of the
// two; it does not matter which is named first; and light added to no light
// is the light, from either side.
func TestAddingLightSumsEveryBand(t *testing.T) {
	rapid.Check(t, addsEveryBand)
}

// addsEveryBand is one draw of the property above.
func addsEveryBand(t *rapid.T) {
	left, right := lamp(t, "left"), lamp(t, "right")
	sum := bands.Add(left, right)
	if !sameLight(sum, bands.Add(right, left)) {
		t.Fatalf("adding in the other order gave another light")
	}
	leftFirst, leftLast := left.Span()
	rightFirst, rightLast := right.Span()
	for band := min(leftFirst, rightFirst); band <= max(leftLast, rightLast); band++ {
		if sum.Power(band) != left.Power(band)+right.Power(band) {
			t.Fatalf("band %d: %v, not %v and %v", band, sum.Power(band),
				left.Power(band), right.Power(band))
		}
	}
	none := bands.Light{}
	if !sameLight(bands.Add(none, left), left) || !sameLight(bands.Add(left, none), left) {
		t.Fatalf("no light changed a light")
	}
}

// The anti-test: no light and no light is no light. It does not become a
// colour, and it has no span.
func TestNoLightAndNoLightIsNoLight(t *testing.T) {
	both := bands.Add(bands.Light{}, bands.Light{})
	if first, last := both.Span(); first <= last {
		t.Errorf("no light and no light spans %d to %d", first, last)
	}
	if both.Swatch() != swatch.NoLight() {
		t.Errorf("no light and no light is %v", both.Swatch())
	}
}

// The anti-test: adding light to another does not change the first, whose
// powers are its own: Add copies, it does not share.
func TestAddingDoesNotChangeEitherLight(t *testing.T) {
	rapid.Check(t, leavesBothAlone)
}

// leavesBothAlone is one draw of the property above.
func leavesBothAlone(t *rapid.T) {
	left, right := lamp(t, "left"), lamp(t, "right")
	first, last := left.Span()
	before := powers(left, first, last)
	bands.Add(left, right)
	bands.Add(bands.Light{}, left).Scale(strongest)
	bands.Add(left, bands.Light{}).Scale(strongest)
	after := powers(left, first, last)
	for place := range before {
		if before[place] != after[place] {
			t.Fatalf("band %d of the light changed", first+bands.Band(place))
		}
	}
}

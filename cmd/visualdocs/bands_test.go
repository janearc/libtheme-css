package main

import (
	"math"
	"testing"

	"pgregory.net/rapid"

	"github.com/janearc/libtheme-css/primitives/swatch"
)

const (
	// dimmest is the least brightness the property puts a light at. It is
	// defined by this test.
	dimmest = 0.01
	// brightest is the most brightness the property puts a light at. It is
	// defined by this test.
	brightest = 1.0
	// violet is the shortest wavelength, in nanometres, the property draws.
	// It is defined by this test.
	violet = 400
	// red is the longest wavelength, in nanometres, the property draws. It
	// is defined by this test.
	red = 700
)

// tiny is how far two floats may differ and be the same here. It is defined
// by this test.
const tiny = 1e-9

// A light with no brightness stays as it is. The anti-test is the property
// below: a light with brightness is put at the one asked for.
func TestNoLightHasNoBrightnessToChange(t *testing.T) {
	dark := swatch.NoLight()
	if got := at(dark, brightest); got != dark {
		t.Errorf("a dark light became %v", got)
	}
}

// The property: a light with brightness, put at a brightness, has that
// brightness, and keeps the proportions of its colour.
func TestAtPutsALightAtABrightness(t *testing.T) {
	rapid.Check(t, putsAtBrightness)
}

// putsAtBrightness is one draw of the property above.
func putsAtBrightness(draw *rapid.T) {
	nm := rapid.IntRange(violet, red).Draw(draw, "nm")
	y := rapid.Float64Range(dimmest, brightest).Draw(draw, "y")
	light := swatch.Monochrome(nm)
	x0, y0, z0 := light.XYZ()
	x, got, z := at(light, y).XYZ()
	if math.Abs(got-y) > tiny || math.Abs(x/got-x0/y0) > tiny ||
		math.Abs(z/got-z0/y0) > tiny {
		draw.Fatalf("%d nm at %g is %g %g %g", nm, y, x, got, z)
	}
}

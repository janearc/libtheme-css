package radiation_test

import (
	"math"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// Numbers the tests use that the arithmetic or the scale defines.
const (
	// none is no wavelength. It is defined by the scale.
	none = 0
	// unitLuminance is a luminance of one. It is defined by the scale.
	unitLuminance = 1
)

// White is no colour's wavelength. It has no direction from itself.
func TestOfWhiteIsNothing(t *testing.T) {
	nm, found := radiation.Seen().Of(swatch.WhitePoint())
	if nm != none || found {
		t.Errorf("white is %v nm, %v", nm, found)
	}
}

// Purples are no single wavelength's colour. They answer with nothing.
func TestOfPurpleIsNothing(t *testing.T) {
	magenta := srgb.MustHex(pickedNumbers(t).Of.Magenta).Swatch()
	nm, found := radiation.Seen().Of(magenta)
	if nm != none || found {
		t.Errorf("magenta is %v nm, %v", nm, found)
	}
}

// Each wavelength's own colour is named by that wavelength to the nanometre.
func TestOfAWavelengthIsItself(t *testing.T) {
	of := pickedNumbers(t).Of
	for _, want := range of.Wavelengths {
		nm, found := radiation.Seen().Of(swatch.Monochrome(want))
		if !found || math.Abs(nm-float64(want)) > of.Within {
			t.Errorf("%d nm is named %v, %v", want, nm, found)
		}
	}
}

// Every row of the observer's table names a wavelength.
// The ends of the sweep are rows. They are in it.
func TestOfNamesEveryRowBeforeTheLast(t *testing.T) {
	visible := radiation.Visible
	for nm := int(visible.Lo); float64(nm) < visible.Hi; nm++ {
		if _, found := radiation.Seen().Of(swatch.Monochrome(nm)); !found {
			t.Fatalf("the colour of %d nm names no wavelength", nm)
		}
	}
}

// The last row is not in the range. Its colour is named by a wavelength before it.
func TestOfTheLastRowIsNotInTheRange(t *testing.T) {
	of := pickedNumbers(t).Of
	last := int(radiation.Visible.Hi)
	nm, found := radiation.Seen().Of(swatch.Monochrome(last))
	if !found || nm >= radiation.Visible.Hi-of.Half {
		t.Errorf("the colour of %d nm is named %v, %v", last, nm, found)
	}
}

// midway is a colour half way between two wavelengths' colours on the chromaticity diagram.
func midway(low, high int) swatch.Swatch {
	first, second := swatch.Monochrome(low).XY(), swatch.Monochrome(high).XY()
	var place swatch.XY
	place.X = (first.X + second.X) / halving
	place.Y = (first.Y + second.Y) / halving
	return swatch.FromXY(place, unitLuminance)
}

// Colours between two wavelengths are named by one of the two. This fails
// if the angle between directions is not worked out as an angle.
func TestOfAColourBetweenTwoWavelengthsIsNamedByOne(t *testing.T) {
	of := pickedNumbers(t).Of
	nm, found := radiation.Seen().Of(midway(of.Between.Low, of.Between.High))
	middle := float64(of.Between.Low+of.Between.High) / halving
	if !found || math.Abs(nm-middle) > of.Within {
		t.Errorf("the colour between %d and %d nm is named %v, %v", of.Between.Low, of.Between.High, nm, found)
	}
}

// Ranges that start before the observer's table use only its rows.
// The wavelengths with no row are not directions.
func TestOfSkipsWavelengthsTheObserverHasNoRowFor(t *testing.T) {
	numbers := pickedNumbers(t)
	short := numbers.Profile.Short
	nm, found := short.Of(swatch.Monochrome(numbers.Of.DeepViolet))
	if !found || math.Abs(nm-float64(numbers.Of.DeepViolet)) > numbers.Of.Within {
		t.Errorf("violet in %s is named %v, %v", short.Name, nm, found)
	}
}

// offTheSweep is a colour in a direction from white that no violet or blue is in.
func offTheSweep(degrees, radius float64, halfTurn float64) swatch.Swatch {
	white := swatch.WhitePoint().XY()
	angle := degrees * math.Pi / halfTurn
	var place swatch.XY
	place.X = white.X + radius*math.Cos(angle)
	place.Y = white.Y + radius*math.Sin(angle)
	return swatch.FromXY(place, unitLuminance)
}

// Directions the rows of a short range do not point are no wavelength's.
func TestOfRefusesADirectionOffTheSweep(t *testing.T) {
	of := pickedNumbers(t).Of
	short := pickedNumbers(t).Profile.Short
	colour := offTheSweep(of.Off.Degrees, of.Off.Radius, of.HalfTurn)
	if nm, found := short.Of(colour); found {
		t.Errorf("a direction off the sweep is named %v nm", nm)
	}
}

// The top of the render range is not in it. The row at the top is no
// direction of the sweep, so its colour names nothing. The row below the
// top is the last of the sweep, and its colour is named.
func TestOfTheRowAtTheTopOfTheRenderIsNotInTheSweep(t *testing.T) {
	numbers := pickedNumbers(t)
	sound := numbers.Profile.Sound
	top := int(sound.Render.Hi)
	nm, found := sound.Of(swatch.Monochrome(top))
	if nm != none || found {
		t.Errorf("the row at the top, %d, is named %v, %v", top, nm, found)
	}
	last := top - numbers.Of.OneRow
	nm, found = sound.Of(swatch.Monochrome(last))
	if !found || sound.Swatch(nm) != swatch.Monochrome(last) {
		t.Errorf("the row %d is named %v, %v", last, nm, found)
	}
}

// Far in the red the rows point the same way from white, to the last bit.
// They tie. The shortest names them all. The colour of the longer row is
// named by the shorter one to the last digits.
func TestOfATieInTheFarRedIsTheShortestRow(t *testing.T) {
	red := pickedNumbers(t).Of.FarRed
	for _, row := range []int{red.Short, red.Long} {
		nm, found := radiation.Seen().Of(swatch.Monochrome(row))
		if !found || math.Abs(nm-float64(red.Short)) > red.Within {
			t.Errorf("the colour of %d nm is named %v, %v, not %d",
				row, nm, found, red.Short)
		}
	}
}

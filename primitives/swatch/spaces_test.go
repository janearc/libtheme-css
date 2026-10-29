package swatch_test

import (
	"math"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
)

// Monochrome is the table's row. The figures are in the JSON with their source.
func TestMonochromeIsTheTablesRow(t *testing.T) {
	rows := pickedNumbers(t).Rows
	for _, want := range rows.At {
		x, y, z := swatch.Monochrome(want.NM).XYZ()
		if !near(x, want.Xbar, rows.Tolerance) || !near(y, want.Ybar, rows.Tolerance) ||
			!near(z, want.Zbar, rows.Tolerance) {
			t.Errorf("%d nm is %v %v %v, not %v %v %v", want.NM, x, y, z, want.Xbar, want.Ybar, want.Zbar)
		}
	}
}

// Outside the table there is no light.
func TestMonochromeOutsideTheTableIsNoLight(t *testing.T) {
	lo, hi := swatch.Observed()
	below := swatch.Monochrome(lo - whole)
	above := swatch.Monochrome(hi + whole)
	if below != swatch.NoLight() || above != swatch.NoLight() {
		t.Errorf("light outside %d to %d", lo, hi)
	}
}

// Blackbody is Planck's law without its leading 2hc squared.
// Put back, it is the spectral radiance. The test works it with expm1.
func TestBlackbodyIsPlancksLaw(t *testing.T) {
	planck := pickedNumbers(t).Planck
	wavelength := float64(planck.NM) * nanometre
	energy := planck.Planck * planck.Light
	exponent := energy / (wavelength * planck.Boltzmann * planck.Kelvin)
	want := twice * energy * planck.Light / (math.Pow(wavelength, fifth) * math.Expm1(exponent))
	relative := swatch.Blackbody(planck.Kelvin)[planck.NM]
	got := relative * twice * energy * planck.Light
	if !near(got, want, planck.Tolerance) {
		t.Errorf("radiance at %d nm is %v, Planck's law gives %v", planck.NM, got, want)
	}
}

// The brightest whole nanometre of a black body is Wien's, to half a nanometre.
func TestBlackbodyPeaksWhereWienSays(t *testing.T) {
	wien := pickedNumbers(t).Wien
	best, peak := none, float64(none)
	for nm, power := range swatch.Blackbody(wien.Kelvin) {
		if power > peak {
			best, peak = nm, power
		}
	}
	want := wien.Constant / wien.Kelvin / nanometre
	if gap := math.Abs(float64(best) - want); gap > wien.SlackNM {
		t.Errorf("the peak is at %d nm, Wien's law says %.2f", best, want)
	}
}

// At one kelvin the spectrum has a row at every wavelength. All of them are nothing.
func TestBlackbodyAtOneKelvinHasEveryRow(t *testing.T) {
	one := pickedNumbers(t).Kelvins.One
	lo, hi := swatch.Observed()
	if got := len(swatch.Blackbody(one)); got != hi-lo+whole {
		t.Errorf("one kelvin has %d rows, the observer has %d", got, hi-lo+whole)
	}
}

// Zero kelvin and below is empty.
func TestBlackbodyAtZeroKelvinIsEmpty(t *testing.T) {
	kelvins := pickedNumbers(t).Kelvins
	zero := swatch.Blackbody(kelvins.Zero)
	below := swatch.Blackbody(kelvins.BelowZero)
	if len(zero) != none || len(below) != none {
		t.Errorf("zero kelvin has %d rows and below it %d", len(zero), len(below))
	}
}

// Chromaticity at a luminance other than one gives X = x/y Y and Z = (1-x-y)/y Y.
func TestFromXYScalesWithLuminance(t *testing.T) {
	d65 := pickedNumbers(t).D65
	chroma, luminance := d65.XY, d65.Luminance
	x, y, z := swatch.FromXY(chroma, luminance).XYZ()
	wantX := chroma.X / chroma.Y * luminance
	wantZ := (whole - chroma.X - chroma.Y) / chroma.Y * luminance
	if !near(x, wantX, d65.Tolerance) || y != luminance || !near(z, wantZ, d65.Tolerance) {
		t.Errorf("got %v %v %v, want %v %v %v", x, y, z, wantX, luminance, wantZ)
	}
}

// At a luminance of one the chromaticity is the published X and Z.
func TestFromXYAtOneIsThePublishedWhite(t *testing.T) {
	d65 := pickedNumbers(t).D65
	x, _, z := swatch.FromXY(d65.XY, whole).XYZ()
	if math.Abs(x-d65.XAtOne) > d65.ToleranceOne || math.Abs(z-d65.ZAtOne) > d65.ToleranceOne {
		t.Errorf("at Y of one: X %v, Z %v, not %v and %v", x, z, d65.XAtOne, d65.ZAtOne)
	}
}

// Swatches dimmer than one have their own chromaticity, not the white's.
func TestDimSwatchHasItsOwnChromaticity(t *testing.T) {
	dim := pickedNumbers(t).Dim
	got := swatch.FromXYZ(dim.XYZ.CieX, dim.XYZ.CieY, dim.XYZ.CieZ).XY()
	if !near(got.X, dim.Chroma.X, dim.Tolerance) || !near(got.Y, dim.Chroma.Y, dim.Tolerance) {
		t.Errorf("chromaticity %v, not %v", got, dim.Chroma)
	}
}

// No light reports the white's chromaticity.
func TestNoLightReportsTheWhite(t *testing.T) {
	got := swatch.NoLight().XY()
	if got != swatch.WhitePoint().XY() {
		t.Errorf("no light is at %v, not the white's", got)
	}
}

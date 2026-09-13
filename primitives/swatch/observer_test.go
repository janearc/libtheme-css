package swatch

import (
	"math"
	"testing"
)

// The derived white lands on the published numbers to five places. The
// published values come from these same tables at 1 nm, which is why the
// tolerance can be this tight; a 5 nm table would miss Z by 1e-4.
func TestWhiteIsDerived(t *testing.T) {
	x, y, z := White.XYZ()
	if math.Abs(x-0.95047) > 5e-5 || y != 1 || math.Abs(z-1.08883) > 5e-5 {
		t.Errorf("white = %.5f %.5f %.5f, want 0.95047 1 1.08883", x, y, z)
	}
}

// A flat spectrum, the same power at every wavelength, is illuminant E,
// and the 1931 observer was built so that it comes out X = Y = Z.
func TestEqualEnergyIsNeutral(t *testing.T) {
	flat := map[int]float64{}
	for nm := 360; nm <= 830; nm++ {
		flat[nm] = 1
	}
	x, _, z := Illuminant(flat).XYZ()
	if math.Abs(x-1) > 2e-3 || math.Abs(z-1) > 2e-3 {
		t.Errorf("equal energy = %.4f 1 %.4f, want about 1 1 1", x, z)
	}
}

// No light is black, not a division by zero.
func TestNoLightIsBlack(t *testing.T) {
	if Illuminant(map[int]float64{}) != Black {
		t.Errorf("an empty spectrum is not black")
	}
}

// Light the observer cannot see contributes nothing. This is a sanity
// check on the range being applied, not a belief that anyone will put
// infrared in a zsh theme: if a future table or a parsing slip let power
// outside the visible range leak into X, Y or Z, every white and every
// swatch downstream would drift, quietly. So: a spectrum entirely in the
// near infrared (900 to 1000 nm) and one entirely in the ultraviolet (200
// to 300 nm) must both come out as no light at all, and adding either to
// daylight must not move the white.
func TestInvisibleLightIsNoLight(t *testing.T) {
	infrared, ultraviolet := map[int]float64{}, map[int]float64{}
	for nm := 900; nm <= 1000; nm++ {
		infrared[nm] = 100
	}
	for nm := 200; nm <= 300; nm++ {
		ultraviolet[nm] = 100
	}
	if Illuminant(infrared) != Black {
		t.Errorf("infrared came out as %v, not black", Illuminant(infrared))
	}
	if Illuminant(ultraviolet) != Black {
		t.Errorf("ultraviolet came out as %v, not black", Illuminant(ultraviolet))
	}
	lit := d65()
	for nm, p := range infrared {
		lit[nm] = p
	}
	for nm, p := range ultraviolet {
		lit[nm] = p
	}
	if Illuminant(lit) != White {
		t.Errorf("daylight plus invisible light moved the white to %v", Illuminant(lit))
	}
}

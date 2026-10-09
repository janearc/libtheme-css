package bands

import (
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/ok"
)

// relative is a swatch put at Y of 1, so two lights of different strength
// can be compared by colour alone.
func relative(s swatch.Swatch) swatch.Swatch {
	x, y, z := s.XYZ()
	if y == 0 {
		return s
	}
	return swatch.FromXYZ(x/y, 1, z/y)
}

func apart(a, b swatch.Swatch) float64 {
	return ok.Distance(ok.FromSwatch(a), ok.FromSwatch(b))
}

// Every PerOctave bands the wavelength doubles, and a band's end is the
// next one's start.
func TestAnOctaveIsPerOctaveBands(t *testing.T) {
	for _, nm := range []float64{1, 380, 555, 780, 28000} {
		if Of(2*nm) != Of(nm)+PerOctave {
			t.Errorf("%v nm is band %d and %v nm is band %d",
				nm, Of(nm), 2*nm, Of(2*nm))
		}
		b := Of(nm)
		lo, hi := b.Edges()
		if nm < lo || nm >= hi {
			t.Errorf("%v nm is in band %d, which is %v to %v",
				nm, b, lo, hi)
		}
		if next, _ := (b + 1).Edges(); next != hi {
			t.Errorf("band %d ends at %v, band %d starts at %v",
				b, hi, b+1, next)
		}
	}
}

// Daylight, into bands and back, is the white, well inside what an eye can
// see, and at the strength its table gives it has Y of 1.
func TestDaylightComesBackWhite(t *testing.T) {
	got := Split(swatch.D65()).Swatch()
	if d := apart(got, swatch.White); d > ok.Eye/10 {
		t.Errorf("d65 through bands is %.5f from white", d)
	}
	if _, y, _ := got.XYZ(); y < 0.999 || y > 1.001 {
		t.Errorf("d65 through bands has Y of %v", y)
	}
}

// Every black body from 1000 to 20000 kelvin, into bands and back, is the
// colour Planckian gives it, well inside what an eye can see. This is the
// check that bands, Planck's law and the observer agree.
func TestBlackBodiesComeBack(t *testing.T) {
	for k := 1000.0; k <= 20000; k += 250 {
		got := relative(Split(swatch.Blackbody(k)).Swatch())
		if d := apart(got, swatch.Planckian(k)); d > ok.Eye/10 {
			t.Errorf("%v K through bands is %.5f from its "+
				"colour", k, d)
		}
	}
}

// Redshift by one step moves every band's light exactly one band redder,
// and the colour moves with it.
func TestShiftByOneMovesEveryBandOne(t *testing.T) {
	l := Split(swatch.D65())
	s := l.Shift(1)
	first, last := l.Span()
	if f, g := s.Span(); f != first+1 || g != last+1 {
		t.Errorf("span %d to %d shifted to %d to %d", first, last,
			f, g)
	}
	for b := first; b <= last; b++ {
		if l.Power(b) != s.Power(b+1) {
			t.Errorf("band %d had %v, band %d has %v", b,
				l.Power(b), b+1, s.Power(b+1))
		}
	}
	if relative(s.Swatch()) == relative(l.Swatch()) {
		t.Error("shifted daylight is the same colour as daylight")
	}
}

// Light adds: two lights on one place are, in every band and in colour,
// the sum of the two. This is what smoke two's red and green lamps rely on.
func TestLightAdds(t *testing.T) {
	blue := Split(map[int]float64{450: 1, 460: 2})
	red := Split(map[int]float64{640: 3, 700: 1})
	both := Add(blue, red)
	for _, b := range []Band{Of(450), Of(460), Of(640), Of(700)} {
		if both.Power(b) != blue.Power(b)+red.Power(b) {
			t.Errorf("band %d: %v is not %v plus %v", b,
				both.Power(b), blue.Power(b), red.Power(b))
		}
	}
	bx, by, bz := blue.Swatch().XYZ()
	rx, ry, rz := red.Swatch().XYZ()
	sum := swatch.FromXYZ(bx+rx, by+ry, bz+rz)
	if d := apart(both.Swatch(), sum); d > ok.Exact {
		t.Errorf("the two lights together are %v from their sum", d)
	}
}

// Twice the light is twice as bright and the same colour.
func TestTwiceTheLight(t *testing.T) {
	l := Split(swatch.D65())
	_, one, _ := l.Swatch().XYZ()
	_, two, _ := l.Scale(2).Swatch().XYZ()
	if d := two - 2*one; d > 1e-12 || d < -1e-12 {
		t.Errorf("doubled daylight has Y %v, once it was %v", two, one)
	}
}

// Light no eye can see is still light: it has power in its bands and no
// colour at all.
func TestInfraredIsLightWithNoColour(t *testing.T) {
	l := Split(map[int]float64{1000: 1, 5000: 1, 28000: 1})
	if l.Power(Of(5000)) != 1 {
		t.Errorf("5000 nm carries %v", l.Power(Of(5000)))
	}
	if l.Swatch() != swatch.Black {
		t.Errorf("infrared has a colour: %v", l.Swatch())
	}
}

// No light is no light, whichever way it is asked for.
func TestNoLight(t *testing.T) {
	for _, l := range []Light{{}, Split(nil), Split(map[int]float64{})} {
		if l.Swatch() != swatch.Black {
			t.Errorf("no light has a colour: %v", l.Swatch())
		}
		if f, g := l.Span(); f <= g {
			t.Errorf("no light spans %d to %d", f, g)
		}
	}
	if got := Add(Light{}, Split(swatch.D65())).Swatch(); got !=
		Split(swatch.D65()).Swatch() {
		t.Errorf("no light plus daylight is %v", got)
	}
}

// The same light is the same number, every time it is asked for, the rule
// the swatch learned in 2026.
func TestTheSameLightIsTheSameNumber(t *testing.T) {
	want := Split(swatch.D65()).Swatch()
	for range 200 {
		if got := Split(swatch.D65()).Swatch(); got != want {
			t.Fatalf("daylight came back as %v, then %v", want, got)
		}
	}
}

// Lines is Split for wavelengths that are not whole nanometres: given whole
// ones, it is Split exactly.
func TestLinesAgreesWithSplit(t *testing.T) {
	d65 := swatch.D65()
	lines := map[float64]float64{}
	for nm, p := range d65 {
		lines[float64(nm)] = p
	}
	a, b := Split(d65), Lines(lines)
	af, al := a.Span()
	bf, bl := b.Span()
	if af != bf || al != bl {
		t.Fatalf("split spans %d to %d, lines %d to %d", af, al, bf, bl)
	}
	for band := af; band <= al; band++ {
		if a.Power(band) != b.Power(band) {
			t.Errorf("band %d: split %v, lines %v", band,
				a.Power(band), b.Power(band))
		}
	}
}

// Lines carries radiation far shorter than a nanometre: a gamma ray at a
// ten-thousandth of one is in its own band, with all its power.
func TestLinesCarriesGamma(t *testing.T) {
	l := Lines(map[float64]float64{1e-4: 2})
	if f, g := l.Span(); f != Of(1e-4) || g != Of(1e-4) {
		t.Errorf("a gamma line spans %d to %d", f, g)
	}
	if l.Power(Of(1e-4)) != 2 || l.Swatch() != swatch.Black {
		t.Errorf("a gamma line: %v in its band, colour %v",
			l.Power(Of(1e-4)), l.Swatch())
	}
}

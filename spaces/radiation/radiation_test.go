package radiation

import (
	"math"
	"testing"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/libtheme-css/primitives/swatch"
)

// near is two numbers the same to a part in a billion.
func near(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9*math.Max(math.Abs(a), math.Abs(b))
}

// mid is the middle of a range the way the ranges are spread: evenly in the
// logarithm of wavelength.
func mid(r Range) float64 { return math.Sqrt(r.Lo * r.Hi) }

// A photon's energy times its wavelength is 1239.84 electronvolt
// nanometres, and the conversions undo each other.
func TestPhotonUnits(t *testing.T) {
	if math.Abs(hc-1239.84198) > 1e-4 {
		t.Errorf("hc is %v eV nm", hc)
	}
	if nm := FromEnergy(1 * KeV); math.Abs(nm-1.23984) > 1e-4 {
		t.Errorf("a 1 keV photon is %v nm", nm)
	}
	for _, nm := range []float64{1e-5, 0.5, 550, 1e6} {
		if !near(FromEnergy(Energy(nm)), nm) {
			t.Errorf("%v nm through energy came back %v", nm,
				FromEnergy(Energy(nm)))
		}
		if !near(HertzToNm(NmToHertz(nm)), nm) {
			t.Errorf("%v nm through hertz came back %v", nm,
				HertzToNm(NmToHertz(nm)))
		}
	}
}

// The hydrogen line, at 1420.405751 megahertz, is the 21 centimetre line.
func TestTheHydrogenLine(t *testing.T) {
	cm := HertzToNm(1420.405751*MHz) / 1e7
	if math.Abs(cm-21.106) > 1e-3 {
		t.Errorf("the hydrogen line is %v cm", cm)
	}
	if !Radio.holds(cm * 1e7) {
		t.Error("radio does not hold the hydrogen line")
	}
}

// Fermi is given in energies, highest first, and comes out a range of
// wavelengths shortest first.
func TestFermiIsARange(t *testing.T) {
	if !(Fermi.Source.Lo < Fermi.Source.Hi) {
		t.Errorf("fermi runs %v to %v nm", Fermi.Source.Lo,
			Fermi.Source.Hi)
	}
	if !near(Energy(Fermi.Source.Lo), 300*GeV) {
		t.Errorf("fermi's short end is %v eV", Energy(Fermi.Source.Lo))
	}
}

// The profiles are listed shortest wavelength first, as their comment says.
func TestProfilesAreInOrder(t *testing.T) {
	for i := 1; i < len(Profiles); i++ {
		a, b := Profiles[i-1], Profiles[i]
		if a.Name != Fermi.Name && b.Name != Fermi.Name &&
			a.Source.Lo > b.Source.Lo {
			t.Errorf("%s starts after %s", a.Name, b.Name)
		}
	}
}

// A profile holds its own range, from its short end up to but not
// including its long end, and nothing outside it.
func TestAProfileHoldsItsOwnRangeOnly(t *testing.T) {
	for _, p := range Profiles {
		s := p.Source
		for _, nm := range []float64{s.Lo, mid(s), s.Hi * 0.999} {
			if !p.holds(nm) {
				t.Errorf("%s does not hold %g nm", p.Name, nm)
			}
		}
		for _, nm := range []float64{s.Lo * 0.999, s.Hi, s.Hi * 2} {
			if p.holds(nm) {
				t.Errorf("%s holds %g nm, outside it",
					p.Name, nm)
			}
		}
	}
}

// Every profile shows its whole range: nothing inside it comes out black.
func TestEveryProfileShowsItsRange(t *testing.T) {
	for _, p := range Profiles {
		if p.Swatch(mid(p.Source)) == swatch.Black {
			t.Errorf("%s shows the middle of its range as black",
				p.Name)
		}
	}
}

// Outside its range a profile shows black, rather than deciding for anyone
// what radiation it did not ask about looks like.
func TestOutsideIsBlack(t *testing.T) {
	for _, p := range Profiles {
		if p.Swatch(p.Source.Hi*2) != swatch.Black ||
			p.Swatch(p.Source.Lo/2) != swatch.Black {
			t.Errorf("%s shows radiation outside its range", p.Name)
		}
	}
}

// The ends of a profile's range land on the ends of what it is shown as,
// and onto and back undo each other everywhere between.
func TestOntoAndBack(t *testing.T) {
	for _, p := range Profiles {
		if !near(p.onto(p.Source.Lo), p.Render.Lo) ||
			!near(p.onto(p.Source.Hi), p.Render.Hi) {
			t.Errorf("%s's ends land at %v and %v", p.Name,
				p.onto(p.Source.Lo), p.onto(p.Source.Hi))
		}
		for _, f := range []float64{0.1, 0.5, 0.9} {
			nm := p.Source.Lo * math.Pow(p.Source.Hi/p.Source.Lo, f)
			if !near(p.back(p.onto(nm)), nm) {
				t.Errorf("%s: %g nm came back %g", p.Name, nm,
					p.back(p.onto(nm)))
			}
		}
	}
}

// A colour a profile shows names the wavelength it came from, to within
// the one nanometre of visible light Swatch rounds to.
//
// Only up to about 690 nm of what it is shown as: past that, every
// wavelength is the same red to the observer, so a colour there cannot
// say which one it was, and Of answers with the first.
func TestAColourNamesItsWavelength(t *testing.T) {
	for _, p := range Profiles {
		for shown := 420.0; shown <= 690; shown += 30 {
			nm := p.back(shown)
			got, ok := p.Of(p.Swatch(nm))
			if !ok {
				t.Errorf("%s: the colour of %g nm names "+
					"nothing", p.Name, nm)
				continue
			}
			if d := math.Abs(p.onto(got) - shown); d > 1.5 {
				t.Errorf("%s: %g nm, shown at %v, named %g, "+
					"shown at %v", p.Name, nm, shown, got,
					p.onto(got))
			}
		}
	}
}

// White, and a purple no single wavelength makes, name no wavelength.
func TestColoursNoWavelengthMakes(t *testing.T) {
	magenta := swatch.FromXYZ(0.59, 0.28, 0.97)
	for _, s := range []swatch.Swatch{swatch.White, magenta} {
		if nm, ok := Infrared.Of(s); ok {
			t.Errorf("%v named %g nm", s, nm)
		}
	}
}

// A single line of radiation, carried in bands and shown by a profile,
// comes out a colour in the same visible band as the line's own shown
// wavelength, or the one beside it.
//
// That is as fine as Light can be: it carries what it shows in bands too,
// 24 an octave, about 16 nm wide at 547, so a colour it makes is known to a
// band of visible light and no finer, though the mapping itself is good to
// a nanometre.
func TestALineInBandsKeepsItsVisibleBand(t *testing.T) {
	for _, p := range Profiles {
		nm := mid(p.Source)
		got, named := p.Of(p.Light(bands.Lines(
			map[float64]float64{nm: 1})))
		if !named {
			t.Errorf("%s: a line at %g nm names nothing",
				p.Name, nm)
			continue
		}
		want, at := bands.Of(p.onto(nm)), bands.Of(p.onto(got))
		if at < want-1 || at > want+1 {
			t.Errorf("%s: a line at %g nm, shown in band %d, "+
				"named %g, in band %d", p.Name, nm, want, got,
				at)
		}
	}
}

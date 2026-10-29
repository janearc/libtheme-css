package si

import (
	"math"
	"testing"
)

// prefixed is a prefix function and the power of ten the SI Brochure's
// table 7 gives it.
type prefixed struct {
	name  string
	apply func(float64) float64
	power int
}

var prefixes = []prefixed{
	{"femto", Femto, -15}, {"pico", Pico, -12}, {"nano", Nano, -9},
	{"micro", Micro, -6}, {"milli", Milli, -3}, {"centi", Centi, -2},
	{"kilo", Kilo, 3}, {"mega", Mega, 6}, {"giga", Giga, 9},
	{"tera", Tera, 12}, {"peta", Peta, 15}, {"exa", Exa, 18},
}

// Each prefix of one unit is its power of ten.
func TestPrefixesAreTheirPowersOfTen(t *testing.T) {
	for _, p := range prefixes {
		if got, want := p.apply(1), math.Pow10(p.power); got != want {
			t.Errorf("%s of one unit is %v, not %v", p.name, got, want)
		}
	}
}

// A prefix scales the unit it is given, not one: three kilo-somethings
// are three thousand somethings.
func TestAPrefixScalesItsUnit(t *testing.T) {
	if got := Kilo(3); got != 3000 {
		t.Errorf("kilo of 3 is %v", got)
	}
	if got := Milli(Kilo(7)); math.Abs(got-7) > 1e-12 {
		t.Errorf("a milli of a kilo of 7 is %v", got)
	}
}

// The anti-test: no two prefixes are the same, so a slip that gave two
// names one power shows.
func TestNoTwoPrefixesAreTheSame(t *testing.T) {
	seen := map[float64]string{}
	for _, p := range prefixes {
		if other, ok := seen[p.apply(1)]; ok {
			t.Errorf("%s and %s are the same", p.name, other)
		}
		seen[p.apply(1)] = p.name
	}
}

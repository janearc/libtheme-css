package main

import "testing"

// The anti-test: a colour with a number that is not finite is refused by
// name, in any of the places a number can go, rather than reported in
// gamut.
func TestParseRefusesWhatIsNotFinite(t *testing.T) {
	for _, bad := range []string{"NaN", "nan", "Inf", "+Inf", "-Inf", "Infinity", "1e999"} {
		for _, spelling := range []string{"oklch(" + bad + " 0.2 30)",
			"oklch(0.5 " + bad + " 30)", "oklch(0.5 0.2 " + bad + ")",
			"oklch(" + bad + "% 0.2 30)"} {
			if _, err := parse(spelling); err == nil {
				t.Errorf("%s was accepted", spelling)
			}
		}
	}
}

// Ordinary colours still parse: a percent, a plain number, a hue of none,
// and a hex code.
func TestParseStillTakesOrdinaryColours(t *testing.T) {
	for _, good := range []string{"oklch(74% 0.2 345)", "oklch(0.74 0.2 345.5)",
		"oklch(50% 0 none)", "#ff6ec7", "#fff"} {
		if _, err := parse(good); err != nil {
			t.Errorf("%s was refused: %v", good, err)
		}
	}
}

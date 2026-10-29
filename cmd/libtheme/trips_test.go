package main

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/janearc/libtheme-css/primitives/swatch"
	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/libtheme-css/spaces/srgb"
)

// onlyOne is how many profiles make a bar of their own. It is defined by this
// test.
const onlyOne = 1

// tripping is a round trip of some colours and some radiation profiles.
type tripping struct {
	list     []labelled
	profiles []radiation.Profile
}

// run takes the trip.
func (trip tripping) run() error { return roundtripOf(trip.list, trip.profiles) }

// barOf is the radiation bars of some profiles.
type barOf []radiation.Profile

// run draws the bars.
func (profiles barOf) run() error {
	roundtripRadiation(profiles)
	return nil
}

// stepsOf is a call of ramp with its arguments.
type stepsOf []string

// run makes the call.
func (arguments stepsOf) run() error { return rampOf(arguments) }

// A colour the oklch spelling cannot carry is lost to the round trip, and the
// verb says so and fails.
func TestRoundtripFailsWhenAColourIsLost(t *testing.T) {
	lost, _ := parse(outOfGamut)
	out, err := printing(t, tripping{[]labelled{{"lost", lost}}, nil}.run)
	if err == nil || !strings.Contains(out, "different to a person") {
		t.Errorf("a lost colour: %q, %v", out, err)
	}
}

// The anti-test: a colour that comes back is not a failure. And a colour that
// cannot be written as oklch at all is, by the reason the parser gives.
func TestRoundtripKeepsWhatComesBack(t *testing.T) {
	kept := []labelled{{"red", srgb.Red.Swatch()}}
	out, err := printing(t, tripping{kept, nil}.run)
	if err != nil || !strings.Contains(out, "same to a person") ||
		strings.Contains(out, "different") {
		t.Errorf("a kept colour: %q, %v", out, err)
	}
	unreadable := swatch.FromXYZ(math.NaN(), math.NaN(), math.NaN())
	_, err = printing(t, tripping{[]labelled{{"nan", unreadable}}, nil}.run)
	if err == nil || !strings.Contains(err.Error(), "finite") {
		t.Errorf("a colour that cannot be written as oklch: %v", err)
	}
}

// A range with one profile is said in the singular, and one with several in
// the plural.
func TestRadiationNamesHowManyProfilesShareABar(t *testing.T) {
	one, _ := printing(t, barOf(radiation.Profiles[:onlyOne]).run)
	if !strings.Contains(one, "1 profile is drawn") {
		t.Errorf("one profile: %q", one)
	}
	every, _ := printing(t, barOf(radiation.Profiles).run)
	if !strings.Contains(every, "profiles are drawn") ||
		strings.Contains(every, "1 profile is") {
		t.Errorf("every profile: %q", every)
	}
}

// The count of steps in a call is read as a number: any is taken, and a word
// that is not one is refused.
func TestRampTakesItsStepsFromTheCall(t *testing.T) {
	for steps := range aFew {
		call := stepsOf{"#000", "#fff", strconv.Itoa(steps)}
		if out, err := printing(t, call.run); err != nil || out == "" {
			t.Errorf("%d steps: %q, %v", steps, out, err)
		}
	}
	if _, err := printing(t, stepsOf{"#000", "#fff", "x"}.run); err == nil {
		t.Error("a word was taken as a count of steps")
	}
}

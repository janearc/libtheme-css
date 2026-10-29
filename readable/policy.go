// Package readable holds a colour scheme to a reader's measured range and
// says, in words, which colour on which surface is outside it.
//
// The rest of libtheme is arithmetic and takes no side. This package is
// the product's acceptance gate, and takes the reader's: prose lives in a
// band of contrast, state colours keep above a floor, and a surface with
// colour in it needs a hue that differs from the text on it.
//
// When a check fails, the colour changes, not the threshold.
//
// Some colours are not graded. The gutter and the border have no rule: they
// are furniture. They do draw text in places, as LineNr does in the gutter
// colour and StatusLineNC does on the panel. The terminal's sixteen are not
// text and have none either.
//
// A policy file is any five numbers that nest. One passed to the check can
// loosen the gate, and nothing here records who may ship one. The check
// is not wired into any pipeline either.
package readable

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

//go:embed reference-dark-2026-09-03.json
var reference []byte

// Policy is a reader's range: the band prose contrast must be inside, the
// floor state colours must be above, and the hue rule. Its numbers are
// measured, not chosen, so a Policy is made whole from a file or not at
// all.
type Policy struct {
	low, high, stateMin float64
	hueApart, neutral   float64
}

// minimumRatio is the contrast of a colour with itself. It is defined by
// the contrast ratio.
const minimumRatio = 1

// halfCircle is the furthest apart two hues can be, in degrees. It is
// defined by the circle.
const halfCircle = 180

// nothing is none of anything. It is defined by counting.
const nothing = 0

// allOfIt is all of a scale from nothing: all the saturation a colour can
// have. It is defined by the scale.
const allOfIt = 1

// Read is a reader's range from a file in the shape of the reference
// range's: contrast low, high and state_min, hue min_separation_degrees
// and neutral_saturation. A newer measurement is a new file.
//
// Each of the five numbers must be there and be a number. The band must
// have room in it, the state floor must be above 1 and no higher than the
// band's low end, the hue gap must be up to the half circle, and the
// neutral saturation between 0 and 1. Anything else is refused.
func Read(source io.Reader) (Policy, error) {
	var document map[string]any
	if err := json.NewDecoder(source).Decode(&document); err != nil {
		return Policy{}, fmt.Errorf("readable: the range: %w", err)
	}
	read := &reading{document: document}
	policy := Policy{
		low:      read.number("contrast", "low"),
		high:     read.number("contrast", "high"),
		stateMin: read.number("contrast", "state_min"),
		hueApart: read.number("hue", "min_separation_degrees"),
		neutral:  read.number("hue", "neutral_saturation"),
	}
	if read.err != nil {
		return Policy{}, read.err
	}
	return policy.sensible()
}

// reading is a decoded range file and the first thing wrong with it.
type reading struct {
	document map[string]any
	err      error
}

// number is the number at a path of names down the document. If there is
// none, it keeps the error, and answers nothing for this and every later
// ask.
func (reading *reading) number(path ...string) (found float64) {
	if reading.err != nil {
		return found
	}
	var level any = reading.document
	for _, name := range path {
		// something that is not an object has nothing under it: a nil
		// map answers nil, and nil is no number.
		object, _ := level.(map[string]any)
		level = object[name]
	}
	found, isNumber := level.(float64)
	if !isNumber {
		reading.err = fmt.Errorf("readable: no number at %s",
			strings.Join(path, "."))
	}
	return found
}

// Reference is the reader's range as it was measured on 2026-09-03, from
// the file beside this one.
func Reference() (Policy, error) {
	return Read(bytes.NewReader(reference))
}

// sensible is the policy if its numbers make sense together.
func (policy Policy) sensible() (Policy, error) {
	nests := policy.stateMin > minimumRatio &&
		policy.stateMin <= policy.low && policy.low < policy.high
	hued := policy.hueApart > nothing && policy.hueApart <= halfCircle
	toned := policy.neutral > nothing && policy.neutral < allOfIt
	if !nests || !hued || !toned {
		return Policy{}, fmt.Errorf("readable: nonsense range: "+
			"floor %g, band %g to %g, hue gap %g, neutral %g",
			policy.stateMin, policy.low, policy.high,
			policy.hueApart, policy.neutral)
	}
	return policy, nil
}
